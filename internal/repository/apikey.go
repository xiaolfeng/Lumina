package repository

import (
	"context"
	"errors"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// ApikeyRepo API 密钥数据访问层
type ApikeyRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewApikeyRepo 创建 API 密钥数据访问层实例
func NewApikeyRepo(db *gorm.DB) *ApikeyRepo {
	return &ApikeyRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "ApikeyRepo"),
	}
}

// Create 创建 API Key
func (r *ApikeyRepo) Create(ctx context.Context, apikey *entity.Apikey) *xError.Error {
	r.log.Info(ctx, "Create - 创建 API Key")
	if err := r.db.WithContext(ctx).Create(apikey).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "创建 API Key 失败", false, err)
	}
	return nil
}

// GetByID 根据 ID 查询 API Key
func (r *ApikeyRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.Apikey, bool, *xError.Error) {
	r.log.Info(ctx, "GetByID - 根据 ID 查询 API Key")
	var apikey entity.Apikey
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&apikey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, xError.NewError(ctx, xError.DatabaseError, "查询 API Key 失败", false, err)
	}
	return &apikey, true, nil
}

// List 查询所有 API Key（按创建时间倒序）
func (r *ApikeyRepo) List(ctx context.Context) ([]entity.Apikey, *xError.Error) {
	r.log.Info(ctx, "List - 查询 API Key 列表")
	var apikeys []entity.Apikey
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&apikeys).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询 API Key 列表失败", false, err)
	}
	return apikeys, nil
}

// Update 更新 API Key（全量更新指定字段）
func (r *ApikeyRepo) Update(ctx context.Context, apikey *entity.Apikey) *xError.Error {
	r.log.Info(ctx, "Update - 更新 API Key")
	if err := r.db.WithContext(ctx).Save(apikey).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "更新 API Key 失败", false, err)
	}
	return nil
}

// Delete 删除 API Key（硬删除）
func (r *ApikeyRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, "Delete - 删除 API Key")
	if err := r.db.WithContext(ctx).Delete(&entity.Apikey{}, id).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "删除 API Key 失败", false, err)
	}
	return nil
}

// GetByKeyHash 根据 KeyHash 查询 API Key（预留，用于未来认证中间件）
func (r *ApikeyRepo) GetByKeyHash(ctx context.Context, keyHash string) (*entity.Apikey, bool, *xError.Error) {
	r.log.Info(ctx, "GetByKeyHash - 根据 KeyHash 查询 API Key")
	var apikey entity.Apikey
	if err := r.db.WithContext(ctx).Where("key_hash = ?", keyHash).First(&apikey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, xError.NewError(ctx, xError.DatabaseError, "查询 API Key 失败", false, err)
	}
	return &apikey, true, nil
}
