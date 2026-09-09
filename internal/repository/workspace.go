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
	"github.com/redis/go-redis/v9"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository/cache"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const cacheTTLWorkspace = 30 * time.Minute

// WorkspaceRepo 工作空间数据访问层，提供 CRUD、默认空间保证与删除搬家
type WorkspaceRepo struct {
	db    *gorm.DB
	cache *cache.WorkspaceCache
	log   *xLog.LogNamedLogger
}

// NewWorkspaceRepo 创建 WorkspaceRepo 实例
func NewWorkspaceRepo(db *gorm.DB, rdb *redis.Client) *WorkspaceRepo {
	return &WorkspaceRepo{
		db: db,
		cache: &cache.WorkspaceCache{
			Base: &cache.Base{RDB: rdb, TTL: cacheTTLWorkspace},
		},
		log: xLog.WithName(xLog.NamedREPO, "WorkspaceRepo"),
	}
}

// Create 创建工作空间，成功后写入缓存
func (r *WorkspaceRepo) Create(ctx context.Context, workspace *entity.Workspace) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Create - 创建工作空间 [%s]", workspace.Slug))

	if err := r.db.WithContext(ctx).Create(workspace).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "创建空间失败", false, err)
	}
	if xErr := r.cache.SetWorkspace(ctx, workspace); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	return nil
}

// GetByID 根据 ID 获取空间，优先读取缓存
func (r *WorkspaceRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.Workspace, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据ID获取空间 [%d]", id.Int64()))

	if workspace, ok, _ := r.cache.GetByID(ctx, id.Int64()); ok {
		r.log.Info(ctx, fmt.Sprintf("GetByID - 缓存命中 [%d]", id.Int64()))
		return workspace, nil
	}

	var workspace entity.Workspace
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&workspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "空间不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询空间失败", false, err)
	}

	if xErr := r.cache.SetWorkspace(ctx, &workspace); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	return &workspace, nil
}

// GetBySlug 根据 slug 获取空间
func (r *WorkspaceRepo) GetBySlug(ctx context.Context, slug string) (*entity.Workspace, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetBySlug - 根据标识获取空间 [%s]", slug))

	if idStr, ok, _ := r.cache.GetIDBySlug(ctx, slug); ok {
		parsedID, err := xSnowflake.ParseSnowflakeID(idStr)
		if err == nil {
			return r.GetByID(ctx, parsedID)
		}
	}

	var workspace entity.Workspace
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&workspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "空间不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询空间失败", false, err)
	}

	if xErr := r.cache.SetWorkspace(ctx, &workspace); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	return &workspace, nil
}

// GetDefault 获取默认空间（is_default=true，按 created_at 升序取最早一行）
func (r *WorkspaceRepo) GetDefault(ctx context.Context) (*entity.Workspace, *xError.Error) {
	r.log.Info(ctx, "GetDefault - 获取默认空间")

	var workspace entity.Workspace
	if err := r.db.WithContext(ctx).
		Where("is_default = ?", true).
		Order("created_at ASC").
		First(&workspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xError.NewError(ctx, xError.NotFound, "默认空间不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询默认空间失败", false, err)
	}
	return &workspace, nil
}

// List 分页获取空间列表（默认空间在前，其余按创建时间升序）
func (r *WorkspaceRepo) List(ctx context.Context, page, size int) ([]*entity.Workspace, int64, *xError.Error) {
	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	page, size = int(pageReq.Page), int(pageReq.Size)
	r.log.Info(ctx, fmt.Sprintf("List - 分页获取空间列表 [page=%d, size=%d]", page, size))

	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.Workspace{}).Count(&total).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "统计空间数量失败", false, err)
	}

	var workspaces []*entity.Workspace
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(size).
		Order("is_default DESC, created_at ASC").
		Find(&workspaces).Error; err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "查询空间列表失败", false, err)
	}
	return workspaces, total, nil
}

// Update 更新空间，成功后刷新缓存；slug 变更时删除旧 slug 键
func (r *WorkspaceRepo) Update(ctx context.Context, workspace *entity.Workspace, oldSlug string) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Update - 更新空间 [%s]", workspace.Slug))

	if err := r.db.WithContext(ctx).Save(workspace).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "更新空间失败", false, err)
	}

	if oldSlug != "" && oldSlug != workspace.Slug {
		r.cache.DeleteWorkspace(ctx, &entity.Workspace{
			BaseEntity: xModels.BaseEntity{ID: workspace.ID},
			Slug:       oldSlug,
		}, "")
	}
	if xErr := r.cache.SetWorkspace(ctx, workspace); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	return nil
}

// Delete 删除空间并清除缓存（调用方须已完成项目搬家）
func (r *WorkspaceRepo) Delete(ctx context.Context, workspace *entity.Workspace) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Delete - 删除空间 [%d]", workspace.ID.Int64()))

	if err := r.db.WithContext(ctx).Delete(&entity.Workspace{}, workspace.ID).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "删除空间失败", false, err)
	}
	r.cache.DeleteWorkspace(ctx, workspace, "")
	return nil
}

// InvalidateCache 清除空间 ID / slug 缓存
func (r *WorkspaceRepo) InvalidateCache(ctx context.Context, workspace *entity.Workspace) {
	r.cache.DeleteWorkspace(ctx, workspace, "")
}

// EnsureDefault 保证恰好存在一条默认空间；已存在时不覆盖 Name
func (r *WorkspaceRepo) EnsureDefault(ctx context.Context) (*entity.Workspace, *xError.Error) {
	r.log.Info(ctx, "EnsureDefault - 保证默认空间存在")

	workspace, xErr := r.GetDefault(ctx)
	if xErr == nil {
		return workspace, nil
	}
	if xErr.GetErrorCode() != xError.NotFound {
		return nil, xErr
	}

	bySlug, slugErr := r.GetBySlug(ctx, bConst.DefaultWorkspaceSlug)
	if slugErr == nil {
		bySlug.IsDefault = true
		if err := r.db.WithContext(ctx).Model(bySlug).Update("is_default", true).Error; err != nil {
			return nil, xError.NewError(ctx, xError.DatabaseError, "标记默认空间失败", false, err)
		}
		if cacheErr := r.cache.SetWorkspace(ctx, bySlug); cacheErr != nil {
			r.log.Warn(ctx, cacheErr.Error())
		}
		return bySlug, nil
	}
	if slugErr.GetErrorCode() != xError.NotFound {
		return nil, slugErr
	}

	id := xSnowflake.GenerateID(bConst.GeneWorkspace)
	created := &entity.Workspace{
		BaseEntity:  xModels.BaseEntity{ID: id},
		Name:        bConst.DefaultWorkspaceName,
		Slug:        bConst.DefaultWorkspaceSlug,
		Description: "",
		Icon:        "",
		IsDefault:   true,
	}
	if createErr := r.Create(ctx, created); createErr != nil {
		return nil, createErr
	}
	return created, nil
}

// BackfillProjectWorkspace 将 workspace_id 为空或 0 的项目回填到指定空间，返回回填前快照
func (r *WorkspaceRepo) BackfillProjectWorkspace(ctx context.Context, defaultID xSnowflake.SnowflakeID) ([]*entity.Project, int64, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("BackfillProjectWorkspace - 回填项目空间 [%d]", defaultID.Int64()))

	var snapshot []*entity.Project
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workspace_id IS NULL OR workspace_id = 0").Find(&snapshot).Error; err != nil {
			return err
		}
		if len(snapshot) == 0 {
			return nil
		}
		return tx.Model(&entity.Project{}).
			Where("workspace_id IS NULL OR workspace_id = 0").
			Update("workspace_id", defaultID).Error
	})
	if err != nil {
		return nil, 0, xError.NewError(ctx, xError.DatabaseError, "回填项目所属空间失败", false, err)
	}
	return snapshot, int64(len(snapshot)), nil
}

// DeduplicateDefaultFlags 多行 is_default=true 时保留 created_at 最早的一行
func (r *WorkspaceRepo) DeduplicateDefaultFlags(ctx context.Context) (int64, *xError.Error) {
	var extras []entity.Workspace
	if err := r.db.WithContext(ctx).
		Where("is_default = ?", true).
		Order("created_at ASC").
		Find(&extras).Error; err != nil {
		return 0, xError.NewError(ctx, xError.DatabaseError, "查询默认空间失败", false, err)
	}
	if len(extras) <= 1 {
		return 0, nil
	}

	ids := make([]xSnowflake.SnowflakeID, 0, len(extras)-1)
	for i := 1; i < len(extras); i++ {
		ids = append(ids, extras[i].ID)
	}
	result := r.db.WithContext(ctx).
		Model(&entity.Workspace{}).
		Where("id IN ?", ids).
		Update("is_default", false)
	if result.Error != nil {
		return 0, xError.NewError(ctx, xError.DatabaseError, "修复重复默认空间失败", false, result.Error)
	}
	return result.RowsAffected, nil
}

// EnsureOneDefaultIndex 创建部分唯一索引，保证最多一行 is_default=true
func (r *WorkspaceRepo) EnsureOneDefaultIndex(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS workspaces_one_default ON workspaces (is_default) WHERE is_default = true`,
	).Error
}

// SetProjectWorkspaceNotNull 将 projects.workspace_id 设为 NOT NULL（已是 NOT NULL 时为 no-op）
func (r *WorkspaceRepo) SetProjectWorkspaceNotNull(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec(
		`ALTER TABLE projects ALTER COLUMN workspace_id SET NOT NULL`,
	).Error
}

// ReassignAndDelete 在同一事务内锁空间行、搬家项目并删除空间，返回搬家前快照与默认空间 ID
func (r *WorkspaceRepo) ReassignAndDelete(ctx context.Context, fromID xSnowflake.SnowflakeID) ([]*entity.Project, xSnowflake.SnowflakeID, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("ReassignAndDelete - 搬家并删除空间 [%d]", fromID.Int64()))

	var (
		snapshot  []*entity.Project
		defaultID xSnowflake.SnowflakeID
		deleted   entity.Workspace
	)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var target entity.Workspace
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", fromID).
			First(&target).Error; err != nil {
			return err
		}
		deleted = target

		if target.IsDefault || target.Slug == bConst.DefaultWorkspaceSlug {
			return errWorkspaceDefaultProtected
		}

		var total int64
		if err := tx.Model(&entity.Workspace{}).Count(&total).Error; err != nil {
			return err
		}
		if total <= 1 {
			return errWorkspaceLastProtected
		}

		var defaultRow entity.Workspace
		if err := tx.Where("is_default = ?", true).
			Order("created_at ASC").
			First(&defaultRow).Error; err != nil {
			return err
		}
		defaultID = defaultRow.ID
		if defaultID == fromID {
			return errWorkspaceDefaultProtected
		}

		if err := tx.Where("workspace_id = ?", fromID).Find(&snapshot).Error; err != nil {
			return err
		}

		if err := tx.Model(&entity.Project{}).
			Where("workspace_id = ?", fromID).
			Update("workspace_id", defaultID).Error; err != nil {
			return err
		}

		if err := tx.Delete(&entity.Workspace{}, fromID).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, 0, xError.NewError(ctx, xError.NotFound, "空间不存在", false, nil)
		case errors.Is(err, errWorkspaceDefaultProtected):
			return nil, 0, xError.NewError(ctx, xError.BusinessError, "默认空间不可删除", false, nil)
		case errors.Is(err, errWorkspaceLastProtected):
			return nil, 0, xError.NewError(ctx, xError.BusinessError, "至少保留一个空间", false, nil)
		default:
			return nil, 0, xError.NewError(ctx, xError.DatabaseError, "删除空间失败", false, err)
		}
	}

	r.cache.DeleteWorkspace(ctx, &deleted, "")
	return snapshot, defaultID, nil
}

var (
	errWorkspaceDefaultProtected = errors.New("default workspace cannot be deleted")
	errWorkspaceLastProtected    = errors.New("last workspace cannot be deleted")
)
