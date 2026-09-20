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

// PageFileRepo 页面快照文件数据访问层
type PageFileRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewPageFileRepo 创建 PageFileRepo 实例
func NewPageFileRepo(db *gorm.DB) *PageFileRepo {
	return &PageFileRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "PageFileRepo"),
	}
}

// BatchCreate 批量写入快照文件（晋升事务内调用）
func (r *PageFileRepo) BatchCreate(ctx context.Context, files []*entity.PageFile) *xError.Error {
	if len(files) == 0 {
		return nil
	}
	r.log.Info(ctx, fmt.Sprintf("BatchCreate - 批量创建页面文件 [count=%d]", len(files)))

	if err := r.db.WithContext(ctx).Create(&files).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "创建页面文件失败", false, err)
	}
	return nil
}

// ListByVersion 按文件名升序列出指定版本的全部文件
func (r *PageFileRepo) ListByVersion(ctx context.Context, versionID xSnowflake.SnowflakeID) ([]*entity.PageFile, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("ListByVersion - 列出版本文件 [%d]", versionID.Int64()))

	files := make([]*entity.PageFile, 0)
	if err := r.db.WithContext(ctx).
		Where("version_id = ?", versionID).
		Order("filename ASC").
		Find(&files).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询页面文件列表失败", false, err)
	}
	return files, nil
}

// GetByVersionAndFilename 根据版本 ID 与文件名获取快照文件
func (r *PageFileRepo) GetByVersionAndFilename(ctx context.Context, versionID xSnowflake.SnowflakeID, filename string) (*entity.PageFile, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByVersionAndFilename - 获取版本文件 [%d/%s]", versionID.Int64(), filename))

	var file entity.PageFile
	if err := r.db.WithContext(ctx).
		Where("version_id = ? AND filename = ?", versionID, filename).
		First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "页面文件不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询页面文件失败", false, err)
	}
	return &file, nil
}
