package repository

import (
	"context"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// ApikeyRepo API密钥数据访问层
type ApikeyRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewApikeyRepo 创建ApikeyRepo实例
func NewApikeyRepo(db *gorm.DB) *ApikeyRepo {
	return &ApikeyRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "ApikeyRepo"),
	}
}

// Create 创建API密钥
func (r *ApikeyRepo) Create(ctx context.Context, apikey *entity.Apikey) *xError.Error {
	r.log.Info(ctx, "Create - 创建API密钥")

	if err := r.db.WithContext(ctx).Create(apikey).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "创建API密钥失败", false, err)
	}

	return nil
}

// GetByID 根据ID获取API密钥
func (r *ApikeyRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.Apikey, *xError.Error) {
	r.log.Info(ctx, "GetByID - 根据ID获取API密钥")

	var apikey entity.Apikey
	if err := r.db.WithContext(ctx).First(&apikey, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "API密钥不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询API密钥失败", false, err)
	}

	return &apikey, nil
}

// List 分页获取API密钥列表
func (r *ApikeyRepo) List(ctx context.Context, page, size int) ([]entity.Apikey, int64, *xError.Error) {
	r.log.Info(ctx, "List - 分页获取API密钥列表")

	var items []entity.Apikey
	var total int64

	if err := r.db.WithContext(ctx).Model(&entity.Apikey{}).Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计API密钥数量失败", false, err)
	}

	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&items).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询API密钥列表失败", false, err)
	}

	return items, total, nil
}

// Update 更新API密钥
func (r *ApikeyRepo) Update(ctx context.Context, apikey *entity.Apikey) *xError.Error {
	r.log.Info(ctx, "Update - 更新API密钥")

	if err := r.db.WithContext(ctx).Save(apikey).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "更新API密钥失败", false, err)
	}

	return nil
}

// Delete 删除API密钥（硬删除）
func (r *ApikeyRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, "Delete - 删除API密钥")

	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Apikey{}).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "删除API密钥失败", false, err)
	}

	return nil
}
