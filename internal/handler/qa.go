package handler

import (
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiQa "github.com/xiaolfeng/Lumina/api/qa"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// ListSessions 分页获取QA会话列表
//
// @Summary      分页获取QA会话列表
// @Tags         QA管理
// @Accept       json
// @Produce      json
// @Param        page     query     int                       false  "页码"
// @Param        size     query     int                       false  "每页数量"
// @Param        status   query     string                    false  "状态过滤(active/expired/deleted)"
// @Param        type     query     string                    false  "类型过滤(temporary/permanent)"
// @Success      200      {object}  apiCommon.BaseResponse    "获取成功"
// @Failure      401      {object}  apiCommon.BaseResponse    "未认证"
// @Security     ApiKeyAuth
// @Router       /api/v1/qa/sessions [get]
func (h *QaHandler) ListSessions(ctx *gin.Context) {
	h.log.Info(ctx, "ListSessions - 获取QA会话列表")

	var req apiQa.ListSessionRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	resp, xErr := h.service.qaLogic.ListSessions(ctx.Request.Context(), &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", resp)
}

// GetSession 获取QA会话详情（含问题列表）
//
// @Summary      获取QA会话详情
// @Tags         QA管理
// @Accept       json
// @Produce      json
// @Param        id       path      string                 true  "会话ID"
// @Success      200      {object}  apiCommon.BaseResponse "获取成功"
// @Failure      400      {object}  apiCommon.BaseResponse "参数错误"
// @Failure      401      {object}  apiCommon.BaseResponse "未认证"
// @Failure      404      {object}  apiCommon.BaseResponse "会话不存在"
// @Security     ApiKeyAuth
// @Router       /api/v1/qa/sessions/{id} [get]
func (h *QaHandler) GetSession(ctx *gin.Context) {
	h.log.Info(ctx, "GetSession - 获取QA会话详情")

	id := ctx.Param("id")

	resp, xErr := h.service.qaLogic.GetSessionDetail(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", resp)
}

// DeleteSession 删除QA会话
//
// @Summary      删除QA会话
// @Tags         QA管理
// @Accept       json
// @Produce      json
// @Param        id       path      string                 true  "会话ID"
// @Success      200      {object}  apiCommon.BaseResponse "删除成功"
// @Failure      400      {object}  apiCommon.BaseResponse "参数错误"
// @Failure      401      {object}  apiCommon.BaseResponse "未认证"
// @Failure      404      {object}  apiCommon.BaseResponse "会话不存在"
// @Security     ApiKeyAuth
// @Router       /api/v1/qa/sessions/{id} [delete]
func (h *QaHandler) DeleteSession(ctx *gin.Context) {
	h.log.Info(ctx, "DeleteSession - 删除QA会话")

	id := ctx.Param("id")

	xErr := h.service.qaLogic.DeleteSession(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "删除成功")
}

// GetQuestion 获取问题详情（含回答+补充内容）
//
// @Summary      获取问题详情
// @Tags         QA管理
// @Accept       json
// @Produce      json
// @Param        id       path      string                 true  "会话ID"
// @Param        qid      path      string                 true  "问题ID"
// @Success      200      {object}  apiCommon.BaseResponse "获取成功"
// @Failure      400      {object}  apiCommon.BaseResponse "参数错误"
// @Failure      401      {object}  apiCommon.BaseResponse "未认证"
// @Failure      404      {object}  apiCommon.BaseResponse "问题不存在"
// @Security     ApiKeyAuth
// @Router       /api/v1/qa/sessions/{id}/questions/{qid} [get]
func (h *QaHandler) GetQuestion(ctx *gin.Context) {
	h.log.Info(ctx, "GetQuestion - 获取问题详情")

	sessionID := ctx.Param("id")
	questionID := ctx.Param("qid")

	resp, xErr := h.service.qaLogic.GetQuestionDetail(ctx.Request.Context(), sessionID, questionID)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", resp)
}

// GetQaConfig 获取Q&A配置
//
// @Summary      获取Q&A配置
// @Tags         QA管理
// @Accept       json
// @Produce      json
// @Success      200      {object}  apiCommon.BaseResponse "获取成功"
// @Failure      401      {object}  apiCommon.BaseResponse "未认证"
// @Security     ApiKeyAuth
// @Router       /api/v1/qa/config [get]
func (h *QaHandler) GetQaConfig(ctx *gin.Context) {
	h.log.Info(ctx, "GetQaConfig - 获取Q&A配置")

	resp, xErr := h.service.qaLogic.GetQaConfig(ctx.Request.Context())
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "获取成功", resp)
}

// UpdateQaConfig 更新Q&A配置
//
// @Summary      更新Q&A配置
// @Tags         QA管理
// @Accept       json
// @Produce      json
// @Param        request  body      apiQa.UpdateQaConfigRequest  true  "配置信息"
// @Success      200      {object}  apiCommon.BaseResponse       "更新成功"
// @Failure      400      {object}  apiCommon.BaseResponse       "参数错误"
// @Failure      401      {object}  apiCommon.BaseResponse       "未认证"
// @Security     ApiKeyAuth
// @Router       /api/v1/qa/config [put]
func (h *QaHandler) UpdateQaConfig(ctx *gin.Context) {
	h.log.Info(ctx, "UpdateQaConfig - 更新Q&A配置")

	var req apiQa.UpdateQaConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	resp, xErr := h.service.qaLogic.UpdateQaConfig(ctx.Request.Context(), &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "更新成功", resp)
}
