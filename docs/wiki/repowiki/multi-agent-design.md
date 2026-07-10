# RepoWiki 多 Agent 协作设计

> **状态**：已审批，待实现  
> **日期**：2026-07-09  
> **替代**：`docs/wiki/repowiki/design.md` 中"阶段三：LLM 分阶段分析"的固定 4 Pass 流程

## 背景与问题

### 当前设计缺陷

RepoWiki 的 Agent 模型分配采用单角色设计：Info 表键 `llm_agent_model:repowiki` 绑定一个 model_id，设置页只渲染一个模型选择器。

然而 RepoWiki 的分析流程涉及多个不同职责的 Agent 协作——代码探索、文档撰写、流程编排——每个职责对模型能力的要求不同：

- **代码探索**：需要快速读文件、低成本、高吞吐
- **文档写作**：需要强写作能力、Markdown/Mermaid 生成
- **编排决策**：需要强推理能力、大上下文窗口

单角色设计无法支持为不同职责分配不同模型。

### 目标

将 RepoWiki 从固定 4 Pass 串行流程升级为 **ReAct 范式的多 Agent 协作**：主 Agent（Coordinator）作为编排者，通过 `agent.NewAgentTool` 将 Explore 和 Write 子 Agent 注册为工具，以 ReAct 循环自主调度整个分析流程。

## 技术基础

bamboo-agent 框架（`github.com/bamboo-services/bamboo-agent v0.0.1`）原生支持三种编排模式：

| 能力 | API | 说明 |
|------|-----|------|
| Agent-as-Tool | `agent.NewAgentTool(name, desc, subAgent)` | 子 Agent 作为父 Agent 的工具调用，独立 session，上下文隔离 |
| Handoff 转移 | `agent.Handoff` | Agent 间转移控制权，内置循环检测 |
| DAG 编排 | `orchestrator.NewOrchestrator()` | 依赖图驱动的任务编排 |

本设计采用 **Agent-as-Tool** 模式，因为它最符合"主 Agent 监工编排 + 子 Agent 专业执行"的 ReAct 范式。

关键框架约束：
- 最大嵌套深度 3 层（框架内置，防止无限递归）
- 子 Agent 使用独立 Session，不共享父 Agent 上下文
- AgentTool 默认只返回 `result.Content`（可配置返回完整 JSON）

## 角色定义

| 角色 | 标识 | 职责 | 工具集 | 模型特点 |
|------|------|------|--------|----------|
| **Coordinator** | `repowiki:coordinator` | 监工/编排：读取仓库扫描摘要，决定探索策略、何时开始写作、文档是否达标。通过 ReAct 循环调用 explore/write 工具 | `explore`(AgentTool), `write`(AgentTool) | 推理能力强、上下文窗口大（如 GPT-4o / Claude Sonnet） |
| **Explore** | `repowiki:explore` | 代码探索专家：读文件、分析依赖、生成结构化 JSON 摘要。被 Coordinator 作为 `explore` 工具调用 | `file_read`, `file_search` | 快速、低成本（如 GPT-4o-mini / Claude Haiku） |
| **Write** | `repowiki:write` | 文档写作专家：基于 Explore 产出的 JSON 摘要，撰写 Markdown 文档和 Mermaid 图表。被 Coordinator 作为 `write` 工具调用 | `save_wiki_page` | 写作能力强（如 Claude Sonnet / GPT-4o） |

### 编排流程（Coordinator 视角的 ReAct 循环）

```
Coordinator 启动（输入：仓库路径 + 文件扫描摘要）
  → Thought: "我需要先了解项目结构"
  → Action: explore("扫描仓库根目录和入口文件，输出项目概览 JSON")
  → Observation: { 项目概览 JSON }
  → Thought: "概览有了，接下来深入分析模块划分"
  → Action: explore("基于概览分析模块间依赖关系，输出模块分析 JSON")
  → Observation: { 模块分析 JSON }
  → Thought: "信息足够了，开始写概览文档"
  → Action: write("基于概览 JSON 编写「项目概览.md」")
  → Observation: { 文档写入成功 }
  → ... 继续直到所有章节完成 ...
  → Thought: "所有章节已完成，结束"
```

输出仍然是四个章节（概览/模块/架构/指南），但生成顺序和深度由 Coordinator 灵活控制，而非硬编码的固定 Pass 顺序。

## 数据模型变更

### 后端 constant 变更（`internal/constant/llm.go`）

```go
// Agent 角色常量
const (
    AgentRoleRepoWiki             = "repowiki"             // RepoWiki 模块标识
    AgentRoleRepoWikiCoordinator  = "repowiki:coordinator" // 主控 Agent（编排决策）
    AgentRoleRepoWikiExplore      = "repowiki:explore"     // 探索 Agent（代码阅读）
    AgentRoleRepoWikiWrite        = "repowiki:write"       // 写作 Agent（文档生成）
)

// RepoWiki 模块下的所有子角色（用于批量查询和设置页渲染）
var AgentRolesRepoWiki = []string{
    AgentRoleRepoWikiCoordinator,
    AgentRoleRepoWikiExplore,
    AgentRoleRepoWikiWrite,
}
```

### Info 表键值结构

```
旧: llm_agent_model:repowiki                 → model_id
新: llm_agent_model:repowiki:coordinator     → model_id
    llm_agent_model:repowiki:explore         → model_id
    llm_agent_model:repowiki:write           → model_id
```

### 迁移策略

启动时在 `prepare/prepare_llm.go` 中检测旧键 `llm_agent_model:repowiki` 是否存在：
- 若存在且值非空，将其值复制到 `llm_agent_model:repowiki:coordinator`（仅当新键值为空时）
- **保留旧键不删除**（作为 coordinator 别名，直到后端编排层重构完成，因为 `repowiki_logic.go` 仍读取旧键）
- 保证幂等性（已迁移则跳过）

### LlmResolver 变更

新增批量解析方法：

```go
// ResolveAgentModels 批量解析多个角色的 LLM 配置
// 返回 map[role]*ResolvedLlmConfig，缺失的角色不出现在 map 中
func (r *LlmResolver) ResolveAgentModels(
    ctx context.Context,
    roles []string,
    keyPrefix string,
) (map[string]*ResolvedLlmConfig, error)
```

### API 变更

**新增端点**：
- `GET /api/v1/llm/agent/models?module=repowiki` — 返回模块下所有子角色的模型分配列表

```json
{
  "code": 0,
  "message": "查询成功",
  "data": {
    "module": "repowiki",
    "assignments": [
      { "role": "repowiki:coordinator", "model_id": "123", "model_name": "GPT-4o" },
      { "role": "repowiki:explore", "model_id": "456", "model_name": "Claude Haiku" },
      { "role": "repowiki:write", "model_id": null, "model_name": null }
    ]
  }
}
```

**保留端点**：
- `PUT /api/v1/llm/agent/:role/model` — body `{ model_id }`，role 值改为子角色标识（如 `repowiki:explore`），路由路径和请求体结构不变

**废弃端点**：
- `GET /api/v1/llm/agent/:role/model`（role=repowiki）— 旧的单角色查询。本迭代**不修改**该 handler（保持向后兼容），前端不再调用该端点，改为使用批量查询端点。后续迭代可选择在 Swagger 注释中标记 @Deprecated

## 后端架构变更

### 整体流程

```
AnalyzeRepo → ResolveAgentModels(["coordinator", "explore", "write"])
            → 3 个 ResolvedLlmConfig → 3 个独立 BambooClient
            → 构建 Explore Agent  (explore client + file_read + file_search)
            → 构建 Write Agent    (write client + save_wiki_page)
            → 构建 Coordinator    (coordinator client + explore_tool + write_tool)
            → Coordinator.Run(ctx, 编排指令) → ReAct 循环自主编排
```

### AgentOrchestrator（替代 AgentPassRunner）

当前 `AgentPassRunner` 硬编码 4 Pass 串行执行，重构为 `AgentOrchestrator`：

```go
type AgentOrchestrator struct {
    coordinatorClient bamboo.BambooClient
    exploreClient     bamboo.BambooClient
    writeClient       bamboo.BambooClient
    storage           *service.WikiStorageService
    log               *xLog.LogNamedLogger
    repoPath          string  // 克隆的仓库路径（Explore Agent 的工具作用域）
    versionID         int64
}
```

### 构建 Agent 的流程

1. **Explore Agent**：使用 explore 的 LLM 配置构建 BambooClient → 创建 Agent → 注入 `file_read`(repoPath) + `file_search`(repoPath)
2. **Write Agent**：使用 write 的 LLM 配置构建 BambooClient → 创建 Agent → 注入 `save_wiki_page`(wikiDir)
3. **Coordinator**：使用 coordinator 的 LLM 配置 → 创建 Agent → 通过 `agent.NewAgentTool("explore", "...", exploreAgent)` 和 `agent.NewAgentTool("write", "...", writeAgent)` 注册子 Agent 为工具

### 工具变更

| 工具 | 注入到 | 作用域 | 状态 |
|------|--------|--------|------|
| `file_read` | Explore | repoPath | 已有，读取仓库内文件 |
| `file_search` | Explore | repoPath | 已有，搜索仓库内文件 |
| `save_wiki_page` | Write | wikiDir | **新增**，将 Markdown 内容写入 Wiki 输出目录（如 `content/项目概览.md`），路径限定在 Wiki 输出目录内，防止路径穿越 |

### Coordinator system prompt 方向

```
你是 RepoWiki 的主编排 Agent。你的任务是协调 Explore 和 Write 两个专家 Agent，
为一个代码仓库生成完整的 Wiki 文档。

可用工具：
- explore(input): 调用探索专家阅读和分析代码库
- write(input): 调用写作专家撰写 Wiki 文档页面

Wiki 文档需要包含以下章节：
1. 项目概览 — 项目定位、技术栈、入口点
2. 模块分析 — 模块清单、职责划分、依赖关系
3. 架构设计 — 整体架构模式、数据流、设计决策
4. 阅读指南 — 新人上手路径、关键代码索引

工作策略：
1. 先用 explore 工具了解项目整体结构和入口文件
2. 根据探索结果，决定需要深入分析哪些模块
3. 当信息充足时，用 write 工具逐个生成 Wiki 章节
4. 所有章节完成后结束

仓库路径: {repoPath}
文件扫描摘要: {fileScanJSON}
```

### 结果提取

Coordinator.Run() 完成后：
- **Wiki 文档**：由 Write Agent 通过 `save_wiki_page` 工具直接写入文件系统，无需额外组装步骤
- **元数据 JSON**（`repowiki-metadata.json`）：由 AgentOrchestrator 在 Coordinator 完成后，基于 Wiki 目录实际产出的文件列表生成
- **工具调用记录**：从 Coordinator session 中提取所有 explore/write 工具调用的记录，用于调试和状态追踪（替代旧的 Pass 结果 JSON）

### 状态机变更

WikiVersion 的 `current_stage` 字段值调整：

```
旧: pending → cloning → analyzing(scan/pass1/pass2/pass3/pass4/assembling) → completed/failed
新: pending → cloning → analyzing(scan/orchestrating) → completed/failed
```

- `scan` — 文件扫描阶段（不变）
- `orchestrating` — Coordinator ReAct 编排阶段（替代 pass1-4 + assembling）

## 前端变更

### 类型定义变更（`web/src/lib/models/response/llm.ts`）

```typescript
// Agent 模型分配项（单个角色，含模型名称，用于批量查询）
export interface AgentModelAssignmentItem {
  role: string               // 如 "repowiki:coordinator"
  model_id: string | null
  model_name: string | null  // 关联的模型显示名
}

// 批量查询响应
export interface AgentModelAssignmentsResponse {
  module: string             // 如 "repowiki"
  assignments: AgentModelAssignmentItem[]
}

// 兼容旧接口（保留，逐步废弃）
export interface AgentModelAssignment {
  role: string
  model_id: string | null
}
```

### API 客户端变更（`web/src/lib/apis/llm.ts`）

```typescript
// 旧: getAgentModel(role: string)
// 新: 批量查询模块下所有角色
export function getAgentModels(module: string): Promise<BaseResponse<AgentModelAssignmentsResponse>>

// PUT 不变，role 值改为子角色标识
export function updateAgentModel(role: string, modelId: string): Promise<BaseResponse<null>>
```

### Hook 变更（`web/src/hooks/useLlmConfig.ts`）

```typescript
// 旧: useAgentModel(role: string)
// 新: 批量查询
export function useAgentModels(module: string)
export function useUpdateAgentModel()  // 不变
```

### 组件变更（`web/src/components/llm/agent-model-assign.tsx`）

**旧设计**：`<AgentModelAssign role="repowiki" />` — 单行选择器

**新设计**：`<AgentModelAssignGroup module="repowiki" />` — 按角色分组的多行选择器

```
┌─────────────────────────────────────────────────────────┐
│  RepoWiki 分析                                           │
│  协调多个 Agent 协作生成 Wiki 文档                         │
│                                                         │
│  🎯 主控 Agent        编排决策，协调子 Agent  [选择模型 ▾] │
│  🔍 探索 Agent        代码阅读与分析        [选择模型 ▾] │
│  ✏️ 写作 Agent        Wiki 文档生成         [选择模型 ▾] │
└─────────────────────────────────────────────────────────┘
```

角色显示映射：

```typescript
const ROLE_DISPLAY_MAP: Record<string, { label: string; desc: string; icon: string }> = {
  'repowiki:coordinator': { label: '主控 Agent', desc: '编排决策，协调子 Agent', icon: '🎯' },
  'repowiki:explore':     { label: '探索 Agent', desc: '代码阅读与分析', icon: '🔍' },
  'repowiki:write':       { label: '写作 Agent', desc: 'Wiki 文档生成', icon: '✏️' },
}
```

### settings.tsx 变更

```tsx
// 旧: <AgentModelAssign role="repowiki" />
// 新:
<AgentModelAssignGroup module="repowiki" />
```

## 文件变更清单

### 后端（Go）

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `internal/constant/llm.go` | 修改 | 新增三个角色常量 + `AgentRolesRepoWiki` 列表 |
| `internal/service/llm_resolver.go` | 修改 | 新增 `ResolveAgentModels` 批量解析方法 |
| `internal/service/repo_tools.go` | 修改 | 新增 `save_wiki_page` 工具 |
| `internal/logic/repowiki_agent.go` | 重构 | `AgentPassRunner` → `AgentOrchestrator`，移除 4 Pass 硬编码 |
| `internal/logic/repowiki_logic.go` | 修改 | `AnalyzeRepo` 改为调用 `ResolveAgentModels` + 构建 `AgentOrchestrator` |
| `internal/constant/repowiki.go` | 修改 | `current_stage` 枚举值调整（`orchestrating` 替代 pass1-4） |
| `internal/handler/llm.go` | 修改 | 新增 `GetAgentModels` 批量查询 handler |
| `internal/app/route/route_llm.go` | 修改 | 注册新端点 `GET /agent/models` |
| `api/llm/llm.go` | 修改 | 新增批量查询请求/响应 DTO |
| `internal/app/startup/prepare/prepare_llm.go` | 修改 | 新增旧键迁移逻辑 |

### 前端（TypeScript）

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `web/src/lib/models/response/llm.ts` | 修改 | `AgentModelAssignment` 结构变更 + 新增 `AgentModelAssignmentsResponse` |
| `web/src/lib/apis/llm.ts` | 修改 | `getAgentModel` → `getAgentModels`（批量查询） |
| `web/src/hooks/useLlmConfig.ts` | 修改 | `useAgentModel` → `useAgentModels`（批量查询） |
| `web/src/components/llm/agent-model-assign.tsx` | 重构 | `AgentModelAssign` → `AgentModelAssignGroup`（多角色选择器） |
| `web/src/routes/console/settings.tsx` | 修改 | Agent 分配 Tab 渲染变更 |

### 设计文档

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `docs/wiki/repowiki/design.md` | 修改 | 更新"阶段三：LLM 分阶段分析"为多 Agent 协作描述 |
| `docs/wiki/repowiki/multi-agent-design.md` | 新增 | 本文档（实际实现时创建到项目 docs 目录） |

## 风险与缓解

| 风险 | 缓解措施 |
|------|----------|
| Coordinator 自主决策可能导致输出不稳定 | system prompt 中明确章节清单和完成条件；设置最大工具调用次数限制 |
| 不同角色的模型来自不同 Provider，需要构建多个 BambooClient | LlmResolver 已支持按角色独立解析，BambooClient 构建逻辑按角色隔离 |
| 旧用户升级后旧键丢失 | `prepare_llm.go` 启动迁移逻辑，幂等保证 |
| ReAct 循环可能消耗更多 Token | Explore Agent 使用低成本模型平衡；Coordinator 使用 summary mode 减少上下文膨胀 |
