// Package service 提供跨业务领域的通用服务（LLM Provider、下载令牌、文件缓存等）。
package service

import (
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

// NewLLMProviderFromEntity 根据数据库实体参数创建 LLM Provider 客户端
//
// 参数说明:
//   - protocol:         协议类型（openai / anthropic）
//   - baseURL:          自定义 API 端点（空字符串表示使用默认端点）
//   - decryptedAPIKey:  已解密的 API Key
//
// 返回值:
//   - bamboo.BambooClient: 对底层协议适配器的统一封装
//   - error:               目前不会返回错误（保留 error 返回值以保持调用方一致性）
func NewLLMProviderFromEntity(protocol, baseURL, decryptedAPIKey string) (bamboo.BambooClient, error) {
	var p provider.Provider
	switch protocol {
	case llmProviderAnthropic:
		opts := []anthropic.Option{anthropic.WithAPIKey(decryptedAPIKey)}
		if baseURL != "" {
			opts = append(opts, anthropic.WithBaseURL(baseURL))
		}
		p = anthropic.NewProviderWithOptions(opts...)
	case llmProviderOpenAI:
		fallthrough
	default:
		opts := []completions.Option{completions.WithAPIKey(decryptedAPIKey)}
		if baseURL != "" {
			opts = append(opts, completions.WithBaseURL(baseURL))
		}
		p = completions.NewCompletionsProviderWithOptions(opts...)
	}

	return bamboo.NewClient(p), nil
}
