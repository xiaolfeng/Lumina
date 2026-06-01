package handler

import (
	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiApikey "github.com/xiaolfeng/Lumina/api/apikey"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// Create 创建 API Key
//
// @Summary		创建 API Key
// @Tags		apikey
// @Accept		json
// @Produce		json
// @Param		request	body		apiApikey.CreateRequest	true	"创建请求"
// @Success		200		{object}	apiCommon.BaseResponse		"创建成功，返回完整Key"
// @Failure		400		{object}	apiCommon.BaseResponse		"参数错误"
// @Failure		401		{object}	apiCommon.BaseResponse		"未授权"
// @Security	ApiKeyAuth
// @Router		/api/v1/apikey [post]
func (h *ApikeyHandler) Create(ctx *gin.Context) {
	h.log.Info(ctx, "Create - 创建 API Key")

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

	xResult.SuccessHasData(ctx, "创建成功", resp)
}

// List 查询 API Key 列表
//
// @Summary		查询 API Key 列表
// @Tags		apikey
// @Accept		json
// @Produce		json
// @Success		200		{object}	apiCommon.BaseResponse		"查询成功，返回脱敏列表"
// @Failure		401		{object}	apiCommon.BaseResponse		"未授权"
// @Security	ApiKeyAuth
// @Router		/api/v1/apikey [get]
func (h *ApikeyHandler) List(ctx *gin.Context) {
	h.log.Info(ctx, "List - 查询 API Key 列表")

	resp, xErr := h.service.apikeyLogic.List(ctx)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// GetByID 查询 API Key 详情
//
// @Summary		查询 API Key 详情
// @Tags		apikey
// @Accept		json
// @Produce		json
// @Param		id	path		string					true	"API Key ID"
// @Success		200	{object}	apiCommon.BaseResponse		"查询成功，返回脱敏详情"
// @Failure		401	{object}	apiCommon.BaseResponse		"未授权"
// @Failure		404	{object}	apiCommon.BaseResponse		"API Key 不存在"
// @Security	ApiKeyAuth
// @Router		/api/v1/apikey/{id} [get]
func (h *ApikeyHandler) GetByID(ctx *gin.Context) {
	h.log.Info(ctx, "GetByID - 查询 API Key 详情")

	idStr := ctx.Param("id")
	id, err := xSnowflake.ParseSnowflakeID(idStr)
	if err != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.BadRequest, "无效的 ID 格式", false, err))
		return
	}

	resp, xErr := h.service.apikeyLogic.GetByID(ctx, id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// Update 更新 API Key
//
// @Summary		更新 API Key
// @Tags		apikey
// @Accept		json
// @Produce		json
// @Param		id		path		string					true	"API Key ID"
// @Param		request	body		apiApikey.UpdateRequest	true	"更新请求"
// @Success		200		{object}	apiCommon.BaseResponse		"更新成功"
// @Failure		400		{object}	apiCommon.BaseResponse		"参数错误"
// @Failure		401		{object}	apiCommon.BaseResponse		"未授权"
// @Failure		404		{object}	apiCommon.BaseResponse		"API Key 不存在"
// @Security	ApiKeyAuth
// @Router		/api/v1/apikey/{id} [put]
func (h *ApikeyHandler) Update(ctx *gin.Context) {
	h.log.Info(ctx, "Update - 更新 API Key")

	idStr := ctx.Param("id")
	id, err := xSnowflake.ParseSnowflakeID(idStr)
	if err != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.BadRequest, "无效的 ID 格式", false, err))
		return
	}

	var req apiApikey.UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	if xErr := h.service.apikeyLogic.Update(ctx, id, &req); xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "更新成功")
}

// Delete 删除 API Key
//
// @Summary		删除 API Key
// @Tags		apikey
// @Accept		json
// @Produce		json
// @Param		id	path		string					true	"API Key ID"
// @Success		200	{object}	apiCommon.BaseResponse		"删除成功"
// @Failure		401	{object}	apiCommon.BaseResponse		"未授权"
// @Failure		404	{object}	apiCommon.BaseResponse		"API Key 不存在"
// @Security	ApiKeyAuth
// @Router		/api/v1/apikey/{id} [delete]
func (h *ApikeyHandler) Delete(ctx *gin.Context) {
	h.log.Info(ctx, "Delete - 删除 API Key")

	idStr := ctx.Param("id")
	id, err := xSnowflake.ParseSnowflakeID(idStr)
	if err != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.BadRequest, "无效的 ID 格式", false, err))
		return
	}

	if xErr := h.service.apikeyLogic.Delete(ctx, id); xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "删除成功")
}

// Reset 重置 API Key
//
// @Summary		重置 API Key
// @Tags		apikey
// @Accept		json
// @Produce		json
// @Param		id	path		string					true	"API Key ID"
// @Success		200	{object}	apiCommon.BaseResponse		"重置成功，返回新Key"
// @Failure		401	{object}	apiCommon.BaseResponse		"未授权"
// @Failure		404	{object}	apiCommon.BaseResponse		"API Key 不存在"
// @Security	ApiKeyAuth
// @Router		/api/v1/apikey/{id}/reset [post]
func (h *ApikeyHandler) Reset(ctx *gin.Context) {
	h.log.Info(ctx, "Reset - 重置 API Key")

	idStr := ctx.Param("id")
	id, err := xSnowflake.ParseSnowflakeID(idStr)
	if err != nil {
		_ = ctx.Error(xError.NewError(ctx, xError.BadRequest, "无效的 ID 格式", false, err))
		return
	}

	resp, xErr := h.service.apikeyLogic.Reset(ctx, id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "重置成功", resp)
}
