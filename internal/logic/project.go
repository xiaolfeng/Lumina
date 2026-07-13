package logic

import (
	"context"
	"fmt"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	apiProject "github.com/xiaolfeng/Lumina/api/project"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

// projectRepo 项目模块依赖的仓储集合
type projectRepo struct {
	project *repository.ProjectRepo
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
			project: repository.NewProjectRepo(db, rdb),
		},
	}
}

// Create 创建项目，校验名称唯一性后构建实体并持久化
func (l *ProjectLogic) Create(ctx context.Context, req *apiProject.CreateProjectRequest) (*apiProject.ProjectResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("Create - 创建项目 [%s]", req.Name))

	// 校验项目名称唯一性
	existing, xErr := l.repo.project.GetByName(ctx, req.Name)
	if xErr != nil {
		// 非 NotFound 错误（数据库异常），直接透传
		if xErr.GetErrorCode() != xError.NotFound {
			return nil, xErr
		}
		// NotFound → 名称可用，继续创建
	} else if existing != nil {
		// 查询成功且记录存在 → 名称重复
		return nil, xError.NewError(ctx, xError.BusinessError, "项目名称已存在", false, nil)
	}

	// 生成雪花 ID 并构建实体
	id := xSnowflake.GenerateID(bConst.GeneProject)
	projectEntity := &entity.Project{
		BaseEntity:  xModels.BaseEntity{ID: id},
		Name:        req.Name,
		AliasName:   req.AliasName,
		MatchPath:   req.MatchPath,
		Description: req.Description,
	}

	// 持久化
	if xErr := l.repo.project.Create(ctx, projectEntity); xErr != nil {
		return nil, xErr
	}

	return l.toResponse(projectEntity), nil
}

// GetByID 根据 ID 获取项目详情
func (l *ProjectLogic) GetByID(ctx context.Context, id string) (*apiProject.ProjectResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetByID - 获取项目 [%s]", id))

	// 解析雪花 ID
	parsedID, err := xSnowflake.ParseSnowflakeID(id)
	if err != nil {
		return nil, xError.NewError(ctx, xError.BusinessError, "无效的项目ID", false, nil)
	}

	// 查询项目
	project, xErr := l.repo.project.GetByID(ctx, parsedID)
	if xErr != nil {
		return nil, xErr
	}

	return l.toResponse(project), nil
}

// List 分页获取项目列表
func (l *ProjectLogic) List(ctx context.Context, page, size int) (*apiProject.ProjectListResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("List - 获取项目列表 [page=%d, size=%d]", page, size))

	// 分页参数规范化
	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	page = int(pageReq.Page)
	size = int(pageReq.Size)

	// 查询列表
	projects, total, xErr := l.repo.project.List(ctx, page, size)
	if xErr != nil {
		return nil, xErr
	}

	// 映射响应
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

	// 解析雪花 ID
	parsedID, err := xSnowflake.ParseSnowflakeID(id)
	if err != nil {
		return nil, xError.NewError(ctx, xError.BusinessError, "无效的项目ID", false, nil)
	}

	// 查询现有项目
	existing, xErr := l.repo.project.GetByID(ctx, parsedID)
	if xErr != nil {
		return nil, xErr
	}

	// 如果名称变更，校验新名称唯一性
	if req.Name != existing.Name {
		conflict, xErr := l.repo.project.GetByName(ctx, req.Name)
		if xErr != nil {
			if xErr.GetErrorCode() != xError.NotFound {
				return nil, xErr
			}
			// NotFound → 新名称可用
		} else if conflict != nil {
			return nil, xError.NewError(ctx, xError.BusinessError, "项目名称已存在", false, nil)
		}
	}

	// 更新字段
	existing.Name = req.Name
	existing.AliasName = req.AliasName
	existing.MatchPath = req.MatchPath
	existing.Description = req.Description

	// 持久化
	if xErr := l.repo.project.Update(ctx, existing); xErr != nil {
		return nil, xErr
	}

	return l.toResponse(existing), nil
}

// Delete 删除项目
func (l *ProjectLogic) Delete(ctx context.Context, id string) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("Delete - 删除项目 [%s]", id))

	// 解析雪花 ID
	parsedID, err := xSnowflake.ParseSnowflakeID(id)
	if err != nil {
		return xError.NewError(ctx, xError.BusinessError, "无效的项目ID", false, nil)
	}

	// 执行删除
	return l.repo.project.Delete(ctx, parsedID)
}

// ResolveByAlias 根据别名查询项目
func (l *ProjectLogic) ResolveByAlias(ctx context.Context, alias string) (*apiProject.ProjectResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("ResolveByAlias - 根据别名查询项目 [%s]", alias))

	// 查询项目
	project, xErr := l.repo.project.FindByAliasName(ctx, alias)
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
// 通过 repo.FindByMatchPath 进行 JSON 数组前缀匹配。
// 例如：match_path=["/home/user/Lumina"] 可以匹配 "/home/user/Lumina/src/main.go"
func (l *ProjectLogic) GetByMatchPath(ctx context.Context, path string) (*apiProject.ProjectResponse, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetByMatchPath - 根据路径匹配项目 [%s]", path))

	project, xErr := l.repo.project.FindByMatchPath(ctx, path)
	if xErr != nil {
		return nil, xErr
	}

	return l.toResponse(project), nil
}

// toResponse 将实体映射为响应 DTO
func (l *ProjectLogic) toResponse(project *entity.Project) *apiProject.ProjectResponse {
	return &apiProject.ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		AliasName:   project.AliasName,
		MatchPath:   project.MatchPath,
		Description: project.Description,
		CreatedAt:   project.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   project.UpdatedAt.Format(time.RFC3339),
	}
}
