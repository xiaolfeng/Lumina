# Q&A 详细设计

## 核心定位

Q&A 是 Agent 与用户之间的富交互式问答通道。Agent 通过 MCP 推送问题，用户在浏览器中实时查看并回答。支持选项、文本、分批推送、高级面板（Markdown 渲染、HTML/JS/CSS 即时预览）。

## Session 生命周期

```
Agent 调用 qa_session_create
        │
        ▼
   [active] ◀─── 返回 session_id + 浏览器连接 URL
        │
        │   Agent 可随时:
        │   ├── qa_push_question（推送新问题）
        │   ├── qa_get_answer（获取用户回答）
        │   │
        │   用户在浏览器中:
        │   ├── 通过 SSE 实时收到问题
        │   ├── 回答问题（选项选择 / 文本输入 / 富内容）
        │   └── 查看高级面板（Markdown / 代码预览）
        │
        ├── 超时（默认 7 天，可通过 QA_SESSION_MAX_DURATION 配置）
        │       │
        │       ▼
        │   [expired]（自动归档，只读）
        │
        └── Agent 调用 qa_session_end
                │
                ▼
            [ended]（手动关闭，只读）
```

### Session 状态说明

| 状态 | 说明 | 可推送新问题 | 可提交回答 |
|------|------|-------------|-----------|
| active | 活跃会话 | ✅ | ✅ |
| expired | 超时自动归档 | ❌ | ❌（只读） |
| ended | 手动关闭 | ❌ | ❌（只读） |

### 超时配置

- 默认最大存活时间：**7 天**（604800 秒）
- 通过环境变量 `QA_SESSION_MAX_DURATION` 配置（单位：秒）
- Session 创建时记录过期时间，后端定时检查并归档过期 Session

## 问题推送类型

### 类型一：选项题

推送多个选项供用户选择，每个选项可配置是否支持"其他"（点击后展开 textarea）。

```json
{
  "type": "option",
  "title": "选择技术栈",
  "description": "请选择项目的后端框架",
  "options": [
    { "label": "Gin", "value": "gin", "allow_other": false },
    { "label": "Echo", "value": "echo", "allow_other": false },
    { "label": "其他", "value": "other", "allow_other": true }
  ]
}
```

### 类型二：文本题

开放式文字输入，用户在 textarea 中自由输入。

```json
{
  "type": "text",
  "title": "描述项目需求",
  "description": "请简要描述项目的核心功能需求",
  "placeholder": "请输入..."
}
```

### 类型三：混合题

选项 + 补充说明，选择后可追加文本。

```json
{
  "type": "mixed",
  "title": "选择部署方式",
  "options": [
    { "label": "Docker", "value": "docker" },
    { "label": "裸机部署", "value": "bare-metal" }
  ],
  "allow_extra_text": true,
  "extra_text_placeholder": "补充说明..."
}
```

### 类型四：分批推送

一个复杂问题分段呈现，通过 `batch` 字段标识同一组问题。

```json
{
  "batch": {
    "group_id": "batch-001",
    "sequence": 1,
    "total": 3,
    "description": "数据库设计（共 3 个问题）"
  }
}
```

### 类型五：高级面板

问题右侧可附带富内容面板，支持 Markdown 渲染和 HTML+JS+CSS 即时预览。

```json
{
  "panels": [
    {
      "type": "markdown",
      "content": "# 架构图\n\n```mermaid\ngraph TD\n    A-->B\n```",
      "position": "right"
    },
    {
      "type": "html_preview",
      "content": "<html><body><h1>预览</h1></body></html>",
      "position": "right"
    }
  ]
}
```

## 问题数据结构

### 完整结构（参考值）

```json
{
  "id": "雪花 ID",
  "session_id": "所属会话 ID",
  "type": "option | text | mixed",
  "title": "问题标题",
  "description": "问题描述（支持 Markdown）",
  "placeholder": "输入提示文本（文本题适用）",
  "options": [
    {
      "label": "选项显示文本",
      "value": "选项值",
      "allow_other": false
    }
  ],
  "allow_extra_text": false,
  "extra_text_placeholder": "补充说明提示文本",
  "panels": [
    {
      "type": "markdown | html_preview",
      "content": "面板内容",
      "position": "right"
    }
  ],
  "batch": {
    "group_id": "分批分组 ID",
    "sequence": 1,
    "total": 3
  },
  "status": "pending | answered | skipped",
  "answer": {
    "value": "用户选择/输入的值",
    "extra_text": "补充文本（如果 allow_other 或 allow_extra_text）",
    "answered_at": "回答时间"
  },
  "created_at": "创建时间"
}
```

## 通信架构

```
Agent (MCP)                        用户 (浏览器)
    │                                   │
    │  qa_push_question                 │
    │ ──────────────▶                   │
    │              Lumina 后端          │
    │           ┌──────────┐            │
    │           │ 持久化问题 │            │
    │           │ 到 PG     │            │
    │           └─────┬────┘            │
    │                 │  SSE 事件推送    │
    │                 │ ──────────────▶ │
    │                 │                 │ 用户看到问题
    │                 │                 │
    │                 │ ◀────────────── │ 用户 POST 提交回答
    │           ┌─────┴────┐            │
    │           │ 保存答案   │            │
    │           │ 到 PG     │            │
    │           └──────────┘            │
    │  qa_get_answer                    │
    │ ◀──────────────                   │
    │  返回用户回答                      │
```

### SSE 事件类型（参考值）

| 事件类型 | 说明 | 数据 |
|----------|------|------|
| `question` | 新问题推送 | 完整的问题数据 |
| `session_end` | 会话结束通知 | 结束原因（timeout / manual） |
| `heartbeat` | 心跳保活 | 无 |

### SSE 连接管理

- 浏览器通过 `GET /api/v1/qa/session/:id/stream` 建立 SSE 连接
- 后端维持 SSE 连接，当有新问题时推送 `question` 事件
- 浏览器原生支持断线自动重连
- Session 状态变为 expired/ended 时，推送 `session_end` 事件并关闭连接

## 数据库模型（逻辑模型）

### Session 表（参考值）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | 雪花 ID | 主键 |
| title | varchar(255) | 会话标题 |
| description | text | 会话描述 |
| status | varchar(16) | 状态（active / expired / ended） |
| connect_url | varchar(512) | 浏览器连接 URL |
| expires_at | timestamp | 过期时间 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 问题表（参考值）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | 雪花 ID | 主键 |
| session_id | 雪花 ID | 关联会话 |
| type | varchar(16) | 类型（option / text / mixed） |
| title | varchar(255) | 问题标题 |
| description | text | 问题描述（Markdown） |
| options | jsonb | 选项配置 |
| panels | jsonb | 高级面板配置 |
| batch | jsonb | 分批信息 |
| status | varchar(16) | 状态（pending / answered / skipped） |
| answer | jsonb | 用户回答数据 |
| created_at | timestamp | 创建时间 |

## Redis 使用（参考值）

| Key 模式 | 用途 | TTL |
|----------|------|-----|
| `qa:session:{id}:status` | Session 状态缓存 | 与 Session 过期时间一致 |
| `qa:session:{id}:ttl` | 超时倒计时 | 与 Session 过期时间一致 |
