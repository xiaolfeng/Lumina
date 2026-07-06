package bConst

// Agent 角色常量（当前仅 repowiki，Memory 预留但不定义常量避免 scope creep）
const (
	AgentRoleRepoWiki = "repowiki" // RepoWiki 分析 Agent 角色
)

// Info 表中 Agent → Model 映射的键前缀
const (
	LlmAgentModelKeyPrefix = "llm_agent_model:" // Agent 模型分配键前缀
)

// LLM Provider 协议类型
const (
	LlmProviderProtocolOpenAI    = "openai"    // OpenAI 协议
	LlmProviderProtocolAnthropic = "anthropic" // Anthropic 协议
)
