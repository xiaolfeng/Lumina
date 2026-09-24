package repository

import (
	"context"
	"errors"
	"fmt"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/gorm"
)

// PageRepo 页面数据访问层，提供 CRUD、项目内 slug 解析与生效指针更新。
//
// 页面快照内容依赖 PageFile 子表，不引入 Redis 缓存，避免指针切换与文件一致性成本。
type PageRepo struct {
	db  *gorm.DB
	log *xLog.LogNamedLogger
}

// NewPageRepo 创建 PageRepo 实例
func NewPageRepo(db *gorm.DB) *PageRepo {
	return &PageRepo{
		db:  db,
		log: xLog.WithName(xLog.NamedREPO, "PageRepo"),
	}
}

// Create 创建页面
func (r *PageRepo) Create(ctx context.Context, page *entity.Page) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Create - 创建页面 [%d/%s]", page.ProjectID.Int64(), page.Slug))

	if err := r.db.WithContext(ctx).Create(page).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "创建页面失败", false, err)
	}
	return nil
}

// GetByID 根据 ID 获取页面
func (r *PageRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.Page, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据 ID 获取页面 [%d]", id.Int64()))

	var page entity.Page
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&page).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "页面不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询页面失败", false, err)
	}
	return &page, nil
}

// GetByProjectAndSlug 根据项目 ID 与 slug 获取页面
func (r *PageRepo) GetByProjectAndSlug(ctx context.Context, projectID xSnowflake.SnowflakeID, slug string) (*entity.Page, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByProjectAndSlug - 根据项目与标识获取页面 [%d/%s]", projectID.Int64(), slug))

	var page entity.Page
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND slug = ?", projectID, slug).
		First(&page).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "页面不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询页面失败", false, err)
	}
	return &page, nil
}

// GetByIDs 根据 ID 列表批量获取页面（空切片安全，返回空切片）
func (r *PageRepo) GetByIDs(ctx context.Context, ids []xSnowflake.SnowflakeID) ([]*entity.Page, *xError.Error) {
	if len(ids) == 0 {
		return []*entity.Page{}, nil
	}
	r.log.Info(ctx, fmt.Sprintf("GetByIDs - 批量获取页面 [%d]", len(ids)))

	pages := make([]*entity.Page, 0, len(ids))
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&pages).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "批量查询页面失败", false, err)
	}
	return pages, nil
}

// List 分页获取页面列表（按更新时间降序；projectID / workspaceID 为零值时不过滤；status 为空不过滤）
func (r *PageRepo) List(ctx context.Context, projectID, workspaceID xSnowflake.SnowflakeID, status string, page, size int) ([]*entity.Page, int64, *xError.Error) {
	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	page, size = int(pageReq.Page), int(pageReq.Size)
	r.log.Info(ctx, fmt.Sprintf("List - 分页获取页面列表 [projectID=%d, workspace=%d, status=%s, page=%d, size=%d]", projectID.Int64(), workspaceID.Int64(), status, page, size))

	query := r.db.WithContext(ctx).Model(&entity.Page{})
	if !projectID.IsZero() {
		query = query.Where("project_id = ?", projectID)
	}
	if !workspaceID.IsZero() {
		projects := r.db.WithContext(ctx).Model(&entity.Project{}).Select("id").Where("workspace_id = ?", workspaceID)
		query = query.Where("project_id IN (?)", projects)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计页面数量失败", false, err)
	}

	pages := make([]*entity.Page, 0)
	offset := (page - 1) * size
	if err := query.
		Offset(offset).
		Limit(size).
		Order("updated_at DESC").
		Find(&pages).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询页面列表失败", false, err)
	}

	return pages, total, nil
}

// UpdateLatestVersionID 原子更新当前线上生效版本指针
func (r *PageRepo) UpdateLatestVersionID(ctx context.Context, id, versionID xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("UpdateLatestVersionID - 更新生效版本指针 [page=%d, version=%d]", id.Int64(), versionID.Int64()))

	result := r.db.WithContext(ctx).Model(&entity.Page{}).
		Where("id = ?", id).
		Update("latest_version_id", versionID)
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新生效版本失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "页面不存在", false, nil)
	}
	return nil
}

// UpdateAccessPolicy 更新访问策略与密码哈希（passwordHash 可为空表示清除）
func (r *PageRepo) UpdateAccessPolicy(ctx context.Context, id xSnowflake.SnowflakeID, accessMode, passwordHash string) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("UpdateAccessPolicy - 更新访问策略 [page=%d, mode=%s]", id.Int64(), accessMode))

	result := r.db.WithContext(ctx).Model(&entity.Page{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"access_mode":   accessMode,
			"password_hash": passwordHash,
		})
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新页面访问策略失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "页面不存在", false, nil)
	}
	return nil
}

// UpdateStatus 更新页面生命周期状态
func (r *PageRepo) UpdateStatus(ctx context.Context, id xSnowflake.SnowflakeID, status string) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("UpdateStatus - 更新页面状态 [page=%d, status=%s]", id.Int64(), status))

	result := r.db.WithContext(ctx).Model(&entity.Page{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		r.log.Warn(ctx, result.Error.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新页面状态失败", false, result.Error)
	}
	if result.RowsAffected == 0 {
		return xError.NewError(ctx, xError.NotFound, "页面不存在", false, nil)
	}
	return nil
}

// errPromoteSessionMissing 晋升事务内源预览会话缺失（并发删除竞态），用于与页面缺失的 404 语义区分
var errPromoteSessionMissing = errors.New("source preview session missing")

// PersistPromotion 在同一事务内写入页面（可选）、版本与快照文件，按需切换生效指针，
// 并将源预览会话软删除（consumeSessionID 零值表示不消耗，供非晋升场景复用）。
//
// 「写快照 + 消耗草稿」同事务原子提交：晋升即消耗草稿，快照为深拷贝不可变，
// 软删除保留预览文件与 SourceSessionID 审计（与 ExpireStaleSessions 一致）。
func (r *PageRepo) PersistPromotion(ctx context.Context, page *entity.Page, createPage bool, version *entity.PageVersion, files []*entity.PageFile, setAsActive bool, consumeSessionID xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("PersistPromotion - 写入晋升快照 [page=%d, version=%s, createPage=%v, active=%v, consumeSession=%d]", page.ID.Int64(), version.Version, createPage, setAsActive, consumeSessionID.Int64()))

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if createPage {
			if setAsActive {
				page.LatestVersionID = version.ID
			}
			if err := tx.Create(page).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		if len(files) > 0 {
			if err := tx.Create(&files).Error; err != nil {
				return err
			}
		}
		if setAsActive && !createPage {
			result := tx.Model(&entity.Page{}).Where("id = ?", page.ID).Update("latest_version_id", version.ID)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
			page.LatestVersionID = version.ID
		}
		if !consumeSessionID.IsZero() {
			result := tx.Model(&entity.PreviewSession{}).
				Where("id = ?", consumeSessionID).
				Update("status", bConst.PreviewSessionStatusDeleted)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errPromoteSessionMissing
			}
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return xError.NewError(ctx, xError.NotFound, "页面不存在", false, nil)
		}
		if errors.Is(err, errPromoteSessionMissing) {
			return xError.NewError(ctx, xError.NotFound, "预览会话不存在", false, nil)
		}
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "写入页面快照失败", false, err)
	}
	return nil
}
