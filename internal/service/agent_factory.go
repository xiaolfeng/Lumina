// Package service 提供跨业务领域的通用服务（LLM Provider、Agent 工厂、下载令牌等）。
package service

import (
	"github.com/bamboo-services/bamboo-agent/agent"
	"github.com/bamboo-services/bamboo-agent/tool"
	"github.com/bamboo-services/bamboo-messages/bamboo"
)

const (
	// repoWikiAgentMaxIterations 是 RepoWiki Agent 的最大 ReAct 迭代次数。
	repoWikiAgentMaxIterations = 30
	// repoWikiAgentMaxConcurrentTools 是 RepoWiki Agent 并发执行工具的最大数量。
	repoWikiAgentMaxConcurrentTools = 5
)

// roleMaxIterations 不同 Agent 角色的最大 ReAct 迭代次数
var roleMaxIterations = map[string]int{
	"repowiki:coordinator": 50, // 调度需要更多轮
	"repowiki:explore":     40, // 探索需要较多轮
	"repowiki:architect":   30, // 架构分析中等
	"repowiki:write":       40, // 写作需要较多轮
	"repowiki:validator":   20, // 校验较少轮
}

// NewRepoWikiAgentFromModel 构建用于分析代码库并生成 Wiki 的 Agent（使用数据库模型参数）
//
// 参数说明:
//   - client:      已配置好的 LLM 客户端（由 NewLLMProviderFromEntity 生成）
//   - modelName:   模型标识（如 gpt-4o）
//   - maxTokens:     单次响应最大输出 Token 数
//   - contextWindow: 模型上下文窗口大小（用于触发上下文压缩的阈值）
//   - temperature:   生成温度
//   - systemPrompt:  系统提示词，用于设定 Agent 的分析角色
//   - tools:         只读分析工具集（禁止包含 shell 等可写工具）
//   - workDir:       会话持久化目录，FileStore 会在此目录下保存会话消息
//
// 返回值:
//   - agent.Agent: 配置好的 Agent 实例
//   - error:       创建 FileStore 失败时返回错误
func NewRepoWikiAgentFromModel(
	client bamboo.BambooClient,
	modelName string,
	maxTokens int64,
	contextWindow int64,
	temperature float64,
	systemPrompt string,
	tools []tool.Tool,
	workDir string,
) (agent.Agent, error) {
	fileStore, err := agent.NewFileStore(workDir)
	if err != nil {
		return nil, err
	}

	temp := temperature
	config := agent.AgentConfig{
		Model:              modelName,
		MaxTokens:          maxTokens,
		MaxContextTokens:   contextWindow,
		Temperature:        &temp,
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

// NewRepoWikiSubAgent 构建用于 RepoWiki 编排引擎的 SubAgent（使用数据库模型参数）
//
// 与 NewRepoWikiAgentFromModel 签名相同，但根据 role 返回不同的 MaxIterations。
// 参数 role 用于查 roleMaxIterations map 决定迭代上限。
func NewRepoWikiSubAgent(
	client bamboo.BambooClient,
	role string,
	modelName string,
	maxTokens int64,
	contextWindow int64,
	temperature float64,
	systemPrompt string,
	tools []tool.Tool,
	workDir string,
) (agent.Agent, error) {
	fileStore, err := agent.NewFileStore(workDir)
	if err != nil {
		return nil, err
	}

	maxIter := roleMaxIterations[role]
	if maxIter == 0 {
		maxIter = repoWikiAgentMaxIterations // 默认值 30
	}

	temp := temperature
	config := agent.AgentConfig{
		Model:              modelName,
		MaxTokens:          maxTokens,
		MaxContextTokens:   contextWindow,
		Temperature:        &temp,
		SystemPrompt:       systemPrompt,
		MaxIterations:      maxIter,
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
