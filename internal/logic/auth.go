package logic

import (
	"context"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xUtil "github.com/bamboo-services/bamboo-base-go/common/utility"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	apiAuth "github.com/xiaolfeng/Lumina/api/auth"
	apiUser "github.com/xiaolfeng/Lumina/api/user"
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

	value, xErr := l.repo.info.GetByKey(ctx, "is_initial")
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

	// 检查是否已初始化
	isInitial, xErr := l.GetInitialStatus(ctx)
	if xErr != nil {
		return xErr
	}
	if !isInitial {
		return xError.NewError(ctx, xError.RepeatOperation, "系统已初始化，不可重复操作", false, nil)
	}

	// 在事务中原子写入：owner 凭据 + 初始化状态
	// bcrypt 加密属于业务逻辑，在 logic 层完成；事务边界下沉到 InfoRepo
	kv := map[string]string{
		"owner_username": req.Username,
		"owner_email":    req.Email,
		"owner_password": xUtil.Password().MustEncryptString(req.Password),
		"is_initial":     "false",
	}
	if xErr := l.repo.info.UpdateValuesInTx(ctx, kv); xErr != nil {
		return xError.NewError(ctx, xError.DatabaseError, "系统初始化失败", false, nil)
	}

	l.log.Info(ctx, "Initialize - 系统初始化成功")
	return nil
}

// Login 用户登录，支持用户名或邮箱登录，返回访问令牌与刷新令牌
func (l *AuthLogic) Login(ctx context.Context, req *apiAuth.LoginRequest) (*apiAuth.TokenResponse, *xError.Error) {
	l.log.Info(ctx, "Login - 用户登录")

	// 从 Info 表读取 owner 用户名
	username, xErr := l.repo.info.GetByKey(ctx, "owner_username")
	if xErr != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "读取用户信息失败", false, nil)
	}

	// 从 Info 表读取 owner 邮箱
	email, xErr := l.repo.info.GetByKey(ctx, "owner_email")
	if xErr != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "读取用户信息失败", false, nil)
	}

	// 根据是否包含 @ 判断登录方式，验证账号匹配
	accountMatched := false
	if strings.Contains(req.Account, "@") {
		accountMatched = req.Account == email
	} else {
		accountMatched = req.Account == username
	}
	if !accountMatched {
		return nil, xError.NewError(ctx, xError.LoginFailed, "账号或密码错误", false, nil)
	}

	// 从 Info 表读取 owner 密码哈希
	passwordHash, xErr := l.repo.info.GetByKey(ctx, "owner_password")
	if xErr != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "读取用户信息失败", false, nil)
	}

	// 验证密码
	if !xUtil.Password().IsValid(req.Password, passwordHash) {
		return nil, xError.NewError(ctx, xError.LoginFailed, "账号或密码错误", false, nil)
	}

	// 使用框架安全密钥生成 AccessToken 和 RefreshToken
	at := xUtil.Security().GenerateKey()
	rt := xUtil.Security().GenerateKey()

	// 存储令牌到 Redis
	if xErr := l.repo.token.SetAccessToken(ctx, at); xErr != nil {
		return nil, xErr
	}
	if xErr := l.repo.token.SetRefreshToken(ctx, rt); xErr != nil {
		return nil, xErr
	}

	l.log.Info(ctx, "Login - 用户登录成功")

	return &apiAuth.TokenResponse{
		AccessToken:  at,
		RefreshToken: rt,
		ExpiresIn:    int64((2 * time.Hour).Seconds()),
	}, nil
}

// Refresh 使用刷新令牌换取新的访问令牌和刷新令牌
func (l *AuthLogic) Refresh(ctx context.Context, req *apiAuth.RefreshRequest) (*apiAuth.TokenResponse, *xError.Error) {
	l.log.Info(ctx, "Refresh - 刷新令牌")

	// 验证刷新令牌
	found, xErr := l.repo.token.GetRefreshToken(ctx, req.RefreshToken)
	if xErr != nil {
		return nil, xErr
	}
	if !found {
		return nil, xError.NewError(ctx, xError.TokenExpired, "刷新令牌无效或已过期", false, nil)
	}

	// 删除旧的刷新令牌
	if xErr := l.repo.token.DeleteRefreshToken(ctx, req.RefreshToken); xErr != nil {
		return nil, xErr
	}

	// 使用框架安全密钥生成新的 AccessToken 和 RefreshToken
	at := xUtil.Security().GenerateKey()
	rt := xUtil.Security().GenerateKey()

	// 存储新令牌到 Redis
	if xErr := l.repo.token.SetAccessToken(ctx, at); xErr != nil {
		return nil, xErr
	}
	if xErr := l.repo.token.SetRefreshToken(ctx, rt); xErr != nil {
		return nil, xErr
	}

	l.log.Info(ctx, "Refresh - 令牌刷新成功")

	return &apiAuth.TokenResponse{
		AccessToken:  at,
		RefreshToken: rt,
		ExpiresIn:    int64((2 * time.Hour).Seconds()),
	}, nil
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

	// 从 Info 表读取 owner 用户名
	username, xErr := l.repo.info.GetByKey(ctx, "owner_username")
	if xErr != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "读取用户信息失败", false, nil)
	}

	// 从 Info 表读取 owner 邮箱
	email, xErr := l.repo.info.GetByKey(ctx, "owner_email")
	if xErr != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "读取用户信息失败", false, nil)
	}

	l.log.Info(ctx, "GetCurrentUser - 获取当前用户信息成功")

	return &apiUser.UserInfoResponse{
		Username: username,
		Email:    email,
	}, nil
}
