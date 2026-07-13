// Package service 提供跨业务领域的通用服务（LLM Provider、Agent 工厂、下载令牌等）。
package service

import (
	"context"
	"fmt"
	"strconv"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"

	"github.com/xiaolfeng/Lumina/internal/repository"
)

// ResolvedLlmConfig 解析后的 LLM 配置（供消费层使用）
//
// 由 LlmResolver.ResolveAgentModel 产出，包含构建 LLM 客户端和 Agent
// 所需的全部参数（协议、端点、解密后的 APIKey、模型参数）。
type ResolvedLlmConfig struct {
	Protocol        string  // 协议类型（openai / anthropic）
	BaseURL         string  // 自定义 API 端点
	DecryptedAPIKey string  // 解密后的 API Key
	ModelName       string  // 模型标识（如 gpt-4o）
	MaxTokens       int64   // 单次响应最大输出 Token 数
	ContextWindow   int64   // 模型上下文窗口大小
	Temperature     float64 // 生成温度
}

// LlmResolver LLM 配置解析器（持有 Repository 引用，避免循环导入）
//
// 在 service 层完成 Info 表 → Model → Provider → 解密 APIKey 的完整解析链路，
// 供 RepoWikiLogic 在 AnalyzeRepo 时调用。
//
// 设计要点：
//   - 仅持有 repository 引用（不持有 logic 引用），避免 service → logic 循环导入
//   - 返回 error（非 *xError.Error），因为 service 层不使用 xError 类型
//   - 解密使用 DecryptAPIKey（同包函数），密钥来自 LLM_ENCRYPT_SECRET 环境变量
type LlmResolver struct {
	modelRepo     *repository.LlmModelRepo
	providerRepo  *repository.LlmProviderRepo
	infoRepo      *repository.InfoRepo
	encryptSecret string
}

// NewLlmResolver 创建 LlmResolver 实例
func NewLlmResolver(
	modelRepo *repository.LlmModelRepo,
	providerRepo *repository.LlmProviderRepo,
	infoRepo *repository.InfoRepo,
	encryptSecret string,
) *LlmResolver {
	return &LlmResolver{
		modelRepo:     modelRepo,
		providerRepo:  providerRepo,
		infoRepo:      infoRepo,
		encryptSecret: encryptSecret,
	}
}

// ResolveAgentModel 解析 Agent 角色对应的 LLM 配置
//
// 解析流程:
//  1. 从 Info 表读取 llm_agent_model:{role} 的值（model_id）
//  2. 查 Model 实体获取模型参数 + ProviderID
//  3. 查 Provider 实体获取协议、端点、加密的 APIKey
//  4. 解密 APIKey
//  5. 组装 ResolvedLlmConfig 返回
//
// 参数说明:
//   - ctx:       上下文
//   - role:      Agent 角色标识（如 "repowiki"）
//   - keyPrefix: Info 表键前缀（如 "llm_agent_model:"）
//
// 返回值:
//   - *ResolvedLlmConfig: 解析后的 LLM 配置
//   - error: 任意步骤失败时返回错误（含可读上下文）
func (r *LlmResolver) ResolveAgentModel(ctx context.Context, role string, keyPrefix string) (*ResolvedLlmConfig, error) {
	// 1. 从 Info 表读取 model_id
	key := keyPrefix + role
	modelIDStr, xErr := r.infoRepo.GetByKey(ctx, key)
	if xErr != nil {
		return nil, fmt.Errorf("未配置 Agent 模型: %w", xErr)
	}
	if modelIDStr == "" {
		return nil, fmt.Errorf("未配置 Agent 模型")
	}

	// 2. 解析 model_id
	modelIDInt, err := strconv.ParseInt(modelIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Agent 模型配置值无效: %w", err)
	}
	modelID := xSnowflake.SnowflakeID(modelIDInt)

	// 3. 查 Model
	model, xErr := r.modelRepo.GetByID(ctx, modelID)
	if xErr != nil {
		return nil, fmt.Errorf("查询模型失败: %w", xErr)
	}

	// 4. 查 Provider
	provider, xErr := r.providerRepo.GetByID(ctx, model.ProviderID)
	if xErr != nil {
		return nil, fmt.Errorf("查询 Provider 失败: %w", xErr)
	}

	// 5. 解密 APIKey
	decryptedKey, xErr := DecryptAPIKey(provider.APIKeyEncrypted, r.encryptSecret)
	if xErr != nil {
		return nil, fmt.Errorf("解密 APIKey 失败: %w", xErr)
	}

	return &ResolvedLlmConfig{
		Protocol:        provider.Protocol,
		BaseURL:         provider.BaseURL,
		DecryptedAPIKey: decryptedKey,
		ModelName:       model.ModelName,
		MaxTokens:       model.MaxTokens,
		ContextWindow:   model.ContextWindow,
		Temperature:     model.Temperature,
	}, nil
}

// ResolveAgentModels 批量解析多个角色的 LLM 配置
//
// 返回 map[role]*ResolvedLlmConfig，缺失角色不出现在 map 中（不报错）。
// keyPrefix 如 "llm_agent_model:"，与单角色方法一致。
func (r *LlmResolver) ResolveAgentModels(
	ctx context.Context,
	roles []string,
	keyPrefix string,
) (map[string]*ResolvedLlmConfig, error) {
	result := make(map[string]*ResolvedLlmConfig, len(roles))
	for _, role := range roles {
		config, err := r.ResolveAgentModel(ctx, role, keyPrefix)
		if err != nil {
			// 单个角色解析失败不中断批量，跳过该角色
			continue
		}
		result[role] = config
	}
	return result, nil
}
