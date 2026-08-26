# RepoWiki 查询约定

## `wiki_id`

来自 `repoWiki_list` 的 `version_id`，数值型雪花 ID。不要传项目名或 commit。

## 状态

只有 `completed` 的版本能 `repoWiki_query`。分析中或失败的版本会报错，换一个已完成版本，不要找「analyze」类 MCP 工具——没有。

## `page`

相对路径，**不要**带 `.mdx` / `.md`：

| 想读 | 传入 |
|---|---|
| 首页 / 总览 | 省略 `page` |
| 架构章 | `content/architecture` |
| 中文路径 | `content/架构设计` |

Wiki 产物是 `.mdx` + YAML frontmatter；旧版 `.md` 已不兼容。路径以 `repoWiki_query` 首页目录为准，不要猜文件系统绝对路径。

## `query`

关键词搜索是预留字段，当前版本尚未启用。需要定位章节时先读首页目录，再带 `page` 拉正文。
