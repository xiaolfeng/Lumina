package repository

import (
	"context"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// LlmProviderRepo LLM Provider 数据访问层
type LlmProviderRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewLlmProviderRepo 创建 LlmProviderRepo 实例
func NewLlmProviderRepo(db *gorm.DB) *LlmProviderRepo {
	return &LlmProviderRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "LlmProviderRepo"),
	}
}

// Create 创建 LLM Provider
func (r *LlmProviderRepo) Create(ctx context.Context, provider *entity.LlmProvider) *xError.Error {
	r.log.Info(ctx, "Create - 创建LLM Provider")

	if err := r.db.WithContext(ctx).Create(provider).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "创建LLM Provider失败", false, err)
	}

	return nil
}

// GetByID 根据ID获取LLM Provider
func (r *LlmProviderRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.LlmProvider, *xError.Error) {
	r.log.Info(ctx, "GetByID - 根据ID获取LLM Provider")

	var provider entity.LlmProvider
	if err := r.db.WithContext(ctx).First(&provider, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "LLM Provider不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询LLM Provider失败", false, err)
	}

	return &provider, nil
}

// List 分页获取LLM Provider列表
func (r *LlmProviderRepo) List(ctx context.Context, page, size int) ([]*entity.LlmProvider, int64, *xError.Error) {
	r.log.Info(ctx, "List - 分页获取LLM Provider列表")

	var total int64

	if err := r.db.WithContext(ctx).Model(&entity.LlmProvider{}).Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计LLM Provider数量失败", false, err)
	}

	var items []*entity.LlmProvider
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&items).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询LLM Provider列表失败", false, err)
	}

	return items, total, nil
}

// Update 更新LLM Provider
func (r *LlmProviderRepo) Update(ctx context.Context, provider *entity.LlmProvider) *xError.Error {
	r.log.Info(ctx, "Update - 更新LLM Provider")

	if err := r.db.WithContext(ctx).Save(provider).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "更新LLM Provider失败", false, err)
	}

	return nil
}

// Delete 删除LLM Provider（硬删除）
func (r *LlmProviderRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, "Delete - 删除LLM Provider")

	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.LlmProvider{}).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "删除LLM Provider失败", false, err)
	}

	return nil
}

// ListActive 查询所有启用的LLM Provider
func (r *LlmProviderRepo) ListActive(ctx context.Context) ([]*entity.LlmProvider, *xError.Error) {
	r.log.Info(ctx, "ListActive - 查询启用的LLM Provider列表")

	var items []*entity.LlmProvider
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询启用的LLM Provider列表失败", false, err)
	}

	return items, nil
}
