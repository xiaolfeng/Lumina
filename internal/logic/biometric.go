package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	xUtil "github.com/bamboo-services/bamboo-base-go/common/utility"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

	apiBiometric "github.com/xiaolfeng/Lumina/api/biometric"
	apiUser "github.com/xiaolfeng/Lumina/api/user"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

// challengeTypeRegister challenge 类型常量：注册流程
const challengeTypeRegister = "reg"

// challengeTypeLogin challenge 类型常量：登录流程
const challengeTypeLogin = "login"

// biometricLogic 是所有 Logic 结构体的公共基础，仅持有日志记录器
//
// 通过嵌入匿名的 [logic] 字段复用公共基础结构（log）。
// 刻意不持有 db/rdb：所有持久化与缓存读写必须经由 repository 层。

// BiometricLogic 生物特征认证业务编排层
//
// 基于 [go-webauthn/webauthn] 实现 WebAuthn 注册/登录 ceremony，
// 对外暴露 RegisterStart/Finish、LoginStart/Finish、GetAvailability、
// ListCredentials、DeleteCredential 等用例方法。
//
// 设计要点:
//   - RP 配置（RPID / RPDisplayName / RPOrigins）从环境变量读取，禁止硬编码
//   - challenge 通过 [BiometricCredentialRepo] 的委托方法写入 Redis，60s TTL，单次使用
//   - 注册/登录成功后更新签名计数器与最后使用时间
//   - 登录复用 [AuthLogic.generateTokens] 生成访问/刷新令牌
//
// 字段说明:
//   - logic:     公共基础（仅持有 log）
//   - repo:      生物特征凭证数据访问层（含 challenge 委托）
//   - info:      Info 表数据访问层（读取 owner 信息 / 更新 biometric_enabled 标记）
//   - auth:      认证业务逻辑层（复用 generateTokens）
//   - webAuthn:  go-webauthn 核心实例（封装 RP 配置）
type BiometricLogic struct {
	logic
	repo     *repository.BiometricCredentialRepo
	info     *repository.InfoRepo
	auth     *AuthLogic
	webAuthn *webauthn.WebAuthn
}

// NewBiometricLogic 创建 BiometricLogic 实例
//
// WebAuthn RP 配置从环境变量读取，提供合理默认值（localhost 场景）:
//   - XLF_BIOMETRIC_RP_ID（默认 localhost）
//   - XLF_BIOMETRIC_RP_NAME（默认 Lumina）
//   - XLF_BIOMETRIC_ORIGIN（默认 http://localhost:8080）
//
// 参数说明:
//   - ctx:      含 db / rdb 注入的上下文（用于构造 repo）
//   - authLogic: AuthLogic 实例（复用 generateTokens 生成登录令牌）
//
// 返回值:
//   - *BiometricLogic: 就绪的业务逻辑实例
//
// 注意: 如果 WebAuthn 配置校验失败（如 RPID 不合法），会 panic。
// 这是启动期故障，应在开发阶段暴露而非运行时吞错。
func NewBiometricLogic(ctx context.Context, authLogic *AuthLogic) *BiometricLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	// 从环境变量读取 WebAuthn RP 配置，禁止硬编码
	rpID := xEnv.GetEnvString(bConst.EnvBiometricRPID, bConst.DefaultBiometricRPID)
	rpName := xEnv.GetEnvString(bConst.EnvBiometricRPName, bConst.DefaultBiometricRPName)
	origin := xEnv.GetEnvString(bConst.EnvBiometricOrigin, bConst.DefaultBiometricOrigin)

	wconfig := &webauthn.Config{
		RPDisplayName: rpName,
		RPID:          rpID,
		RPOrigins:     []string{origin},
	}

	// webauthn.New 返回 (*WebAuthn, error)，配置校验失败需在启动期暴露
	wa, err := webauthn.New(wconfig)
	if err != nil {
		panic(fmt.Sprintf("初始化 WebAuthn 失败（请检查 RP 配置）: %v", err))
	}

	return &BiometricLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "BiometricLogic"),
		},
		repo:     repository.NewBiometricCredentialRepo(db, rdb),
		info:     repository.NewInfoRepo(db),
		auth:     authLogic,
		webAuthn: wa,
	}
}

// RegisterStart 开始注册流程（需认证用户调用）
//
// 流程:
//  1. 从 Info 表读取 owner 信息，组装 [LuminaWebAuthnUser]
//  2. 调用 [webauthn.WebAuthn.BeginRegistration] 生成 PublicKeyCredentialCreationOptions + SessionData
//  3. 序列化 SessionData 写入 Redis（通过 repo 委托，60s TTL）
//  4. 返回 sessionToken + options JSON 给前端
//
// 前端拿到 options 后调用 navigator.credentials.create() 进行认证器交互。
func (l *BiometricLogic) RegisterStart(ctx context.Context, req *apiBiometric.RegisterStartRequest) (*apiBiometric.RegisterStartResponse, *xError.Error) {
	l.log.Info(ctx, "RegisterStart - 开始生物特征注册")

	user, xErr := l.getWebAuthnUser(ctx)
	if xErr != nil {
		return nil, xErr
	}

	// 生成注册选项与会话数据
	creation, sessionData, err := l.webAuthn.BeginRegistration(user)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "生成注册选项失败", false, err)
	}

	// 序列化 sessionData 存入 Redis（通过 repo 委托，60s TTL）
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "序列化会话数据失败", false, err)
	}

	// 生成会话令牌作为 Redis key 的 sessionID 标识
	sessionToken := xUtil.Security().GenerateKey()
	if xErr := l.repo.SetChallenge(ctx, challengeTypeRegister, sessionToken, sessionDataJSON); xErr != nil {
		return nil, xErr
	}

	// 序列化 PublicKeyCredentialCreationOptions（透传给前端）
	// 必须取 .Response：creation 是 protocol.CredentialCreation，
	// 其 JSON tag 为 "publicKey"，直接序列化整个对象会产生多余包装层，
	// 导致浏览器 parseCreationOptionsFromJSON 找不到 challenge 字段。
	optionsJSON, err := json.Marshal(creation.Response)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "序列化注册选项失败", false, err)
	}

	return &apiBiometric.RegisterStartResponse{
		SessionToken: sessionToken,
		Options:      optionsJSON,
	}, nil
}

// RegisterFinish 完成注册流程
//
// 流程:
//  1. 从 Redis 读取并删除 challenge（单次使用，防重放）
//  2. 反序列化 SessionData
//  3. 调用 [protocol.ParseCredentialCreationResponseBody] 解析前端提交的凭证 JSON
//  4. 调用 [webauthn.WebAuthn.CreateCredential] 完成服务端验证
//  5. 持久化凭证到数据库（通过 repo）
//  6. 更新 Info 表 biometric_enabled 标记为 true
//
// 任何步骤失败都会立即返回错误，challenge 已在读取时删除确保不被重放。
func (l *BiometricLogic) RegisterFinish(ctx context.Context, req *apiBiometric.RegisterFinishRequest) (*apiBiometric.RegisterFinishResponse, *xError.Error) {
	l.log.Info(ctx, "RegisterFinish - 完成生物特征注册")

	// 读取并删除 challenge（单次使用）
	sessionDataJSON, ok, xErr := l.repo.GetChallenge(ctx, challengeTypeRegister, req.SessionToken)
	if xErr != nil {
		return nil, xErr
	}
	if !ok {
		return nil, xError.NewError(ctx, xError.TokenExpired, "注册会话已过期，请重新注册", false, nil)
	}
	// 验证后立即删除，防止 challenge 被重放
	l.repo.DeleteChallenge(ctx, challengeTypeRegister, req.SessionToken)

	// 反序列化 SessionData
	var sessionData webauthn.SessionData
	if err := json.Unmarshal(sessionDataJSON, &sessionData); err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "解析会话数据失败", false, err)
	}

	user, xErr := l.getWebAuthnUser(ctx)
	if xErr != nil {
		return nil, xErr
	}

	// 解析前端提交的凭证 JSON（navigator.credentials.create() 的返回值）
	parsedResponse, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(req.Credential))
	if err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, "解析凭证数据失败", false, err)
	}

	// 服务端验证凭证（签名、challenge、origin 等）
	credential, err := l.webAuthn.CreateCredential(user, sessionData, parsedResponse)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "凭证验证失败", false, err)
	}

	// 持久化凭证到数据库
	credEntity := &entity.BiometricCredential{
		CredentialID:   credential.ID,
		PublicKey:      credential.PublicKey,
		AAGUID:         string(credential.Authenticator.AAGUID), // []byte → string
		SignCount:      credential.Authenticator.SignCount,
		DeviceName:     req.DeviceName,
		TransportTypes: transportsToString(credential.Transport),
	}
	if xErr := l.repo.Create(ctx, credEntity); xErr != nil {
		return nil, xErr
	}

	// 注册成功后更新 Info 表标记（非致命，失败仅记录日志）
	_ = l.info.UpdateValue(ctx, "biometric_enabled", "true")

	l.log.Info(ctx, "RegisterFinish - 生物特征注册成功")
	return &apiBiometric.RegisterFinishResponse{
		Success: true,
		Message: "生物特征注册成功",
	}, nil
}

// LoginStart 开始登录流程（公开接口，无需认证）
//
// 流程:
//  1. 从 Info 表读取 owner 信息 + 所有凭证，组装 [LuminaWebAuthnUser]
//  2. 调用 [webauthn.WebAuthn.BeginLogin] 生成 PublicKeyCredentialRequestOptions + SessionData
//  3. 序列化 SessionData 写入 Redis（60s TTL）
//  4. 返回 sessionToken + options JSON 给前端
//
// 前端拿到 options 后调用 navigator.credentials.get() 进行认证器交互。
func (l *BiometricLogic) LoginStart(ctx context.Context) (*apiBiometric.LoginStartResponse, *xError.Error) {
	l.log.Info(ctx, "LoginStart - 开始生物特征登录")

	user, xErr := l.getWebAuthnUser(ctx)
	if xErr != nil {
		return nil, xErr
	}

	// 生成登录选项与会话数据
	assertion, sessionData, err := l.webAuthn.BeginLogin(user)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "生成登录选项失败", false, err)
	}

	// 序列化 SessionData 存入 Redis（60s TTL）
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "序列化会话数据失败", false, err)
	}

	sessionToken := xUtil.Security().GenerateKey()
	if xErr := l.repo.SetChallenge(ctx, challengeTypeLogin, sessionToken, sessionDataJSON); xErr != nil {
		return nil, xErr
	}

	// 序列化 PublicKeyCredentialRequestOptions（透传给前端）
	// 必须取 .Response：assertion 是 protocol.CredentialAssertion，
	// 其 JSON tag 为 "publicKey"，直接序列化整个对象会产生多余包装层，
	// 导致浏览器 parseRequestOptionsFromJSON 找不到 challenge 字段。
	optionsJSON, err := json.Marshal(assertion.Response)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "序列化登录选项失败", false, err)
	}

	return &apiBiometric.LoginStartResponse{
		SessionToken: sessionToken,
		Options:      optionsJSON,
	}, nil
}

// LoginFinish 完成登录流程
//
// 流程:
//  1. 从 Redis 读取并删除 challenge（单次使用）
//  2. 反序列化 SessionData
//  3. 调用 [protocol.ParseCredentialRequestResponseBody] 解析前端提交的断言 JSON
//  4. 调用 [webauthn.WebAuthn.ValidateLogin] 验证断言签名
//  5. 通过 credential.ID 反查 DB 实体，更新签名计数器与最后使用时间
//  6. 复用 [AuthLogic.generateTokens] 生成访问/刷新令牌
//
// 验证失败返回 LoginFailed 错误；challenge 在读取时已删除。
func (l *BiometricLogic) LoginFinish(ctx context.Context, req *apiBiometric.LoginFinishRequest) (*apiBiometric.LoginFinishResponse, *xError.Error) {
	l.log.Info(ctx, "LoginFinish - 完成生物特征登录")

	// 读取并删除 challenge（单次使用）
	sessionDataJSON, ok, xErr := l.repo.GetChallenge(ctx, challengeTypeLogin, req.SessionToken)
	if xErr != nil {
		return nil, xErr
	}
	if !ok {
		return nil, xError.NewError(ctx, xError.TokenExpired, "登录会话已过期，请重新登录", false, nil)
	}
	// 验证后立即删除，防止 challenge 被重放
	l.repo.DeleteChallenge(ctx, challengeTypeLogin, req.SessionToken)

	// 反序列化 SessionData
	var sessionData webauthn.SessionData
	if err := json.Unmarshal(sessionDataJSON, &sessionData); err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "解析会话数据失败", false, err)
	}

	user, xErr := l.getWebAuthnUser(ctx)
	if xErr != nil {
		return nil, xErr
	}

	// 解析前端提交的断言 JSON（navigator.credentials.get() 的返回值）
	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(req.Credential))
	if err != nil {
		return nil, xError.NewError(ctx, xError.ParameterError, "解析凭证数据失败", false, err)
	}

	// 验证断言（签名、challenge、origin 等）
	credential, err := l.webAuthn.ValidateLogin(user, sessionData, parsedResponse)
	if err != nil {
		return nil, xError.NewError(ctx, xError.LoginFailed, "生物特征验证失败", false, err)
	}

	// 通过 WebAuthn credential.ID（[]byte）反查 DB 实体，获取雪花 ID
	// 用于更新签名计数器（反克隆检测）与最后使用时间
	if credEntity, xErr := l.repo.GetByCredentialID(ctx, credential.ID); xErr == nil && credEntity != nil {
		_ = l.repo.UpdateSignCount(ctx, credEntity.ID, credential.Authenticator.SignCount)
		_ = l.repo.UpdateLastUsedAt(ctx, credEntity.ID)
	}

	// 复用 AuthLogic.generateTokens 生成令牌
	tokenResp, xErr := l.auth.generateTokens(ctx)
	if xErr != nil {
		return nil, xErr
	}

	l.log.Info(ctx, "LoginFinish - 生物特征登录成功")
	return &apiBiometric.LoginFinishResponse{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		// ExpiresIn 是秒数，转换为 Unix 时间戳
		ExpiresAt: time.Now().Unix() + tokenResp.ExpiresIn,
	}, nil
}

// GetAvailability 获取生物特征登录可用性
//
// 用于登录页判断是否展示「生物特征登录」入口。返回 true 表示至少有一个已注册凭证。
func (l *BiometricLogic) GetAvailability(ctx context.Context) (*apiBiometric.AvailabilityResponse, *xError.Error) {
	available, xErr := l.repo.IsAvailable(ctx)
	if xErr != nil {
		return nil, xErr
	}
	return &apiBiometric.AvailabilityResponse{
		IsAvailable: available,
	}, nil
}

// ListCredentials 获取已注册的凭证列表（管理页用）
//
// 返回所有凭证的元信息（不包含公钥等敏感字段），按创建时间降序排列。
func (l *BiometricLogic) ListCredentials(ctx context.Context) (*apiUser.BiometricCredentialListResponse, *xError.Error) {
	creds, xErr := l.repo.ListAll(ctx)
	if xErr != nil {
		return nil, xErr
	}

	items := make([]apiUser.BiometricCredentialItem, 0, len(creds))
	for _, c := range creds {
		var lastUsedAt *int64
		if c.LastUsedAt != nil {
			ts := c.LastUsedAt.Unix()
			lastUsedAt = &ts
		}
		items = append(items, apiUser.BiometricCredentialItem{
			ID:         strconv.FormatInt(c.ID.Int64(), 10),
			DeviceName: c.DeviceName,
			AAGUID:     c.AAGUID,
			LastUsedAt: lastUsedAt,
			CreatedAt:  c.CreatedAt.Unix(),
		})
	}

	return &apiUser.BiometricCredentialListResponse{
		Total: len(items),
		Items: items,
	}, nil
}

// DeleteCredential 删除指定的凭证
//
// 流程:
//  1. 将字符串 ID 解析为雪花 ID
//  2. 通过 repo 删除凭证（会同步清除缓存）
//  3. 检查是否还有剩余凭证，若无则将 Info 表 biometric_enabled 标记置为 false
//
// 删除最后一个凭证后，「生物特征登录」入口会自动隐藏。
func (l *BiometricLogic) DeleteCredential(ctx context.Context, idStr string) *xError.Error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return xError.NewError(ctx, xError.ParameterError, "无效的凭证 ID", false, err)
	}
	snowflakeID := xSnowflake.SnowflakeID(id)

	if xErr := l.repo.Delete(ctx, snowflakeID); xErr != nil {
		return xErr
	}

	// 检查是否还有剩余凭证，更新 Info 标记（非致命，失败仅记录日志）
	if available, _ := l.repo.IsAvailable(ctx); !available {
		_ = l.info.UpdateValue(ctx, "biometric_enabled", "false")
	}

	l.log.Info(ctx, fmt.Sprintf("DeleteCredential - 凭证删除成功 [%d]", id))
	return nil
}

// ── 私有辅助方法 ──

// getWebAuthnUser 从 Info 表读取 owner 信息并构造 WebAuthn User 适配器
//
// 该方法在注册与登录的 start 阶段都会被调用，用于组装 go-webauthn 需要的
// [webauthn.User] 接口实现。读取失败返回 DatabaseError。
func (l *BiometricLogic) getWebAuthnUser(ctx context.Context) (*LuminaWebAuthnUser, *xError.Error) {
	// 从 Info 表读取 owner 用户名与邮箱
	username, email, xErr := l.auth.GetOwnerInfo(ctx)
	if xErr != nil {
		return nil, xErr
	}

	creds, xErr := l.repo.ListAll(ctx)
	if xErr != nil {
		return nil, xErr
	}

	return NewLuminaWebAuthnUser(username, email, entitiesToWebAuthnCredentials(creds)), nil
}
