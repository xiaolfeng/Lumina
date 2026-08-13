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

// PreviewFileRepo 预览文件数据访问层，提供按 (session, filename) 覆写写入与按会话查询能力。
//
// 字段说明:
//   - db:  GORM 数据库实例，执行持久化操作
//   - log: 带命名空间的结构化日志记录器
type PreviewFileRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewPreviewFileRepo 创建 PreviewFileRepo 实例
//
// 参数说明:
//   - db: 已初始化的 GORM 数据库实例
//
// 返回值:
//   - *PreviewFileRepo: 配置完成的 PreviewFileRepo 实例指针
func NewPreviewFileRepo(db *gorm.DB) *PreviewFileRepo {
	return &PreviewFileRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "PreviewFileRepo"),
	}
}

// CreateOrUpdate 创建或覆写预览文件（按 (session_id, filename) 唯一约束）
//
//   - 已存在：更新 MimeType、Content、Size、UpdatedAt 字段
//   - 不存在：创建新记录
//
// 参数:
//   - ctx:  上下文对象
//   - file: 待创建或覆写的预览文件实体
//
// 返回值:
//   - *entity.PreviewFile: 持久化后的实体
//   - *xError.Error:        操作过程中的错误
func (r *PreviewFileRepo) CreateOrUpdate(ctx context.Context, file *entity.PreviewFile) (*entity.PreviewFile, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("CreateOrUpdate - 创建或覆写预览文件 [%d/%s]", file.SessionID.Int64(), file.Filename))

	var existing entity.PreviewFile
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND filename = ?", file.SessionID, file.Filename).
		First(&existing).Error

	if err == nil {
		// 已存在 → 覆写
		existing.MimeType = file.MimeType
		existing.Content = file.Content
		existing.Size = file.Size
		if saveErr := r.db.WithContext(ctx).Save(&existing).Error; saveErr != nil {
			r.log.Warn(ctx, saveErr.Error())
			return nil, xError.NewError(ctx, xError.DatabaseError, "更新预览文件失败", false, saveErr)
		}
		return &existing, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 不存在 → 创建
		if createErr := r.db.WithContext(ctx).Create(file).Error; createErr != nil {
			r.log.Warn(ctx, createErr.Error())
			return nil, xError.NewError(ctx, xError.DatabaseError, "创建预览文件失败", false, createErr)
		}
		return file, nil
	}

	// 其他数据库错误
	r.log.Warn(ctx, err.Error())
	return nil, xError.NewError(ctx, xError.DatabaseError, "查询预览文件失败", false, err)
}

// GetByID 根据 ID 获取预览文件
//
// 参数:
//   - ctx: 上下文对象
//   - id:  预览文件雪花 ID
//
// 返回值:
//   - *entity.PreviewFile: 查询到的预览文件实体
//   - *xError.Error:        查询过程中的错误
func (r *PreviewFileRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.PreviewFile, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据 ID 获取预览文件 [%d]", id.Int64()))

	var file entity.PreviewFile
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "预览文件不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询预览文件失败", false, err)
	}
	return &file, nil
}

// GetBySessionAndFilename 根据会话 ID 与文件名获取预览文件
//
// 参数:
//   - ctx:       上下文对象
//   - sessionID: 预览会话雪花 ID
//   - filename:  文件名（扁平单层）
//
// 返回值:
//   - *entity.PreviewFile: 查询到的预览文件实体
//   - *xError.Error:        查询过程中的错误
func (r *PreviewFileRepo) GetBySessionAndFilename(ctx context.Context, sessionID xSnowflake.SnowflakeID, filename string) (*entity.PreviewFile, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetBySessionAndFilename - 根据会话与文件名获取预览文件 [%d/%s]", sessionID.Int64(), filename))

	var file entity.PreviewFile
	if err := r.db.WithContext(ctx).
		Where("session_id = ? AND filename = ?", sessionID, filename).
		First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "预览文件不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询预览文件失败", false, err)
	}
	return &file, nil
}

// ListBySession 根据会话 ID 获取全部预览文件列表（按文件名升序）
//
// 参数:
//   - ctx:       上下文对象
//   - sessionID: 预览会话雪花 ID
//
// 返回值:
//   - []*entity.PreviewFile: 预览文件列表（无记录时返回空切片）
//   - *xError.Error:          查询过程中的错误
func (r *PreviewFileRepo) ListBySession(ctx context.Context, sessionID xSnowflake.SnowflakeID) ([]*entity.PreviewFile, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("ListBySession - 根据会话获取预览文件列表 [%d]", sessionID.Int64()))

	files := make([]*entity.PreviewFile, 0)
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("filename ASC").
		Find(&files).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询预览文件列表失败", false, err)
	}

	return files, nil
}
