package handler

import (
	"strconv"

	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiApikey "github.com/xiaolfeng/Lumina/api/apikey"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// 确保 xModels 分页类型被 swag 识别（List 接口返回 PageResponse）
var _ = xModels.PageResponse[[]apiApikey.ApikeyItem]{}

// Create 创建API密钥
//
// @Summary     [管理] 创建API密钥
// @Description 提交名称与可选描述/过期时间创建 API 密钥，返回完整密钥（仅此一次）
// @Tags        API密钥接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                              true  "Bearer Access Token"
// @Param       request        body      apiApikey.CreateRequest             true  "创建请求"
// @Success     200            {object}  apiCommon.BaseResponse{data=apiApikey.CreateResponse}  "创建成功，返回完整密钥（仅此一次）"
// @Failure     400            {object}  apiCommon.BaseResponse              "请求参数错误"
// @Failure     401            {object}  apiCommon.BaseResponse              "未授权"
// @Failure     500            {object}  apiCommon.BaseResponse              "服务器内部错误"
// @Router      /api/v1/apikey [POST]
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
// @Summary     [管理] 分页获取API密钥列表
// @Description 按 page/size 分页查询 API 密钥列表，密钥以脱敏形式展示
// @Tags        API密钥接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       page           query     int      false  "页码"  default(1)
// @Param       size           query     int      false  "每页数量（最大200）"  default(10)
// @Success     200  {object}  apiCommon.BaseResponse{data=xModels.PageResponse[[]apiApikey.ApikeyItem]}  "查询成功，密钥脱敏展示"
// @Failure     401  {object}  apiCommon.BaseResponse              "未授权"
// @Failure     500  {object}  apiCommon.BaseResponse              "服务器内部错误"
// @Router      /api/v1/apikey [GET]
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
// @Summary     [管理] 获取API密钥详情
// @Description 根据密钥 ID 查询单个 API 密钥详情，密钥以脱敏形式展示
// @Tags        API密钥接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "API密钥ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiApikey.DetailResponse}  "查询成功，密钥脱敏展示"
// @Failure     400  {object}  apiCommon.BaseResponse              "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse              "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse              "密钥不存在"
// @Router      /api/v1/apikey/{id} [GET]
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
// @Summary     [管理] 更新API密钥
// @Description 更新指定 ID 的 API 密钥信息（名称、描述、过期时间、启用状态，字段可选）
// @Tags        API密钥接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                       true  "Bearer Access Token"
// @Param       id             path      string                       true  "API密钥ID"
// @Param       request        body      apiApikey.UpdateRequest      true  "更新请求"
// @Success     200  {object}  apiCommon.BaseResponse                "更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse                "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse                "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse                "密钥不存在"
// @Failure     500  {object}  apiCommon.BaseResponse                "服务器内部错误"
// @Router      /api/v1/apikey/{id} [PUT]
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
// @Summary     [管理] 删除API密钥
// @Description 根据密钥 ID 删除指定的 API 密钥，删除后不可恢复
// @Tags        API密钥接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "API密钥ID"
// @Success     200  {object}  apiCommon.BaseResponse  "删除成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "密钥不存在"
// @Router      /api/v1/apikey/{id} [DELETE]
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
// @Summary     [管理] 重置API密钥
// @Description 根据密钥 ID 重置密钥，生成新密钥并返回完整密钥（仅此一次）
// @Tags        API密钥接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "API密钥ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiApikey.ResetResponse}  "重置成功，返回新完整密钥（仅此一次）"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "密钥不存在"
// @Failure     500  {object}  apiCommon.BaseResponse  "服务器内部错误"
// @Router      /api/v1/apikey/{id}/reset [POST]
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
