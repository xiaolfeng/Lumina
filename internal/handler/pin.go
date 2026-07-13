package handler

import (
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiPin "github.com/xiaolfeng/Lumina/api/pin"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// CreatePin 创建 Pin 约束
//
// @Summary     [管理] 创建 Pin 约束
// @Description 提交标题、内容、分类、优先级与目标项目创建跨项目依赖约束，推送到目标项目待消费队列
// @Tags        Pin接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                     true  "Bearer Access Token"
// @Param       request        body      apiPin.CreatePinRequest    true  "创建 Pin 请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPin.PinResponse}  "创建成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "目标项目不存在"
// @Router      /api/v1/pin [POST]
func (h *PinHandler) CreatePin(ctx *gin.Context) {
	h.log.Info(ctx, "CreatePin - 创建 Pin 约束")

	var req apiPin.CreatePinRequest
	if !BindJSON(ctx, &req) {
		return
	}

	resp, xErr := h.service.pinLogic.Push(ctx.Request.Context(), &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "创建成功", resp)
}

// ListPins 获取 Pin 列表
//
// @Summary     [管理] 获取 Pin 列表
// @Description 按目标项目、来源项目、状态、分类、优先级多条件动态过滤，page/size 分页查询
// @Tags        Pin接口
// @Accept      json
// @Produce     json
// @Param       Authorization    header    string   true   "Bearer Access Token"
// @Param       to_project_id    query     string   false  "目标项目ID筛选"
// @Param       from_project_id  query     string   false  "来源项目ID筛选"
// @Param       status           query     string   false  "状态筛选 (pending/consumed)"
// @Param       category         query     string   false  "分类筛选 (notice/dependency/api_change/other)"
// @Param       priority         query     string   false  "优先级筛选 (high/medium/low)"
// @Param       page             query     int      false  "页码"  default(1)
// @Param       size             query     int      false  "每页数量"  default(20)
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPin.PinListResponse}  "查询成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/pin [GET]
func (h *PinHandler) ListPins(ctx *gin.Context) {
	h.log.Info(ctx, "ListPins - 获取 Pin 列表")

	var req apiPin.PinListRequest
	if !BindQuery(ctx, &req) {
		return
	}

	resp, xErr := h.service.pinLogic.List(ctx.Request.Context(), &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// GetPin 获取 Pin 详情
//
// @Summary     [管理] 获取 Pin 详情
// @Description 根据 Pin ID 查询单个约束详情（只读，不改变状态）
// @Tags        Pin接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "Pin ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPin.PinResponse}  "查询成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "Pin 不存在"
// @Router      /api/v1/pin/{id} [GET]
func (h *PinHandler) GetPin(ctx *gin.Context) {
	h.log.Info(ctx, "GetPin - 获取 Pin 详情")

	id, xErr := ParseSnowflakeID(ctx, ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	resp, xErr := h.service.pinLogic.Peek(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// UpdatePin 更新 Pin 元数据
//
// @Summary     [管理] 更新 Pin 元数据
// @Description 更新指定 ID 的 Pin 优先级与分类（可选字段），不支持更新状态
// @Tags        Pin接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                     true  "Bearer Access Token"
// @Param       id             path      string                     true  "Pin ID"
// @Param       request        body      apiPin.UpdatePinRequest    true  "更新 Pin 请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiPin.PinResponse}  "更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "Pin 不存在"
// @Router      /api/v1/pin/{id} [PUT]
func (h *PinHandler) UpdatePin(ctx *gin.Context) {
	h.log.Info(ctx, "UpdatePin - 更新 Pin 元数据")

	id, xErr := ParseSnowflakeID(ctx, ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	var req apiPin.UpdatePinRequest
	if !BindJSON(ctx, &req) {
		return
	}

	resp, xErr := h.service.pinLogic.Update(ctx.Request.Context(), id, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "更新成功", resp)
}

// DeletePin 删除 Pin 约束
//
// @Summary     [管理] 删除 Pin 约束
// @Description 根据 Pin ID 删除指定约束，删除后不可恢复
// @Tags        Pin接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "Pin ID"
// @Success     200  {object}  apiCommon.BaseResponse  "删除成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "Pin 不存在"
// @Router      /api/v1/pin/{id} [DELETE]
func (h *PinHandler) DeletePin(ctx *gin.Context) {
	h.log.Info(ctx, "DeletePin - 删除 Pin 约束")

	id, xErr := ParseSnowflakeID(ctx, ctx.Param("id"))
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xErr = h.service.pinLogic.Delete(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "删除成功")
}
