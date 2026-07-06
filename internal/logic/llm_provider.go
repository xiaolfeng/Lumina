package logic

import (
	"context"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	apiLlm "github.com/xiaolfeng/Lumina/api/llm"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// llmProviderRepo LLM Provider 模块依赖的仓储集合（不持有 InfoRepo，Agent 分配由 LlmModelLogic 管理）
type llmProviderRepo struct {
	provider *repository.LlmProviderRepo
	model    *repository.LlmModelRepo
}

// LlmProviderLogic LLM Provider 业务逻辑层
type LlmProviderLogic struct {
	logic
	repo llmProviderRepo
}

// NewLlmProviderLogic 创建 LLM Provider 业务逻辑层实例
func NewLlmProviderLogic(ctx context.Context) *LlmProviderLogic {
	db := xCtxUtil.MustGetDB(ctx)

	return &LlmProviderLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "LlmProviderLogic"),
		},
		repo: llmProviderRepo{
			provider: repository.NewLlmProviderRepo(db),
			model:    repository.NewLlmModelRepo(db),
		},
	}
}

// Create 创建 LLM Provider，加密 APIKey 后存库
func (l *LlmProviderLogic) Create(ctx context.Context, req *apiLlm.CreateProviderRequest) (*apiLlm.ProviderDetailResponse, *xError.Error) {
	l.log.Info(ctx, "Create - 创建LLM Provider")

	secret := xEnv.GetEnvString("LLM_ENCRYPT_SECRET", "")
	encrypted, xErr := service.EncryptAPIKey(req.APIKey, secret)
	if xErr != nil {
		return nil, xErr
	}

	provider := &entity.LlmProvider{
		Name:            req.Name,
		Protocol:        req.Protocol,
		BaseURL:         req.BaseURL,
		APIKeyEncrypted: encrypted,
		IsActive:        true,
		Description:     req.Description,
	}

	if xErr := l.repo.provider.Create(ctx, provider); xErr != nil {
		return nil, xErr
	}

	l.log.Info(ctx, "Create - LLM Provider创建成功")

	return l.toDetailResponse(provider), nil
}

// List 分页获取 LLM Provider 列表
func (l *LlmProviderLogic) List(ctx context.Context, page, size int) (*xModels.PageResponse[[]apiLlm.ProviderListItem], *xError.Error) {
	l.log.Info(ctx, "List - 分页获取LLM Provider列表")

	items, total, xErr := l.repo.provider.List(ctx, page, size)
	if xErr != nil {
		return nil, xErr
	}

	listItems := make([]apiLlm.ProviderListItem, 0, len(items))
	for _, item := range items {
		listItems = append(listItems, apiLlm.ProviderListItem{
			ID:          item.BaseEntity.ID.String(),
			Name:        item.Name,
			Protocol:    item.Protocol,
			BaseURL:     item.BaseURL,
			HasKey:      item.APIKeyEncrypted != "",
			IsActive:    item.IsActive,
			Description: item.Description,
			CreatedAt:   item.BaseEntity.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		})
	}

	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	return xModels.NewPageFromRequest(pageReq, total, listItems), nil
}

// GetByID 根据 ID 获取 LLM Provider 详情（不返回 APIKey 明文）
func (l *LlmProviderLogic) GetByID(ctx context.Context, id string) (*apiLlm.ProviderDetailResponse, *xError.Error) {
	l.log.Info(ctx, "GetByID - 根据ID获取LLM Provider详情")

	parsedID, xErr := parseSnowflakeID(ctx, id)
	if xErr != nil {
		return nil, xErr
	}

	provider, xErr := l.repo.provider.GetByID(ctx, parsedID)
	if xErr != nil {
		return nil, xErr
	}

	return l.toDetailResponse(provider), nil
}

// Update 更新 LLM Provider 信息（APIKey 为空时不更新加密字段）
func (l *LlmProviderLogic) Update(ctx context.Context, id string, req *apiLlm.UpdateProviderRequest) *xError.Error {
	l.log.Info(ctx, "Update - 更新LLM Provider")

	parsedID, xErr := parseSnowflakeID(ctx, id)
	if xErr != nil {
		return xErr
	}

	provider, xErr := l.repo.provider.GetByID(ctx, parsedID)
	if xErr != nil {
		return xErr
	}

	if req.Name != nil {
		provider.Name = *req.Name
	}
	if req.Protocol != nil {
		provider.Protocol = *req.Protocol
	}
	if req.BaseURL != nil {
		provider.BaseURL = *req.BaseURL
	}
	if req.IsActive != nil {
		provider.IsActive = *req.IsActive
	}
	if req.Description != nil {
		provider.Description = *req.Description
	}

	// APIKey 非空时重新加密
	if req.APIKey != nil && *req.APIKey != "" {
		secret := xEnv.GetEnvString("LLM_ENCRYPT_SECRET", "")
		encrypted, xErr := service.EncryptAPIKey(*req.APIKey, secret)
		if xErr != nil {
			return xErr
		}
		provider.APIKeyEncrypted = encrypted
	}

	if xErr := l.repo.provider.Update(ctx, provider); xErr != nil {
		return xErr
	}

	l.log.Info(ctx, "Update - LLM Provider更新成功")
	return nil
}

// Delete 删除 LLM Provider（有关联 Model 时拒绝）
func (l *LlmProviderLogic) Delete(ctx context.Context, id string) *xError.Error {
	l.log.Info(ctx, "Delete - 删除LLM Provider")

	parsedID, xErr := parseSnowflakeID(ctx, id)
	if xErr != nil {
		return xErr
	}

	// 校验关联 Model 存在性
	models, xErr := l.repo.model.ListByProviderID(ctx, parsedID)
	if xErr != nil {
		return xErr
	}
	if len(models) > 0 {
		return xError.NewError(ctx, xError.ValidationError, "该 Provider 下存在关联模型，无法删除", false, nil)
	}

	if xErr := l.repo.provider.Delete(ctx, parsedID); xErr != nil {
		return xErr
	}

	l.log.Info(ctx, "Delete - LLM Provider删除成功")
	return nil
}

// GetDecryptedAPIKey 解密返回 Provider 的 APIKey 明文，供 LLM 消费层调用
func (l *LlmProviderLogic) GetDecryptedAPIKey(ctx context.Context, providerID xSnowflake.SnowflakeID) (string, *xError.Error) {
	l.log.Info(ctx, "GetDecryptedAPIKey - 解密Provider APIKey")

	provider, xErr := l.repo.provider.GetByID(ctx, providerID)
	if xErr != nil {
		return "", xErr
	}

	if provider.APIKeyEncrypted == "" {
		return "", xError.NewError(ctx, xError.ValidationError, "该 Provider 未配置 APIKey", false, nil)
	}

	secret := xEnv.GetEnvString("LLM_ENCRYPT_SECRET", "")
	return service.DecryptAPIKey(provider.APIKeyEncrypted, secret)
}

// toDetailResponse 将实体映射为详情响应（不返回 APIKey 明文）
func (l *LlmProviderLogic) toDetailResponse(provider *entity.LlmProvider) *apiLlm.ProviderDetailResponse {
	return &apiLlm.ProviderDetailResponse{
		ID:          provider.BaseEntity.ID.String(),
		Name:        provider.Name,
		Protocol:    provider.Protocol,
		BaseURL:     provider.BaseURL,
		HasKey:      provider.APIKeyEncrypted != "",
		IsActive:    provider.IsActive,
		Description: provider.Description,
		CreatedAt:   provider.BaseEntity.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		UpdatedAt:   provider.BaseEntity.UpdatedAt.Format("2006-01-02T15:04:05-07:00"),
	}
}
