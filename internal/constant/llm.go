package bConst

// Agent 角色常量（当前仅 repowiki，Memory 预留但不定义常量避免 scope creep）
//
// 角色名统一以 AgentRoleRepoWikiPrefix 为前缀（点号分隔），
// 与 resources/prompts/{role}.md 内嵌文件经 service/prompt_loader.go 对应，
// 禁止散落裸字符串字面量。
const (
	AgentRoleRepoWikiPrefix      = "repowiki."                             // RepoWiki 角色统一前缀
	AgentRoleRepoWikiCoordinator = AgentRoleRepoWikiPrefix + "coordinator" // RepoWiki 主控 Agent（编排决策）
	AgentRoleRepoWikiExplore     = AgentRoleRepoWikiPrefix + "explore"     // RepoWiki 探索 Agent（读代码）
	AgentRoleRepoWikiWrite       = AgentRoleRepoWikiPrefix + "write"       // RepoWiki 写作 Agent（写文档）
	AgentRoleRepoWikiArchitect   = AgentRoleRepoWikiPrefix + "architect"   // RepoWiki 架构 Agent（架构梳理）
	AgentRoleRepoWikiValidator   = AgentRoleRepoWikiPrefix + "validator"   // RepoWiki 校验 Agent（校验审阅）
)

// AgentRolesRepoWiki RepoWiki 模块的子 Agent 角色列表
var AgentRolesRepoWiki = []string{
	AgentRoleRepoWikiCoordinator,
	AgentRoleRepoWikiExplore,
	AgentRoleRepoWikiWrite,
	AgentRoleRepoWikiArchitect,
	AgentRoleRepoWikiValidator,
}

// Info 表中 Agent → Model 映射的键前缀
const (
	LlmAgentModelKeyPrefix = "llm.agent." // Agent 模型分配键前缀
)

// LLM Provider 协议类型
const (
	LlmProviderProtocolOpenAI    = "openai"    // OpenAI 协议
	LlmProviderProtocolAnthropic = "anthropic" // Anthropic 协议
)
