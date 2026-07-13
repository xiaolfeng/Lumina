# RepoWiki 多 Agent 协作设计

> **状态**：已过时（superseded by 5-role SubAgent orchestration）  
> **日期**：2026-07-09  
> **替代**：本文档描述的旧版 3 角色 ReAct Coordinator 设计已废弃，RepoWiki 当前采用预定义 5 阶段 SubAgent 编排。详见 `docs/wiki/repowiki/design.md`。

---

# RepoWiki 5 角色 SubAgent 编排设计

> **状态**：当前实现  
> **日期**：2026-07-14  
> **替代**：`docs/wiki/repowiki/design.md` 中旧版"4 Pass 串行流程"与"3 角色 ReAct Coordinator"设计

## 背景与问题

### 旧设计缺陷

RepoWiki 经历了三代设计演进：

1. **固定 4 Pass 串行流程**：概览 → 模块 → 架构 → 指南。阶段硬编码，无法根据仓库规模动态调整。
2. **3 角色 ReAct Coordinator**：Coordinator 通过 Agent-as-Tool 调用 Explore/Write，自主决定探索策略。但 Coordinator 自主决策导致输出不稳定，Token 消耗不可控。
3. **单角色模型配置**：Info 表键 `llm_agent_model:repowiki` 仅绑定一个 model_id，无法为不同职责分配合适模型。

### 目标

将 RepoWiki 升级为**预定义 5 阶段 SubAgent 编排**：

- 阶段固定，每个阶段职责明确
- 5 个角色各自使用独立模型配置
- 并发阶段（Explore/Writer）使用信号量限流
- 失败可局部重试（Architect JSON 解析重试、Writer 失败重驱动）
- 存储按版本隔离，支持多版本共存和切换

## 技术基础

bamboo-agent 框架（`github.com/bamboo-services/bamboo-agent`）提供 Agent 运行能力，但编排不再依赖框架的 Agent-as-Tool 或 ReAct 循环，而是由 Lumina 内部的 `SubAgentOrchestrator` 直接控制阶段流转。

每个子 Agent 仍使用 bamboo-agent 的 `agent.Agent.Run()` 执行，但调用时机和参数由 Orchestrator 决定。

## 角色定义

| 角色 | 标识 | 职责 | 工具集 | 模型特点 |
|------|------|------|--------|----------|
| **Coordinator** | `repowiki:coordinator` | 项目概要分析 | `file_read`, `file_search`, `list_dir` | 上下文窗口大，理解力强 |
| **Explore** | `repowiki:explore` | 代码探索 | `file_read`, `file_search` | 快速、低成本，并发执行 |
| **Architect** | `repowiki:architect` | 架构规划 | `file_read` | 结构化输出能力强 |
| **Write** | `repowiki:write` | 文档写作 | `file_read`, `save_wiki_page` | 写作能力强，懂 Markdown |
| **Validator** | `repowiki:validator` | 文档校验 | `file_read`, `file_search`, `save_wiki_page` | 严谨，能输出结构化 JSON |

## 编排流程

```
SubAgentOrchestrator.Execute
  │
  ├─▶ runOverview (Coordinator)
  │     产出 overview.md
  │
  ├─▶ runExploreConcurrent (Explore × 4)
  │     产出 versions/{vid}/explore/*.xml
  │
  ├─▶ runArchitect (Architect)
  │     产出 architecture.json
  │
  ├─▶ runWritersConcurrent (Writer × 4)
  │     产出 versions/{vid}/wiki/*.md
  │
  └─▶ runValidator (Validator)
        校验 → 失败则重驱动 Writer（最多 2 次）
```

### 阶段 1：Coordinator 概要分析

- 构建 Coordinator Agent，作用域为仓库根目录
- 使用 `file_read` / `file_search` / `list_dir` 阅读 README、清单文件、入口文件
- 产出自由 Markdown 格式的项目概要
- 写入 `versions/{vid}/overview.md`

### 阶段 2：Explore 并发探索

- 从 `overview.md` 正则提取顶层目录路径作为分析 scope
- 若提取失败，回退到仓库根目录的顶层子目录列表
- 最多 `exploreMaxScopes = 8` 个 scope，每路独立超时
- 并发数由 `RepoWikiExploreMaxConcurrent` 控制（默认 4）
- 失败比例超过 50% 则整体失败
- 每个 scope 产出 XML 骨架，写入 `versions/{vid}/explore/{scope}.xml`

### 阶段 3：Architect 架构规划

- 读取 `overview.md` 和全部 Explore 产出
- 输出 JSON 数组 `[]WikiEntry`，字段：`title` / `path` / `description` / `explore_refs` / `complexity`
- JSON 解析失败时自动重试 2 次，每次追加格式提醒
- 写入 `versions/{vid}/architecture.json`

### 阶段 4：Writer 并发撰写

- 按 `complexity` 分组：
  - `high`：每组 1 个 WikiEntry
  - `medium` / `low`：每组最多 2 个 WikiEntry
- 分批并发，每批最多 `RepoWikiWriterMaxConcurrent` 个 Writer（默认 4）
- 每批独立超时，等待一批完成后再启动下一批
- 通过 `save_wiki_page` 工具写入 `versions/{vid}/wiki/`

### 阶段 5：Validator 校验与重试

- 扫描 `versions/{vid}/wiki/` 目录
- 输出 `{valid, errors, metadata_generated}`
- 错误类型包括：`missing_file`, `missing_metadata`, `empty_page`, `orphan_file`
- `valid=false` 时，从 errors 提取缺失 path，重驱动 Writer 补写对应 WikiEntry
- 最多重试 `RepoWikiWriterMaxRetry` 次（默认 2）

## 数据模型变更

### 后端 constant 变更（`internal/constant/llm.go`）

```go
// Agent 角色常量
const (
    AgentRoleRepoWiki             = "repowiki"             // RepoWiki 模块标识
    AgentRoleRepoWikiCoordinator  = "repowiki:coordinator" // 概要分析
    AgentRoleRepoWikiExplore      = "repowiki:explore"     // 代码探索
    AgentRoleRepoWikiArchitect    = "repowiki:architect"   // 架构规划
    AgentRoleRepoWikiWrite        = "repowiki:write"       // 文档写作
    AgentRoleRepoWikiValidator    = "repowiki:validator"   // 文档校验
)

// RepoWiki 模块下的所有子角色（用于批量查询和设置页渲染）
var AgentRolesRepoWiki = []string{
    AgentRoleRepoWikiCoordinator,
    AgentRoleRepoWikiExplore,
    AgentRoleRepoWikiArchitect,
    AgentRoleRepoWikiWrite,
    AgentRoleRepoWikiValidator,
}
```

### Info 表键值结构

```
llm_agent_model:repowiki:coordinator  → model_id
llm_agent_model:repowiki:explore      → model_id
llm_agent_model:repowiki:architect     → model_id
llm_agent_model:repowiki:write        → model_id
llm_agent_model:repowiki:validator   → model_id
```

### 迁移策略

启动时在 `prepare/prepare_llm.go` 中检测旧键 `llm_agent_model:repowiki` 是否存在：
- 若存在且值非空，将其值复制到 `llm_agent_model:repowiki:coordinator`（仅当新键值为空时）
- 保证幂等性（已迁移则跳过）

### LlmResolver 变更

使用 `ResolveAgentModel` 逐个解析 5 个角色，构建 `map[role]*ResolvedLlmConfig`。

### 存储结构变更

新版存储按 WikiVersion ID 隔离：

```
{basePath}/
├── repos/{configID}/              # 复用的 Git 克隆目录
└── versions/{versionID}/           # 版本隔离数据
    ├── overview.md
    ├── explore/
    ├── architecture.json
    ├── wiki/
    └── sessions/
```

旧版 `wiki/{configID}/` 目录已废弃，新版本完成后由 Pipeline 清理。

### 版本切换

`RepoWikiConfig.SelectedVersionID` 字段指向当前对外服务的版本。查询时优先使用该版本；未设置时回退到最新完成版本。

## 后端架构变更

### 整体流程

```
AnalyzeRepo
  → 创建 WikiVersion 记录
  → EnsureCloned / FetchAndCheckout
  → resolveOrchestrator: 解析 5 角色模型配置 → 5 个 BambooClient
  → NewSubAgentOrchestrator
  → NewAnalysisPipeline
  → pipeline.Execute
       → SubAgentOrchestrator.Execute
            → overview → explore → architect → writers → validator
  → 完成后更新 SelectedVersionID
  → 清理旧版 wiki 目录
```

### SubAgentOrchestrator

```go
type SubAgentOrchestrator struct {
    roleClients map[string]bamboo.BambooClient // 5 角色 LLM client
    roleModels  map[string]*ModelRunConfig     // 5 角色模型配置
    storage     *service.WikiStorageService    // Wiki 存储服务
    log         *xLog.LogNamedLogger
    versionID   int64
    repoPath    string
}
```

主要方法：

| 方法 | 说明 |
|------|------|
| `Execute` | 完整 5 阶段流水线入口 |
| `runOverview` | 阶段 1：Coordinator 概要 |
| `runExploreConcurrent` | 阶段 2：Explore 并发 |
| `runArchitect` | 阶段 3：Architect 大纲 |
| `runWritersConcurrent` | 阶段 4：Writer 并发 |
| `runValidator` | 阶段 5：Validator 校验 |

### AnalysisPipeline

`AnalysisPipeline` 串联 Git 准备和 `SubAgentOrchestrator`：

```go
type AnalysisPipeline struct {
    logic           *RepoWikiLogic
    log             *xLog.LogNamedLogger
    orchestrator    *SubAgentOrchestrator
    llmProviderName string
    llmModelName    string
}
```

### 状态机变更

```
旧: pending → cloning → analyzing(scan/pass1/pass2/pass3/pass4/assembling) → completed/failed
新: pending → cloning → scanning → analyzing(exploring/architecting/writing/validating) → completed/failed/cancelled
```

## 前端变更

### 角色显示映射

```typescript
const ROLE_DISPLAY_MAP: Record<string, { label: string; desc: string; icon: string }> = {
  'repowiki:coordinator': { label: '概要 Agent', desc: '项目整体概要分析', icon: '📋' },
  'repowiki:explore':     { label: '探索 Agent', desc: '代码阅读与分析', icon: '🔍' },
  'repowiki:architect':   { label: '架构 Agent', desc: '构建 Wiki 目录大纲', icon: '🏗️' },
  'repowiki:write':       { label: '写作 Agent', desc: 'Wiki 文档生成', icon: '✏️' },
  'repowiki:validator':   { label: '校验 Agent', desc: '完整性校验与修复', icon: '✅' },
}
```

### 设置页组件

`<AgentModelAssignGroup module="repowiki" />` 渲染 5 个角色的模型选择器。

## 文件变更清单

### 后端（Go）

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `internal/constant/llm.go` | 修改 | 新增 5 个角色常量 + `AgentRolesRepoWiki` 列表 |
| `internal/constant/repowiki.go` | 修改 | 新增阶段枚举 `exploring/architecting/writing/validating` |
| `internal/service/llm_resolver.go` | 修改 | 支持按角色解析模型配置 |
| `internal/service/wiki_storage.go` | 修改 | 支持版本隔离路径（versions/{vid}/） |
| `internal/service/repo_tools.go` | 修改 | 新增 `save_wiki_page` 工具 |
| `internal/logic/repowiki_orchestrator.go` | 新增 | SubAgentOrchestrator：5 阶段预定义编排 |
| `internal/logic/repowiki_subagent_prompts.go` | 新增 | 5 角色 system prompt 和 user prompt 构建 |
| `internal/logic/repowiki_types.go` | 新增 | `WikiEntry` / `ValidationError` / `ExploreOutput` / `ModelRunConfig` |
| `internal/logic/repowiki_pipeline.go` | 新增 | AnalysisPipeline：Git 准备 + 状态机驱动 |
| `internal/logic/repowiki_logic.go` | 修改 | AnalyzeRepo 调用 Pipeline + Orchestrator |
| `internal/entity/repowiki_config.go` | 修改 | 新增 `SelectedVersionID` 字段 |
| `internal/handler/repowiki.go` | 修改 | 新增版本管理/查询接口 |
| `internal/app/route/route_repowiki.go` | 新增 | RepoWiki REST API 路由 |
| `internal/mcp/repowiki_tools.go` | 新增 | MCP 只读工具：repoWiki_query / repoWiki_list |
| `internal/app/startup/prepare/prepare_llm.go` | 修改 | 旧键迁移逻辑 |

### 设计文档

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `docs/wiki/repowiki/design.md` | 修改 | 更新核心流程为 5 角色 SubAgent 编排 |
| `docs/wiki/repowiki/multi-agent-design.md` | 重写 | 本文档 |

## 风险与缓解

| 风险 | 缓解措施 |
|------|----------|
| Explore 并发导致 Token 激增 | 单 scope 独立超时，失败比例阈值控制，使用低成本模型 |
| Architect JSON 输出不稳定 | 自动重试 2 次，追加格式提醒；仍失败标记整体失败 |
| Writer 产出空页面或缺失 | Validator 校验 + 重驱动 Writer，最多 2 次 |
| 版本目录磁盘占用增长 | 每个 config 保留最多 10 个版本，后台清理旧版本 |
| 不同角色模型来自不同 Provider | 每个角色独立构建 BambooClient，配置隔离 |
| 旧版 config 级 Wiki 目录残留 | 新版本完成后 Pipeline 清理旧目录 |
