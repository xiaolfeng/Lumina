// Package service 提供跨业务领域的通用服务（LLM Provider、Agent 工厂、下载令牌等）。
package service

import (
	"github.com/bamboo-services/bamboo-agent/agent"
	"github.com/bamboo-services/bamboo-agent/tool"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	"github.com/bamboo-services/bamboo-messages/bamboo"
)

const (
	// repoWikiAgentMaxIterations 是 RepoWiki Agent 的最大 ReAct 迭代次数。
	repoWikiAgentMaxIterations = 30
	// repoWikiAgentMaxTokens 是 RepoWiki Agent 单次调用的默认最大 token 数。
	repoWikiAgentMaxTokens = 8192
	// repoWikiAgentMaxConcurrentTools 是 RepoWiki Agent 并发执行工具的最大数量。
	repoWikiAgentMaxConcurrentTools = 5
)

// NewRepoWikiAgent 构建用于分析代码库并生成 Wiki 的 Agent。
//
// 参数说明:
//   - client: 已配置好的 LLM 客户端（由 NewLLMProvider 生成）
//   - systemPrompt: 系统提示词，用于设定 Agent 的分析角色
//   - tools: 只读分析工具集（禁止包含 shell 等可写工具）
//   - workDir: 会话持久化目录，FileStore 会在此目录下保存会话消息
//
// 配置读取:
//   - LLM_MODEL: 模型名称（默认 gpt-4o）
//   - LLM_MAX_TOKENS: 最大 token 数（默认 8192）
//   - LLM_TEMPERATURE: 生成温度（默认 0.3）
//
// 返回:
//   - agent.Agent: 配置好的 Agent 实例
//   - error: 创建 FileStore 失败时返回错误
func NewRepoWikiAgent(
	client bamboo.BambooClient,
	systemPrompt string,
	tools []tool.Tool,
	workDir string,
) (agent.Agent, error) {
	fileStore, err := agent.NewFileStore(workDir)
	if err != nil {
		return nil, err
	}

	model := xEnv.GetEnvString("LLM_MODEL", "gpt-4o")
	maxTokens := int64(xEnv.GetEnvInt("LLM_MAX_TOKENS", repoWikiAgentMaxTokens))
	temperature := xEnv.GetEnvFloat("LLM_TEMPERATURE", 0.3)

	config := agent.AgentConfig{
		Model:              model,
		MaxTokens:          maxTokens,
		Temperature:        &temperature,
		SystemPrompt:       systemPrompt,
		MaxIterations:      repoWikiAgentMaxIterations,
		MaxConcurrentTools: repoWikiAgentMaxConcurrentTools,
	}

	ag := agent.NewAgentWithOptions(
		client,
		agent.WithConfig(config),
		agent.WithTools(tools...),
		agent.WithSessionStore(fileStore),
	)

	return ag, nil
}
