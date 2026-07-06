package handler

import (
	"strconv"

	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	"github.com/gin-gonic/gin"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiLlm "github.com/xiaolfeng/Lumina/api/llm"
)

// 确保 apiCommon 包被编译器识别（swag 注释依赖此导入）
var _ = apiCommon.BaseResponse{}

// 确保 xModels 分页类型被 swag 识别（List 接口返回 PageResponse）
var _ = xModels.PageResponse[[]apiLlm.ProviderListItem]{}
var _ = xModels.PageResponse[[]apiLlm.ModelListItem]{}

// ──────────────────────────── Provider ────────────────────────────

// CreateProvider 创建 LLM Provider
//
// @Summary     [管理] 创建LLM Provider
// @Description 提交名称、协议、API密钥等创建 LLM Provider，密钥加密存储不返回明文
// @Tags        LLM配置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                              true  "Bearer Access Token"
// @Param       request        body      apiLlm.CreateProviderRequest        true  "创建Provider请求"
// @Success     200            {object}  apiCommon.BaseResponse{data=apiLlm.ProviderDetailResponse}  "创建成功"
// @Failure     400            {object}  apiCommon.BaseResponse              "请求参数错误"
// @Failure     401            {object}  apiCommon.BaseResponse              "未授权"
// @Failure     500            {object}  apiCommon.BaseResponse              "服务器内部错误"
// @Router      /api/v1/llm/provider [POST]
func (h *LlmHandler) CreateProvider(ctx *gin.Context) {
	h.log.Info(ctx, "CreateProvider - 创建LLM Provider")

	var req apiLlm.CreateProviderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	resp, xErr := h.service.llmProviderLogic.Create(ctx.Request.Context(), &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "LLM Provider创建成功", resp)
}

// ListProviders 分页获取 LLM Provider 列表
//
// @Summary     [管理] 分页获取LLM Provider列表
// @Description 按 page/size 分页查询 LLM Provider 列表，密钥不返回明文
// @Tags        LLM配置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       page           query     int      false  "页码"  default(1)
// @Param       size           query     int      false  "每页数量（最大200）"  default(20)
// @Success     200  {object}  apiCommon.BaseResponse{data=xModels.PageResponse[[]apiLlm.ProviderListItem]}  "查询成功"
// @Failure     401  {object}  apiCommon.BaseResponse              "未授权"
// @Failure     500  {object}  apiCommon.BaseResponse              "服务器内部错误"
// @Router      /api/v1/llm/provider [GET]
func (h *LlmHandler) ListProviders(ctx *gin.Context) {
	h.log.Info(ctx, "ListProviders - 分页获取LLM Provider列表")

	page, _ := strconv.Atoi(ctx.Query("page"))
	size, _ := strconv.Atoi(ctx.Query("size"))
	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	page = int(pageReq.Page)
	size = int(pageReq.Size)

	resp, xErr := h.service.llmProviderLogic.List(ctx.Request.Context(), page, size)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// GetProvider 根据 ID 获取 LLM Provider 详情
//
// @Summary     [管理] 获取LLM Provider详情
// @Description 根据 Provider ID 查询单个 LLM Provider 详情，密钥不返回明文
// @Tags        LLM配置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "Provider ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiLlm.ProviderDetailResponse}  "查询成功"
// @Failure     400  {object}  apiCommon.BaseResponse              "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse              "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse              "Provider不存在"
// @Router      /api/v1/llm/provider/{id} [GET]
func (h *LlmHandler) GetProvider(ctx *gin.Context) {
	h.log.Info(ctx, "GetProvider - 获取LLM Provider详情")

	id := ctx.Param("id")
	resp, xErr := h.service.llmProviderLogic.GetByID(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// UpdateProvider 更新 LLM Provider 信息
//
// @Summary     [管理] 更新LLM Provider
// @Description 更新指定 ID 的 LLM Provider 信息（名称、协议、端点、密钥、启用状态等，字段可选；密钥为空时不更新）
// @Tags        LLM配置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                       true  "Bearer Access Token"
// @Param       id             path      string                       true  "Provider ID"
// @Param       request        body      apiLlm.UpdateProviderRequest true  "更新Provider请求"
// @Success     200  {object}  apiCommon.BaseResponse                "更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse                "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse                "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse                "Provider不存在"
// @Failure     500  {object}  apiCommon.BaseResponse                "服务器内部错误"
// @Router      /api/v1/llm/provider/{id} [PUT]
func (h *LlmHandler) UpdateProvider(ctx *gin.Context) {
	h.log.Info(ctx, "UpdateProvider - 更新LLM Provider")

	id := ctx.Param("id")
	var req apiLlm.UpdateProviderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	xErr := h.service.llmProviderLogic.Update(ctx.Request.Context(), id, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "LLM Provider更新成功")
}

// DeleteProvider 删除 LLM Provider
//
// @Summary     [管理] 删除LLM Provider
// @Description 根据 Provider ID 删除指定的 LLM Provider，存在关联模型时拒绝删除
// @Tags        LLM配置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "Provider ID"
// @Success     200  {object}  apiCommon.BaseResponse  "删除成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "Provider不存在"
// @Failure     409  {object}  apiCommon.BaseResponse  "存在关联模型，无法删除"
// @Router      /api/v1/llm/provider/{id} [DELETE]
func (h *LlmHandler) DeleteProvider(ctx *gin.Context) {
	h.log.Info(ctx, "DeleteProvider - 删除LLM Provider")

	id := ctx.Param("id")
	xErr := h.service.llmProviderLogic.Delete(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "LLM Provider删除成功")
}

// ──────────────────────────── Model ────────────────────────────

// CreateModel 创建 LLM 模型
//
// @Summary     [管理] 创建LLM模型
// @Description 提交关联Provider ID、模型标识、显示名称等创建 LLM 模型
// @Tags        LLM配置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                              true  "Bearer Access Token"
// @Param       request        body      apiLlm.CreateModelRequest           true  "创建模型请求"
// @Success     200            {object}  apiCommon.BaseResponse{data=apiLlm.ModelDetailResponse}  "创建成功"
// @Failure     400            {object}  apiCommon.BaseResponse              "请求参数错误"
// @Failure     401            {object}  apiCommon.BaseResponse              "未授权"
// @Failure     404            {object}  apiCommon.BaseResponse              "关联Provider不存在"
// @Failure     500            {object}  apiCommon.BaseResponse              "服务器内部错误"
// @Router      /api/v1/llm/model [POST]
func (h *LlmHandler) CreateModel(ctx *gin.Context) {
	h.log.Info(ctx, "CreateModel - 创建LLM模型")

	var req apiLlm.CreateModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	resp, xErr := h.service.llmModelLogic.Create(ctx.Request.Context(), &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "LLM模型创建成功", resp)
}

// ListModels 分页获取 LLM 模型列表
//
// @Summary     [管理] 分页获取LLM模型列表
// @Description 按 page/size 分页查询 LLM 模型列表，支持 provider_id 可选筛选
// @Tags        LLM配置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true   "Bearer Access Token"
// @Param       page           query     int      false  "页码"  default(1)
// @Param       size           query     int      false  "每页数量（最大200）"  default(20)
// @Param       provider_id    query     string   false  "Provider ID（可选筛选）"
// @Success     200  {object}  apiCommon.BaseResponse{data=xModels.PageResponse[[]apiLlm.ModelListItem]}  "查询成功"
// @Failure     401  {object}  apiCommon.BaseResponse              "未授权"
// @Failure     500  {object}  apiCommon.BaseResponse              "服务器内部错误"
// @Router      /api/v1/llm/model [GET]
func (h *LlmHandler) ListModels(ctx *gin.Context) {
	h.log.Info(ctx, "ListModels - 分页获取LLM模型列表")

	page, _ := strconv.Atoi(ctx.Query("page"))
	size, _ := strconv.Atoi(ctx.Query("size"))
	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	page = int(pageReq.Page)
	size = int(pageReq.Size)
	providerID := ctx.Query("provider_id")

	resp, xErr := h.service.llmModelLogic.List(ctx.Request.Context(), page, size, providerID)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// GetModel 根据 ID 获取 LLM 模型详情
//
// @Summary     [管理] 获取LLM模型详情
// @Description 根据模型 ID 查询单个 LLM 模型详情
// @Tags        LLM配置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "模型ID"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiLlm.ModelDetailResponse}  "查询成功"
// @Failure     400  {object}  apiCommon.BaseResponse              "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse              "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse              "模型不存在"
// @Router      /api/v1/llm/model/{id} [GET]
func (h *LlmHandler) GetModel(ctx *gin.Context) {
	h.log.Info(ctx, "GetModel - 获取LLM模型详情")

	id := ctx.Param("id")
	resp, xErr := h.service.llmModelLogic.GetByID(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// UpdateModel 更新 LLM 模型信息
//
// @Summary     [管理] 更新LLM模型
// @Description 更新指定 ID 的 LLM 模型信息（模型标识、显示名称、参数等，字段可选）
// @Tags        LLM配置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                       true  "Bearer Access Token"
// @Param       id             path      string                       true  "模型ID"
// @Param       request        body      apiLlm.UpdateModelRequest    true  "更新模型请求"
// @Success     200  {object}  apiCommon.BaseResponse                "更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse                "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse                "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse                "模型不存在"
// @Failure     500  {object}  apiCommon.BaseResponse                "服务器内部错误"
// @Router      /api/v1/llm/model/{id} [PUT]
func (h *LlmHandler) UpdateModel(ctx *gin.Context) {
	h.log.Info(ctx, "UpdateModel - 更新LLM模型")

	id := ctx.Param("id")
	var req apiLlm.UpdateModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	xErr := h.service.llmModelLogic.Update(ctx.Request.Context(), id, &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "LLM模型更新成功")
}

// DeleteModel 删除 LLM 模型
//
// @Summary     [管理] 删除LLM模型
// @Description 根据模型 ID 删除指定的 LLM 模型，删除后不可恢复
// @Tags        LLM配置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       id             path      string   true  "模型ID"
// @Success     200  {object}  apiCommon.BaseResponse  "删除成功"
// @Failure     400  {object}  apiCommon.BaseResponse  "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse  "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse  "模型不存在"
// @Router      /api/v1/llm/model/{id} [DELETE]
func (h *LlmHandler) DeleteModel(ctx *gin.Context) {
	h.log.Info(ctx, "DeleteModel - 删除LLM模型")

	id := ctx.Param("id")
	xErr := h.service.llmModelLogic.Delete(ctx.Request.Context(), id)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "LLM模型删除成功")
}

// ──────────────────────────── Agent ────────────────────────────

// GetAgentModel 获取 Agent 角色对应的模型分配
//
// @Summary     [管理] 获取Agent模型分配
// @Description 根据 Agent 角色查询当前分配的模型 ID，未分配时 model_id 为 null
// @Tags        LLM配置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string   true  "Bearer Access Token"
// @Param       role           path      string   true  "Agent角色"
// @Success     200  {object}  apiCommon.BaseResponse{data=apiLlm.AgentModelAssignment}  "查询成功"
// @Failure     401  {object}  apiCommon.BaseResponse              "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse              "未配置Agent模型"
// @Router      /api/v1/llm/agent/{role}/model [GET]
func (h *LlmHandler) GetAgentModel(ctx *gin.Context) {
	h.log.Info(ctx, "GetAgentModel - 获取Agent模型分配")

	role := ctx.Param("role")
	config, xErr := h.service.llmModelLogic.GetAgentModelConfig(ctx.Request.Context(), role)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	modelID := config.Model.BaseEntity.ID.String()
	resp := &apiLlm.AgentModelAssignment{
		Role:    role,
		ModelID: &modelID,
	}

	xResult.SuccessHasData(ctx, "查询成功", resp)
}

// UpdateAgentModel 更新 Agent 角色对应的模型分配
//
// @Summary     [管理] 更新Agent模型分配
// @Description 为指定 Agent 角色分配模型，需先验证模型存在
// @Tags        LLM配置接口
// @Accept      json
// @Produce     json
// @Param       Authorization  header    string                       true  "Bearer Access Token"
// @Param       role           path      string                       true  "Agent角色"
// @Param       request        body      apiLlm.UpdateAgentModelRequest true "更新Agent模型分配请求"
// @Success     200  {object}  apiCommon.BaseResponse                "更新成功"
// @Failure     400  {object}  apiCommon.BaseResponse                "请求参数错误"
// @Failure     401  {object}  apiCommon.BaseResponse                "未授权"
// @Failure     404  {object}  apiCommon.BaseResponse                "模型不存在"
// @Router      /api/v1/llm/agent/{role}/model [PUT]
func (h *LlmHandler) UpdateAgentModel(ctx *gin.Context) {
	h.log.Info(ctx, "UpdateAgentModel - 更新Agent模型分配")

	role := ctx.Param("role")
	var req apiLlm.UpdateAgentModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		_ = ctx.Error(err)
		return
	}

	xErr := h.service.llmModelLogic.SetAgentModel(ctx.Request.Context(), role, req.ModelID)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}

	xResult.Success(ctx, "Agent模型分配更新成功")
}
