package logic

import (
	"context"
	"crypto/subtle"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xUtil "github.com/bamboo-services/bamboo-base-go/common/utility"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"
	apiAuth "github.com/xiaolfeng/Lumina/api/auth"
	apiUser "github.com/xiaolfeng/Lumina/api/user"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

// authRepo 认证模块依赖的仓储集合
type authRepo struct {
	token *repository.TokenRepo
	info  *repository.InfoRepo
}

// AuthLogic 认证业务逻辑层，负责初始化、登录、令牌管理与校验
type AuthLogic struct {
	logic
	repo authRepo
}

// NewAuthLogic 创建认证业务逻辑层实例
func NewAuthLogic(ctx context.Context) *AuthLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	return &AuthLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "AuthLogic"),
		},
		repo: authRepo{
			token: repository.NewTokenRepo(rdb),
			info:  repository.NewInfoRepo(db),
		},
	}
}

// GetInitialStatus 获取系统是否为初始状态（未初始化）
func (l *AuthLogic) GetInitialStatus(ctx context.Context) (bool, *xError.Error) {
	l.log.Info(ctx, "GetInitialStatus - 检查系统初始化状态")

	value, xErr := l.repo.info.GetByKey(ctx, bConst.InfoKeyAuthIsInitial)
	if xErr != nil {
		// NotFound 视为未初始化
		if xErr.GetErrorCode() == xError.NotFound {
			return true, nil
		}
		return false, xError.NewError(ctx, xError.DatabaseError, "查询初始化状态失败", false, nil)
	}

	return value == "true", nil
}

// Initialize 执行系统初始化，将 owner 凭据写入 Info 表并标记已初始化
func (l *AuthLogic) Initialize(ctx context.Context, req *apiAuth.InitializeRequest) *xError.Error {
	l.log.Info(ctx, "Initialize - 执行系统初始化")

	// 一次性初始化令牌校验：部署时若配置 XLF_INITIALIZE_TOKEN，则请求必须携带匹配值，
	// 防止公开的 /auth/initialize 在初始化窗口内被未认证攻击者抢先接管管理员账户。
	if expected := xEnv.GetEnvString(bConst.EnvInitializeToken, ""); expected != "" {
		if subtle.ConstantTimeCompare([]byte(req.SetupToken), []byte(expected)) != 1 {
			return xError.NewError(ctx, xError.TokenInvalid, "初始化令牌无效", false, nil)
		}
	}

	// 密码长度校验（防 bcrypt 72 字节上限触发 panic）
	if xErr := validatePasswordLength(ctx, req.Password); xErr != nil {
		return xErr
	}

	// bcrypt 加密属于业务逻辑，在 logic 层完成；事务边界下沉到 InfoRepo
	kv := map[string]string{
		bConst.InfoKeyOwnerUsername: req.Username,
		bConst.InfoKeyOwnerEmail:    req.Email,
		bConst.InfoKeyOwnerPassword: xUtil.Password().MustEncryptString(req.Password),
		bConst.InfoKeyAuthIsInitial: "false",
	}

	// 原子「检查-写入」：通过行锁杜绝并发 TOCTOU，返回 false 表示已被他人初始化
	initialized, xErr := l.repo.info.InitializeIfNotInitialized(ctx, bConst.InfoKeyAuthIsInitial, kv)
	if xErr != nil {
		return xError.NewError(ctx, xError.DatabaseError, "系统初始化失败", false, nil)
	}
	if !initialized {
		return xError.NewError(ctx, xError.RepeatOperation, "系统已初始化，不可重复操作", false, nil)
	}

	l.log.Info(ctx, "Initialize - 系统初始化成功")
	return nil
}

// GetOwnerInfo 从 Info 表读取 owner 用户名与邮箱
func (l *AuthLogic) GetOwnerInfo(ctx context.Context) (username, email string, xErr *xError.Error) {
	username, xErr = l.repo.info.GetByKey(ctx, bConst.InfoKeyOwnerUsername)
	if xErr != nil {
		return "", "", xError.NewError(ctx, xError.DatabaseError, "读取用户信息失败", false, nil)
	}

	email, xErr = l.repo.info.GetByKey(ctx, bConst.InfoKeyOwnerEmail)
	if xErr != nil {
		return "", "", xError.NewError(ctx, xError.DatabaseError, "读取用户信息失败", false, nil)
	}

	return username, email, nil
}

// Login 用户登录，支持用户名或邮箱登录，返回访问令牌与刷新令牌
func (l *AuthLogic) Login(ctx context.Context, req *apiAuth.LoginRequest) (*apiAuth.TokenResponse, *xError.Error) {
	l.log.Info(ctx, "Login - 用户登录")

	// 从 Info 表读取 owner 用户名与邮箱
	username, email, xErr := l.GetOwnerInfo(ctx)
	if xErr != nil {
		return nil, xErr
	}

	// 根据是否包含 @ 判断登录方式，验证账号匹配
	accountMatched := false
	if strings.Contains(req.Account, "@") {
		accountMatched = req.Account == email
	} else {
		accountMatched = req.Account == username
	}
	if !accountMatched {
		// 账号不匹配时仍执行一次 bcrypt 比较（用当前存储的密码哈希），
		// 消除「账号存在与否」的响应时延差异，防时序侧信道枚举账号
		if hash, hErr := l.repo.info.GetByKey(ctx, bConst.InfoKeyOwnerPassword); hErr == nil && hash != "" {
			_ = xUtil.Password().IsValid(req.Password, hash)
		}
		return nil, xError.NewError(ctx, xError.LoginFailed, "账号或密码错误", false, nil)
	}

	// 从 Info 表读取 owner 密码哈希
	passwordHash, xErr := l.repo.info.GetByKey(ctx, bConst.InfoKeyOwnerPassword)
	if xErr != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "读取用户信息失败", false, nil)
	}

	// 验证密码
	if !xUtil.Password().IsValid(req.Password, passwordHash) {
		return nil, xError.NewError(ctx, xError.LoginFailed, "账号或密码错误", false, nil)
	}

	l.log.Info(ctx, "Login - 用户登录成功")
	return l.generateTokens(ctx)
}

// Refresh 使用刷新令牌换取新的访问令牌和刷新令牌
func (l *AuthLogic) Refresh(ctx context.Context, req *apiAuth.RefreshRequest) (*apiAuth.TokenResponse, *xError.Error) {
	l.log.Info(ctx, "Refresh - 刷新令牌")

	// 原子消费刷新令牌（GETDEL）：并发请求对同一 RT 仅一个成功，防止会话克隆
	consumed, xErr := l.repo.token.ConsumeRefreshToken(ctx, req.RefreshToken)
	if xErr != nil {
		return nil, xErr
	}
	if !consumed {
		return nil, xError.NewError(ctx, xError.TokenExpired, "刷新令牌无效或已过期", false, nil)
	}

	l.log.Info(ctx, "Refresh - 令牌刷新成功")
	return l.generateTokens(ctx)
}

// Logout 用户登出，清除刷新令牌（访问令牌等待自然过期）
func (l *AuthLogic) Logout(ctx context.Context, refreshToken string) *xError.Error {
	l.log.Info(ctx, "Logout - 用户登出")

	// 删除刷新令牌
	if xErr := l.repo.token.DeleteRefreshToken(ctx, refreshToken); xErr != nil {
		return xErr
	}

	l.log.Info(ctx, "Logout - 用户登出成功")
	return nil
}

// ValidateAccessToken 验证访问令牌的有效性
func (l *AuthLogic) ValidateAccessToken(ctx context.Context, accessToken string) (bool, *xError.Error) {
	l.log.Info(ctx, "ValidateAccessToken - 验证访问令牌")

	// 从 Redis 检查令牌是否存在
	found, xErr := l.repo.token.GetAccessToken(ctx, accessToken)
	if xErr != nil {
		return false, xErr
	}
	if !found {
		return false, xError.NewError(ctx, xError.TokenInvalid, "访问令牌无效或已过期", false, nil)
	}

	return true, nil
}

// GetCurrentUser 获取当前用户信息（单用户模式，从 Info 表读取）
func (l *AuthLogic) GetCurrentUser(ctx context.Context) (*apiUser.UserInfoResponse, *xError.Error) {
	l.log.Info(ctx, "GetCurrentUser - 获取当前用户信息")

	// 从 Info 表读取 owner 用户名与邮箱
	username, email, xErr := l.GetOwnerInfo(ctx)
	if xErr != nil {
		return nil, xErr
	}

	l.log.Info(ctx, "GetCurrentUser - 获取当前用户信息成功")

	// 读取生物特征状态（Info 表标记，由 BiometricLogic 维护）
	biometricEnabled := false
	if val, xErr := l.repo.info.GetByKey(ctx, bConst.InfoKeyOwnerBiometricEnabled); xErr == nil {
		biometricEnabled = val == "true"
	}

	return &apiUser.UserInfoResponse{
		Username:                 username,
		Email:                    email,
		BiometricEnabled:         biometricEnabled,
		BiometricCredentialCount: 0, // TODO(Task 5): 由 BiometricLogic 注入真实数量
	}, nil
}

// generateTokens 生成新的 AccessToken 和 RefreshToken 并存储到 Redis
// Login、Refresh、BiometricLogin 共用此方法
func (l *AuthLogic) generateTokens(ctx context.Context) (*apiAuth.TokenResponse, *xError.Error) {
	at := xUtil.Security().GenerateKey()
	rt := xUtil.Security().GenerateKey()

	if xErr := l.repo.token.SetAccessToken(ctx, at); xErr != nil {
		return nil, xErr
	}
	if xErr := l.repo.token.SetRefreshToken(ctx, rt); xErr != nil {
		return nil, xErr
	}

	return &apiAuth.TokenResponse{
		AccessToken:  at,
		RefreshToken: rt,
		ExpiresIn:    int64((2 * time.Hour).Seconds()),
	}, nil
}

// UpdateProfile 更新个人资料（用户名 + 邮箱）
func (l *AuthLogic) UpdateProfile(ctx context.Context, req *apiUser.UpdateProfileRequest) *xError.Error {
	l.log.Info(ctx, "UpdateProfile - 更新个人资料")

	kv := map[string]string{
		bConst.InfoKeyOwnerUsername: req.Username,
		bConst.InfoKeyOwnerEmail:    req.Email,
	}
	if xErr := l.repo.info.UpdateValuesInTx(ctx, kv); xErr != nil {
		return xError.NewError(ctx, xError.DatabaseError, "更新个人资料失败", false, nil)
	}

	l.log.Info(ctx, "UpdateProfile - 个人资料更新成功")
	return nil
}

// UpdatePassword 修改登录密码（验证旧密码 + 更新新密码 + 撤销所有现有 token）
func (l *AuthLogic) UpdatePassword(ctx context.Context, req *apiUser.UpdatePasswordRequest) *xError.Error {
	l.log.Info(ctx, "UpdatePassword - 修改密码")

	oldHash, xErr := l.repo.info.GetByKey(ctx, bConst.InfoKeyOwnerPassword)
	if xErr != nil {
		return xError.NewError(ctx, xError.DatabaseError, "读取密码信息失败", false, nil)
	}
	if !xUtil.Password().IsValid(req.OldPassword, oldHash) {
		return xError.NewError(ctx, xError.LoginFailed, "旧密码错误", false, nil)
	}

	// 新密码长度校验（防 bcrypt 72 字节上限触发 panic）
	if xErr := validatePasswordLength(ctx, req.NewPassword); xErr != nil {
		return xErr
	}

	newHash := xUtil.Password().MustEncryptString(req.NewPassword)
	if xErr := l.repo.info.UpdateValue(ctx, bConst.InfoKeyOwnerPassword, newHash); xErr != nil {
		return xError.NewError(ctx, xError.DatabaseError, "更新密码失败", false, nil)
	}

	// 撤销全部 access/refresh token，强制所有会话重新登录
	// （防止已泄露的 refresh_token 跨改密无限续期）
	if xErr := l.repo.token.ClearAllTokens(ctx); xErr != nil {
		return xError.NewError(ctx, xError.DatabaseError, "撤销令牌失败", false, nil)
	}

	l.log.Info(ctx, "UpdatePassword - 密码修改成功")
	return nil
}

// validatePasswordLength 校验密码长度，防止 base64 编码后超过 bcrypt 72 字节上限
// 导致 MustEncryptString panic（框架 Password.Encrypt 先 base64 再 bcrypt，
// 54 字符明文经 base64 后约 72 字节，55 字符起即超限）。
func validatePasswordLength(ctx context.Context, password string) *xError.Error {
	if len(password) > 54 {
		return xError.NewError(ctx, xError.ValidationError, "密码长度不能超过 54 个字符", false, nil)
	}
	return nil
}
