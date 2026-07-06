package repository

import (
	"context"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// LlmModelRepo LLM 模型数据访问层
type LlmModelRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewLlmModelRepo 创建 LlmModelRepo 实例
func NewLlmModelRepo(db *gorm.DB) *LlmModelRepo {
	return &LlmModelRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "LlmModelRepo"),
	}
}

// Create 创建 LLM Model
func (r *LlmModelRepo) Create(ctx context.Context, model *entity.LlmModel) *xError.Error {
	r.log.Info(ctx, "Create - 创建LLM Model")

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "创建LLM Model失败", false, err)
	}

	return nil
}

// GetByID 根据ID获取LLM Model
func (r *LlmModelRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.LlmModel, *xError.Error) {
	r.log.Info(ctx, "GetByID - 根据ID获取LLM Model")

	var model entity.LlmModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "LLM Model不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询LLM Model失败", false, err)
	}

	return &model, nil
}

// List 分页获取LLM Model列表
func (r *LlmModelRepo) List(ctx context.Context, page, size int) ([]*entity.LlmModel, int64, *xError.Error) {
	r.log.Info(ctx, "List - 分页获取LLM Model列表")

	var total int64

	if err := r.db.WithContext(ctx).Model(&entity.LlmModel{}).Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计LLM Model数量失败", false, err)
	}

	var items []*entity.LlmModel
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&items).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询LLM Model列表失败", false, err)
	}

	return items, total, nil
}

// Update 更新LLM Model
func (r *LlmModelRepo) Update(ctx context.Context, model *entity.LlmModel) *xError.Error {
	r.log.Info(ctx, "Update - 更新LLM Model")

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "更新LLM Model失败", false, err)
	}

	return nil
}

// Delete 删除LLM Model（硬删除）
func (r *LlmModelRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, "Delete - 删除LLM Model")

	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.LlmModel{}).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "删除LLM Model失败", false, err)
	}

	return nil
}

// ListByProviderID 根据 ProviderID 获取全部 LLM Model
func (r *LlmModelRepo) ListByProviderID(ctx context.Context, providerID xSnowflake.SnowflakeID) ([]*entity.LlmModel, *xError.Error) {
	r.log.Info(ctx, "ListByProviderID - 根据ProviderID获取LLM Model列表")

	var items []*entity.LlmModel
	if err := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "根据ProviderID查询LLM Model列表失败", false, err)
	}

	return items, nil
}

// GetActiveByProviderID 根据 ProviderID 获取启用的 LLM Model
func (r *LlmModelRepo) GetActiveByProviderID(ctx context.Context, providerID xSnowflake.SnowflakeID) ([]*entity.LlmModel, *xError.Error) {
	r.log.Info(ctx, "GetActiveByProviderID - 根据ProviderID获取启用的LLM Model列表")

	var items []*entity.LlmModel
	if err := r.db.WithContext(ctx).
		Where("provider_id = ? AND is_active = ?", providerID, true).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "根据ProviderID查询启用的LLM Model列表失败", false, err)
	}

	return items, nil
}
