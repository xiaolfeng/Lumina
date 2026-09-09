package handler

import (
	"strconv"

	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiWorkspace "github.com/xiaolfeng/Lumina/api/workspace"
)

var _ = apiCommon.BaseResponse{}

// CreateWorkspace 创建工作空间
//
// @Summary     [管理] 创建工作空间
// @Description 创建非默认工作空间，用于拆分生活/工作项目分组
// @Tags        工作空间接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                                  true  "Bearer Access Token"
// @Param       request        body      apiWorkspace.CreateWorkspaceRequest     true  "创建空间请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiWorkspace.WorkspaceResponse}  "创建成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/workspace [POST]
func (h *WorkspaceHandler) CreateWorkspace(ctx *gin.Context) {
	h.log.Info(ctx, "CreateWorkspace - 创建工作空间")

	var req apiWorkspace.CreateWorkspaceRequest
	if !BindJSON(ctx, &req) {
		return
	}

	resp, xErr := h.service.workspaceLogic.Create(ctx.Request.Context(), &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "创建成功", resp)
}

// ListWorkspaces 获取工作空间列表
//
// @Summary     [管理] 获取工作空间列表
// @Description 按 page/size 分页查询工作空间，默认空间排在最前
// @Tags        工作空间接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true   "Bearer Access Token"
// @Param       page           query     int      false  "页码"
// @Param       size           query     int      false  "每页数量"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiWorkspace.WorkspaceListResponse}  "获取成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/workspace [GET]
func (h *WorkspaceHandler) ListWorkspaces(ctx *gin.Context) {
	h.log.Info(ctx, "ListWorkspaces - 获取工作空间列表")

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "50"))

	resp, xErr := h.service.workspaceLogic.List(ctx.Request.Context(), page, size)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "获取成功", resp)
}

// GetWorkspace 获取工作空间详情
//
// @Summary     [管理] 获取工作空间详情
// @Description 根据空间 ID 查询单个工作空间
// @Tags        工作空间接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "空间ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiWorkspace.WorkspaceResponse}  "获取成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "空间不存在"
// @Router      /api/v1/workspace/{id} [GET]
func (h *WorkspaceHandler) GetWorkspace(ctx *gin.Context) {
	h.log.Info(ctx, "GetWorkspace - 获取工作空间详情")

	resp, xErr := h.service.workspaceLogic.GetByID(ctx.Request.Context(), ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "获取成功", resp)
}

// UpdateWorkspace 更新工作空间
//
// @Summary     [管理] 更新工作空间
// @Description 更新名称、描述、图标；非默认空间可改标识
// @Tags        工作空间接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                                  true  "Bearer Access Token"
// @Param       id             path      string                                  true  "空间ID"
// @Param       request        body      apiWorkspace.UpdateWorkspaceRequest     true  "更新空间请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiWorkspace.WorkspaceResponse}  "更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "空间不存在"
// @Router      /api/v1/workspace/{id} [PUT]
func (h *WorkspaceHandler) UpdateWorkspace(ctx *gin.Context) {
	h.log.Info(ctx, "UpdateWorkspace - 更新工作空间")

	var req apiWorkspace.UpdateWorkspaceRequest
	if !BindJSON(ctx, &req) {
		return
	}

	resp, xErr := h.service.workspaceLogic.Update(ctx.Request.Context(), ctx.Param("id"), &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "更新成功", resp)
}

// DeleteWorkspace 删除工作空间
//
// @Summary     [管理] 删除工作空间
// @Description 删除非默认空间，其中的项目将迁到默认空间
// @Tags        工作空间接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "空间ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiWorkspace.DeleteWorkspaceResponse}  "删除成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "空间不存在"
// @Router      /api/v1/workspace/{id} [DELETE]
func (h *WorkspaceHandler) DeleteWorkspace(ctx *gin.Context) {
	h.log.Info(ctx, "DeleteWorkspace - 删除工作空间")

	resp, xErr := h.service.workspaceLogic.Delete(ctx.Request.Context(), ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "删除成功", resp)
}
