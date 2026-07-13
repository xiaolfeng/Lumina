package bConst

import (
	_ "github.com/bamboo-services/bamboo-agent/agent"
	_ "github.com/bamboo-services/bamboo-messages/provider"
	_ "github.com/go-git/go-git/v5"
)

// ── RepoWiki 分析状态枚举 ──
const (
	RepoWikiStatusPending    = "pending"    // 等待执行
	RepoWikiStatusCloning    = "cloning"    // Git 克隆中
	RepoWikiStatusScanning   = "scanning"   // 文件扫描中
	RepoWikiStatusAnalyzing  = "analyzing"  // Agent 分析中
	RepoWikiStatusCompleted  = "completed"  // 分析完成
	RepoWikiStatusFailed     = "failed"     // 分析失败
	RepoWikiStatusCancelled  = "cancelled"  // 已取消
)

// ── RepoWiki 分析阶段枚举（current_stage 字段用）──
const (
	RepoWikiStageScan          = "scan"          // 文件扫描阶段
	RepoWikiStageExploring     = "exploring"     // 代码探索阶段
	RepoWikiStageArchitecting  = "architecting"  // 架构设计阶段
	RepoWikiStageWriting       = "writing"       // 文档编写阶段
	RepoWikiStageValidating    = "validating"    // 结果校验阶段
)

// ── RepoWiki 默认值 ──
const (
	RepoWikiDefaultMaxFileSize = 1024 * 1024 // 默认最大文件大小 1MB
	RepoWikiDefaultLanguage    = "zh"        // 默认 Wiki 语言
	RepoWikiCookieMaxAge       = 2 * 60 * 60 // Wiki Cookie 有效期 2 小时（秒）
)

// ── RepoWiki Redis 缓存 Key ──
const (
	CacheRepoWikiConfigByID    RedisKey = "repowiki:config:%d"         // CacheRepoWikiConfigByID 配置 ID→详情缓存（%d = configID）
	CacheRepoWikiVersionStatus RedisKey = "repowiki:version:%d:status" // CacheRepoWikiVersionStatus 版本状态缓存（%d = versionID，TTL 30s 轮询优化）
)

// ── Webhook Provider 枚举 ──
const (
	WebhookProviderGitHub = "github" // GitHub
	WebhookProviderGitee  = "gitee"  // Gitee
	WebhookProviderGitLab = "gitlab" // GitLab
	WebhookProviderGitea  = "gitea"  // Gitea
)

// ── Webhook 事件状态枚举 ──
const (
	WebhookEventStatusReceived = "received" // 已接收
	WebhookEventStatusAccepted = "accepted" // 已接受
	WebhookEventStatusIgnored  = "ignored"  // 已忽略
	WebhookEventStatusFailed   = "failed"   // 已失败
)

// ── Webhook 请求头名称 ──
const (
	WebhookHeaderGitHub = "X-Hub-Signature-256" // GitHub Webhook签名头
	WebhookHeaderGitee  = "X-Gitee-Token"       // Gitee Webhook Token头
	WebhookHeaderGitLab = "X-Gitlab-Token"      // GitLab Webhook Token头
	WebhookHeaderGitea  = "X-Gitea-Signature"   // Gitea Webhook签名头
)

// ── RepoWiki 编排引擎超时/配额/重试常量 ──
const (
	RepoWikiPipelineTimeoutMin   = 60        // Pipeline 总超时（分钟）
	RepoWikiOverviewTimeoutMin   = 10        // Coordinator 概要阶段超时（分钟）
	RepoWikiExploreTimeoutMin    = 15        // Explore 每路超时（分钟）
	RepoWikiArchitectTimeoutMin  = 10        // Architect 阶段超时（分钟）
	RepoWikiWriterTimeoutMin     = 20        // Writer 每批超时（分钟）
	RepoWikiValidatorTimeoutMin  = 5         // Validator 阶段超时（分钟）
	RepoWikiExploreMaxConcurrent = 4         // Explore 最大并发数
	RepoWikiWriterMaxConcurrent  = 4         // Writer 最大并发数
	RepoWikiWriterMaxRetry       = 2         // Writer 失败最大重试次数
	RepoWikiMaxTokens            = 2000000   // 每个 version 总 token 上限
	RepoWikiMaxRepoSizeMB        = 2048      // 单仓库磁盘上限（MB）
	RepoWikiMaxVersions          = 10        // 每个 config 保留版本数上限
)
