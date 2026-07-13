package handler

import (
	"strconv"

	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiProject "github.com/xiaolfeng/Lumina/api/project"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// CreateProject 创建项目
//
// @Summary     [管理] 创建项目
// @Description 提交项目名称、别名、匹配路径与描述创建项目，名称需唯一
// @Tags        项目接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                              true  "Bearer Access Token"
// @Param       request        body      apiProject.CreateProjectRequest     true  "创建项目请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiProject.ProjectResponse}  "创建成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     409  {object}  apiCommon.BaseResponse  "项目名称已存在"
// @Router      /api/v1/project [POST]
func (h *ProjectHandler) CreateProject(ctx *gin.Context) {
	h.log.Info(ctx, "CreateProject - 创建项目")

	var req apiProject.CreateProjectRequest
	if !BindJSON(ctx, &req) {
		return
	}

	resp, xErr := h.service.projectLogic.Create(ctx, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "创建成功", resp)
}

// ListProjects 获取项目列表
//
// @Summary     [管理] 获取项目列表
// @Description 按 page/size 分页查询项目列表，返回项目信息与总数
// @Tags        项目接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       page           query     int      false  "页码"
// @Param       size           query     int      false  "每页数量"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiProject.ProjectListResponse}  "获取成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/project [GET]
func (h *ProjectHandler) ListProjects(ctx *gin.Context) {
	h.log.Info(ctx, "ListProjects - 获取项目列表")

	pageStr := ctx.DefaultQuery("page", "1")
	sizeStr := ctx.DefaultQuery("size", "20")
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	resp, xErr := h.service.projectLogic.List(ctx, page, size)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", resp)
}

// GetProject 获取项目详情
//
// @Summary     [管理] 获取项目详情
// @Description 根据项目 ID 查询单个项目详情
// @Tags        项目接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "项目ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiProject.ProjectResponse}  "获取成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "项目不存在"
// @Router      /api/v1/project/{id} [GET]
func (h *ProjectHandler) GetProject(ctx *gin.Context) {
	h.log.Info(ctx, "GetProject - 获取项目详情")

	id := ctx.Param("id")

	resp, xErr := h.service.projectLogic.GetByID(ctx, id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", resp)
}

// UpdateProject 更新项目
//
// @Summary     [管理] 更新项目
// @Description 更新指定 ID 的项目信息，名称需保持唯一
// @Tags        项目接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                              true  "Bearer Access Token"
// @Param       id             path      string                              true  "项目ID"
// @Param       request        body      apiProject.UpdateProjectRequest     true  "更新项目请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiProject.ProjectResponse}  "更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "项目不存在"
// @Failure     409  {object}  apiCommon.BaseResponse  "项目名称已存在"
// @Router      /api/v1/project/{id} [PUT]
func (h *ProjectHandler) UpdateProject(ctx *gin.Context) {
	h.log.Info(ctx, "UpdateProject - 更新项目")

	id := ctx.Param("id")

	var req apiProject.UpdateProjectRequest
	if !BindJSON(ctx, &req) {
		return
	}

	resp, xErr := h.service.projectLogic.Update(ctx, id, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "更新成功", resp)
}

// DeleteProject 删除项目
//
// @Summary     [管理] 删除项目
// @Description 根据项目 ID 删除指定项目，删除后不可恢复
// @Tags        项目接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "项目ID"
// @Success     200  {object}  apiCommon.BaseResponse  "删除成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "项目不存在"
// @Router      /api/v1/project/{id} [DELETE]
func (h *ProjectHandler) DeleteProject(ctx *gin.Context) {
	h.log.Info(ctx, "DeleteProject - 删除项目")

	id := ctx.Param("id")

	xErr := h.service.projectLogic.Delete(ctx, id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "删除成功")
}
