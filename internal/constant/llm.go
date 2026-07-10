package bConst

// Agent 角色常量（当前仅 repowiki，Memory 预留但不定义常量避免 scope creep）
const (
	AgentRoleRepoWiki            = "repowiki"             // RepoWiki 模块别名（兼容旧逻辑）
	AgentRoleRepoWikiCoordinator = "repowiki:coordinator" // RepoWiki 主控 Agent（编排决策）
	AgentRoleRepoWikiExplore     = "repowiki:explore"     // RepoWiki 探索 Agent（读代码）
	AgentRoleRepoWikiWrite       = "repowiki:write"       // RepoWiki 写作 Agent（写文档）
)

// AgentRolesRepoWiki RepoWiki 模块的子 Agent 角色列表
var AgentRolesRepoWiki = []string{
	AgentRoleRepoWikiCoordinator,
	AgentRoleRepoWikiExplore,
	AgentRoleRepoWikiWrite,
}

// Info 表中 Agent → Model 映射的键前缀
const (
	LlmAgentModelKeyPrefix = "llm_agent_model:" // Agent 模型分配键前缀
)

// LLM Provider 协议类型
const (
	LlmProviderProtocolOpenAI    = "openai"    // OpenAI 协议
	LlmProviderProtocolAnthropic = "anthropic" // Anthropic 协议
)
