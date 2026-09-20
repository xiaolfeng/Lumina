package repository

import (
	"context"
	"errors"
	"fmt"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// PageVersionRepo 页面版本数据访问层
type PageVersionRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewPageVersionRepo 创建 PageVersionRepo 实例
func NewPageVersionRepo(db *gorm.DB) *PageVersionRepo {
	return &PageVersionRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "PageVersionRepo"),
	}
}

// Create 创建页面版本
func (r *PageVersionRepo) Create(ctx context.Context, version *entity.PageVersion) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Create - 创建页面版本 [%d/%s]", version.PageID.Int64(), version.Version))

	if err := r.db.WithContext(ctx).Create(version).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "创建页面版本失败", false, err)
	}
	return nil
}

// GetByID 根据 ID 获取页面版本
func (r *PageVersionRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.PageVersion, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据 ID 获取页面版本 [%d]", id.Int64()))

	var version entity.PageVersion
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "页面版本不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询页面版本失败", false, err)
	}
	return &version, nil
}

// ListByPage 按创建时间降序列出页面全部版本
func (r *PageVersionRepo) ListByPage(ctx context.Context, pageID xSnowflake.SnowflakeID) ([]*entity.PageVersion, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("ListByPage - 列出页面版本 [%d]", pageID.Int64()))

	versions := make([]*entity.PageVersion, 0)
	if err := r.db.WithContext(ctx).
		Where("page_id = ?", pageID).
		Order("created_at DESC").
		Find(&versions).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询页面版本列表失败", false, err)
	}
	return versions, nil
}

// GetByPageAndVersionLabel 根据页面 ID 与语义化版本号获取版本
func (r *PageVersionRepo) GetByPageAndVersionLabel(ctx context.Context, pageID xSnowflake.SnowflakeID, version string) (*entity.PageVersion, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByPageAndVersionLabel - 根据版本号获取页面版本 [%d/%s]", pageID.Int64(), version))

	var item entity.PageVersion
	if err := r.db.WithContext(ctx).
		Where("page_id = ? AND version = ?", pageID, version).
		First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "页面版本不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询页面版本失败", false, err)
	}
	return &item, nil
}

// GetLatestByPage 获取指定页面最近创建的版本（用于递增建议）
func (r *PageVersionRepo) GetLatestByPage(ctx context.Context, pageID xSnowflake.SnowflakeID) (*entity.PageVersion, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetLatestByPage - 获取页面最近版本 [%d]", pageID.Int64()))

	var version entity.PageVersion
	if err := r.db.WithContext(ctx).
		Where("page_id = ?", pageID).
		Order("created_at DESC").
		First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "页面版本不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询页面版本失败", false, err)
	}
	return &version, nil
}
