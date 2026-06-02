package handler

import (
	"strconv"

	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiApikey "github.com/xiaolfeng/Lumina/api/apikey"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// Create 创建API密钥
//
// @Summary		创建API密钥
// @Tags		apikey
// @Accept		json
// @Produce		json
// @Param		request	body		apiApikey.CreateRequest	true	"创建请求"
// @Success		200		{object}	apiCommon.BaseResponse		"创建成功，返回完整密钥（仅此一次）"
// @Failure		400		{object}	apiCommon.BaseResponse		"参数错误"
// @Failure		401		{object}	apiCommon.BaseResponse		"未授权"
// @Failure		500		{object}	apiCommon.BaseResponse		"服务器内部错误"
// @Security	ApiKeyAuth
// @Router		/api/v1/apikey [post]
func (h *ApikeyHandler) Create(ctx *gin.Context) {
	h.log.Info(ctx, "Create - 创建API密钥")

	var req apiApikey.CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	resp, xErr := h.service.apikeyLogic.Create(ctx, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "API密钥创建成功", resp)
}

// List 分页获取API密钥列表
//
// @Summary		分页获取API密钥列表
// @Tags		apikey
// @Accept		json
// @Produce		json
// @Param		page	query		int	false	"页码"	default(1)
// @Param		size	query		int	false	"每页数量"	default(10)
// @Success		200		{object}	apiCommon.BaseResponse		"查询成功，密钥脱敏展示"
// @Failure		401		{object}	apiCommon.BaseResponse		"未授权"
// @Failure		500		{object}	apiCommon.BaseResponse		"服务器内部错误"
// @Security	ApiKeyAuth
// @Router		/api/v1/apikey [get]
func (h *ApikeyHandler) List(ctx *gin.Context) {
	h.log.Info(ctx, "List - 分页获取API密钥列表")

	// 从查询参数解析分页，提供默认值
	pageStr := ctx.DefaultQuery("page", "1")
	sizeStr := ctx.DefaultQuery("size", "10")
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 10
	}

	resp, xErr := h.service.apikeyLogic.List(ctx, page, size)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// GetByID 根据ID获取API密钥详情
//
// @Summary		获取API密钥详情
// @Tags		apikey
// @Accept		json
// @Produce		json
// @Param		id	path		string	true	"API密钥ID"
// @Success		200	{object}	apiCommon.BaseResponse		"查询成功，密钥脱敏展示"
// @Failure		400	{object}	apiCommon.BaseResponse		"参数错误"
// @Failure		401	{object}	apiCommon.BaseResponse		"未授权"
// @Failure		404	{object}	apiCommon.BaseResponse		"密钥不存在"
// @Security	ApiKeyAuth
// @Router		/api/v1/apikey/{id} [get]
func (h *ApikeyHandler) GetByID(ctx *gin.Context) {
	h.log.Info(ctx, "GetByID - 获取API密钥详情")

	id := ctx.Param("id")
	resp, xErr := h.service.apikeyLogic.GetByID(ctx, id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// Update 更新API密钥信息
//
// @Summary		更新API密钥
// @Tags		apikey
// @Accept		json
// @Produce		json
// @Param		id		path		string						true	"API密钥ID"
// @Param		request	body		apiApikey.UpdateRequest		true	"更新请求"
// @Success		200		{object}	apiCommon.BaseResponse		"更新成功"
// @Failure		400		{object}	apiCommon.BaseResponse		"参数错误"
// @Failure		401		{object}	apiCommon.BaseResponse		"未授权"
// @Failure		404		{object}	apiCommon.BaseResponse		"密钥不存在"
// @Failure		500		{object}	apiCommon.BaseResponse		"服务器内部错误"
// @Security	ApiKeyAuth
// @Router		/api/v1/apikey/{id} [put]
func (h *ApikeyHandler) Update(ctx *gin.Context) {
	h.log.Info(ctx, "Update - 更新API密钥")

	id := ctx.Param("id")
	var req apiApikey.UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	xErr := h.service.apikeyLogic.Update(ctx, id, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "API密钥更新成功")
}

// Delete 删除API密钥
//
// @Summary		删除API密钥
// @Tags		apikey
// @Accept		json
// @Produce		json
// @Param		id	path		string	true	"API密钥ID"
// @Success		200	{object}	apiCommon.BaseResponse		"删除成功"
// @Failure		400	{object}	apiCommon.BaseResponse		"参数错误"
// @Failure		401	{object}	apiCommon.BaseResponse		"未授权"
// @Failure		404	{object}	apiCommon.BaseResponse		"密钥不存在"
// @Security	ApiKeyAuth
// @Router		/api/v1/apikey/{id} [delete]
func (h *ApikeyHandler) Delete(ctx *gin.Context) {
	h.log.Info(ctx, "Delete - 删除API密钥")

	id := ctx.Param("id")
	xErr := h.service.apikeyLogic.Delete(ctx, id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "API密钥删除成功")
}

// Reset 重置API密钥
//
// @Summary		重置API密钥
// @Tags		apikey
// @Accept		json
// @Produce		json
// @Param		id	path		string	true	"API密钥ID"
// @Success		200	{object}	apiCommon.BaseResponse		"重置成功，返回新完整密钥（仅此一次）"
// @Failure		400	{object}	apiCommon.BaseResponse		"参数错误"
// @Failure		401	{object}	apiCommon.BaseResponse		"未授权"
// @Failure		404	{object}	apiCommon.BaseResponse		"密钥不存在"
// @Failure		500	{object}	apiCommon.BaseResponse		"服务器内部错误"
// @Security	ApiKeyAuth
// @Router		/api/v1/apikey/{id}/reset [post]
func (h *ApikeyHandler) Reset(ctx *gin.Context) {
	h.log.Info(ctx, "Reset - 重置API密钥")

	id := ctx.Param("id")
	resp, xErr := h.service.apikeyLogic.Reset(ctx, id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "API密钥重置成功", resp)
}
