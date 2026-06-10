# Project MCP 工具与 REST API

> **注意**：以下所有工具名称、参数、API 路径均为设计参考值，实际实现可能调整。

## MCP 工具定义

### project_list

列出所有项目，支持分页。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 页码，默认 1 |
| `size` | int | 否 | 每页数量，默认 20 |

**返回值**：`ProjectListResponse`
```json
{
  "items": [
    {
      "id": "proj_789",
      "name": "Lumina",
      "alias_name": ["lumina", "微明"],
      "description": "AI 知识中枢后端服务",
      "created_at": "2026-06-09T00:00:00+08:00",
      "updated_at": "2026-06-09T00:00:00+08:00"
    }
  ],
  "total": 1
}
```

---

### project_get

获取项目详情。通过项目 ID 或项目名称（精确匹配）查询。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 否* | 项目 ID（与 project_name 二选一） |
| `project_name` | string | 否* | 项目名称精确匹配（与 id 二选一） |

> *`id` 和 `project_name` 至少提供一个，优先使用 `id`。

**返回值**：`ProjectResponse`
```json
{
  "id": "proj_789",
  "name": "Lumina",
  "alias_name": ["lumina", "微明"],
  "description": "AI 知识中枢后端服务",
  "created_at": "2026-06-09T00:00:00+08:00",
  "updated_at": "2026-06-09T00:00:00+08:00"
}
```

---

## REST API 定义

### POST `/api/v1/project`

创建项目（需要认证）。

**请求体**：
```json
{
  "name": "Lumina",
  "alias_name": ["lumina", "微明"],
  "description": "AI 知识中枢后端服务"
}
```

**响应**：
```json
{
  "id": "proj_789",
  "name": "Lumina",
  "alias_name": ["lumina", "微明"],
  "description": "AI 知识中枢后端服务",
  "created_at": "2026-06-09T00:00:00+08:00",
  "updated_at": "2026-06-09T00:00:00+08:00"
}
```

---

### GET `/api/v1/project`

列出所有项目，支持分页（需要认证）。

查询参数：`page`（默认1）、`size`（默认20）

响应：ProjectListResponse

---

### GET `/api/v1/project/:id`

获取项目详情（需要认证）。

响应：ProjectResponse

---

### PUT `/api/v1/project/:id`

更新项目信息（需要认证）。

请求体同 POST，响应为更新后的 ProjectResponse。

---

### DELETE `/api/v1/project/:id`

删除项目（需要认证）。

响应：
```json
{
  "message": "已删除"
}
```
