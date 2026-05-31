# RepoWiki MCP 工具与 REST API

> **注意**：以下所有工具名称、参数、API 路径均为设计参考值，实际实现可能调整。

## MCP 工具定义

### repoWiki_analyze

克隆项目并生成 Wiki 文档。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `git_url` | string | 是 | Git 仓库地址 |
| `language` | string | 否 | Wiki 语言，默认 `zh` |
| `branch` | string | 否 | 目标分支，默认仓库默认分支 |

**返回值**：项目 ID、分析状态、预计耗时

---

### repoWiki_query

查询已有 Wiki 内容。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `project_id` | string | 是 | 项目 ID |
| `keyword` | string | 否 | 关键词搜索 |
| `page` | string | 否 | 指定 Wiki 页面路径（如 `content/架构设计.md`） |

**返回值**：Wiki 页面内容（Markdown）或搜索结果列表

---

### repoWiki_update

增量更新指定项目的 Wiki。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `project_id` | string | 是 | 项目 ID |

**返回值**：更新状态、变更文件数、重新生成的页面列表

---

### repoWiki_list

列出所有已分析的项目。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 页码，默认 1 |
| `size` | int | 否 | 每页数量，默认 20 |

**返回值**：项目列表（ID、名称、Git URL、状态、最后更新时间）

---

### repoWiki_delete

删除项目及其 Wiki 文档。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `project_id` | string | 是 | 项目 ID |

**返回值**：删除确认

---

## REST API 定义

### POST `/api/v1/repowiki/analyze`

触发项目分析。

**请求体（参考值）**：
```json
{
  "git_url": "https://github.com/org/repo",
  "language": "zh",
  "branch": "main"
}
```

**响应**：
```json
{
  "project_id": "123456",
  "status": "analyzing"
}
```

---

### GET `/api/v1/repowiki/list`

获取项目列表。

**查询参数**：`page`、`size`

**响应**：
```json
{
  "items": [
    {
      "id": "123456",
      "name": "repo-name",
      "git_url": "https://github.com/org/repo",
      "status": "completed",
      "updated_at": "2026-05-31T00:00:00Z"
    }
  ],
  "total": 1
}
```

---

### GET `/api/v1/repowiki/:id`

获取项目详情。

**响应**：
```json
{
  "id": "123456",
  "name": "repo-name",
  "git_url": "https://github.com/org/repo",
  "commit_hash": "abc1234",
  "status": "completed",
  "language": "zh",
  "created_at": "2026-05-31T00:00:00Z",
  "updated_at": "2026-05-31T00:00:00Z"
}
```

---

### GET `/api/v1/repowiki/:id/wiki/:path`

获取 Wiki 文档内容。

**路径参数**：`id`（项目 ID）、`path`（文档路径，如 `content/架构设计.md`）

**响应**：
```json
{
  "title": "架构设计",
  "content": "# 架构设计\n\n...",
  "path": "content/架构设计.md",
  "language": "zh"
}
```

---

### PUT `/api/v1/repowiki/:id/update`

触发增量更新。

**响应**：
```json
{
  "status": "analyzing",
  "changed_files": 5,
  "updated_pages": ["content/模块分析.md"]
}
```

---

### DELETE `/api/v1/repowiki/:id`

删除项目及其 Wiki。

**响应**：
```json
{
  "message": "已删除"
}
```
