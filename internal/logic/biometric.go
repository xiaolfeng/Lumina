package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"golang.org/x/net/publicsuffix"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xUtil "github.com/bamboo-services/bamboo-base-go/common/utility"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"

	apiBiometric "github.com/xiaolfeng/Lumina/api/biometric"
	apiUser "github.com/xiaolfeng/Lumina/api/user"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
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
//   - challenge 通过 [BiometricCredentialRepo] 的委托方法写入 Redis，动态 TTL，原子单次消费
//   - 注册/登录成功后更新签名计数器与最后使用时间
//   - 登录复用 [AuthLogic.generateTokens] 生成访问/刷新令牌
//
// 字段说明:
//   - logic:     公共基础（仅持有 log）
//   - repo:      生物特征凭证数据访问层（含 challenge 委托）
//   - info:      Info 表数据访问层（读取 owner 信息 / 更新 owner.biometric-enabled 标记）
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
//   - XLF_BIOMETRIC_ORIGIN（逗号分隔，默认覆盖 localhost:8080 与 localhost:3000）
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
	originValue := xEnv.GetEnvString(bConst.EnvBiometricOrigin, bConst.DefaultBiometricOrigin)

	wconfig := &webauthn.Config{
		RPDisplayName: rpName,
		RPID:          rpID,
		RPOrigins:     parseWebAuthnOrigins(originValue),
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementPreferred,
			UserVerification: protocol.VerificationRequired,
		},
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
//  3. 序列化 SessionData 写入 Redis（通过 repo 委托，TTL 覆盖浏览器 ceremony 超时）
//  4. 返回 sessionToken + options JSON 给前端
//
// 前端拿到 options 后调用 navigator.credentials.create() 进行认证器交互。
func (l *BiometricLogic) RegisterStart(ctx context.Context, req *apiBiometric.RegisterStartRequest) (*apiBiometric.RegisterStartResponse, *xError.Error) {
	l.log.Info(ctx, "RegisterStart - 开始生物特征注册")

	user, xErr := l.getWebAuthnUser(ctx)
	if xErr != nil {
		return nil, xErr
	}

	timeout := l.getWebAuthnTimeout(ctx)
	exclusions := webauthn.Credentials(user.WebAuthnCredentials()).CredentialDescriptors()

	// 生成注册选项与会话数据
	wa, xErr := l.resolveWebAuthn(ctx)
	if xErr != nil {
		return nil, xErr
	}
	creation, sessionData, err := wa.BeginRegistration(
		user,
		webauthn.WithExclusions(exclusions),
		withRegistrationTimeout(timeout),
	)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "生成注册选项失败", false, err)
	}

	// 序列化 sessionData 存入 Redis（通过 repo 委托，TTL 覆盖浏览器 ceremony 超时）
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "序列化会话数据失败", false, err)
	}

	// 生成会话令牌作为 Redis key 的 sessionID 标识
	sessionToken := xUtil.Security().GenerateKey()
	if xErr := l.repo.SetChallenge(ctx, challengeTypeRegister, sessionToken, sessionDataJSON, challengeTTL(timeout)); xErr != nil {
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
//  6. 更新 Info 表 owner.biometric-enabled 标记为 true
//
// 任何步骤失败都会立即返回错误，challenge 已在读取时删除确保不被重放。
func (l *BiometricLogic) RegisterFinish(ctx context.Context, req *apiBiometric.RegisterFinishRequest) (*apiBiometric.RegisterFinishResponse, *xError.Error) {
	l.log.Info(ctx, "RegisterFinish - 完成生物特征注册")

	// 读取并删除 challenge（单次使用）
	sessionDataJSON, ok, xErr := l.repo.ConsumeChallenge(ctx, challengeTypeRegister, req.SessionToken)
	if xErr != nil {
		return nil, xErr
	}
	if !ok {
		return nil, xError.NewError(ctx, xError.TokenExpired, "注册会话已过期，请重新注册", false, nil)
	}
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

	// 服务端验证凭证（签名、challenge、origin 等），
	// RP ID 必须与 Start 阶段写入 SessionData 的值一致
	wa, xErr := l.resolveWebAuthnForFinish(ctx, sessionData.RelyingPartyID)
	if xErr != nil {
		return nil, xErr
	}
	credential, err := wa.CreateCredential(user, sessionData, parsedResponse)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "凭证验证失败", false, err)
	}

	// 持久化凭证到数据库
	credEntity, err := newBiometricCredentialEntity(credential, req.DeviceName)
	if err != nil {
		return nil, xError.NewError(ctx, xError.SerializeError, "序列化凭证记录失败", false, err)
	}
	if xErr := l.repo.Create(ctx, credEntity); xErr != nil {
		return nil, xErr
	}

	// 注册成功后更新 Info 表标记（非致命，失败仅记录日志）
	_ = l.info.UpdateValue(ctx, bConst.InfoKeyOwnerBiometricEnabled, "true")

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
//  3. 序列化 SessionData 写入 Redis（TTL 覆盖浏览器 ceremony 超时）
//  4. 返回 sessionToken + options JSON 给前端
//
// 前端拿到 options 后调用 navigator.credentials.get() 进行认证器交互。
func (l *BiometricLogic) LoginStart(ctx context.Context) (*apiBiometric.LoginStartResponse, *xError.Error) {
	l.log.Info(ctx, "LoginStart - 开始生物特征登录")

	user, xErr := l.getWebAuthnUser(ctx)
	if xErr != nil {
		return nil, xErr
	}

	timeout := l.getWebAuthnTimeout(ctx)

	// 生成登录选项与会话数据
	wa, xErr := l.resolveWebAuthn(ctx)
	if xErr != nil {
		return nil, xErr
	}
	assertion, sessionData, err := wa.BeginLogin(
		user,
		webauthn.WithUserVerification(protocol.VerificationRequired),
		withLoginTimeout(timeout),
	)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "生成登录选项失败", false, err)
	}

	// 序列化 SessionData 存入 Redis（TTL 覆盖浏览器 ceremony 超时）
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "序列化会话数据失败", false, err)
	}

	sessionToken := xUtil.Security().GenerateKey()
	if xErr := l.repo.SetChallenge(ctx, challengeTypeLogin, sessionToken, sessionDataJSON, challengeTTL(timeout)); xErr != nil {
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
	sessionDataJSON, ok, xErr := l.repo.ConsumeChallenge(ctx, challengeTypeLogin, req.SessionToken)
	if xErr != nil {
		return nil, xErr
	}
	if !ok {
		return nil, xError.NewError(ctx, xError.TokenExpired, "登录会话已过期，请重新登录", false, nil)
	}
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

	// 历史记录没有持久化 CredentialFlags。仅对这类记录从当前断言临时恢复标志，
	// 随后仍需通过已存公钥完成签名校验；成功后会写回完整凭证记录。
	credEntity, xErr := l.repo.GetByCredentialID(ctx, parsedResponse.RawID)
	if xErr != nil {
		return nil, xErr
	}
	if len(credEntity.CredentialData) == 0 {
		hydrateLegacyCredentialFlags(user, parsedResponse.RawID, parsedResponse.Response.AuthenticatorData.Flags)
	}

	// 验证断言（签名、challenge、origin 等），
	// RP ID 必须与 Start 阶段写入 SessionData 的值一致
	wa, xErr := l.resolveWebAuthnForFinish(ctx, sessionData.RelyingPartyID)
	if xErr != nil {
		return nil, xErr
	}
	credential, err := wa.ValidateLogin(user, sessionData, parsedResponse)
	if err != nil {
		return nil, xError.NewError(ctx, xError.LoginFailed, "生物特征验证失败", false, err)
	}

	// 持久化更新后的签名计数器、备份状态与最后使用时间。
	credentialData, err := marshalWebAuthnCredential(credential)
	if err != nil {
		return nil, xError.NewError(ctx, xError.SerializeError, "序列化凭证记录失败", false, err)
	}
	if xErr := l.repo.UpdateAfterLogin(ctx, credEntity, credentialData, credential.Authenticator.SignCount); xErr != nil {
		return nil, xErr
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
			ID:         c.ID,
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
//  3. 检查是否还有剩余凭证，若无则将 Info 表 owner.biometric-enabled 标记置为 false
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
		_ = l.info.UpdateValue(ctx, bConst.InfoKeyOwnerBiometricEnabled, "false")
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

	credentials, err := entitiesToWebAuthnCredentials(creds)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "解析 WebAuthn 凭证记录失败", false, err)
	}

	return NewLuminaWebAuthnUser(username, email, credentials), nil
}

// parseWebAuthnOrigins 解析逗号分隔的允许 Origin 列表。
func parseWebAuthnOrigins(value string) []string {
	origins := make([]string, 0, 2)
	for _, origin := range strings.Split(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return []string{"http://localhost:8080"}
	}
	return origins
}

// resolveWebAuthn 按请求动态解析 WebAuthn RP 配置并构造实例
//
// WebAuthn 要求 rp.id 为当前域的可注册域后缀（或相等），否则浏览器在
// navigator.credentials.create()/get() 阶段即拒绝（典型错误：
// 「The relying party ID is not a registrable domain suffix of, nor equal to
// the current domain」）。启动期静态配置无法感知部署域名，这里根据
// middleware.WebAuthnOrigin 注入的浏览器 Origin 推导：
//
//   - RPID: XLF_BIOMETRIC_RP_ID 未配置时自动取 Origin 的 Hostname；
//     已配置且为可注册域后缀时使用配置值，否则返回配置错误
//     （禁止静默生成会被浏览器拒绝的 rp.id）
//   - RPOrigins: 环境变量配置与请求 Origin 取并集（服务端 origin 校验必达）
//
// 仅无 Origin 注入的场景回退启动期静态实例；其余失败一律返回显式错误，
// 避免把 localhost 等静态 rp.id 下发给线上页面。
func (l *BiometricLogic) resolveWebAuthn(ctx context.Context) (*webauthn.WebAuthn, *xError.Error) {
	origin, _ := ctx.Value(bConst.WebAuthnOriginContextKey).(string)
	if origin == "" {
		return l.webAuthn, nil
	}

	u, err := url.Parse(origin)
	if err != nil || u.Hostname() == "" {
		return nil, xError.NewError(ctx, xError.ConfigError,
			xError.ErrMessage(fmt.Sprintf("无法识别访问 Origin [%s]，请检查反向代理是否透传 Origin/Referer/Host", origin)), true)
	}

	reqOrigin, ok := normalizeOrigin(origin)
	if !ok {
		return nil, xError.NewError(ctx, xError.ConfigError,
			xError.ErrMessage(fmt.Sprintf("非法的访问 Origin [%s]", origin)), true)
	}

	hostname := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))

	// 可选域名白名单：仅当配置了 XLF_BIOMETRIC_ALLOWED_ORIGINS 时校验。
	// 配置白名单后，不在名单内的请求被直接拒绝，防止 DNS-rebinding 使凭证
	// 绑定到攻击者可控域名。校验基于归一化完整 Origin（scheme://host[:port]），
	// scheme 或端口不同即为不同 Origin，不接受仅 hostname 匹配。
	var allowedOrigins []string
	for _, o := range strings.Split(xEnv.GetEnvString(bConst.EnvBiometricAllowedOrigins, ""), ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowedOrigins = append(allowedOrigins, o)
		}
	}
	if len(allowedOrigins) > 0 && !isAllowedRPOrigin(allowedOrigins, reqOrigin) {
		if l.log != nil {
			l.log.Warn(ctx, fmt.Sprintf("resolveWebAuthn - 访问 Origin 不在白名单 [origin=%s]", reqOrigin))
		}
		return nil, xError.NewError(ctx, xError.Forbidden,
			xError.ErrMessage(fmt.Sprintf("访问 Origin [%s] 不在 XLF_BIOMETRIC_ALLOWED_ORIGINS 白名单内", reqOrigin)), true)
	}

	rpID, xErr := l.resolveRPID(ctx, hostname)
	if xErr != nil || rpID == "" {
		return nil, xErr
	}

	// 基于启动期配置复制，仅覆盖请求相关的 RPID/RPOrigins。
	// 追加归一化后的 Origin：浏览器 clientData 的 origin 恒为规范形式
	// （不携带默认端口），用原始请求串会造成 Finish 阶段比对失败或重复项。
	config := *l.webAuthn.Config
	config.RPID = rpID
	config.RPOrigins = appendWebAuthnOrigins(config.RPOrigins, reqOrigin)

	wa, err := webauthn.New(&config)
	if err != nil {
		if l.log != nil {
			l.log.Warn(ctx, fmt.Sprintf("resolveWebAuthn - 动态构建 WebAuthn 实例失败: %v", err))
		}
		return nil, xError.NewError(ctx, xError.ServerInternalError, "构建 WebAuthn 实例失败", true, err)
	}
	return wa, nil
}

// resolveWebAuthnForFinish 构建 Finish 阶段的验证实例。
//
// Start 阶段的 SessionData.RelyingPartyID 已随 Challenge 固化进 Redis，
// Finish 阶段必须以同一 RP ID 计算 rpIdHash 才能通过校验；按请求头二次推导
// 在部署域名变更、多域名入口等场景下会产生偏差。sessionRPID 为空（历史数据）
// 时退化为按当前请求推导。
func (l *BiometricLogic) resolveWebAuthnForFinish(ctx context.Context, sessionRPID string) (*webauthn.WebAuthn, *xError.Error) {
	wa, xErr := l.resolveWebAuthn(ctx)
	if xErr != nil || sessionRPID == "" || sessionRPID == wa.Config.RPID {
		return wa, xErr
	}

	// Config.validate() 带 validated 记忆化，浅拷贝启动配置会让动态 New()
	// 跳过 RP ID 合法性校验，这里对会话固化值显式补验后再构建
	if err := protocol.ValidateRPID(sessionRPID); err != nil {
		return nil, xError.NewError(ctx, xError.ConfigError,
			xError.ErrMessage(fmt.Sprintf("会话中的 RP ID 非法 [%s]", sessionRPID)), true, err)
	}

	config := *wa.Config
	config.RPID = sessionRPID
	finishWa, err := webauthn.New(&config)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "构建 WebAuthn 验证实例失败", true, err)
	}
	return finishWa, nil
}

// resolveRPID 推导当前请求的 RPID
//
// 规则:
//   - XLF_BIOMETRIC_RP_ID 为空或等于内置默认值（视为未配置）时，
//     使用请求 Hostname 自动推导（rp.id 等于当前域必然合法）
//   - 配置值与 Hostname 相等时使用配置值
//   - 配置值为 Hostname 后缀时，必须是可注册域后缀（eTLD+1）才允许，
//     支持子域共享凭证；co.uk/com 等公共后缀会被浏览器直接拒绝，返回配置错误
//   - 其余情况均返回配置错误，交由管理员修正部署域名或 RP ID 配置
func (l *BiometricLogic) resolveRPID(ctx context.Context, hostname string) (string, *xError.Error) {
	configured := strings.TrimSpace(xEnv.GetEnvString(bConst.EnvBiometricRPID, ""))
	autoDerive := configured == "" || configured == bConst.DefaultBiometricRPID

	// IP 主机只能以自身作为 rp.id（浏览器侧限制）
	if ip := net.ParseIP(hostname); ip != nil {
		if autoDerive || strings.EqualFold(configured, hostname) {
			return hostname, nil
		}
		return "", xError.NewError(ctx, xError.ConfigError,
			xError.ErrMessage(fmt.Sprintf("XLF_BIOMETRIC_RP_ID=%s 与访问地址 %s 不匹配", configured, hostname)), true)
	}

	switch {
	case autoDerive:
		return hostname, nil
	case strings.EqualFold(configured, hostname):
		return strings.ToLower(configured), nil
	case strings.HasSuffix(hostname, "."+strings.ToLower(configured)):
		etdp1, err := publicsuffix.EffectiveTLDPlusOne(hostname)
		if err == nil && strings.EqualFold(etdp1, configured) {
			return strings.ToLower(configured), nil
		}
		return "", xError.NewError(ctx, xError.ConfigError,
			xError.ErrMessage(fmt.Sprintf("XLF_BIOMETRIC_RP_ID=%s 不是 %s 的可注册域后缀（不允许 com/co.uk 等公共后缀）", configured, hostname)), true)
	default:
		return "", xError.NewError(ctx, xError.ConfigError,
			xError.ErrMessage(fmt.Sprintf("XLF_BIOMETRIC_RP_ID=%s 与访问域名 %s 不匹配，如需自动推导请清除该配置", configured, hostname)), true)
	}
}

// appendWebAuthnOrigins 在配置的 Origin 基础上补充请求 Origin（去重，保持顺序）。
func appendWebAuthnOrigins(configured []string, origin string) []string {
	origins := make([]string, 0, len(configured)+1)
	seen := make(map[string]struct{}, len(configured)+1)
	for _, o := range configured {
		if _, ok := seen[o]; !ok {
			seen[o] = struct{}{}
			origins = append(origins, o)
		}
	}
	if _, ok := seen[origin]; !ok {
		origins = append(origins, origin)
	}
	return origins
}

// isAllowedRPOrigin 校验请求 Origin 是否命中允许的 Origin 白名单。
//
// 基于归一化后的完整 Origin 精确匹配，不比较裸 hostname：不同 scheme 或端口
// 的 Origin 在浏览器侧互不相同，宽松匹配会削弱 XLF_BIOMETRIC_ALLOWED_ORIGINS
// 对 DNS-rebinding 的防护边界。配置项非法时跳过而非放行。
func isAllowedRPOrigin(configured []string, origin string) bool {
	target, ok := normalizeOrigin(origin)
	if !ok {
		return false
	}
	for _, entry := range configured {
		normalized, ok := normalizeOrigin(entry)
		if ok && strings.EqualFold(normalized, target) {
			return true
		}
	}
	return false
}

// normalizeOrigin 将 Origin 字符串归一化为 scheme://host[:port] 的严格形式。
//
// http 与 https、不同端口的 Origin 在 WebAuthn Origin 校验中互不相同；
// 默认端口（http=80 / https=443）与显式端口视为同一 Origin。
// 携带 userinfo / path / query / fragment 或非 HTTP(S) 协议的值视为非法，
// 返回 false 交由调用方拒绝。
func normalizeOrigin(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	u, err := url.Parse(value)
	if err != nil || u.User != nil ||
		(u.Path != "" && u.Path != "/") || u.RawQuery != "" ||
		u.Fragment != "" || u.RawFragment != "" {
		return "", false
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", false
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return "", false
	}

	port := u.Port()
	switch {
	case port == "80" && scheme == "http", port == "443" && scheme == "https":
		port = ""
	}

	hostPart := host
	if strings.Contains(host, ":") { // IPv6 地址补回方括号
		hostPart = "[" + host + "]"
	}
	if port != "" {
		hostPart += ":" + port
	}
	return scheme + "://" + hostPart, true
}

// getWebAuthnTimeout 读取安全设置中的 ceremony 超时，并限制在安全范围内。
func (l *BiometricLogic) getWebAuthnTimeout(ctx context.Context) int {
	value, xErr := l.info.GetByKey(ctx, bConst.InfoKeySecurityWebAuthnTimeout)
	if xErr != nil {
		return bConst.DefaultBiometricTimeout
	}
	timeout, err := strconv.Atoi(value)
	if err != nil || timeout < bConst.MinBiometricTimeout || timeout > bConst.MaxBiometricTimeout {
		l.log.Warn(ctx, fmt.Sprintf("WebAuthn 超时配置无效 [%s]，使用默认值", value))
		return bConst.DefaultBiometricTimeout
	}
	return timeout
}

// challengeTTL 比浏览器 ceremony 超时多保留 30 秒网络与提交余量。
func challengeTTL(timeout int) time.Duration {
	return time.Duration(timeout)*time.Millisecond + 30*time.Second
}

func withRegistrationTimeout(timeout int) webauthn.RegistrationOption {
	return func(options *protocol.PublicKeyCredentialCreationOptions) {
		options.Timeout = timeout
	}
}

func withLoginTimeout(timeout int) webauthn.LoginOption {
	return func(options *protocol.PublicKeyCredentialRequestOptions) {
		options.Timeout = timeout
	}
}
