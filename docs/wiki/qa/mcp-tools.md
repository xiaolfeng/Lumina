# Q&A MCP 工具与 REST API

> **注意**：以下所有工具名称、参数、API 路径均为设计参考值，实际实现可能调整。

## MCP 工具定义

### qa_session_create

创建问答会话，返回 session ID 和浏览器连接 URL。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | 是 | 会话标题 |
| `description` | string | 否 | 会话描述 |

**返回值**：

```json
{
  "session_id": "123456",
  "connect_url": "http://localhost:8080/qa/session/123456",
  "expires_at": "2026-06-07T00:00:00Z"
}
```

---

### qa_push_question

向指定会话推送问题。推送后通过 SSE 实时发送到浏览器。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `session_id` | string | 是 | 会话 ID |
| `type` | string | 是 | 问题类型：`option` / `text` / `mixed` |
| `title` | string | 是 | 问题标题 |
| `description` | string | 否 | 问题描述（Markdown） |
| `placeholder` | string | 否 | 输入提示文本（文本题） |
| `options` | array | 否 | 选项列表（选项题/混合题） |
| `allow_extra_text` | bool | 否 | 是否允许补充文本 |
| `panels` | array | 否 | 高级面板列表 |
| `batch` | object | 否 | 分批信息（group_id, sequence, total） |

**选项对象结构**：

```json
{
  "label": "选项显示文本",
  "value": "选项值",
  "allow_other": false
}
```

**面板对象结构**：

```json
{
  "type": "markdown | html_preview",
  "content": "面板内容",
  "position": "right"
}
```

**分批对象结构**：

```json
{
  "group_id": "batch-001",
  "sequence": 1,
  "total": 3
}
```

**返回值**：问题 ID、推送状态

---

### qa_get_answer

获取用户对指定问题的回答。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `session_id` | string | 是 | 会话 ID |
| `question_id` | string | 是 | 问题 ID |

**返回值**：

```json
{
  "question_id": "789012",
  "status": "answered | pending | skipped",
  "answer": {
    "value": "gin",
    "extra_text": "补充说明文本",
    "answered_at": "2026-05-31T12:00:00Z"
  }
}
```

当 `status` 为 `pending` 时，表示用户尚未回答，Agent 可稍后重试。

---

### qa_get_session

获取会话状态信息。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `session_id` | string | 是 | 会话 ID |

**返回值**：

```json
{
  "session_id": "123456",
  "title": "技术选型讨论",
  "status": "active",
  "expires_at": "2026-06-07T00:00:00Z",
  "questions": [
    {
      "id": "789012",
      "title": "选择技术栈",
      "status": "answered"
    }
  ],
  "total_questions": 1,
  "answered_count": 1
}
```

---

### qa_session_end

手动结束会话。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `session_id` | string | 是 | 会话 ID |

**返回值**：结束确认

---

## REST API 定义

### GET `/api/v1/qa/session/:id`

获取会话信息与问题列表。

**响应**：
```json
{
  "id": "123456",
  "title": "技术选型讨论",
  "description": "讨论项目的技术选型",
  "status": "active",
  "expires_at": "2026-06-07T00:00:00Z",
  "questions": [
    {
      "id": "789012",
      "type": "option",
      "title": "选择技术栈",
      "description": "请选择项目的后端框架",
      "options": [
        { "label": "Gin", "value": "gin", "allow_other": false },
        { "label": "其他", "value": "other", "allow_other": true }
      ],
      "panels": [],
      "status": "pending",
      "batch": null
    }
  ]
}
```

---

### GET `/api/v1/qa/session/:id/stream`

SSE 连接端点。建立后，后端实时推送新问题到浏览器。

**SSE 事件格式**：

```
event: question
data: {"id":"789012","type":"option","title":"选择技术栈",...}

event: session_end
data: {"reason":"manual"}

event: heartbeat
data: {}
```

**行为说明**：
- 浏览器原生支持断线自动重连
- Session 状态变为 expired/ended 时推送 `session_end` 并关闭连接
- 定期发送 `heartbeat` 保活

---

### POST `/api/v1/qa/session/:id/question/:qid/answer`

用户提交回答。

**请求体（参考值）**：
```json
{
  "value": "gin",
  "extra_text": "考虑社区活跃度和性能表现"
}
```

**响应**：
```json
{
  "message": "回答已提交",
  "question_id": "789012",
  "answered_at": "2026-05-31T12:00:00Z"
}
```

---

### GET `/api/v1/qa/sessions`

列出当前会话列表。

**查询参数**：`page`、`size`、`status`（可选过滤）

**响应**：
```json
{
  "items": [
    {
      "id": "123456",
      "title": "技术选型讨论",
      "status": "active",
      "expires_at": "2026-06-07T00:00:00Z",
      "total_questions": 3,
      "answered_count": 1
    }
  ],
  "total": 1
}
```
