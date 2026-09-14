---
name: lumina-pages
description: Lumina Pages 把核对完成的 Preview 草稿晋升为项目级不可变快照，并支持 Fork 后再发新版本。用户确认把原型做成可分享的路径式页面时使用。不要在晋升时设置密码。
license: MIT
compatibility: Requires Lumina MCP (Streamable HTTP) and network access to the Lumina instance.
metadata:
  author: lumina
  version: "0.1.0"
argument-hint: [ session-id | slug | page-id ]
allowed-tools: Read, Write, Edit, Bash, AskUserQuestion, mcp__lumina__project_get, mcp__lumina__project_list, mcp__lumina__preview_file_list, mcp__lumina__preview_file_upload, mcp__lumina__preview_file_get, mcp__lumina__pages_list, mcp__lumina__pages_promote, mcp__lumina__pages_fork, mcp__plugin_lumina_lumina__project_get, mcp__plugin_lumina_lumina__project_list, mcp__plugin_lumina_lumina__preview_file_list, mcp__plugin_lumina_lumina__preview_file_upload, mcp__plugin_lumina_lumina__preview_file_get, mcp__plugin_lumina_lumina__pages_list, mcp__plugin_lumina_lumina__pages_promote, mcp__plugin_lumina_lumina__pages_fork
---

# Lumina 持久即时页面晋升与版本迭代指南 (lumina-pages)

把已经核对过的 Preview 草稿发布成项目级、路径式、不可变快照；需要继续改时先 Fork 成新草稿，再晋升新版本。

项目解析见 [`../_shared/project-resolver.md`](../_shared/project-resolver.md)。Preview 文件约束见 [`../lumina-preview/references/file-layout.md`](../lumina-preview/references/file-layout.md)。

## 按需加载

| 需要 | 读取 |
|---|---|
| 晋升 / Fork / OCC 字段 | [`references/promote.md`](./references/promote.md) |
| 从草稿到线上走查 | [`examples/promote-and-fork.md`](./examples/promote-and-fork.md) |

---

## 核心定位

- **Preview** 是 7 天 TTL 的登录草稿工作台。
- **Pages** 是项目级持久快照：`/pages/<project_name>/<slug>/<file>`，可选密码，密码只在控制台 `/console/pages` 配置。
- MCP 不设置密码，也不改访问策略。

## 标准流程

```text
preview_file_list 核对 HTML 入口
        │
        ▼
pages_list(project_id) 确认 slug
        │
        ▼
pages_promote(session_id, slug, title)
        │
        ▼ 用户要继续改
pages_fork(page_id) → 新 Preview 草稿
        │
        ▼
preview_file_upload 迭代 → pages_promote（slug 可省略）
```

### 新建页面

1. Preview 文件齐全且 `workflow.state` 为 `ready_for_review`。
2. `pages_list` 看 slug 是否占用。
3. `pages_promote`：
   ```json
   {
     "session_id": "<preview session id>",
     "slug": "design-system",
     "title": "设计系统",
     "set_as_active": true
   }
   ```
4. 用返回的 `page.page_url` 打开路径式页面。无 GUI 时才把 URL 交给用户。

### 基于线上页面继续改

1. `pages_fork({ "page_id": "<id>" })`，默认拷贝生效版本。
2. 返回会话带 `source_page_id` / `source_page_slug`。用 `preview_file_get` / `preview_file_upload` 改草稿。
3. 再 `pages_promote`。Fork 会话 **不必再传 slug**，后端锁定来源页面并递增 patch（`v1.0.0` → `v1.1.0`）。
4. 若返回 `conflict`，向用户说明线上已有更新，确认后再带 `confirm_conflict: true`。

## 红线

1. **MUST**: 无 HTML 入口不得晋升。
2. **MUST**: 晋升前用 `pages_list` 核对 slug；新建必须给合法 slug（小写字母、数字、短横线）。
3. **NEVER**: 不要在 MCP 调用里传密码或访问策略。
4. **NEVER**: 不要把 Pages 当成热更新工作台；改内容先 Fork。
5. **MUST**: 独立打开 `page_url` / Fork 后的 `preview_url` 时，立即用系统命令弹出浏览器。
