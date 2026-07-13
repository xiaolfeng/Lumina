# RepoWiki 详细设计

## 核心流程

```
Git URL 输入
    │
    ▼
┌──────────┐     ┌──────────────────────────────────────────────────────────────┐
│ Git Clone │────▶│ 5 角色 SubAgent 预定义编排                                     │
│ (go-git)  │     │                                                               │
└──────────┘     │  1. Coordinator: 项目概要分析（overview.md）                  │
                 │  2. Explore ×4:  并发代码探索（versions/{vid}/explore/*.xml）  │
                 │  3. Architect:   构建 Wiki 目录大纲（architecture.json）       │
                 │  4. Writer ×4:   并发文档撰写（versions/{vid}/wiki/*.md）        │
                 │  5. Validator:   完整性校验 + 失败重驱动 Writer（最多 2 次）     │
                 └────────────────────────┬──────────────────────────────────────┘
                                          │
                                          ▼
                               ┌──────────────────────┐
                               │  版本隔离 Wiki 输出    │
                               │  versions/{vid}/wiki/ │
                               └──────────────────────┘
```

## 阶段一：Git 克隆与文件扫描

### Git 克隆

- 使用 go-git 库将目标仓库克隆到本地持久化目录
- 支持公开仓库和私有仓库（通过配置 SSH 密钥或 credentials）
- 克隆路径按 `RepoWikiConfig.ID` 隔离：`{basePath}/repos/{configID}/`
- 跨版本复用同一克隆目录，通过 `fetch + checkout` 进行增量更新

### 文件扫描

- 遍历仓库目录树，收集文件列表
- 过滤规则：
  - 排除 `.git`、`node_modules`、`vendor`、`dist`、`build` 等目录
  - 排除二进制文件、图片、字体等非文本内容
  - 尊重 `.gitignore` 规则
- 识别编程语言（根据文件扩展名统计）
- 检测入口文件（`main.go`、`index.ts`、`app.py` 等）
- 扫描结果持久化到 `versions/{vid}/file_scan.json`

## 阶段二：5 角色 SubAgent 预定义编排

RepoWiki 不再采用固定的 4 Pass 串行流程，也不再使用 ReAct 循环让 Coordinator 自主决策。整个 Wiki 生成流程由 `SubAgentOrchestrator` 按预定义阶段驱动，每个阶段调用对应角色的子 Agent。

### 角色定义

| 角色 | 标识 | 职责 | 说明 |
|------|------|------|------|
| **Coordinator** | `repowiki:coordinator` | 概要分析 | 阅读仓库关键文件，产出项目整体概览 Markdown |
| **Explore** | `repowiki:explore` | 代码探索 | 按模块/目录并发分析，产出 XML 骨架 |
| **Architect** | `repowiki:architect` | 架构规划 | 综合概要与探索产出，构建 Wiki 目录大纲 JSON |
| **Write** | `repowiki:write` | 文档写作 | 按大纲分批并发撰写 Markdown 页面 |
| **Validator** | `repowiki:validator` | 文档校验 | 校验完整性，失败时重驱动 Writer 补写（最多 2 次） |

### 编排流程

```
Coordinator（单实例）
  → 产出 overview.md
  → 从 overview 提取顶层目录作为分析 scope

Explore（并发，最多 4 路）
  → 每个 scope 一个 Explore Agent
  → 产出 explore/{scope}.xml
  → 失败比例超过 50% 则整体失败

Architect（单实例）
  → 读取 overview + 全部 explore 产出
  → 输出 architecture.json（WikiEntry 数组）
  → JSON 解析失败自动重试 2 次

Writer（分批并发，每批最多 4 路）
  → 按 complexity 分组：high 每路 1 个条目，medium/low 每路 2 个条目
  → 通过 save_wiki_page 工具写入 versions/{vid}/wiki/

Validator（单实例）
  → 扫描 wiki 目录校验完整性
  → 输出 {valid, errors, metadata_generated}
  → valid=false 时重驱动 Writer 补写缺失条目，最多 2 次
```

### 阶段状态枚举

`WikiVersion.current_stage` 字段枚举值：

| 阶段 | 说明 |
|------|------|
| `scan` | 文件扫描阶段 |
| `exploring` | 代码探索阶段 |
| `architecting` | 架构规划阶段 |
| `writing` | 文档撰写阶段 |
| `validating` | 结果校验阶段 |

## 阶段三：文档组装与输出

### 目录结构

```
{basePath}/versions/{version_id}/
├── overview.md                 # Coordinator 产出的项目概要
├── explore/                    # Explore 阶段产出
│   ├── internal_logic.xml
│   ├── web_src.xml
│   └── ...
├── architecture.json           # Architect 产出的 Wiki 目录大纲
├── wiki/                       # 最终 Wiki 文档
│   ├── meta/
│   │   └── repowiki-metadata.json
│   ├── content/
│   │   ├── overview.md
│   │   ├── architecture.md
│   │   ├── modules/
│   │   └── ...
│   └── home.md
└── sessions/                   # 子 Agent FileStore 会话目录
```

### 元数据 JSON 结构（参考值）

```json
{
  "navigation": [
    {
      "title": "项目概览",
      "path": "content/overview.md",
      "order": 1
    },
    {
      "title": "架构设计",
      "path": "content/architecture.md",
      "order": 2
    }
  ],
  "home": "home.md",
  "language": "zh",
  "project_name": "项目名称",
  "generated_at": "2026-07-14T00:00:00Z",
  "commit_hash": "abc1234"
}
```

## 版本切换策略

### 多版本隔离

每个 Wiki 版本拥有独立的文件系统目录 `versions/{vid}/`。版本之间互不覆盖，便于回滚和对比。

### 版本切换

`RepoWikiConfig.SelectedVersionID` 字段指向当前对外服务的 Wiki 版本。查询 Wiki 内容时，优先使用选中版本；未设置时默认使用最新完成版本。

### 版本清理

每个 `RepoWikiConfig` 最多保留 `RepoWikiMaxVersions` 个完成版本（默认 10），旧版本由后台清理任务删除。

## 增量更新策略

### 触发条件

- Git Provider Webhook 推送触发
- 对比当前 HEAD commit hash 与上次分析的 commit hash
- 如果无变更，跳过分析直接返回现有 Wiki

### 更新流程

1. 接收 Webhook 推送事件
2. 验证签名/Token 并匹配分支
3. 创建新的 `WikiVersion` 记录
4. 执行 Git 增量更新（fetch + checkout）
5. 运行 5 阶段 SubAgent 编排生成新 Wiki
6. 新版本完成后自动更新 `SelectedVersionID`
7. 清理旧版 config 级 Wiki 目录（兼容性清理）

## Webhook 更新机制

RepoWiki 的 Wiki 更新由 Git Provider Webhook 触发。当 Git 仓库发生 push 事件时，对应的 Provider 向 Lumina Webhook 端点推送事件，Lumina 验证签名、匹配分支后自动创建新版本并执行分析。

详细设计参见 [webhook-design.md](webhook-design.md)。

## 数据库模型（物理模型）

### 设计决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 与 Project 表关系 | RepoWikiConfig 1:1 → Project（uniqueIndex） | 一个项目最多一份 Wiki 配置，避免冗余关联表 |
| 版本管理策略 | WikiVersion 表 + version 级文件隔离 | 每个版本独立目录，支持多版本共存和切换 |
| 当前服务版本 | RepoWikiConfig.SelectedVersionID | 通过配置字段指向当前对外服务的版本 |
| 任务管理 | WikiVersion 记录 status + current_stage + error_msg | 不新建任务表，版本即任务 |
| 页面索引 | 纯文件系统 metadata.json | 避免数据库存储文档内容，保持轻量 |
| Git 配置存储 | RepoWikiConfig 表（1对1） | 含 last_accessed_at 用于 TTL 清理 |
| 源码存储 | 持久化 + TTL 30 天清理 | 平衡磁盘占用与增量更新效率 |
| 中间产物 | 文件系统缓存 | overview/explore/architecture JSON/XML 文件 |
| Agent Session | versions/{vid}/sessions/ | 按 WikiVersion ID 隔离，FileStore 持久化 |

### 表关系图

```
┌──────────────────┐     1:1      ┌──────────────────────┐     1:N      ┌──────────────────┐
│    Project       │◄────────────│  RepoWikiConfig      │◄────────────│  WikiVersion     │
│                  │             │  (Gene=39)           │             │  (Gene=40)       │
│  id (PK)         │             │  id (PK)             │             │  id (PK)         │
│  name            │             │  project_id (FK→1:1) │             │  config_id (FK)  │
│  alias_name      │             │  git_url             │             │  commit_hash     │
│  match_path      │             │  default_branch      │             │  branch          │
│  description     │             │  default_language    │             │  language        │
│                  │             │  selected_version_id │             │  status          │
│                  │             │  webhook_token       │             │  current_stage   │
│                  │             │  webhook_secret      │             │  ...             │
│                  │             │  webhook_branches    │             │                  │
└──────────────────┘             └──────────────────────┘             └──────────────────┘
```

### RepoWikiConfig 表（Gene=39）

RepoWiki 配置表，与 Project 表 1:1 关联，存储 Git 仓库配置、Wiki 生成偏好和当前服务版本。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK, 雪花ID | 主键 |
| project_id | bigint | NOT NULL, UNIQUE INDEX | 关联项目ID（1:1） |
| name | varchar(128) | NOT NULL | 配置名称 |
| git_url | varchar(512) | NOT NULL, INDEX | Git仓库地址 |
| default_branch | varchar(128) | NOT NULL, DEFAULT 'main' | 默认分支 |
| local_path | varchar(512) | — | 本地克隆路径 |
| default_language | varchar(16) | NOT NULL, DEFAULT 'zh' | 默认Wiki语言 |
| ssh_key_id | bigint | INDEX | 关联SSH密钥ID |
| wiki_password_hash | varchar(128) | — | Wiki访问密码bcrypt哈希 |
| status | varchar(16) | NOT NULL, DEFAULT 'pending', INDEX | 当前分析状态（快速查询用） |
| selected_version_id | bigint | INDEX | 当前选中的Wiki版本ID |
| last_accessed_at | timestamptz | INDEX | 最后访问时间（TTL清理用） |
| webhook_token | varchar(128) | UNIQUE INDEX | Webhook访问令牌 |
| webhook_secret | varchar(128) | — | Webhook签名密钥 |
| webhook_branches | jsonb | — | Webhook监听分支列表 |
| created_at | timestamptz | — | 创建时间 |
| updated_at | timestamptz | — | 更新时间 |

#### Go Entity 参考

```go
type RepoWikiConfig struct {
    xModels.BaseEntity
    ProjectID          xSnowflake.SnowflakeID `gorm:"type:bigint;not null;uniqueIndex;comment:关联项目ID" json:"project_id"`
    Name               string                 `gorm:"type:varchar(128);not null;comment:配置名称" json:"name"`
    GitURL             string                 `gorm:"type:varchar(512);not null;index;comment:Git仓库地址" json:"git_url"`
    DefaultBranch      string                 `gorm:"type:varchar(128);not null;default:main;comment:默认分支" json:"default_branch"`
    LocalPath          string                 `gorm:"type:varchar(512);comment:本地克隆路径" json:"local_path"`
    DefaultLanguage    string                 `gorm:"type:varchar(16);not null;default:zh;comment:默认Wiki语言" json:"default_language"`
    SSHKeyID          *xSnowflake.SnowflakeID `gorm:"type:bigint;index;comment:关联SSH密钥ID" json:"ssh_key_id,omitempty"`
    WikiPasswordHash   string                 `gorm:"type:varchar(128);comment:Wiki访问密码bcrypt哈希" json:"wiki_password_hash,omitempty"`
    Status             string                 `gorm:"type:varchar(16);not null;default:pending;index;comment:当前分析状态（快速查询用）" json:"status"`
    SelectedVersionID *xSnowflake.SnowflakeID `gorm:"type:bigint;index;comment:当前选中的Wiki版本ID" json:"selected_version_id,omitempty"`
    LastAccessedAt     *time.Time             `gorm:"type:timestamptz;index;comment:最后访问时间" json:"last_accessed_at,omitempty"`
    WebhookToken       string                 `gorm:"type:varchar(128);uniqueIndex;comment:Webhook访问令牌" json:"webhook_token,omitempty"`
    WebhookSecret      string                 `gorm:"type:varchar(128);comment:Webhook签名密钥" json:"webhook_secret,omitempty"`
    WebhookBranches    datatypes.JSON         `gorm:"type:jsonb;comment:Webhook监听分支列表" json:"webhook_branches,omitempty"`
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
    xModels.BaseEntity
    ConfigID         xSnowflake.SnowflakeID `gorm:"type:bigint;not null;index;comment:关联配置ID" json:"config_id"`
    CommitHash       string                 `gorm:"type:varchar(64);not null;comment:Git commit hash" json:"commit_hash"`
    Branch           string                 `gorm:"type:varchar(128);not null;default:main;comment:分析分支" json:"branch"`
    Language         string                 `gorm:"type:varchar(16);not null;default:zh;comment:Wiki语言" json:"language"`
    LLMModel         string                 `gorm:"type:varchar(128);comment:LLM模型名称" json:"llm_model"`
    LLMProvider      string                 `gorm:"type:varchar(64);comment:LLM Provider" json:"llm_provider"`
    Status           string                 `gorm:"type:varchar(16);not null;default:pending;index;comment:分析状态" json:"status"`
    CurrentStage     string                 `gorm:"type:varchar(32);comment:当前阶段" json:"current_stage"`
    AgentSessionPath string                 `gorm:"type:varchar(512);comment:bamboo-agent Session路径" json:"agent_session_path"`
    FileCount        int                    `gorm:"type:int;default:0;comment:分析文件数" json:"file_count"`
    TokenCount       int64                  `gorm:"type:bigint;default:0;comment:LLM token消耗" json:"token_count"`
    DurationMs       int                    `gorm:"type:int;default:0;comment:分析耗时毫秒" json:"duration_ms"`
    ErrorMsg         string                 `gorm:"type:text;comment:失败原因" json:"error_msg,omitempty"`
    StartedAt        *time.Time             `gorm:"type:timestamptz;comment:分析开始时间" json:"started_at,omitempty"`
    CompletedAt      *time.Time             `gorm:"type:timestamptz;comment:分析完成时间" json:"completed_at,omitempty"`
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
           │ 克隆失败       │ 克隆成功      │ 克隆失败
           │              │              │
           │        ┌─────▼─────┐        │
           │        │  scanning │        │
           │        └─────┬─────┘        │
           │              │              │
           │        ┌─────▼─────┐        │
           │        │ exploring │        │
           │        └─────┬─────┘        │
           │              │              │
           │        ┌─────▼─────┐        │
           │        │architecting│       │
           │        └─────┬─────┘        │
           │              │              │
           │        ┌─────▼─────┐        │
           │        │  writing  │        │
           │        └─────┬─────┘        │
           │              │              │
           │        ┌─────▼─────┐        │
           │        │ validating│        │
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
           │                             │
           ▼                             ▼
     ┌──────────┐               ┌──────────┐
     │cancelled │               │cancelled │
     └──────────┘               └──────────┘
```

**Status 枚举值**：`pending` → `cloning` → `scanning` / `analyzing` → `completed` / `failed` / `cancelled`

**CurrentStage 枚举值**（status=analyzing 时）：
- `scan` — 文件扫描阶段
- `exploring` — 代码探索阶段
- `architecting` — 架构规划阶段
- `writing` — 文档撰写阶段
- `validating` — 结果校验阶段

### 文件系统结构

```
.lumina/repowiki/
├── repos/                          # Git 克隆持久化目录
│   └── {config_id}/                # 按 RepoWikiConfig ID 隔离
│       └── ...                     # 完整 Git 仓库（TTL 30天清理）
├── versions/                       # 版本隔离数据
│   └── {version_id}/               # 按 WikiVersion ID 隔离
│       ├── overview.md             # Coordinator 概要产出
│       ├── file_scan.json          # 文件扫描结果
│       ├── dep_summary.json        # 依赖摘要
│       ├── architecture.json       # Architect 目录大纲
│       ├── explore/                # Explore 阶段 XML 产出
│       │   ├── {scope}.xml
│       │   └── ...
│       ├── wiki/                   # 最终 Wiki 文档
│       │   ├── home.md
│       │   ├── meta/
│       │   │   └── repowiki-metadata.json
│       │   └── content/
│       │       └── ...
│       └── sessions/               # Agent FileStore 会话目录
│           └── ...
└── wiki/                           # 旧版 config 级 Wiki 目录（已废弃，保留清理用）
    └── {config_id}/
```

### 基因编号注册

| 基因编号 | 常量名 | 实体 |
|----------|--------|------|
| 39 | `GeneRepoWikiConfig` | RepoWikiConfig |
| 40 | `GeneWikiVersion` | WikiVersion |

> 基因编号定义于 `internal/constant/gene_number.go`，用于雪花算法分布式 ID 生成。
