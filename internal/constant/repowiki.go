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
	RepoWikiStatusAssembling = "assembling" // 文档组装中
	RepoWikiStatusCompleted  = "completed"  // 分析完成
	RepoWikiStatusFailed     = "failed"     // 分析失败
	RepoWikiStatusCancelled  = "cancelled"  // 已取消
)

// ── RepoWiki 分析阶段枚举（current_stage 字段用）──
const (
	RepoWikiStageScan       = "scan"        // 文件扫描阶段
	RepoWikiStageDepExtract = "dep_extract" // 依赖提取阶段
	RepoWikiStagePass1      = "pass1"       // Pass 1: 项目概览
	RepoWikiStagePass2      = "pass2"       // Pass 2: 模块分析
	RepoWikiStagePass3      = "pass3"       // Pass 3: 架构设计
	RepoWikiStagePass4      = "pass4"       // Pass 4: 阅读指南
	RepoWikiStageAssemble   = "assemble"    // 文档组装
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
