package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xUtil "github.com/bamboo-services/bamboo-base-go/common/utility"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	"github.com/google/uuid"
	apiAuth "github.com/xiaolfeng/Lumina/api/auth"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"gorm.io/gorm"
)

// authRepo 认证模块依赖的仓储集合
type authRepo struct {
	token *repository.TokenRepo
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
			db:  db,
			rdb: rdb,
			log: xLog.WithName(xLog.NamedLOGC, "AuthLogic"),
		},
		repo: authRepo{
			token: repository.NewTokenRepo(rdb),
		},
	}
}

// GetInitialStatus 获取系统是否为初始状态（未初始化）
func (l *AuthLogic) GetInitialStatus(ctx context.Context) (bool, *xError.Error) {
	l.log.Info(ctx, "GetInitialStatus - 检查系统初始化状态")

	var info entity.Info
	if err := l.db.WithContext(ctx).Model(&entity.Info{}).Where("key = ?", "is_initial").First(&info).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return true, nil
		}
		return false, xError.NewError(ctx, xError.DatabaseError, "查询初始化状态失败", false, nil)
	}

	return info.Value == "true", nil
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

	// 在事务中执行：写入 owner 凭据 + 更新初始化状态
	if err := l.db.Transaction(func(tx *gorm.DB) error {
		// 写入 owner 用户名
		if err := tx.WithContext(ctx).Model(&entity.Info{}).
			Where("key = ?", "owner_username").
			Update("value", req.Username).Error; err != nil {
			return err
		}

		// 写入 owner 邮箱
		if err := tx.WithContext(ctx).Model(&entity.Info{}).
			Where("key = ?", "owner_email").
			Update("value", req.Email).Error; err != nil {
			return err
		}

		// 写入 owner 密码（bcrypt 加密）
		if err := tx.WithContext(ctx).Model(&entity.Info{}).
			Where("key = ?", "owner_password").
			Update("value", xUtil.Password().MustEncryptString(req.Password)).Error; err != nil {
			return err
		}

		// 标记系统已完成初始化
		if err := tx.WithContext(ctx).Model(&entity.Info{}).
			Where("key = ?", "is_initial").
			Update("value", "false").Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "系统初始化失败", false, err)
	}

	l.log.Info(ctx, "Initialize - 系统初始化成功")
	return nil
}

// Login 用户登录，支持用户名或邮箱登录，返回访问令牌与刷新令牌
func (l *AuthLogic) Login(ctx context.Context, req *apiAuth.LoginRequest) (*apiAuth.TokenResponse, *xError.Error) {
	l.log.Info(ctx, "Login - 用户登录")

	// 从 Info 表读取 owner 用户名
	var usernameInfo entity.Info
	if err := l.db.WithContext(ctx).Model(&entity.Info{}).Where("key = ?", "owner_username").First(&usernameInfo).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "读取用户信息失败", false, err)
	}

	// 从 Info 表读取 owner 邮箱
	var emailInfo entity.Info
	if err := l.db.WithContext(ctx).Model(&entity.Info{}).Where("key = ?", "owner_email").First(&emailInfo).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "读取用户信息失败", false, err)
	}

	// 根据是否包含 @ 判断登录方式，验证账号匹配
	accountMatched := false
	if strings.Contains(req.Account, "@") {
		accountMatched = req.Account == emailInfo.Value
	} else {
		accountMatched = req.Account == usernameInfo.Value
	}
	if !accountMatched {
		return nil, xError.NewError(ctx, xError.LoginFailed, "账号或密码错误", false, nil)
	}

	// 从 Info 表读取 owner 密码哈希
	var passwordInfo entity.Info
	if err := l.db.WithContext(ctx).Model(&entity.Info{}).Where("key = ?", "owner_password").First(&passwordInfo).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "读取用户信息失败", false, err)
	}

	// 验证密码
	if !xUtil.Password().IsValid(req.Password, passwordInfo.Value) {
		return nil, xError.NewError(ctx, xError.LoginFailed, "账号或密码错误", false, nil)
	}

	// 生成 AccessToken 和 RefreshToken
	at := uuid.New().String()
	atMD5Hash := md5.Sum([]byte(at))
	atMD5 := hex.EncodeToString(atMD5Hash[:])

	rt := uuid.New().String()

	// 存储令牌到 Redis
	if xErr := l.repo.token.SetAccessToken(ctx, atMD5); xErr != nil {
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

	// 生成新的 AccessToken 和 RefreshToken
	at := uuid.New().String()
	atMD5Hash := md5.Sum([]byte(at))
	atMD5 := hex.EncodeToString(atMD5Hash[:])

	rt := uuid.New().String()

	// 存储新令牌到 Redis
	if xErr := l.repo.token.SetAccessToken(ctx, atMD5); xErr != nil {
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

	// 计算 AccessToken 的 MD5 摘要
	atMD5Hash := md5.Sum([]byte(accessToken))
	atMD5 := hex.EncodeToString(atMD5Hash[:])

	// 从 Redis 检查令牌是否存在
	found, xErr := l.repo.token.GetAccessToken(ctx, atMD5)
	if xErr != nil {
		return false, xErr
	}
	if !found {
		return false, xError.NewError(ctx, xError.TokenInvalid, "访问令牌无效或已过期", false, nil)
	}

	return true, nil
}
