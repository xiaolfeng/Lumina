# Pin MCP 工具与 REST API

> **注意**：以下所有工具名称、参数、API 路径均为设计参考值，实际实现可能调整。

## MCP 工具定义

### pin_push

创建并向目标项目推送一条约束（Pin）。点对点定向推送（from → to）。

`to_project_name` 的匹配逻辑：优先通过 `alias_name` JSON 数组进行模糊包含匹配（不区分大小写），若未命中则回退到精确的 `project_id` 匹配。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | 是 | 约束标题 |
| `content` | string | 是 | 详细内容（Markdown 格式） |
| `category` | string | 否 | 分类：`notice` / `dependency` / `api_change` / `other`，默认 `notice` |
| `priority` | string | 是 | 优先级：`high` / `medium` / `low` |
| `from_project_id` | string | 否 | 来源项目 ID |
| `to_project_name` | string | 是 | 目标项目标识（通过 `alias_name` 模糊匹配或 `project_id` 精确匹配） |
| `expire_days` | int | 否 | 超时天数，不传则使用系统默认配置 `pin_expire_days` |

**返回值**：创建的 Pin ID、完整数据

---

### pin_get

消费指定项目下最早的一条待处理 Pin。按 FIFO 顺序读取，读取后自动标记为已消费。这是主要的消费工具。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `project_name` | string | 是 | 项目标识（当前项目） |
| `category` | string | 否 | 按分类筛选，不传返回所有分类 |
| `priority` | string | 否 | 按优先级筛选，不传返回所有优先级 |
| `force` | bool | 否 | 强制重读已消费 Pin，默认 `false` |

**返回值**：最早的一条待处理 Pin 数据；无待处理 Pin 时返回 `null`

**消费行为说明**：

- 当 `force=false`（默认）：读取最早的待处理 Pin → 自动标记为已消费 → 返回数据。若无待处理 Pin，返回 `null` 并附带提示 `"暂无待处理约束"`。
- 当 `force=true`：读取最近一条 Pin（无论待处理或已消费），不修改状态。若该 Pin 已被消费，返回数据中附带警告提示 `"⚠️ 此 Pin 已被消费（消费时间：{consumed_at}），使用 force 模式重新查看。状态不变。"`
- 已过期 Pin（当前时间 - `created_at` > `expire_days`）：`pin_get` 默认不返回过期 Pin；过期 Pin 仅通过 `pin_list` 展示并附带状态提示。

---

### pin_list

列出指定项目的 Pin，支持过滤、分页和基于索引的跳过。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `project_name` | string | 是 | 项目标识 |
| `status` | string | 否 | 状态过滤：`pending` / `consumed`，不传返回所有 |
| `category` | string | 否 | 按分类筛选 |
| `priority` | string | 否 | 按优先级筛选 |
| `from_project_id` | string | 否 | 按来源项目筛选 |
| `page` | int | 否 | 页码，默认 1 |
| `size` | int | 否 | 每页数量，默认 20 |

**返回值**：Pin 列表（含分页信息），每条数据包含过期状态提示

**过期状态提示**：列表中的过期 Pin 会附带 `"expired": true` 标记。Agent 应检查该标记来决定是否需要通过 `force` 模式重新读取。

---

### pin_update

更新已有 Pin 的元数据（仅支持状态、优先级、分类，`content` 字段推送后不可变更）。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | Pin ID |
| `status` | string | 否 | 状态：`pending` / `consumed` |
| `priority` | string | 否 | 优先级 |
| `category` | string | 否 | 分类 |

**返回值**：更新后的完整 Pin 数据

**说明**：`content` 字段在推送后不可更新，以确保约束内容的完整性。仅允许修改元数据字段。

---

## REST API 定义

### POST `/api/v1/project/:id/pins`

在指定项目下创建 Pin（Web 管理端）。

**请求体（参考值）**：
```json
{
  "title": "升级 Go 版本至 1.25",
  "content": "## 背景\n\n上游依赖要求 Go 1.25+...",
  "category": "dependency",
  "priority": "high",
  "from_project_id": "proj_123",
  "expire_days": 30
}
```

**响应**：
```json
{
  "id": "pin_456",
  "title": "升级 Go 版本至 1.25",
  "category": "dependency",
  "priority": "high",
  "status": "pending",
  "project_id": "proj_789",
  "from_project_id": "proj_123",
  "expire_at": "2026-07-09T00:00:00Z",
  "created_at": "2026-06-09T00:00:00Z"
}
```

---

### GET `/api/v1/project/:id/pins`

列出指定项目下的 Pin，支持过滤和分页。

**查询参数**：`status`、`category`、`priority`、`from_project_id`、`page`、`size`

**响应**：
```json
{
  "items": [
    {
      "id": "pin_456",
      "title": "升级 Go 版本至 1.25",
      "category": "dependency",
      "priority": "high",
      "status": "pending",
      "from_project_id": "proj_123",
      "expired": false,
      "created_at": "2026-06-09T00:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "size": 20
}
```

---

### GET `/api/v1/project/:id/pins/:pin_id`

获取指定 Pin 的详情。

**响应**：
```json
{
  "id": "pin_456",
  "title": "升级 Go 版本至 1.25",
  "content": "## 背景\n\n上游依赖要求 Go 1.25+...",
  "category": "dependency",
  "priority": "high",
  "status": "pending",
  "project_id": "proj_789",
  "from_project_id": "proj_123",
  "consumed_at": null,
  "expire_at": "2026-07-09T00:00:00Z",
  "created_at": "2026-06-09T00:00:00Z",
  "updated_at": "2026-06-09T00:00:00Z"
}
```

---

### PUT `/api/v1/project/:id/pins/:pin_id`

更新 Pin 元数据（状态、优先级、分类）。

**请求体（参考值）**：
```json
{
  "status": "consumed",
  "priority": "medium",
  "category": "notice"
}
```

**响应**：更新后的完整 Pin 数据。

---

### DELETE `/api/v1/project/:id/pins/:pin_id`

删除指定 Pin（仅 Web 管理端可用）。

**响应**：
```json
{
  "message": "已删除"
}
```

---

### POST `/api/v1/project`

创建项目。

**请求体（参考值）**：
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
  "created_at": "2026-06-09T00:00:00Z"
}
```

---

### GET `/api/v1/project`

列出所有项目。

**响应**：
```json
{
  "items": [
    {
      "id": "proj_789",
      "name": "Lumina",
      "alias_name": ["lumina", "微明"],
      "description": "AI 知识中枢后端服务",
      "pin_count": 5,
      "created_at": "2026-06-09T00:00:00Z"
    }
  ],
  "total": 1
}
```

---

### GET `/api/v1/project/:id`

获取项目详情。

**响应**：
```json
{
  "id": "proj_789",
  "name": "Lumina",
  "alias_name": ["lumina", "微明"],
  "description": "AI 知识中枢后端服务",
  "pin_count": 5,
  "created_at": "2026-06-09T00:00:00Z",
  "updated_at": "2026-06-09T00:00:00Z"
}
```

---

### PUT `/api/v1/project/:id`

更新项目信息。

**请求体（参考值）**：
```json
{
  "name": "Lumina",
  "alias_name": ["lumina", "微明", "lumina-server"],
  "description": "AI 知识中枢后端服务（已更新）"
}
```

**响应**：更新后的完整项目数据。

---

### DELETE `/api/v1/project/:id`

删除项目（同步删除该项目下的所有 Pin）。

**响应**：
```json
{
  "message": "已删除"
}
```

---

### GET `/api/v1/settings/pin`

获取 Pin 模块系统配置。

**响应**：
```json
{
  "pin_expire_days": 30
}
```

---

### PUT `/api/v1/settings/pin`

更新 Pin 模块系统配置。

**请求体（参考值）**：
```json
{
  "pin_expire_days": 60
}
```

**响应**：更新后的完整配置数据。
