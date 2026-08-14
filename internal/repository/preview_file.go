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

// CountBySessions 批量统计各会话的文件数量（按 session_id 分组）。
//
// 用于列表场景一次性填充文件数，避免 N+1 逐会话查询。
//
// 参数:
//   - ctx:        上下文对象
//   - sessionIDs: 会话雪花 ID 切片（空切片安全，返回空 map）
//
// 返回值:
//   - map[int64]int64: 会话 ID → 文件数映射
//   - *xError.Error:   查询过程中的错误
func (r *PreviewFileRepo) CountBySessions(ctx context.Context, sessionIDs []xSnowflake.SnowflakeID) (map[int64]int64, *xError.Error) {
	result := make(map[int64]int64)
	if len(sessionIDs) == 0 {
		return result, nil
	}

	type countRow struct {
		SessionID int64 `gorm:"column:session_id"`
		Count     int64 `gorm:"column:count"`
	}

	rows := make([]countRow, 0)
	if err := r.db.WithContext(ctx).
		Model(&entity.PreviewFile{}).
		Select("session_id, COUNT(*) AS count").
		Where("session_id IN ?", sessionIDs).
		Group("session_id").
		Scan(&rows).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "批量统计预览文件数量失败", false, err)
	}

	for _, row := range rows {
		result[row.SessionID] = row.Count
	}
	return result, nil
}

// DeleteBySession 物理删除指定会话的全部预览文件（级联清理用）
//
// 参数:
//   - ctx:       上下文对象
//   - sessionID: 预览会话雪花 ID
//
// 返回值:
//   - *xError.Error: 删除过程中的错误
func (r *PreviewFileRepo) DeleteBySession(ctx context.Context, sessionID xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("DeleteBySession - 删除会话全部预览文件 [%d]", sessionID.Int64()))

	if err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Delete(&entity.PreviewFile{}).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "删除会话预览文件失败", false, err)
	}
	return nil
}

// Delete 物理删除单个预览文件
//
// 参数:
//   - ctx: 上下文对象
//   - id:  待删除的预览文件雪花 ID
//
// 返回值:
//   - *xError.Error: 删除过程中的错误
func (r *PreviewFileRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Delete - 删除预览文件 [%d]", id.Int64()))

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.PreviewFile{})
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "删除预览文件失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "预览文件不存在", false, nil)
	}
	return nil
}
