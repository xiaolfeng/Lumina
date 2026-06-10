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
// @Summary      创建项目
// @Tags         project
// @Accept       json
// @Produce      json
// @Param        request  body      apiProject.CreateProjectRequest  true  "创建项目请求"
// @Success      200      {object}  apiCommon.BaseResponse           "创建成功"
// @Failure      400      {object}  apiCommon.BaseResponse           "参数错误"
// @Failure      401      {object}  apiCommon.BaseResponse           "未认证"
// @Failure      409      {object}  apiCommon.BaseResponse           "项目名称已存在"
// @Security     ApiKeyAuth
// @Router       /api/v1/project [post]
func (h *ProjectHandler) CreateProject(ctx *gin.Context) {
	h.log.Info(ctx, "CreateProject - 创建项目")

	var req apiProject.CreateProjectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
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
// @Summary      获取项目列表
// @Tags         project
// @Accept       json
// @Produce      json
// @Param        page     query     int                    false  "页码"
// @Param        size     query     int                    false  "每页数量"
// @Success      200      {object}  apiCommon.BaseResponse "获取成功"
// @Failure      401      {object}  apiCommon.BaseResponse "未认证"
// @Security     ApiKeyAuth
// @Router       /api/v1/project [get]
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
// @Summary      获取项目详情
// @Tags         project
// @Accept       json
// @Produce      json
// @Param        id       path      string                 true  "项目ID"
// @Success      200      {object}  apiCommon.BaseResponse "获取成功"
// @Failure      400      {object}  apiCommon.BaseResponse "参数错误"
// @Failure      401      {object}  apiCommon.BaseResponse "未认证"
// @Failure      404      {object}  apiCommon.BaseResponse "项目不存在"
// @Security     ApiKeyAuth
// @Router       /api/v1/project/{id} [get]
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
// @Summary      更新项目
// @Tags         project
// @Accept       json
// @Produce      json
// @Param        id       path      string                          true  "项目ID"
// @Param        request  body      apiProject.UpdateProjectRequest true  "更新项目请求"
// @Success      200      {object}  apiCommon.BaseResponse         "更新成功"
// @Failure      400      {object}  apiCommon.BaseResponse         "参数错误"
// @Failure      401      {object}  apiCommon.BaseResponse         "未认证"
// @Failure      404      {object}  apiCommon.BaseResponse         "项目不存在"
// @Failure      409      {object}  apiCommon.BaseResponse         "项目名称已存在"
// @Security     ApiKeyAuth
// @Router       /api/v1/project/{id} [put]
func (h *ProjectHandler) UpdateProject(ctx *gin.Context) {
	h.log.Info(ctx, "UpdateProject - 更新项目")

	id := ctx.Param("id")

	var req apiProject.UpdateProjectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
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
// @Summary      删除项目
// @Tags         project
// @Accept       json
// @Produce      json
// @Param        id       path      string                 true  "项目ID"
// @Success      200      {object}  apiCommon.BaseResponse "删除成功"
// @Failure      400      {object}  apiCommon.BaseResponse "参数错误"
// @Failure      401      {object}  apiCommon.BaseResponse "未认证"
// @Failure      404      {object}  apiCommon.BaseResponse "项目不存在"
// @Security     ApiKeyAuth
// @Router       /api/v1/project/{id} [delete]
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
