# RepoWiki 详细设计

## 核心流程

```
Git URL 输入
    │
    ▼
┌──────────┐     ┌──────────┐     ┌──────────┐
│ Git Clone │────▶│ 文件扫描  │────▶│ 依赖图   │
│ (go-git)  │     │ + 过滤   │     │ + PageRank│
└──────────┘     └──────────┘     └──────────┘
                                         │
                                         ▼
                                  ┌──────────────┐
                                  │ LLM 分阶段分析 │
                                  │              │
                                  │ Pass 1: 概览  │
                                  │ Pass 2: 模块  │
                                  │ Pass 3: 架构  │
                                  │ Pass 4: 指南  │
                                  └──────┬───────┘
                                         │
                                         ▼
                                  ┌──────────────┐
                                  │ 文档组装      │
                                  │ Markdown 输出 │
                                  │ Mermaid 图表  │
                                  │ 元数据 JSON   │
                                  └──────┬───────┘
                                         │
                              ┌──────────┴──────────┐
                              ▼                     ▼
                       PostgreSQL 元数据       文件系统 Markdown
                       (项目、版本、状态)       (.lumina/repowiki/)
```

## 阶段一：Git 克隆与文件扫描

### Git 克隆

- 使用 go-git 库克隆目标仓库到本地临时目录
- 支持公开仓库和私有仓库（通过配置 credentials）
- 克隆完成后记录当前 commit hash 用于增量更新

### 文件扫描

- 遍历仓库目录树，收集文件列表
- 过滤规则：
  - 排除 `.git`、`node_modules`、`vendor`、`build` 等目录
  - 排除二进制文件、图片、字体等非文本内容
  - 尊重 `.gitignore` 规则
- 识别编程语言（根据文件扩展名统计）
- 检测入口文件（`main.go`、`index.ts`、`app.py` 等）

## 阶段二：依赖图与 PageRank

### 依赖图构建

- 解析 import / require 语句，建立文件间依赖关系
- 构建有向图（文件 → 文件、模块 → 模块）
- 识别核心模块（被依赖最多的节点）

### PageRank 排序

- 对文件按重要性排序（基于依赖图的 PageRank 算法）
- 高分文件优先送入 LLM 分析
- 确保在 token 限制内覆盖最重要的代码

## 阶段三：多 Agent 协作分析

RepoWiki 不再采用固定的 4 Pass 串行流程，而是使用 ReAct 范式的多 Agent 协作。主 Agent 以 Coordinator 角色监工，通过 `agent.NewAgentTool` 把 Explore 和 Write 两个子 Agent 注册为工具，自主决定何时探索、何时写作、何时结束。

### 角色定义

| 角色 | 标识 | 职责 | 说明 |
|------|------|------|------|
| **Coordinator** | `repowiki:coordinator` | 编排决策 | 读取仓库扫描摘要，规划探索策略，调度 explore 和 write 工具，判断 Wiki 是否完成 |
| **Explore** | `repowiki:explore` | 代码探索 | 读文件、分析依赖、产出结构化 JSON 摘要，供 Coordinator 和 Write 使用 |
| **Write** | `repowiki:write` | 文档写作 | 基于 Explore 的摘要，撰写 Markdown 页面和 Mermaid 图表 |

### 编排流程

Coordinator 进入 ReAct 循环后，按如下方式工作：

1. 先用 `explore` 工具扫描仓库根目录和入口文件，获得项目概览 JSON
2. 根据概览决定深入哪些模块，继续调用 `explore` 产出模块分析 JSON
3. 当信息足够时，调用 `write` 工具逐个生成 Wiki 章节
4. 所有章节完成后结束

Wiki 仍包含四个章节（项目概览、模块分析、架构设计、阅读指南），但生成顺序和深度由 Coordinator 灵活控制，而不是硬编码的 Pass 顺序。

详细设计（数据模型、后端架构、前端变更、风险缓解）参见 [multi-agent-design.md](multi-agent-design.md)。

## 阶段四：文档组装与输出

### 目录结构

```
.lumina/repowiki/{project_id}/
├── zh/
│   ├── 主页.md                    # Wiki 首页，项目概览与导航
│   ├── content/
│   │   ├── 项目概览.md            # Pass 1 产出
│   │   ├── 模块分析.md            # Pass 2 产出
│   │   ├── 架构设计.md            # Pass 3 产出
│   │   └── 阅读指南.md            # Pass 4 产出
│   └── meta/
│       └── repowiki-metadata.json # 侧边栏导航配置
└── en/
    └── ...                        # 英文版本（可选）
```

### 元数据 JSON 结构（参考值）

```json
{
  "navigation": [
    {
      "title": "项目概览",
      "path": "content/项目概览.md",
      "order": 1
    },
    {
      "title": "模块分析",
      "path": "content/模块分析.md",
      "order": 2
    },
    {
      "title": "架构设计",
      "path": "content/架构设计.md",
      "order": 3
    },
    {
      "title": "阅读指南",
      "path": "content/阅读指南.md",
      "order": 4
    }
  ],
  "home": "主页.md",
  "language": "zh",
  "project_name": "项目名称",
  "generated_at": "2026-05-31T00:00:00Z",
  "commit_hash": "abc1234"
}
```

## 增量更新策略

### 触发条件

- Git Provider Webhook 推送触发
- 对比当前 HEAD commit hash 与上次分析的 commit hash
- 如果无变更，跳过分析直接返回现有 Wiki

### 更新流程

1. 获取当前 HEAD commit hash
2. 与上次记录的 commit hash 对比
3. 计算变更文件列表（`git diff --name-only`）
4. 分析变更文件的依赖链（影响范围）
5. 仅对变更影响范围内的模块重新执行 LLM 分析
6. 未变更部分保留原 Wiki 内容
7. 更新 commit hash 记录

## Webhook 更新机制

RepoWiki 的 Wiki 增量更新由 Git Provider Webhook 触发。当 Git 仓库发生 push 事件时，对应的 Provider 向 Lumina Webhook 端点推送事件，Lumina 验证签名、匹配分支后自动创建新版本并执行增量分析。

详细设计参见 [webhook-design.md](webhook-design.md)。

## 数据库模型（物理模型）

### 设计决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 与 Project 表关系 | RepoWikiConfig 1:1 → Project（uniqueIndex） | 一个项目最多一份 Wiki 配置，避免冗余关联表 |
| 版本管理策略 | WikiVersion 表 + Language 列 | zh/en 独立版本记录，同一 commit 可有多语言版本 |
| 任务管理 | WikiVersion 记录 status + current_stage + error_msg | 不新建任务表，版本即任务 |
| 页面索引 | 纯文件系统 metadata.json | 避免数据库存储文档内容，保持轻量 |
| Git 配置存储 | RepoWikiConfig 表（1对1） | 含 last_accessed_at 用于 TTL 清理 |
| 源码存储 | 持久化 + TTL 30 天清理 | 平衡磁盘占用与增量更新效率 |
| 中间产物 | 文件系统缓存 | file_scan/dep_graph/pagerank JSON 文件 |
| Agent Session | WikiVersion 记 agent_session_path | bamboo-agent FileStore 持久化（SDK 未发布，仅预留字段） |

### 表关系图

```
┌──────────────────┐     1:1      ┌──────────────────────┐     1:N      ┌──────────────────┐
│    Project       │◄────────────│  RepoWikiConfig      │◄────────────│  WikiVersion     │
│                  │             │  (Gene=39)           │             │  (Gene=40)       │
│  id (PK)         │             │  id (PK)             │             │  id (PK)         │
│  name            │             │  project_id (FK→1:1) │             │  config_id (FK)  │
│  alias_name      │             │  git_url             │             │  commit_hash     │
│  match_path      │             │  default_branch      │             │  branch          │
│  description     │             │  local_path          │             │  language        │
│                  │             │  default_language    │             │  status          │
│                  │             │  last_accessed_at    │             │  current_stage   │
│                  │             │                      │             │  ...             │
└──────────────────┘             └──────────────────────┘             └──────────────────┘
```

### RepoWikiConfig 表（Gene=39）

RepoWiki 配置表，与 Project 表 1:1 关联，存储 Git 仓库配置和 Wiki 生成偏好。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, 雪花ID | 主键 |
| project_id | bigint | NOT NULL, UNIQUE INDEX | 关联项目ID（1:1） |
| git_url | varchar(512) | NOT NULL, INDEX | Git仓库地址 |
| default_branch | varchar(128) | NOT NULL, DEFAULT 'main' | 默认分支 |
| local_path | varchar(512) | — | 本地克隆路径 |
| default_language | varchar(16) | NOT NULL, DEFAULT 'zh' | 默认Wiki语言 |
| last_accessed_at | timestamptz | INDEX | 最后访问时间（TTL清理用） |
| created_at | timestamptz | — | 创建时间 |
| updated_at | timestamptz | — | 更新时间 |

#### Go Entity 参考

```go
type RepoWikiConfig struct {
    xModels.BaseEntity                                                                                       // 基础实体
    ProjectID       xSnowflake.SnowflakeID `gorm:"type:bigint;not null;uniqueIndex;comment:关联项目ID" json:"project_id"`       // 关联项目ID
    GitURL          string                 `gorm:"type:varchar(512);not null;index;comment:Git仓库地址" json:"git_url"`           // Git仓库地址
    DefaultBranch   string                 `gorm:"type:varchar(128);not null;default:main;comment:默认分支" json:"default_branch"` // 默认分支
    LocalPath       string                 `gorm:"type:varchar(512);comment:本地克隆路径" json:"local_path"`                         // 本地克隆路径
    DefaultLanguage string                 `gorm:"type:varchar(16);not null;default:zh;comment:默认Wiki语言" json:"default_language"` // 默认Wiki语言
    LastAccessedAt  *time.Time             `gorm:"type:timestamptz;index;comment:最后访问时间" json:"last_accessed_at,omitempty"`     // 最后访问时间
}
```

### WikiVersion 表（Gene=40）

Wiki 版本表，每次分析生成一条版本记录，同时承载任务状态管理。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, 雪花ID | 主键 |
| config_id | bigint | NOT NULL, INDEX | 关联配置ID |
| commit_hash | varchar(64) | NOT NULL | Git commit hash |
| branch | varchar(128) | NOT NULL, DEFAULT 'main' | 分析分支 |
| language | varchar(16) | NOT NULL, DEFAULT 'zh' | Wiki语言 |
| llm_model | varchar(128) | — | LLM模型名称 |
| llm_provider | varchar(64) | — | LLM Provider |
| status | varchar(16) | NOT NULL, DEFAULT 'pending', INDEX | 分析状态 |
| current_stage | varchar(32) | — | 当前阶段 |
| agent_session_path | varchar(512) | — | bamboo-agent Session路径 |
| file_count | int | DEFAULT 0 | 分析文件数 |
| token_count | bigint | DEFAULT 0 | LLM token消耗 |
| duration_ms | int | DEFAULT 0 | 分析耗时毫秒 |
| error_msg | text | — | 失败原因 |
| started_at | timestamptz | — | 分析开始时间 |
| completed_at | timestamptz | — | 分析完成时间 |
| created_at | timestamptz | — | 创建时间 |
| updated_at | timestamptz | — | 更新时间 |

#### Go Entity 参考

```go
type WikiVersion struct {
    xModels.BaseEntity                                                                                              // 基础实体
    ConfigID         xSnowflake.SnowflakeID `gorm:"type:bigint;not null;index;comment:关联配置ID" json:"config_id"`                // 关联配置ID
    CommitHash       string                 `gorm:"type:varchar(64);not null;comment:Git commit hash" json:"commit_hash"`           // Git commit hash
    Branch           string                 `gorm:"type:varchar(128);not null;default:main;comment:分析分支" json:"branch"`           // 分析分支
    Language         string                 `gorm:"type:varchar(16);not null;default:zh;comment:Wiki语言" json:"language"`           // Wiki语言
    LLMModel         string                 `gorm:"type:varchar(128);comment:LLM模型名称" json:"llm_model"`                           // LLM模型名称
    LLMProvider      string                 `gorm:"type:varchar(64);comment:LLM Provider" json:"llm_provider"`                      // LLM Provider
    Status           string                 `gorm:"type:varchar(16);not null;default:pending;index;comment:分析状态" json:"status"`   // 分析状态
    CurrentStage     string                 `gorm:"type:varchar(32);comment:当前阶段" json:"current_stage"`                           // 当前阶段
    AgentSessionPath string                 `gorm:"type:varchar(512);comment:bamboo-agent Session路径" json:"agent_session_path"`    // bamboo-agent Session路径
    FileCount        int                    `gorm:"type:int;default:0;comment:分析文件数" json:"file_count"`                           // 分析文件数
    TokenCount       int64                  `gorm:"type:bigint;default:0;comment:LLM token消耗" json:"token_count"`                  // LLM token消耗
    DurationMs       int                    `gorm:"type:int;default:0;comment:分析耗时毫秒" json:"duration_ms"`                         // 分析耗时毫秒
    ErrorMsg         string                 `gorm:"type:text;comment:失败原因" json:"error_msg,omitempty"`                            // 失败原因
    StartedAt        *time.Time             `gorm:"type:timestamptz;comment:分析开始时间" json:"started_at,omitempty"`                  // 分析开始时间
    CompletedAt      *time.Time             `gorm:"type:timestamptz;comment:分析完成时间" json:"completed_at,omitempty"`                // 分析完成时间
}
```

### 状态流转图

```
                    ┌─────────┐
                    │ pending │ ← 初始状态
                    └────┬────┘
                         │ 开始克隆
                         ▼
                   ┌───────────┐
          ┌────────│  cloning  │────────┐
          │        └─────┬─────┘        │
          │ 克隆失败      │ 克隆成功       │ 克隆失败
          │              │              │
          │        ┌─────▼─────┐        │
          │        │ analyzing │        │
          │        └─────┬─────┘        │
          │              │              │
          │        ┌─────▼─────┐        │
          │        │    scan   │        │
          │        └─────┬─────┘        │
          │              │              │
          │        ┌─────▼─────┐        │
          │        │orchestrating│      │
          │        └─────┬─────┘        │
          │              │              │
          │        ┌─────▼─────┐        │
          │        │ completed │        │
          │        └───────────┘        │
          │                             │
          ▼                             ▼
    ┌──────────┐               ┌──────────┐
    │  failed  │◄──────────────│  failed  │
    └──────────┘               └──────────┘
```

**Status 枚举值**：`pending` → `cloning` → `analyzing` → `completed` / `failed`

**CurrentStage 枚举值**（status=analyzing 时）：
- `scan` — 文件扫描阶段
- `orchestrating` — Coordinator 多 Agent ReAct 编排阶段（替代原 pass1-4 + assembling）

### 文件系统结构

```
.lumina/repowiki/
├── repos/                          # Git 克隆持久化目录
│   └── {config_id}/                # 按 RepoWikiConfig ID 隔离
│       └── {branch}/               # 按分支隔离
│           └── ...                 # 完整 Git 仓库（TTL 30天清理）
├── intermediate/                   # 中间产物缓存
│   └── {version_id}/               # 按 WikiVersion ID 隔离
│       ├── file_scan.json          # 文件扫描结果
│       ├── dep_graph.json          # 依赖图数据
│       └── pagerank.json           # PageRank 排序结果
├── sessions/                       # bamboo-agent Session 持久化
│   └── {version_id}/               # 按 WikiVersion ID 隔离
│       └── ...                     # Agent Session 文件（FileStore）
└── wiki/                           # 最终 Wiki 文档输出
    └── {project_id}/
        ├── zh/
        │   ├── 主页.md
        │   ├── content/
        │   │   ├── 项目概览.md
        │   │   ├── 模块分析.md
        │   │   ├── 架构设计.md
        │   │   └── 阅读指南.md
        │   └── meta/
        │       └── repowiki-metadata.json
        └── en/
            └── ...
```

### 基因编号注册

| 基因编号 | 常量名 | 实体 |
|----------|--------|------|
| 39 | `GeneRepoWikiConfig` | RepoWikiConfig |
| 40 | `GeneWikiVersion` | WikiVersion |

> 基因编号定义于 `internal/constant/gene_number.go`，用于雪花算法分布式 ID 生成。
