package repository

import (
	"context"
	"errors"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// UserRepo 用户数据访问层
type UserRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewUserRepo 创建用户数据访问层实例
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "UserRepo"),
	}
}

// Create 创建用户
func (r *UserRepo) Create(ctx context.Context, user *entity.User) *xError.Error {
	r.log.Info(ctx, "Create - 创建用户")

	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "创建用户失败", false, err)
	}

	return nil
}

// GetByUsername 根据用户名查询用户
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*entity.User, bool, *xError.Error) {
	r.log.Info(ctx, "GetByUsername - 根据用户名查询用户")

	var user entity.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, xError.NewError(ctx, xError.DatabaseError, "查询用户失败", false, err)
	}

	return &user, true, nil
}

// GetByEmail 根据邮箱查询用户
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, bool, *xError.Error) {
	r.log.Info(ctx, "GetByEmail - 根据邮箱查询用户")

	var user entity.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, xError.NewError(ctx, xError.DatabaseError, "查询用户失败", false, err)
	}

	return &user, true, nil
}

// GetByID 根据 ID 查询用户
func (r *UserRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.User, bool, *xError.Error) {
	r.log.Info(ctx, "GetByID - 根据 ID 查询用户")

	var user entity.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, xError.NewError(ctx, xError.DatabaseError, "查询用户失败", false, err)
	}

	return &user, true, nil
}

// UpdateLastLoginAt 更新用户最后登录时间
func (r *UserRepo) UpdateLastLoginAt(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, "UpdateLastLoginAt - 更新用户最后登录时间")

	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("last_login_at", time.Now()).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "更新登录时间失败", false, err)
	}

	return nil
}
