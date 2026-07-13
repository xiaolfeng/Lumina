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
// @Summary     [管理] 分页获取QA会话列表
// @Description 按 page/size 分页查询会话列表，支持按状态与类型过滤
// @Tags        Q&A接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true   "Bearer Access Token"
// @Param       page           query     int      false  "页码"
// @Param       size           query     int      false  "每页数量"
// @Param       status         query     string   false  "状态过滤(active/expired/deleted)"
// @Param       type           query     string   false  "类型过滤(temporary/permanent)"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiQa.SessionListResponse}  "获取成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/qa/sessions [GET]
func (h *QaHandler) ListSessions(ctx *gin.Context) {
	h.log.Info(ctx, "ListSessions - 获取QA会话列表")

	var req apiQa.ListSessionRequest
	if !BindQuery(ctx, &req) {
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
// @Summary     [管理] 获取QA会话详情
// @Description 根据会话 ID 查询会话详情，含完整问题列表（含回答与补充内容）
// @Tags        Q&A接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "会话ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiQa.SessionDetailResponse}  "获取成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "会话不存在"
// @Router      /api/v1/qa/sessions/{id} [GET]
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
// @Summary     [管理] 删除QA会话
// @Description 根据会话 ID 删除指定会话及其关联问题，删除后不可恢复
// @Tags        Q&A接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "会话ID"
// @Success     200  {object}  apiCommon.BaseResponse  "删除成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "会话不存在"
// @Router      /api/v1/qa/sessions/{id} [DELETE]
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

// CreateSession 创建QA会话
//
// @Summary     [管理] 创建QA会话
// @Description 关联项目 ID 创建 Q&A 会话，支持指定标题、Agent 名称与会话类型，返回会话详情
// @Tags        Q&A接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                          true  "Bearer Access Token"
// @Param       request        body      apiQa.CreateSessionRequest      true  "创建请求"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiQa.SessionDetailResponse}  "创建成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Router      /api/v1/qa/sessions [POST]
func (h *QaHandler) CreateSession(ctx *gin.Context) {
	h.log.Info(ctx, "CreateSession - 创建QA会话")

	var req apiQa.CreateSessionRequest
	if !BindJSON(ctx, &req) {
		return
	}

	agent := req.Agent
	if agent == "" {
		agent = "web"
	}
	sessionType := req.Type
	if sessionType == "" {
		sessionType = "temporary"
	}

	id, _, xErr := h.service.qaLogic.CreateSession(ctx.Request.Context(), req.Title, agent, sessionType, req.ProjectID.String())
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	detailResp, xErr := h.service.qaLogic.GetSessionDetail(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "创建成功", detailResp)
}

// GetQuestion 获取问题详情（含回答+补充内容）
//
// @Summary     [管理] 获取问题详情
// @Description 根据会话 ID 与问题 ID 查询问题详情，含回答内容与关联补充内容
// @Tags        Q&A接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "会话ID"
// @Param       qid            path      string   true  "问题ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiQa.QuestionDetailResponse}  "获取成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "问题不存在"
// @Router      /api/v1/qa/sessions/{id}/questions/{qid} [GET]
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
// @Summary     [管理] 获取Q&A配置
// @Description 查询当前 Q&A 模块运行配置（会话TTL、运行时域名、轮询参数、重试次数）
// @Tags        Q&A接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiQa.QaConfigResponse}  "获取成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/qa/config [GET]
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
// @Summary     [管理] 更新Q&A配置
// @Description 更新 Q&A 模块运行配置，所有字段可选（nil 表示不更新）
// @Tags        Q&A接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                          true  "Bearer Access Token"
// @Param       request        body      apiQa.UpdateQaConfigRequest     true  "配置信息"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiQa.QaConfigResponse}  "更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Router      /api/v1/qa/config [PUT]
func (h *QaHandler) UpdateQaConfig(ctx *gin.Context) {
	h.log.Info(ctx, "UpdateQaConfig - 更新Q&A配置")

	var req apiQa.UpdateQaConfigRequest
	if !BindJSON(ctx, &req) {
		return
	}

	resp, xErr := h.service.qaLogic.UpdateQaConfig(ctx.Request.Context(), &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "更新成功", resp)
}
