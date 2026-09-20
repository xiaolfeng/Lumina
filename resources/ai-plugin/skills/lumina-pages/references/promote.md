# Pages 晋升与 Fork 字段

## 静态运行时继承

晋升会深拷贝 Preview 文件内容与 MIME 类型，因此原生 HTML/CSS/JS 以及浏览器端 React/Vue 页面可保持相同运行方式。Pages 不执行 npm 安装、构建命令、SSR 或服务端代码，也不会把外部 CDN 资源收进快照。框架 CDN 必须固定版本，并确保页面受众所在网络可访问；完整边界见 [`../../lumina-preview/references/framework-runtime.md`](../../lumina-preview/references/framework-runtime.md)。

## pages_promote

| 字段 | 必填 | 说明 |
|---|---|---|
| `session_id` | 是 | Preview 会话雪花 ID |
| `title` | 是 | 页面标题，≤255 |
| `slug` | 新建时是 | `^[a-z0-9-]+$`；Fork 会话可省略 |
| `description` | 否 | 页面描述 |
| `version` | 否 | 空则首次 `v1.0.0`，后续递增 patch |
| `changelog` | 否 | 版本说明 |
| `set_as_active` | 否 | 默认 true，立即切换线上指针 |
| `confirm_conflict` | 否 | OCC：线上指针已前进时必须确认 |

返回：

- 成功：`page` + `version`，含路径式 `page_url`
- 冲突：`conflict.message` 与当前线上版本信息；未确认不改指针

## pages_fork

| 字段 | 必填 | 说明 |
|---|---|---|
| `page_id` | 是 | Pages 雪花 ID |
| `version_id` | 否 | 省略则拷贝当前生效版本 |

返回新 Preview 会话：`source_page_id`、`source_page_slug`、`source_version_id`、路径式 `preview_url`。

## pages_list

按 `project_id` 分页列出 slug、生效版本、`page_url`。晋升前用它确认标识。
