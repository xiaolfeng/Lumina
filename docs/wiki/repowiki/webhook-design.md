# RepoWiki Webhook 更新机制设计

> **状态**：设计定稿（决策已确认）  
> **日期**：2026-07-07  
> **背景**：原设计中 RepoWiki 的更新由 Agent 通过 MCP 工具 `repoWiki_update` 触发。  
> 实际场景中，Wiki 更新应由 **Git 代码推送事件** 驱动（GitHub/Gitee/GitLab/Gitea Webhook），  
> 而非 Agent 主动决定。因此需要将更新触发权从 MCP 迁移到 Webhook 端点。

---

## 一、设计原则

| 通道 | 职责 | 使用者 |
|------|------|--------|
| **MCP** | **仅读取** Wiki 内容与版本列表 | AI Agent |
| **Webhook** | **仅触发更新**（增量分析） | Git Provider（GitHub/Gitee/GitLab/Gitea） |
| **REST API** | **完整 CRUD**（创建配置、首次分析、删除、Webhook 管理、分支监听管理） | 前端管理界面 |

核心理念：**Agent 是 Wiki 内容的消费者，不是更新的触发者。** 更新由代码仓库的 push 事件自动驱动。

---

## 二、MCP 工具调整

### 2.1 工具变更矩阵

| 工具 | 当前状态 | 调整后 | 理由 |
|------|----------|--------|------|
| `repoWiki_query` | ✅ 已实现 | ✅ 保留 | Agent 查询 Wiki 内容（核心读取能力） |
| `repoWiki_list` | ✅ 已实现（列配置） | ✅ **语义变更**（列 Wiki 版本） | Agent 需要的是「有哪些 Wiki 可读」，不是「有哪些项目配置」。项目列表由 Project 模块提供 |
| `repoWiki_analyze` | ✅ 已实现 | ❌ 删除 | 初始分析通过前端 REST API 触发 |
| `repoWiki_update` | ✅ 已实现 | ❌ 删除 | 更新通过 Webhook 触发 |
| `repoWiki_delete` | ✅ 已实现 | ❌ 删除 | 删除通过前端 REST API 操作 |

### 2.2 `repoWiki_list` 语义变更

**调整前**：列出所有 RepoWikiConfig（配置列表 — config_id、git_url、branch、status）

**调整后**：列出所有 **已完成的 Wiki 版本**（Wiki 列表 — version_id、项目名称、分支、语言、commit_hash、完成时间）

Agent 通过 `repoWiki_list` 获取「哪些 Wiki 可以读」，然后通过 `repoWiki_query` 读取具体内容。两个工具构成完整的 Wiki 消费链路。

**新返回值结构（参考）**：

```
Wiki 版本列表（共 3 个，第 1/1 页）：

1. [version_id: 123] Lumina · 微明
   分支: main | 语言: zh | commit: acfbafb
   完成时间: 2026-07-07T12:00:00Z

2. [version_id: 456] bamboo-base-go
   分支: master | 语言: zh | commit: 29a74f0
   完成时间: 2026-07-04T18:00:00Z
```

### 2.3 MCP Server Instructions 更新

`internal/mcp/server.go` 中的 `Instructions` 字段需要：
- 移除 RepoWiki 写操作描述（analyze/update/delete）
- 明确说明 MCP 仅提供 RepoWiki 读取能力（query Wiki 内容、list 可用 Wiki 版本）
- 说明更新由 Git Webhook 自动触发，Agent 无需关心

---

## 三、Webhook 端点设计

### 3.1 URL 设计

```
POST /api/v1/webhooks/repowiki/:token
```

- **token**：使用 `xUtil.Security().GenerateLongKey()` 生成（格式：`cs_` + 64 位十六进制，共 67 字符）
- 不暴露 config_id，通过 token 反查配置（安全性 + 隐蔽性）
- Token 具有唯一索引，支持快速反查

### 3.2 认证方式

支持四种 Git Provider 的认证方式，按请求 Header 自动适配：

| Provider | 认证 Header | 认证方式 | 验证逻辑 |
|----------|-------------|----------|----------|
| GitHub | `X-Hub-Signature-256` | HMAC-SHA256 | `sha256=hex(HMAC-SHA256(secret, body))` |
| Gitee | `X-Gitee-Token` | 静态 Token 比对 | `token == config.WebhookSecret` |
| GitLab | `X-Gitlab-Token` | 静态 Token 比对 | `token == config.WebhookSecret` |
| Gitea | `X-Gitea-Signature` | HMAC-SHA256 | `hex(HMAC-SHA256(secret, body))` |

**自动适配逻辑**：
1. 检查 `X-Gitee-Token` → Gitee 模式（静态 Token）
2. 检查 `X-Gitlab-Token` → GitLab 模式（静态 Token）
3. 检查 `X-Hub-Signature-256` → GitHub 模式（HMAC）
4. 检查 `X-Gitea-Signature` → Gitea 模式（HMAC）
5. 都没有 → 401 Unauthorized

### 3.3 Payload 解析

Webhook handler 需要解析四种 Provider 的 push 事件 payload，统一提取：

| 字段 | GitHub | Gitee | GitLab | Gitea |
|------|--------|-------|--------|-------|
| 事件类型 | `X-GitHub-Event: push` | `X-Gitee-Event: Push Hook` | `object_kind: "push"` | `X-Gitea-Event: push` |
| 分支 | `ref: "refs/heads/main"` | `ref: "refs/heads/main"` | `ref: "refs/heads/main"` | `ref: "refs/heads/main"` |
| 仓库 URL | `repository.clone_url` | `project.clone_url` | `project.git_http_url` | `repository.clone_url` |
| 前置 commit | `before` | `before` | `before` | `before` |
| 当前 commit | `after` | `after` | `after` | `after` |
| 变更文件 | `commits[].added/modified/removed` | `commits[].added/modified/removed` | `commits[].added/modified/removed` | `commits[].added/modified/removed` |

**统一内部结构**：

```go
type WebhookPushEvent struct {
    Provider     string   // "github" | "gitee" | "gitlab" | "gitea"
    Branch       string   // 推送的目标分支（去掉 refs/heads/ 前缀）
    RepoURL      string   // 仓库 URL（用于二次验证）
    BeforeHash   string   // 变更前的 commit hash
    AfterHash    string   // 变更后的 commit hash
    ChangedFiles []string // 变更文件列表（added + modified + removed）
}
```

### 3.4 处理流程

```
Git Provider Push Event
    │
    ▼
┌──────────────────────────┐
│  1. Token 反查 Config     │ ← 从 URL :token 查 RepoWikiConfig
│  （失败 → 404）            │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│  2. 签名/Token 验证       │ ← 按 Provider 自动适配（HMAC 或 静态比对）
│  （失败 → 401，记录日志）  │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│  3. 解析 Payload          │ ← 统一提取 branch/repo_url/files
│  （非 push 事件 → 200 跳过 │   记录日志
│   ，记录日志）             │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│  4. 分支匹配检查          │ ← 与 config.WebhookBranches 列表比对
│  （不在监听列表 → 200 跳过 │   用户在 UI 中配置的监听分支
│   ，记录日志）             │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│  5.5 仓库 URL 二次验证     │ ← 标准化并比对 payload 中的仓库 URL
│  （与 config.GitURL 不匹配 │   与 config.GitURL 比对
│   → 200 跳过，记录日志）    │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│  6. 创建 WebhookEvent     │ ← status = received
│  （写入审计日志）           │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│  7. 检查是否已有进行中分析 │ ← 存在 pending/cloning/analyzing 则排队
│  （复用现有并发控制）         │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│  8. 创建 WikiVersion 记录 │ ← status = pending
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│  8.5 重复 commit 检查      │ ← 与最新 completed 版本 commit hash 比对
│  （一致 → 200 跳过，        │   避免重复分析同一版本
│   记录日志）                │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│  9. 触发增量分析           │ ← 创建后台管道，异步执行
│  （异步，立即返回 200）     │   记录 version_id 到日志
└──────────┬───────────────┘
           │
           ▼
     HTTP 200 + JSON Response
     {
       "status": "accepted",
       "version_id": 123456,
       "message": "增量分析已触发"
     }
```

**关键设计决策**：

1. **快速返回**：Webhook handler 在验证通过后立即创建 WikiVersion 记录并返回 200，分析管道在后台 goroutine 异步执行。Git Provider 的 webhook 超时通常为 10-30 秒，必须快速响应。
2. **非 push 事件跳过**：对于 `ping`、`pull_request` 等非 push 事件，返回 200 但不触发分析。**记录到 WebhookEvent 日志**。
3. **分支过滤（多分支监听）**：只处理 `config.WebhookBranches` 列表中包含的分支。**默认为空列表（不监听任何分支）**，用户必须通过 UI 显式添加要监听的分支。这避免了 webhook 自动触发未授权的分析。
4. **去重**：如果同一 config 已有正在进行的分析（status 为 pending/cloning/analyzing），新 webhook 触发的分析排队等待（信号量已存在，复用现有并发控制）。
5. **全链路日志**：每个 webhook 请求都记录到 `WebhookEvent` 表，包含：接收时间、Provider、事件类型、分支、处理结果（accepted/ignored/failed）、关联的 version_id。

---

### WebhookEvent 状态生命周期

| 状态 | 含义 | 下一状态 |
|------|------|----------|
| `received` | 已接收，正在处理 | `accepted` / `ignored` / `failed` |
| `accepted` | 已接受，触发了分析 | — |
| `ignored` | 已跳过（非 push 事件 / 分支不匹配 / 仓库 URL 不匹配 / 重复 commit / 监听列表为空） | — |
| `failed` | 处理失败（签名验证失败 / 分析触发失败） | — |

## 四、实体变更

### 4.1 RepoWikiConfig 新增字段

```go
// 新增到 RepoWikiConfig 实体
WebhookToken    string   `gorm:"type:varchar(128);uniqueIndex;comment:Webhook访问令牌" json:"webhook_token,omitempty"`      // Webhook 访问令牌（URL 路由用，xUtil.Security().GenerateLongKey() 生成）
WebhookSecret   string   `gorm:"type:varchar(128);comment:Webhook签名密钥" json:"webhook_secret,omitempty"`               // Webhook HMAC 签名密钥 / 静态 Token（xUtil.Security().GenerateLongKey() 生成）
WebhookBranches datatypes.JSON `gorm:"type:jsonb;comment:Webhook监听分支列表" json:"webhook_branches,omitempty"`        // Webhook 监听分支列表（JSON 数组，如 ["main","develop"]，空数组=不监听）
```

**字段说明**：

- `WebhookToken`：使用 `xUtil.Security().GenerateLongKey()` 生成（`cs_` + 64 位十六进制，67 字符）。用于 URL 路由匹配，唯一索引。**明文存储**（非敏感信息，仅用于路由）。
- `WebhookSecret`：使用 `xUtil.Security().GenerateLongKey()` 生成。用于 HMAC 签名验证（GitHub/Gitea）或静态 Token 比对（Gitee/GitLab）。**明文存储**（HMAC 验证需要明文密钥，无法哈希）。
- `WebhookBranches`：JSON 数组，存储用户配置的监听分支列表。**默认为空数组 `[]`**（不监听任何分支，webhook 不会触发分析）。用户通过 UI 管理此列表。

### 4.2 新增 WebhookEvent 实体

```go
// internal/entity/webhook_event.go
type WebhookEvent struct {
    xModels.BaseEntity                                                                                              // 基础实体
    ConfigID      xSnowflake.SnowflakeID `gorm:"type:bigint;not null;index;comment:关联RepoWikiConfig ID" json:"config_id"`        // 关联配置 ID
    Provider      string                 `gorm:"type:varchar(16);not null;comment:Git Provider" json:"provider"`                    // Git Provider（github/gitee/gitlab/gitea）
    EventType     string                 `gorm:"type:varchar(32);not null;comment:事件类型" json:"event_type"`                         // 事件类型（push/ping/pull_request 等）
    Branch        string                 `gorm:"type:varchar(128);comment:推送分支" json:"branch,omitempty"`                            // 推送分支（仅 push 事件）
    CommitBefore  string                 `gorm:"type:varchar(64);comment:变更前commit" json:"commit_before,omitempty"`                 // 变更前 commit hash
    CommitAfter   string                 `gorm:"type:varchar(64);comment:变更后commit" json:"commit_after,omitempty"`                  // 变更后 commit hash
    ChangedCount  int                    `gorm:"type:int;default:0;comment:变更文件数" json:"changed_count"`                             // 变更文件数量
    Status        string                 `gorm:"type:varchar(16);not null;default:received;index;comment:处理状态" json:"status"`       // 处理状态（received/accepted/ignored/failed）
    Reason        string                 `gorm:"type:varchar(256);comment:跳过/失败原因" json:"reason,omitempty"`                         // 跳过或失败的原因
    VersionID     xSnowflake.SnowflakeID `gorm:"type:bigint;comment:关联WikiVersion ID" json:"version_id,omitempty"`                  // 触发分析时关联的版本 ID
    ResponseCode  int                    `gorm:"type:int;comment:HTTP响应码" json:"response_code"`                                       // 返回给 Provider 的 HTTP 状态码
    ReceivedAt    time.Time              `gorm:"type:timestamptz;not null;comment:接收时间" json:"received_at"`                           // 接收时间
    ProcessedAt   *time.Time             `gorm:"type:timestamptz;comment:处理完成时间" json:"processed_at,omitempty"`                     // 处理完成时间
}

func (w *WebhookEvent) GetGene() xSnowflake.Gene {
    return bConst.GeneWebhookEvent
}
```

**基因编号**：`GeneWebhookEvent = 43`（定义于 `internal/constant/gene_number.go`）

**Status 枚举值**：
- `received` — 已接收，正在处理
- `accepted` — 已接受，触发了分析
- `ignored` — 已跳过（非 push 事件 / 分支不匹配 / 监听列表为空）
- `failed` — 处理失败（签名验证失败 / 分析触发失败）

### 4.3 字段生成时机

- **创建配置时**：自动生成 `WebhookToken` + `WebhookSecret`（使用 `xUtil.Security().GenerateLongKey()`），`WebhookBranches` 默认为 `[]`
- **更新配置时**：不自动重新生成 token/secret（需显式调用 regenerate 端点）
- **Regenerate 端点**：`POST /api/v1/repowiki/configs/:id/webhook/regenerate` — 重新生成 token + secret，旧 token 立即失效
- **现有记录迁移**：现有 `RepoWikiConfig` 记录在启动时自动生成 `WebhookToken`/`WebhookSecret`，`WebhookBranches` 默认设为空数组

---

## 五、REST API 变更

### 5.1 新增端点

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| `GET` | `/api/v1/repowiki/configs/:id/webhook` | 获取 Webhook 配置信息（URL、监听分支、Provider 指南） | Auth |
| `PUT` | `/api/v1/repowiki/configs/:id/webhook/branches` | 更新 Webhook 监听分支列表 | Auth |
| `POST` | `/api/v1/repowiki/configs/:id/webhook/regenerate` | 重新生成 Webhook Token + Secret | Auth |
| `GET` | `/api/v1/repowiki/configs/:id/webhook/events` | 获取 Webhook 事件日志（分页） | Auth |
| `POST` | `/api/v1/webhooks/repowiki/:token` | Webhook 接收端点（Git Provider 调用） | **无 Auth**（签名验证） |

### 5.2 Webhook 接收端点详细设计

**路径**：`POST /api/v1/webhooks/repowiki/:token`

**请求**：
- Content-Type: `application/json`
- Body: Git Provider 的原始 webhook payload
- Headers: 各 Provider 的签名/Token Header

**响应**：

```json
// 成功触发分析
{
  "code": 0,
  "message": "增量分析已触发",
  "data": {
    "version_id": 123456789,
    "status": "pending"
  }
}

// 监听分支列表为空（未配置监听）
{
  "code": 0,
  "message": "事件已接收但未触发分析",
  "data": {
    "reason": "no_branches_configured",
    "hint": "请在管理界面添加监听分支"
  }
}

// 分支不在监听列表中
{
  "code": 0,
  "message": "事件已接收但未触发分析",
  "data": {
    "reason": "branch_not_monitored",
    "branch": "develop",
    "monitored_branches": ["main", "release"]
  }
}

// 非 push 事件（跳过）
{
  "code": 0,
  "message": "事件已接收但未触发分析",
  "data": {
    "reason": "non_push_event",
    "event_type": "ping"
  }
}
```

**错误响应**：

```json
// Token 无效
404 { "code": 404, "message": "webhook not found" }

// 签名验证失败
401 { "code": 401, "message": "signature verification failed" }
```

### 5.3 路由注册位置

Webhook 端点**不经过 Auth 中间件**，但需要独立的签名验证逻辑。注册位置参考 MCP 端点（在 `engine.Use()` 之前注册，绕开 `ResponseMiddleware`）。

```go
// internal/app/route/route_webhook.go
func (r *route) webhookRouter(engine *gin.Engine) {
    // Webhook 端点：无 Auth，签名验证在 handler 内部完成
    engine.POST("/api/v1/webhooks/repowiki/:token", handler.NewWebhookHandler(r.context).HandleRepoWikiWebhook)
}
```

---

## 六、分层架构影响

### 6.1 新增文件清单

| 层 | 文件 | 职责 |
|----|------|------|
| **Entity** | `internal/entity/webhook_event.go` | WebhookEvent 实体（Gene=43） |
| **Constant** | `internal/constant/gene_number.go` | 追加 `GeneWebhookEvent = 43` |
| **Route** | `internal/app/route/route_webhook.go` | Webhook 接收路由注册（无 Auth） |
| **Handler** | `internal/handler/webhook.go` | Webhook 请求处理（Token 反查 + 签名验证 + Payload 解析 + 触发分析 + 日志记录） |
| **Logic** | `internal/logic/repowiki_webhook.go` | Webhook 业务逻辑（Payload 统一解析 + 事件分发 + 日志写入） |
| **Repository** | `internal/repository/webhook_event.go` | WebhookEvent 持久化（CRUD + 按 ConfigID 分页查询） |
| **Service** | `internal/service/webhook_parser.go` | Git Provider Payload 解析器（GitHub/Gitee/GitLab/Gitea 四种格式 → 统一结构） |
| **Service** | `internal/service/webhook_signer.go` | 签名验证器（HMAC-SHA256 for GitHub/Gitea + 静态 Token for Gitee/GitLab） |
| **DTO** | `api/webhook/webhook.go` | Webhook 请求/响应 DTO |

### 6.2 修改文件清单

| 文件 | 变更内容 |
|------|----------|
| `internal/entity/repowiki_config.go` | 新增 `WebhookToken` + `WebhookSecret` + `WebhookBranches` 字段 |
| `internal/constant/gene_number.go` | 追加 `GeneWebhookEvent = 43` |
| `internal/mcp/repowiki_tools.go` | 删除 `repoWiki_analyze`/`repoWiki_update`/`repoWiki_delete`；`repoWiki_list` 语义改为列 Wiki 版本 |
| `internal/mcp/server.go` | 更新 `Instructions` 描述（移除写操作，说明 MCP 仅读取） |
| `internal/handler/repowiki.go` | 新增 `GetWebhookConfig`/`UpdateWebhookBranches`/`RegenerateWebhook`/`ListWebhookEvents` 方法 |
| `internal/logic/repowiki_logic.go` | 新增 `GenerateWebhookCredentials`/`RegenerateWebhook`/`HandleWebhookPush`/`ListWebhookEvents` 方法 |
| `internal/repository/repowiki_config.go` | 新增 `GetByWebhookToken` 方法 |
| `internal/app/route/route_repowiki.go` | 新增 webhook 管理路由 |
| `internal/app/route/route.go` | 注册 webhook 接收路由 |
| `internal/app/startup/startup_database.go` | `migrateTables` 追加 `WebhookEvent` |
| `api/repowiki/config.go` | 新增 `WebhookConfigResponse`/`WebhookBranchesRequest`/`WebhookEventResponse` DTO |

### 6.3 MCP 工具变更明细

`internal/mcp/repowiki_tools.go` 变更：

**删除**：
- `repoWikiToolDefs` 中的 `repoWiki_analyze`、`repoWiki_update`、`repoWiki_delete` 定义
- `handleRepoWikiAnalyze`、`handleRepoWikiUpdate`、`handleRepoWikiDelete` handler 函数
- `RegisterRepoWikiTools` 中对应的 case 分支

**保留并修改**：
- `repoWiki_query` 定义 + `handleRepoWikiQuery`（不变）
- `repoWiki_list`：**修改描述 + handler 实现**，从调用 `ListConfigs` 改为调用 `ListCompletedWikis`（新方法，返回已完成的 Wiki 版本列表）
- `SetRepoWikiLogic` / `repoWikiLogic` 变量
- `parseSnowflakeInt` helper

---

## 七、前端变更

### 7.1 RepoWiki 配置详情页

新增 **Webhook 配置** 区块：

- **Webhook URL**：显示完整 URL（`https://your-lumina.com/api/v1/webhooks/repowiki/{token}`），可一键复制
- **Webhook Secret**：显示遮蔽值（`cs_••••••••`），仅在创建/regenerate 时显示明文
- **监听分支管理**：
  - 分支列表展示（Tag 形式，可删除）
  - 添加分支输入框（精确匹配，不支持通配符，如 `main`、`develop`）
  - 空列表时显示提示：「未配置监听分支，Webhook 不会触发分析」
- **Provider 指南**：下拉选择 GitHub/Gitee/GitLab/Gitea，显示对应的配置步骤
- **Regenerate 按钮**：重新生成 token + secret（需二次确认，警告旧 token 立即失效）
- **Webhook 事件日志**：最近 N 条 webhook 事件记录（时间、Provider、分支、状态、原因）

### 7.2 前端新增文件

| 文件 | 职责 |
|------|------|
| `web/src/components/repowiki/webhook-config.tsx` | Webhook 配置展示与操作组件 |
| `web/src/components/repowiki/webhook-branches.tsx` | 监听分支管理组件（Tag 列表 + 添加/删除） |
| `web/src/components/repowiki/webhook-events.tsx` | Webhook 事件日志展示组件 |
| `web/src/lib/apis/webhook.ts` | Webhook 管理 API 封装 |
| `web/src/hooks/useWebhook.ts` | Webhook 配置 + 事件日志 TanStack Query Hook |

---

## 八、Git Provider 配置指南

### 8.1 GitHub

1. 进入仓库 **Settings → Webhooks → Add webhook**
2. **Payload URL**: `https://your-lumina.com/api/v1/webhooks/repowiki/{token}`
3. **Content type**: `application/json`
4. **Secret**: 填入 Webhook Secret（从 Lumina 配置页获取）
5. **Events**: Just the `push` event
6. **Active**: ✅

### 8.2 Gitee

1. 进入仓库 **管理 → WebHooks → 添加 WebHook**
2. **URL**: `https://your-lumina.com/api/v1/webhooks/repowiki/{token}`
3. **密码**: 填入 Webhook Secret
4. **触发事件**: Push 事件
5. **激活**: ✅

### 8.3 GitLab

1. 进入仓库 **Settings → Webhooks**
2. **URL**: `https://your-lumina.com/api/v1/webhooks/repowiki/{token}`
3. **Secret token**: 填入 Webhook Secret
4. **Trigger**: Push events
5. **SSL verification**: Enable

### 8.4 Gitea

1. 进入仓库 **Settings → Webhooks → Add Webhook → Gitea**
2. **URL**: `https://your-lumina.com/api/v1/webhooks/repowiki/{token}`
3. **HTTP method**: POST
4. **Content-Type**: application/json
5. **Secret**: 填入 Webhook Secret
6. **Trigger On**: Push events

---

## 九、安全考量

| 风险 | 缓解措施 |
|------|----------|
| Token 泄露 | Token 仅在 URL 中传输，配合 HTTPS；Regenerate 可随时轮换 |
| 伪造 Webhook | HMAC-SHA256 签名验证（GitHub/Gitea）；静态 Token 比对（Gitee/GitLab） |
| 仓库 URL 伪造 | Payload 中的仓库 URL 与 config.GitURL 二次比对（不匹配则跳过） |
| DDoS 攻击 | Token 反查使用唯一索引（O(1)）；非 push 事件快速跳过；分析异步执行 |
| 未授权触发分析 | `WebhookBranches` 默认为空，必须通过 UI 显式添加监听分支才会触发分析 |
| 事件追溯 | WebhookEvent 表记录所有 webhook 请求的完整审计日志 |

---

## 十、决策确认

| 编号 | 问题 | 决策 | 说明 |
|------|------|------|------|
| Q1 | 初始分析是否保留在 MCP？ | **完全移除** | 初始分析通过前端 REST API 触发，Agent 不负责建配置 |
| Q2 | 是否支持多分支监听？ | **支持，UI 管理** | `WebhookBranches` JSON 数组，默认空，用户通过 UI 添加监听分支 |
| Q3 | Webhook 事件日志是否落库？ | **落库** | 新增 `WebhookEvent` 实体（Gene=43），全链路审计 |
| Q4 | Webhook Secret 存储策略 | **bamboo-base SDK 密钥生成器** | `xUtil.Security().GenerateLongKey()` 生成，明文存储（HMAC 需要明文） |
| Q5 | 支持哪些 Git Provider？ | **GitHub / Gitee / GitLab / Gitea** | 四种主流 Provider，覆盖国内国际场景 |

---

## 十一、实施计划（建议顺序）

| 阶段 | 任务 | 说明 |
|------|------|------|
| **P1** | 实体 + 常量 + 迁移 | RepoWikiConfig 新增 3 字段 + WebhookEvent 实体 + GeneWebhookEvent=43 + migrateTables |
| **P2** | Service 层 | webhook_parser.go（四 Provider Payload 解析）+ webhook_signer.go（签名验证） |
| **P3** | Repository 层 | webhook_event.go（CRUD + 分页）+ repowiki_config.go 新增 GetByWebhookToken |
| **P4** | Logic 层 | repowiki_webhook.go（HandleWebhookPush + 日志写入 + 凭证管理）+ repowiki_logic.go 新增方法 |
| **P5** | Handler + Route | webhook.go handler + route_webhook.go + route_repowiki.go 扩展 |
| **P6** | MCP 裁剪 | 删除 3 个写工具，repoWiki_list 语义变更，更新 Instructions |
| **P7** | DTO | api/webhook/ 新包 + api/repowiki/config.go 扩展 |
| **P8** | 前端 | webhook 配置组件 + 分支管理 + 事件日志 + API + Hook |
| **P9** | 文档更新 | 更新 `mcp-tools.md`、`design.md`、`architecture.md` |
| **P10** | 测试验证 | 单元测试（Payload 解析 + 签名验证）+ 集成测试（模拟 webhook 调用） |

---

> **本文档为设计定稿，决策已确认。**  
> 实施时按 P1-P10 顺序执行，每个阶段完成后验证。
