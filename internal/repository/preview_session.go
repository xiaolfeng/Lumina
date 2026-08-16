package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// PreviewSessionRepo 预览会话数据访问层，提供 CRUD 与按项目/哈希查询能力。
//
// 预览会话为活动工作区（1:N 多工作区），无 Redis 缓存（与 Pin 一致），
// 因为会话内容依赖文件子表实时变更，缓存层会引入一致性成本。
//
// 字段说明:
//   - db:  GORM 数据库实例，执行持久化操作
//   - log: 带命名空间的结构化日志记录器
type PreviewSessionRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewPreviewSessionRepo 创建 PreviewSessionRepo 实例
//
// 参数说明:
//   - db: 已初始化的 GORM 数据库实例
//
// 返回值:
//   - *PreviewSessionRepo: 配置完成的 PreviewSessionRepo 实例指针
func NewPreviewSessionRepo(db *gorm.DB) *PreviewSessionRepo {
	return &PreviewSessionRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "PreviewSessionRepo"),
	}
}

// Create 创建预览会话
//
// 参数:
//   - ctx:     上下文对象
//   - session: 待创建的预览会话实体（ID 由雪花算法自动生成）
//
// 返回值:
//   - *xError.Error: 创建过程中的错误
func (r *PreviewSessionRepo) Create(ctx context.Context, session *entity.PreviewSession) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Create - 创建预览会话 [%s]", session.Title))

	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "创建预览会话失败", false, err)
	}
	return nil
}

// GetByID 根据 ID 获取预览会话
//
// 参数:
//   - ctx: 上下文对象
//   - id:  预览会话雪花 ID
//
// 返回值:
//   - *entity.PreviewSession: 查询到的预览会话实体
//   - *xError.Error:          查询过程中的错误
func (r *PreviewSessionRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.PreviewSession, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据 ID 获取预览会话 [%d]", id.Int64()))

	var session entity.PreviewSession
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "预览会话不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询预览会话失败", false, err)
	}
	return &session, nil
}

// GetByHash 根据访问哈希获取预览会话（公开访问鉴权用）
//
// 参数:
//   - ctx:  上下文对象
//   - hash: 预览会话访问哈希（32 位 hex）
//
// 返回值:
//   - *entity.PreviewSession: 查询到的预览会话实体
//   - *xError.Error:          查询过程中的错误
func (r *PreviewSessionRepo) GetByHash(ctx context.Context, hash string) (*entity.PreviewSession, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByHash - 根据哈希获取预览会话 [%s]", hash))

	var session entity.PreviewSession
	if err := r.db.WithContext(ctx).Where("hash = ?", hash).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "预览会话不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询预览会话失败", false, err)
	}
	return &session, nil
}

// List 分页获取预览会话列表（按创建时间降序，projectID 为零值时不过滤）
//
// 参数:
//   - ctx:       上下文对象
//   - projectID: 项目雪花 ID（零值表示不按项目过滤）
//   - page:      页码（从 1 开始）
//   - size:      每页数量
//
// 返回值:
//   - []*entity.PreviewSession: 当前页的预览会话列表
//   - int64:                    符合条件的总记录数
//   - *xError.Error:            查询过程中的错误
func (r *PreviewSessionRepo) List(ctx context.Context, projectID xSnowflake.SnowflakeID, page, size int) ([]*entity.PreviewSession, int64, *xError.Error) {
	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	page, size = int(pageReq.Page), int(pageReq.Size)
	r.log.Info(ctx, fmt.Sprintf("List - 分页获取预览会话列表 [projectID=%d, page=%d, size=%d]", projectID.Int64(), page, size))

	query := r.db.WithContext(ctx).Model(&entity.PreviewSession{})
	if !projectID.IsZero() {
		query = query.Where("project_id = ?", projectID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计预览会话数量失败", false, err)
	}

	sessions := make([]*entity.PreviewSession, 0)
	offset := (page - 1) * size
	if err := query.
		Offset(offset).
		Limit(size).
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询预览会话列表失败", false, err)
	}

	return sessions, total, nil
}

// Delete 物理删除预览会话
//
// 参数:
//   - ctx: 上下文对象
//   - id:  待删除的预览会话雪花 ID
//
// 返回值:
//   - *xError.Error: 删除过程中的错误
func (r *PreviewSessionRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Delete - 删除预览会话 [%d]", id.Int64()))

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.PreviewSession{})
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "删除预览会话失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "预览会话不存在", false, nil)
	}
	return nil
}

// DeleteCascade 事务级联删除预览会话及其下全部预览文件
//
// 在同一数据库事务内先删除会话下全部文件，再删除会话本体；
// 会话不存在时（RowsAffected==0）返回 NotFound，避免遗留孤儿文件。
//
// 参数:
//   - ctx: 上下文对象
//   - id:  待删除的预览会话雪花 ID
//
// 返回值:
//   - *xError.Error: 删除过程中的错误
func (r *PreviewSessionRepo) DeleteCascade(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("DeleteCascade - 事务级联删除预览会话及其文件 [%d]", id.Int64()))

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除会话下全部文件，避免外键关联孤儿
		if err := tx.Where("session_id = ?", id).Delete(&entity.PreviewFile{}).Error; err != nil {
			return err
		}

		// 再删除会话本体；RowsAffected==0 表示会话不存在，回滚事务
		result := tx.Where("id = ?", id).Delete(&entity.PreviewSession{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return xError.NewError(ctx, xError.NotFound, "预览会话不存在", false, nil)
		}
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "删除预览会话失败", false, err)
	}
	return nil
}

// TouchUpdatedAt 触摸预览会话更新时间（文件变更后同步会话 updated_at）
//
// 文件上传/覆写/删除后调用，使会话 updated_at 与内容变更保持一致，
// 供前端预览页以会话时间为变更依据刷新缓存。
//
// 参数:
//   - ctx:       上下文对象
//   - sessionID: 预览会话雪花 ID
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *PreviewSessionRepo) TouchUpdatedAt(ctx context.Context, sessionID xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("TouchUpdatedAt - 更新预览会话更新时间 [%d]", sessionID.Int64()))

	if err := r.db.WithContext(ctx).Model(&entity.PreviewSession{}).
		Where("id = ?", sessionID).
		Update("updated_at", time.Now()).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新预览会话时间失败", false, err)
	}
	return nil
}
