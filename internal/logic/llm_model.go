package logic

import (
	"context"
	"strconv"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	apiLlm "github.com/xiaolfeng/Lumina/api/llm"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

// llmModelRepo LLM Model 模块依赖的仓储集合（持有 InfoRepo 用于 Agent 分配）
type llmModelRepo struct {
	model    *repository.LlmModelRepo
	provider *repository.LlmProviderRepo
	info     *repository.InfoRepo
}

// LlmModelLogic LLM Model 业务逻辑层
type LlmModelLogic struct {
	logic
	repo llmModelRepo
}

// AgentModelConfig Agent 模型配置复合返回类型，供 LLM 消费层使用
type AgentModelConfig struct {
	Model    *entity.LlmModel
	Provider *entity.LlmProvider
}

// NewLlmModelLogic 创建 LLM Model 业务逻辑层实例
func NewLlmModelLogic(ctx context.Context) *LlmModelLogic {
	db := xCtxUtil.MustGetDB(ctx)

	return &LlmModelLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "LlmModelLogic"),
		},
		repo: llmModelRepo{
			model:    repository.NewLlmModelRepo(db),
			provider: repository.NewLlmProviderRepo(db),
			info:     repository.NewInfoRepo(db),
		},
	}
}

// Create 创建 LLM Model（校验 ProviderID 对应的 Provider 存在）
func (l *LlmModelLogic) Create(ctx context.Context, req *apiLlm.CreateModelRequest) (*apiLlm.ModelDetailResponse, *xError.Error) {
	l.log.Info(ctx, "Create - 创建LLM Model")

	providerID, xErr := parseSnowflakeID(ctx, req.ProviderID)
	if xErr != nil {
		return nil, xErr
	}

	// 校验 Provider 存在
	if _, xErr := l.repo.provider.GetByID(ctx, providerID); xErr != nil {
		return nil, xErr
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}
	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.3
	}

	model := &entity.LlmModel{
		ProviderID:  providerID,
		ModelName:   req.ModelName,
		DisplayName: req.DisplayName,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		IsActive:    true,
		Description: req.Description,
	}

	if xErr := l.repo.model.Create(ctx, model); xErr != nil {
		return nil, xErr
	}

	l.log.Info(ctx, "Create - LLM Model创建成功")

	return l.toDetailResponse(model), nil
}

// List 分页获取 LLM Model 列表（支持 providerId 可选筛选）
func (l *LlmModelLogic) List(ctx context.Context, page, size int, providerID string) (*xModels.PageResponse[[]apiLlm.ModelListItem], *xError.Error) {
	l.log.Info(ctx, "List - 分页获取LLM Model列表")

	if providerID != "" {
		parsedID, xErr := parseSnowflakeID(ctx, providerID)
		if xErr != nil {
			return nil, xErr
		}
		models, xErr := l.repo.model.ListByProviderID(ctx, parsedID)
		if xErr != nil {
			return nil, xErr
		}
		listItems := make([]apiLlm.ModelListItem, 0, len(models))
		for _, m := range models {
			listItems = append(listItems, l.toListListItem(m))
		}
		pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
		return xModels.NewPageFromRequest(pageReq, int64(len(listItems)), listItems), nil
	}

	items, total, xErr := l.repo.model.List(ctx, page, size)
	if xErr != nil {
		return nil, xErr
	}

	listItems := make([]apiLlm.ModelListItem, 0, len(items))
	for _, item := range items {
		listItems = append(listItems, l.toListListItem(item))
	}

	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	return xModels.NewPageFromRequest(pageReq, total, listItems), nil
}

// GetByID 根据 ID 获取 LLM Model 详情
func (l *LlmModelLogic) GetByID(ctx context.Context, id string) (*apiLlm.ModelDetailResponse, *xError.Error) {
	l.log.Info(ctx, "GetByID - 根据ID获取LLM Model详情")

	parsedID, xErr := parseSnowflakeID(ctx, id)
	if xErr != nil {
		return nil, xErr
	}

	model, xErr := l.repo.model.GetByID(ctx, parsedID)
	if xErr != nil {
		return nil, xErr
	}

	return l.toDetailResponse(model), nil
}

// Update 更新 LLM Model 信息
func (l *LlmModelLogic) Update(ctx context.Context, id string, req *apiLlm.UpdateModelRequest) *xError.Error {
	l.log.Info(ctx, "Update - 更新LLM Model")

	parsedID, xErr := parseSnowflakeID(ctx, id)
	if xErr != nil {
		return xErr
	}

	model, xErr := l.repo.model.GetByID(ctx, parsedID)
	if xErr != nil {
		return xErr
	}

	if req.ModelName != nil {
		model.ModelName = *req.ModelName
	}
	if req.DisplayName != nil {
		model.DisplayName = *req.DisplayName
	}
	if req.MaxTokens != nil {
		model.MaxTokens = *req.MaxTokens
	}
	if req.Temperature != nil {
		model.Temperature = *req.Temperature
	}
	if req.IsActive != nil {
		model.IsActive = *req.IsActive
	}
	if req.Description != nil {
		model.Description = *req.Description
	}

	if xErr := l.repo.model.Update(ctx, model); xErr != nil {
		return xErr
	}

	l.log.Info(ctx, "Update - LLM Model更新成功")
	return nil
}

// Delete 删除 LLM Model
func (l *LlmModelLogic) Delete(ctx context.Context, id string) *xError.Error {
	l.log.Info(ctx, "Delete - 删除LLM Model")

	parsedID, xErr := parseSnowflakeID(ctx, id)
	if xErr != nil {
		return xErr
	}

	if xErr := l.repo.model.Delete(ctx, parsedID); xErr != nil {
		return xErr
	}

	l.log.Info(ctx, "Delete - LLM Model删除成功")
	return nil
}

// SetActive 快捷启用/禁用 LLM Model
func (l *LlmModelLogic) SetActive(ctx context.Context, id string, active bool) *xError.Error {
	l.log.Info(ctx, "SetActive - 设置LLM Model启用状态")

	parsedID, xErr := parseSnowflakeID(ctx, id)
	if xErr != nil {
		return xErr
	}

	model, xErr := l.repo.model.GetByID(ctx, parsedID)
	if xErr != nil {
		return xErr
	}

	model.IsActive = active
	if xErr := l.repo.model.Update(ctx, model); xErr != nil {
		return xErr
	}

	l.log.Info(ctx, "SetActive - LLM Model启用状态设置成功")
	return nil
}

// GetAgentModelConfig 根据 Agent 角色读取配置的 Model + Provider
//
// 从 Info 表读取 llm_agent_model:{role} 的值，解析 model_id 后查询关联的 Model 和 Provider。
// 未配置时返回 NotFound 错误。
func (l *LlmModelLogic) GetAgentModelConfig(ctx context.Context, role string) (*AgentModelConfig, *xError.Error) {
	l.log.Info(ctx, "GetAgentModelConfig - 获取Agent模型配置 ["+role+"]")

	key := bConst.LlmAgentModelKeyPrefix + role
	modelIDStr, xErr := l.repo.info.GetByKey(ctx, key)
	if xErr != nil {
		return nil, xError.NewError(ctx, xError.NotFound, "未配置 Agent 模型", false, nil)
	}

	if modelIDStr == "" {
		return nil, xError.NewError(ctx, xError.NotFound, "未配置 Agent 模型", false, nil)
	}

	modelIDInt, err := strconv.ParseInt(modelIDStr, 10, 64)
	if err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "Agent 模型配置值无效", false, err)
	}
	modelID := xSnowflake.SnowflakeID(modelIDInt)

	model, xErr := l.repo.model.GetByID(ctx, modelID)
	if xErr != nil {
		return nil, xErr
	}

	provider, xErr := l.repo.provider.GetByID(ctx, model.ProviderID)
	if xErr != nil {
		return nil, xErr
	}

	return &AgentModelConfig{
		Model:    model,
		Provider: provider,
	}, nil
}

// SetAgentModel 设置 Agent 角色对应的模型（先验证 Model 存在，再 UpsertValue）
func (l *LlmModelLogic) SetAgentModel(ctx context.Context, role, modelID string) *xError.Error {
	l.log.Info(ctx, "SetAgentModel - 设置Agent模型 ["+role+"]")

	parsedID, xErr := parseSnowflakeID(ctx, modelID)
	if xErr != nil {
		return xErr
	}

	// 验证 Model 存在
	if _, xErr := l.repo.model.GetByID(ctx, parsedID); xErr != nil {
		return xErr
	}

	key := bConst.LlmAgentModelKeyPrefix + role
	return l.repo.info.UpsertValue(ctx, key, modelID)
}

// toDetailResponse 将实体映射为详情响应
func (l *LlmModelLogic) toDetailResponse(model *entity.LlmModel) *apiLlm.ModelDetailResponse {
	return &apiLlm.ModelDetailResponse{
		ID:          model.BaseEntity.ID.String(),
		ProviderID:  model.ProviderID.String(),
		ModelName:   model.ModelName,
		DisplayName: model.DisplayName,
		MaxTokens:   model.MaxTokens,
		Temperature: model.Temperature,
		IsActive:    model.IsActive,
		Description: model.Description,
		CreatedAt:   model.BaseEntity.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		UpdatedAt:   model.BaseEntity.UpdatedAt.Format("2006-01-02T15:04:05-07:00"),
	}
}

// toListListItem 将实体映射为列表项
func (l *LlmModelLogic) toListListItem(model *entity.LlmModel) apiLlm.ModelListItem {
	return apiLlm.ModelListItem{
		ID:          model.BaseEntity.ID.String(),
		ProviderID:  model.ProviderID.String(),
		ModelName:   model.ModelName,
		DisplayName: model.DisplayName,
		MaxTokens:   model.MaxTokens,
		Temperature: model.Temperature,
		IsActive:    model.IsActive,
		Description: model.Description,
		CreatedAt:   model.BaseEntity.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
	}
}
