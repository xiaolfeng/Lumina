# Memory MCP 工具与 REST API

> **注意**：以下所有工具名称、参数、API 路径均为设计参考值，实际实现可能调整。

## MCP 工具定义

### memory_create

创建一条记忆决策卡片。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | 是 | 决策标题 |
| `content` | string | 是 | 详细内容（Markdown 格式） |
| `reason` | string | 否 | 决策原因 |
| `scope` | string | 否 | 适用范围 |
| `priority` | string | 是 | 优先级：`high` / `medium` / `low` |
| `tags` | string[] | 否 | 标签列表 |
| `source_project` | string | 否 | 关联项目 ID |

**返回值**：创建的记忆 ID、完整卡片数据

---

### memory_query

按条件检索记忆。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyword` | string | 否 | 关键词搜索（标题 + 内容） |
| `tags` | string[] | 否 | 标签过滤（多个标签取交集） |
| `priority` | string | 否 | 优先级过滤 |
| `status` | string | 否 | 状态过滤，默认 `active` |
| `source_project` | string | 否 | 按关联项目过滤 |
| `page` | int | 否 | 页码，默认 1 |
| `size` | int | 否 | 每页数量，默认 20 |

**返回值**：记忆卡片列表（含分页信息）

---

### memory_get

获取单条记忆详情。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | 记忆 ID |

**返回值**：完整的记忆卡片数据

---

### memory_update

更新已有记忆。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | 记忆 ID |
| `title` | string | 否 | 新标题 |
| `content` | string | 否 | 新内容 |
| `reason` | string | 否 | 新原因 |
| `scope` | string | 否 | 新范围 |
| `priority` | string | 否 | 新优先级 |
| `tags` | string[] | 否 | 新标签列表（全量替换） |
| `status` | string | 否 | 新状态（`active` / `deprecated` / `archived`） |

**返回值**：更新后的完整卡片数据

---

### memory_delete

删除记忆（物理删除）。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | 记忆 ID |

**返回值**：删除确认

---

## REST API 定义

### POST `/api/v1/memory/`

创建记忆。

**请求体（参考值）**：
```json
{
  "title": "采用严格分层架构",
  "content": "## 详细说明\n\n本项目采用 route→handler→logic→repository 四层架构...",
  "reason": "确保代码职责清晰",
  "scope": "所有新增接口",
  "priority": "high",
  "tags": ["架构", "规范"],
  "source_project": "123456"
}
```

**响应**：
```json
{
  "id": "789012",
  "title": "采用严格分层架构",
  "status": "active",
  "created_at": "2026-05-31T00:00:00Z"
}
```

---

### GET `/api/v1/memory/`

检索记忆列表。

**查询参数**：`keyword`、`tags`（逗号分隔）、`priority`、`status`、`source_project`、`page`、`size`

**响应**：
```json
{
  "items": [
    {
      "id": "789012",
      "title": "采用严格分层架构",
      "priority": "high",
      "tags": ["架构", "规范"],
      "status": "active",
      "updated_at": "2026-05-31T00:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "size": 20
}
```

---

### GET `/api/v1/memory/:id`

获取记忆详情。

**响应**：
```json
{
  "id": "789012",
  "title": "采用严格分层架构",
  "content": "## 详细说明\n\n...",
  "reason": "确保代码职责清晰",
  "scope": "所有新增接口",
  "priority": "high",
  "tags": ["架构", "规范"],
  "status": "active",
  "source_project": "123456",
  "created_at": "2026-05-31T00:00:00Z",
  "updated_at": "2026-05-31T00:00:00Z"
}
```

---

### PUT `/api/v1/memory/:id`

更新记忆。

**请求体**：与创建相同，所有字段可选。

**响应**：更新后的完整卡片数据。

---

### DELETE `/api/v1/memory/:id`

删除记忆。

**响应**：
```json
{
  "message": "已删除"
}
```

---

### GET `/api/v1/memory/tags`

获取所有标签列表。

**响应**：
```json
{
  "tags": [
    { "id": "1", "name": "架构", "count": 5 },
    { "id": "2", "name": "规范", "count": 3 }
  ]
}
```
