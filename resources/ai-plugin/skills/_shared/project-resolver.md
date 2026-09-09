# Lumina 项目上下文解析标准 (Project Resolution Standard)

所有依赖项目上下文的业务模块（Q&A 会话、Preview 预览会话等）均需要关联一个有效的 `project_id`（雪花 ID 字符串）。为了避免项目重复膨胀与无效创建，必须严格遵循以下标准流程进行项目解析。

---

## 🎯 标准解析流

```text
① workspace_list / workspace_get 选定空间
       │
       ▼
② 获取项目路径 (pwd)
       │
       ▼
③ project_get(match_path=pwd, workspace_id=...)
       │─── 命中 ───▶ [获取 project_id 结束]
       │
       ▼ 未直接命中
④ project_list(match_path=pwd, workspace_id=...)
       │─── 找到匹配项目 ───▶ [获取 project_id 结束]
       │
       ▼ 无任何匹配
⑤ project_create(name, match_path=[pwd], workspace_id=...)  [仅在当前任务明确需要时创建]
```

---

## 📋 详细操作步骤

### 第零步：选定工作空间
先调用 `workspace_list`，或用已知 slug（默认空间为 `default`）调用 `workspace_get`。后续 `project_get` / `project_list` / `project_create` 在不以 `project_id` 查询时必须带 `workspace_id` 或 `workspace_slug`。

### 第一步：获取当前项目路径
执行命令（如 Unix/macOS 下的 `pwd` 或获取当前项目路径根），确保拿到的是**标准化绝对路径**（例如 `/Users/username/projects/my-app`）。

### 第二步：优先调用 `project_get`
使用获取到的绝对路径作为 `match_path`，并带上当前空间：
```json
{
  "match_path": "/Users/username/projects/my-app",
  "workspace_id": "1234567890123456789"
}
```
- **匹配机制**：若已有项目的 `MatchPath` 列表中任一前缀匹配当前路径（例如项目配置了 `/Users/username/projects`），则直接命中并返回完整项目信息。
- **命中动作**：提取返回的 `project_id`（或 `id`），直接用于后续 Q&A 或 Preview 操作。

### 第三步：未命中时调用 `project_list` 过滤
若 `project_get` 未命中，调用 `project_list` 并传入 `match_path` 与空间：
```json
{
  "match_path": "/Users/username/projects/my-app",
  "workspace_id": "1234567890123456789"
}
```
- **匹配机制**：应用层双向前缀匹配（已有项目前缀包含当前路径，或当前路径包含已有项目前缀）。
- **命中动作**：从列表中选取最匹配的已有项目，获取其 `project_id`。

### 第四步：谨慎创建新项目 (`project_create`)
**仅当**前两步均未找到已有项目，且**当前用户任务确实需要**建立长期会话或预览会话时，才调用 `project_create` 注册项目：
```json
{
  "name": "my-app",
  "match_path": ["/Users/username/projects/my-app"],
  "alias_name": "我的应用",
  "description": "项目简要说明",
  "workspace_id": "1234567890123456789"
}
```

---

## ⛔ 核心红线 (Hard Rules)

1. **MUST**: 优先查后建。严禁在未经 `project_get` / `project_list` 探测前直接调用 `project_create`。
2. **MUST**: `match_path` 必须为绝对路径数组，不得传入相对路径（如 `./` 或 `../`）。
3. **MUST**: 用 `name` / `match_path` 解析或列出项目时必须带空间；不要假设实例里只有一个空间。
4. **NEVER**: 严禁因为工具存在就无理由自动创建项目。只有当用户明确要求或任务确实需要持久化存储时才触发。
