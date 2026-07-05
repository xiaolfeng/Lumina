// Package service 提供跨业务领域的通用服务（LLM Provider、Agent 工厂、下载令牌等）。
package service

import (
	"context"

	"github.com/bamboo-services/bamboo-messages/provider"
)

// stubResponseJSON 是 StubLLMProvider 在同步调用中返回的固定 JSON 内容。
const stubResponseJSON = `{"status":"ok","message":"stub response"}`

// StubLLMProvider 是一个用于单元测试的 LLM Provider 桩实现。
//
// 不调用真实 LLM 服务，所有方法都返回固定 JSON 内容，便于在测试环境中
// 验证 Agent 编排逻辑而不消耗实际 API 配额。
type StubLLMProvider struct {
	// ProviderType 返回的提供商类型，默认为 OpenAI Completions。
	ProviderType provider.ProviderType
	// ResponseJSON 同步调用返回的内容，为空时使用默认 stubResponseJSON。
	ResponseJSON string
}

// responseContent 返回实际使用的响应内容。
func (s *StubLLMProvider) responseContent() string {
	if s.ResponseJSON != "" {
		return s.ResponseJSON
	}
	return stubResponseJSON
}

// providerType 返回实际使用的 ProviderType。
func (s *StubLLMProvider) providerType() provider.ProviderType {
	if s.ProviderType != "" {
		return s.ProviderType
	}
	return provider.ProviderOpenAICompletions
}

// Chat 实现 provider.Provider 接口，返回固定 JSON 作为流式事件。
func (s *StubLLMProvider) Chat(ctx context.Context, messages []provider.Message, config *provider.ChatConfig) <-chan provider.StreamEvent {
	return s.chatWithSystem(ctx, "", messages, config)
}

// ChatWithSystem 实现 provider.Provider 接口，返回固定 JSON 作为流式事件。
func (s *StubLLMProvider) ChatWithSystem(ctx context.Context, systemPrompt string, messages []provider.Message, config *provider.ChatConfig) <-chan provider.StreamEvent {
	return s.chatWithSystem(ctx, systemPrompt, messages, config)
}

// chatWithSystem 是 Chat 和 ChatWithSystem 的公共实现。
func (s *StubLLMProvider) chatWithSystem(_ context.Context, _ string, _ []provider.Message, _ *provider.ChatConfig) <-chan provider.StreamEvent {
	ch := make(chan provider.StreamEvent, 3)
	ch <- provider.StreamEvent{Type: provider.StreamTypeStart}
	ch <- provider.StreamEvent{
		Type: provider.StreamTypeDelta,
		Delta: provider.StreamDelta[any]{
			Type: provider.StreamDeltaTypeTextOutput,
			Data: provider.TextData(s.responseContent()),
		},
	}
	ch <- provider.StreamEvent{
		Type:         provider.StreamTypeStop,
		FinishReason: provider.FinishReasonStop,
	}
	close(ch)
	return ch
}

// Complete 实现 provider.Provider 接口，返回固定 JSON 作为完整响应。
func (s *StubLLMProvider) Complete(ctx context.Context, messages []provider.Message, config *provider.ChatConfig) (*provider.CompletionResult, error) {
	return s.completeWithSystem(ctx, "", messages, config)
}

// CompleteWithSystem 实现 provider.Provider 接口，返回固定 JSON 作为完整响应。
func (s *StubLLMProvider) CompleteWithSystem(ctx context.Context, systemPrompt string, messages []provider.Message, config *provider.ChatConfig) (*provider.CompletionResult, error) {
	return s.completeWithSystem(ctx, systemPrompt, messages, config)
}

// completeWithSystem 是 Complete 和 CompleteWithSystem 的公共实现。
func (s *StubLLMProvider) completeWithSystem(_ context.Context, _ string, _ []provider.Message, _ *provider.ChatConfig) (*provider.CompletionResult, error) {
	return &provider.CompletionResult{
		Content:      s.responseContent(),
		FinishReason: provider.FinishReasonStop,
	}, nil
}

// GetProviderType 返回当前 Provider 的类型标识。
func (s *StubLLMProvider) GetProviderType() provider.ProviderType {
	return s.providerType()
}

// GetAvailableModels 返回可用模型列表（桩实现，返回固定的测试模型名）。
func (s *StubLLMProvider) GetAvailableModels() []string {
	return []string{"stub-model"}
}
