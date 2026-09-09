package logic

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"
	apiWorkspace "github.com/xiaolfeng/Lumina/api/workspace"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

var workspaceSlugRegexp = regexp.MustCompile(bConst.WorkspaceSlugPattern)

type workspaceRepo struct {
	workspace *repository.WorkspaceRepo
	project   *repository.ProjectRepo
}

// WorkspaceLogic 工作空间业务逻辑层，负责空间 CRUD、校验与删除搬家
type WorkspaceLogic struct {
	logic
	repo workspaceRepo
}

// NewWorkspaceLogic 创建工作空间业务逻辑层实例
func NewWorkspaceLogic(ctx context.Context) *WorkspaceLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	return &WorkspaceLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "WorkspaceLogic"),
		},
		repo: workspaceRepo{
			workspace: repository.NewWorkspaceRepo(db, rdb),
			project:   repository.NewProjectRepo(db, rdb),
		},
	}
}

// Create 创建工作空间；忽略客户端传入的 is_default
func (l *WorkspaceLogic) Create(ctx context.Context, req *apiWorkspace.CreateWorkspaceRequest) (*apiWorkspace.WorkspaceResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("Create - 创建工作空间 [%s]", req.Slug))

	if xErr := l.validateName(ctx, req.Name); xErr != nil {
		return nil, xErr
	}
	if xErr := l.validateSlug(ctx, req.Slug); xErr != nil {
		return nil, xErr
	}
	if xErr := l.validateIcon(ctx, req.Icon); xErr != nil {
		return nil, xErr
	}

	existing, xErr := l.repo.workspace.GetBySlug(ctx, req.Slug)
	if xErr != nil {
		if xErr.GetErrorCode() != xError.NotFound {
			return nil, xErr
		}
	} else if existing != nil {
		return nil, xError.NewError(ctx, xError.BusinessError, "空间标识已存在", false, nil)
	}

	id := xSnowflake.GenerateID(bConst.GeneWorkspace)
	workspace := &entity.Workspace{
		BaseEntity:  xModels.BaseEntity{ID: id},
		Name:        strings.TrimSpace(req.Name),
		Slug:        req.Slug,
		Description: req.Description,
		Icon:        req.Icon,
		IsDefault:   false,
	}
	if xErr := l.repo.workspace.Create(ctx, workspace); xErr != nil {
		return nil, xErr
	}
	return l.toResponse(workspace), nil
}

// GetByID 根据 ID 获取空间详情
func (l *WorkspaceLogic) GetByID(ctx context.Context, id string) (*apiWorkspace.WorkspaceResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetByID - 获取空间 [%s]", id))

	parsedID, err := xSnowflake.ParseSnowflakeID(id)
	if err != nil {
		return nil, xError.NewError(ctx, xError.BusinessError, "无效的空间ID", false, nil)
	}
	workspace, xErr := l.repo.workspace.GetByID(ctx, parsedID)
	if xErr != nil {
		return nil, xErr
	}
	return l.toResponse(workspace), nil
}

// GetBySlug 根据 slug 获取空间详情
func (l *WorkspaceLogic) GetBySlug(ctx context.Context, slug string) (*apiWorkspace.WorkspaceResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetBySlug - 获取空间 [%s]", slug))

	workspace, xErr := l.repo.workspace.GetBySlug(ctx, slug)
	if xErr != nil {
		return nil, xErr
	}
	return l.toResponse(workspace), nil
}

// List 分页获取空间列表
func (l *WorkspaceLogic) List(ctx context.Context, page, size int) (*apiWorkspace.WorkspaceListResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("List - 获取空间列表 [page=%d, size=%d]", page, size))

	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.NormalizeWithConfig(xModels.PageConfig{
		DefaultPage: 1,
		DefaultSize: 50,
		MaxSize:     100,
		DefaultSort: xModels.SortAsc,
	})
	workspaces, total, xErr := l.repo.workspace.List(ctx, int(pageReq.Page), int(pageReq.Size))
	if xErr != nil {
		return nil, xErr
	}

	items := make([]apiWorkspace.WorkspaceResponse, 0, len(workspaces))
	for _, workspace := range workspaces {
		items = append(items, *l.toResponse(workspace))
	}
	return &apiWorkspace.WorkspaceListResponse{Items: items, Total: total}, nil
}

// Update 更新空间；默认空间禁止改 slug，忽略 is_default
func (l *WorkspaceLogic) Update(ctx context.Context, id string, req *apiWorkspace.UpdateWorkspaceRequest) (*apiWorkspace.WorkspaceResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("Update - 更新空间 [%s]", id))

	parsedID, err := xSnowflake.ParseSnowflakeID(id)
	if err != nil {
		return nil, xError.NewError(ctx, xError.BusinessError, "无效的空间ID", false, nil)
	}
	if xErr := l.validateName(ctx, req.Name); xErr != nil {
		return nil, xErr
	}
	if xErr := l.validateSlug(ctx, req.Slug); xErr != nil {
		return nil, xErr
	}
	if xErr := l.validateIcon(ctx, req.Icon); xErr != nil {
		return nil, xErr
	}

	existing, xErr := l.repo.workspace.GetByID(ctx, parsedID)
	if xErr != nil {
		return nil, xErr
	}
	if (existing.IsDefault || existing.Slug == bConst.DefaultWorkspaceSlug) && req.Slug != existing.Slug {
		return nil, xError.NewError(ctx, xError.BusinessError, "默认空间标识不可修改", false, nil)
	}

	if req.Slug != existing.Slug {
		conflict, conflictErr := l.repo.workspace.GetBySlug(ctx, req.Slug)
		if conflictErr != nil {
			if conflictErr.GetErrorCode() != xError.NotFound {
				return nil, conflictErr
			}
		} else if conflict != nil && conflict.ID != existing.ID {
			return nil, xError.NewError(ctx, xError.BusinessError, "空间标识已存在", false, nil)
		}
	}

	oldSlug := existing.Slug
	existing.Name = strings.TrimSpace(req.Name)
	existing.Slug = req.Slug
	existing.Description = req.Description
	existing.Icon = req.Icon
	if xErr := l.repo.workspace.Update(ctx, existing, oldSlug); xErr != nil {
		return nil, xErr
	}
	return l.toResponse(existing), nil
}

// Delete 删除非默认空间，项目迁到默认空间后再刷缓存
func (l *WorkspaceLogic) Delete(ctx context.Context, id string) (*apiWorkspace.DeleteWorkspaceResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("Delete - 删除空间 [%s]", id))

	parsedID, err := xSnowflake.ParseSnowflakeID(id)
	if err != nil {
		return nil, xError.NewError(ctx, xError.BusinessError, "无效的空间ID", false, nil)
	}

	snapshot, defaultID, xErr := l.repo.workspace.ReassignAndDelete(ctx, parsedID)
	if xErr != nil {
		return nil, xErr
	}
	if cacheErr := l.repo.project.ReplaceWorkspaceCache(ctx, snapshot, defaultID); cacheErr != nil {
		l.log.Warn(ctx, cacheErr.Error())
	}
	l.log.Info(ctx, fmt.Sprintf("Delete - 已移动项目数 [%d]", len(snapshot)))
	return &apiWorkspace.DeleteWorkspaceResponse{MovedProjectCount: int64(len(snapshot))}, nil
}

func (l *WorkspaceLogic) validateName(ctx context.Context, name string) *xError.Error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 128 {
		return xError.NewError(ctx, xError.ParameterError, "空间名称不合法", false, nil)
	}
	return nil
}

func (l *WorkspaceLogic) validateSlug(ctx context.Context, slug string) *xError.Error {
	if len(slug) == 0 || len(slug) > 63 || !workspaceSlugRegexp.MatchString(slug) {
		return xError.NewError(ctx, xError.ParameterError, "空间标识不合法", false, nil)
	}
	return nil
}

func (l *WorkspaceLogic) validateIcon(ctx context.Context, icon string) *xError.Error {
	if icon == "" {
		return nil
	}
	if utf8.RuneCountInString(icon) > 64 || strings.Contains(icon, "://") || strings.Contains(icon, "/") {
		return xError.NewError(ctx, xError.ParameterError, "空间图标不合法", false, nil)
	}
	return nil
}

func (l *WorkspaceLogic) toResponse(workspace *entity.Workspace) *apiWorkspace.WorkspaceResponse {
	return &apiWorkspace.WorkspaceResponse{
		ID:          workspace.ID,
		Name:        workspace.Name,
		Slug:        workspace.Slug,
		Description: workspace.Description,
		Icon:        workspace.Icon,
		IsDefault:   workspace.IsDefault,
		CreatedAt:   workspace.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   workspace.UpdatedAt.Format(time.RFC3339),
	}
}
