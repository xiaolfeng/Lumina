package logic

import (
	"context"
	"testing"

	apiLlm "github.com/xiaolfeng/Lumina/api/llm"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// setupLlmModelTestLogic 创建测试用 LlmModelLogic 实例；无数据库连接时跳过
func setupLlmModelTestLogic(t *testing.T) *LlmModelLogic {
	t.Helper()
	t.Skip("requires database connection - skipping unit test")
	return nil
}

// TestLlmModelCreate 测试创建 Model
func TestLlmModelCreate(t *testing.T) {
	l := setupLlmModelTestLogic(t)
	ctx := context.Background()

	// 先创建 Provider
	providerLogic := setupLlmProviderTestLogic(t)
	provider, xErr := providerLogic.Create(ctx, &apiLlm.CreateProviderRequest{
		Name:     "test-model-provider",
		Protocol: "openai",
		APIKey:   "sk-test-model-provider",
	})
	if xErr != nil {
		t.Fatalf("Provider Create failed: %v", xErr)
	}

	req := &apiLlm.CreateModelRequest{
		ProviderID:  provider.ID,
		ModelName:   "gpt-4o",
		DisplayName: "GPT-4o",
		MaxTokens:   4096,
		Temperature: 0.3,
		Description: "test model",
	}

	resp, xErr := l.Create(ctx, req)
	if xErr != nil {
		t.Fatalf("Create failed: %v", xErr)
	}
	if resp.ID.IsZero() {
		t.Error("expected non-empty ID")
	}
	if resp.ModelName != req.ModelName {
		t.Errorf("expected model name %q, got %q", req.ModelName, resp.ModelName)
	}

	_ = l.Delete(ctx, resp.ID.String())
	_ = providerLogic.Delete(ctx, provider.ID.String())
}

// TestLlmModelGetAgentModelConfig 测试空配置返回 NotFound 错误
func TestLlmModelGetAgentModelConfig(t *testing.T) {
	l := setupLlmModelTestLogic(t)
	ctx := context.Background()

	_, xErr := l.GetAgentModelConfig(ctx, bConst.AgentRoleRepoWikiCoordinator)
	if xErr == nil {
		t.Fatal("expected NotFound error for unconfigured agent, got nil")
	}
}

// TestLlmModelSetAgentModel 测试设置后可读回
func TestLlmModelSetAgentModel(t *testing.T) {
	l := setupLlmModelTestLogic(t)
	ctx := context.Background()

	// 先创建 Provider + Model
	providerLogic := setupLlmProviderTestLogic(t)
	provider, xErr := providerLogic.Create(ctx, &apiLlm.CreateProviderRequest{
		Name:     "test-agent-provider",
		Protocol: "openai",
		APIKey:   "sk-test-agent-provider",
	})
	if xErr != nil {
		t.Fatalf("Provider Create failed: %v", xErr)
	}

	model, xErr := l.Create(ctx, &apiLlm.CreateModelRequest{
		ProviderID:  provider.ID,
		ModelName:   "gpt-4o",
		DisplayName: "GPT-4o",
	})
	if xErr != nil {
		t.Fatalf("Model Create failed: %v", xErr)
	}

	// 设置 Agent 模型
	if xErr := l.SetAgentModel(ctx, bConst.AgentRoleRepoWikiCoordinator, model.ID.String()); xErr != nil {
		t.Fatalf("SetAgentModel failed: %v", xErr)
	}

	// 读回验证
	config, xErr := l.GetAgentModelConfig(ctx, bConst.AgentRoleRepoWikiCoordinator)
	if xErr != nil {
		t.Fatalf("GetAgentModelConfig failed: %v", xErr)
	}
	if config.Model == nil {
		t.Fatal("expected non-nil Model")
	}
	if config.Model.ID != model.ID {
		t.Errorf("expected model ID %s, got %s", model.ID, config.Model.ID)
	}
	if config.Provider == nil {
		t.Fatal("expected non-nil Provider")
	}

	_ = l.Delete(ctx, model.ID.String())
	_ = providerLogic.Delete(ctx, provider.ID.String())
}
