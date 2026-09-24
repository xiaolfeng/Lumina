package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"
	apiProject "github.com/xiaolfeng/Lumina/api/project"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

// projectRepo 项目模块依赖的仓储集合
type projectRepo struct {
	project   *repository.ProjectRepo
	workspace *repository.WorkspaceRepo
}

// ProjectLogic 项目业务逻辑层，负责项目 CRUD 编排与校验
type ProjectLogic struct {
	logic
	repo projectRepo
}

// NewProjectLogic 创建项目业务逻辑层实例
func NewProjectLogic(ctx context.Context) *ProjectLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	return &ProjectLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "ProjectLogic"),
		},
		repo: projectRepo{
			project:   repository.NewProjectRepo(db, rdb),
			workspace: repository.NewWorkspaceRepo(db, rdb),
		},
	}
}

// Create 创建项目，校验名称唯一性后构建实体并持久化。
// 未带合法 workspace_id 时写入默认空间，禁止落 0。
func (l *ProjectLogic) Create(ctx context.Context, req *apiProject.CreateProjectRequest) (*apiProject.ProjectResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("Create - 创建项目 [%s]", req.Name))

	workspaceID, xErr := l.resolveCreateWorkspaceID(ctx, req.WorkspaceID)
	if xErr != nil {
		return nil, xErr
	}

	existing, xErr := l.repo.project.GetByName(ctx, req.Name)
	if xErr != nil {
		if xErr.GetErrorCode() != xError.NotFound {
			return nil, xErr
		}
	} else if existing != nil {
		return nil, xError.NewError(ctx, xError.BusinessError, "项目名称已存在", false, nil)
	}

	id := xSnowflake.GenerateID(bConst.GeneProject)
	projectEntity := &entity.Project{
		BaseEntity:  xModels.BaseEntity{ID: id},
		WorkspaceID: workspaceID,
		Name:        req.Name,
		AliasName:   req.AliasName,
		MatchPath:   req.MatchPath,
		Description: req.Description,
	}

	if xErr := l.repo.project.Create(ctx, projectEntity); xErr != nil {
		return nil, xErr
	}

	return l.toResponse(projectEntity), nil
}

func (l *ProjectLogic) resolveCreateWorkspaceID(ctx context.Context, workspaceID xSnowflake.SnowflakeID) (xSnowflake.SnowflakeID, *xError.Error) {
	if workspaceID.IsZero() {
		return 0, xError.NewError(ctx, xError.ParameterError, "缺少所属空间", false, nil)
	}

	workspace, xErr := l.repo.workspace.GetByID(ctx, workspaceID)
	if xErr != nil {
		if xErr.GetErrorCode() == xError.NotFound {
			return 0, xError.NewError(ctx, xError.NotFound, "空间不存在", false, nil)
		}
		return 0, xErr
	}
	if workspace.ID.IsZero() {
		return 0, xError.NewError(ctx, xError.NotFound, "空间不存在", false, nil)
	}
	return workspace.ID, nil
}

// GetByID 根据 ID 获取项目详情
func (l *ProjectLogic) GetByID(ctx context.Context, id string) (*apiProject.ProjectResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetByID - 获取项目 [%s]", id))

	parsedID, err := xSnowflake.ParseSnowflakeID(id)
	if err != nil {
		return nil, xError.NewError(ctx, xError.BusinessError, "无效的项目ID", false, nil)
	}

	project, xErr := l.repo.project.GetByID(ctx, parsedID)
	if xErr != nil {
		return nil, xErr
	}

	return l.toResponse(project), nil
}

// List 分页获取项目列表；workspaceID 为零时不过滤，search 不为空时模糊搜索
func (l *ProjectLogic) List(ctx context.Context, page, size int, workspaceID xSnowflake.SnowflakeID, search string) (*apiProject.ProjectListResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("List - 获取项目列表 [page=%d, size=%d, workspace=%d, search=%s]", page, size, workspaceID.Int64(), search))

	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	page = int(pageReq.Page)
	size = int(pageReq.Size)

	projects, total, xErr := l.repo.project.List(ctx, page, size, workspaceID, search)
	if xErr != nil {
		return nil, xErr
	}

	items := make([]apiProject.ProjectResponse, 0, len(projects))
	for _, p := range projects {
		items = append(items, *l.toResponse(p))
	}

	return &apiProject.ProjectListResponse{
		Items: items,
		Total: total,
	}, nil
}

// Update 更新项目，校验名称唯一性后更新字段并持久化
func (l *ProjectLogic) Update(ctx context.Context, id string, req *apiProject.UpdateProjectRequest) (*apiProject.ProjectResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("Update - 更新项目 [%s]", id))

	parsedID, err := xSnowflake.ParseSnowflakeID(id)
	if err != nil {
		return nil, xError.NewError(ctx, xError.BusinessError, "无效的项目ID", false, nil)
	}

	existing, xErr := l.repo.project.GetByID(ctx, parsedID)
	if xErr != nil {
		return nil, xErr
	}

	if req.Name != existing.Name {
		conflict, xErr := l.repo.project.GetByName(ctx, req.Name)
		if xErr != nil {
			if xErr.GetErrorCode() != xError.NotFound {
				return nil, xErr
			}
		} else if conflict != nil {
			return nil, xError.NewError(ctx, xError.BusinessError, "项目名称已存在", false, nil)
		}
	}

	if req.WorkspaceID != nil {
		workspaceID, xErr := l.ResolveWorkspace(ctx, req.WorkspaceID.String(), "")
		if xErr != nil {
			return nil, xErr
		}
		existing.WorkspaceID = workspaceID
	}

	existing.Name = req.Name
	existing.AliasName = req.AliasName
	existing.MatchPath = req.MatchPath
	existing.Description = req.Description

	if xErr := l.repo.project.Update(ctx, existing); xErr != nil {
		return nil, xErr
	}

	return l.toResponse(existing), nil
}

// Delete 删除项目
func (l *ProjectLogic) Delete(ctx context.Context, id string) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("Delete - 删除项目 [%s]", id))

	parsedID, err := xSnowflake.ParseSnowflakeID(id)
	if err != nil {
		return xError.NewError(ctx, xError.BusinessError, "无效的项目ID", false, nil)
	}

	return l.repo.project.Delete(ctx, parsedID)
}

// ResolveByAlias 根据别名查询项目
func (l *ProjectLogic) ResolveByAlias(ctx context.Context, alias string) (*apiProject.ProjectResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("ResolveByAlias - 根据别名查询项目 [%s]", alias))

	project, xErr := l.repo.project.FindByAliasName(ctx, alias, 0)
	if xErr != nil {
		return nil, xErr
	}

	return l.toResponse(project), nil
}

// GetByName 根据名称获取项目详情
func (l *ProjectLogic) GetByName(ctx context.Context, name string) (*apiProject.ProjectResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetByName - 根据名称获取项目 [%s]", name))

	project, xErr := l.repo.project.GetByName(ctx, name)
	if xErr != nil {
		return nil, xErr
	}

	return l.toResponse(project), nil
}

// GetByMatchPath 根据路径匹配查询项目（用于 MCP 工具 project_get）
//
// workspaceID 为零时不过滤空间。
func (l *ProjectLogic) GetByMatchPath(ctx context.Context, path string, workspaceID xSnowflake.SnowflakeID) (*apiProject.ProjectResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetByMatchPath - 根据路径匹配项目 [%s, workspace=%d]", path, workspaceID.Int64()))

	project, xErr := l.repo.project.FindByMatchPath(ctx, path, workspaceID)
	if xErr != nil {
		return nil, xErr
	}

	return l.toResponse(project), nil
}

// ResolveWorkspace 将 workspace_id 或 slug 解析为空间 ID；id 优先于 slug
func (l *ProjectLogic) ResolveWorkspace(ctx context.Context, workspaceID, slug string) (xSnowflake.SnowflakeID, *xError.Error) {
	workspaceID = strings.TrimSpace(workspaceID)
	slug = strings.TrimSpace(slug)
	if workspaceID == "" && slug == "" {
		return 0, xError.NewError(ctx, xError.ParameterError, "缺少必填参数: workspace_id 或 workspace_slug", false, nil)
	}

	if workspaceID != "" {
		parsedID, err := xSnowflake.ParseSnowflakeID(workspaceID)
		if err != nil {
			return 0, xError.NewError(ctx, xError.BusinessError, "无效的空间ID", false, nil)
		}
		workspace, xErr := l.repo.workspace.GetByID(ctx, parsedID)
		if xErr != nil {
			return 0, xErr
		}
		return workspace.ID, nil
	}

	workspace, xErr := l.repo.workspace.GetBySlug(ctx, slug)
	if xErr != nil {
		return 0, xErr
	}
	return workspace.ID, nil
}

// toResponse 将实体映射为响应 DTO
func (l *ProjectLogic) toResponse(project *entity.Project) *apiProject.ProjectResponse {
	return &apiProject.ProjectResponse{
		ID:          project.ID,
		WorkspaceID: project.WorkspaceID,
		Name:        project.Name,
		AliasName:   project.AliasName,
		MatchPath:   project.MatchPath,
		Description: project.Description,
		CreatedAt:   project.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   project.UpdatedAt.Format(time.RFC3339),
	}
}
