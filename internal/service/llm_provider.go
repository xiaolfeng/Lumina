// Package service 提供跨业务领域的通用服务（LLM Provider、下载令牌、文件缓存等）。
package service

import (
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	"github.com/bamboo-services/bamboo-messages/bamboo"
	"github.com/bamboo-services/bamboo-messages/provider"
	"github.com/bamboo-services/bamboo-messages/provider/anthropic"
	"github.com/bamboo-services/bamboo-messages/provider/openai/completions"
)

const (
	// llmProviderOpenAI 是 OpenAI Chat Completions 协议标识。
	llmProviderOpenAI = "openai"
	// llmProviderAnthropic 是 Anthropic Messages 协议标识。
	llmProviderAnthropic = "anthropic"
)

// NewLLMProvider 根据环境变量创建 LLM Provider 客户端。
//
// 读取的环境变量：
//   - LLM_PROVIDER: 协议类型，支持 "openai"（默认）和 "anthropic"
//   - LLM_API_KEY: API 密钥（必填）
//   - LLM_BASE_URL: 自定义端点（可选）
//   - LLM_MODEL: 模型名称（默认 gpt-4o，在 Agent 配置中使用）
//   - LLM_MAX_TOKENS: 最大 token 数（默认 4096，在 Agent 配置中使用）
//   - LLM_TEMPERATURE: 生成温度（默认 0.3，在 Agent 配置中使用）
//
// 返回的 bamboo.BambooClient 是对底层协议适配器的统一封装，
// 可直接交给 bamboo-agent 的 Agent 作为对话客户端。
func NewLLMProvider() (bamboo.BambooClient, error) {
	providerType := xEnv.GetEnvString("LLM_PROVIDER", llmProviderOpenAI)
	apiKey := xEnv.GetEnvString("LLM_API_KEY", "")
	baseURL := xEnv.GetEnvString("LLM_BASE_URL", "")

	var p provider.Provider
	switch providerType {
	case llmProviderAnthropic:
		opts := []anthropic.Option{anthropic.WithAPIKey(apiKey)}
		if baseURL != "" {
			opts = append(opts, anthropic.WithBaseURL(baseURL))
		}
		p = anthropic.NewProviderWithOptions(opts...)
	case llmProviderOpenAI:
		fallthrough
	default:
		opts := []completions.Option{completions.WithAPIKey(apiKey)}
		if baseURL != "" {
			opts = append(opts, completions.WithBaseURL(baseURL))
		}
		p = completions.NewCompletionsProviderWithOptions(opts...)
	}

	return bamboo.NewClient(p), nil
}
