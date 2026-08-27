---
name: lumina-repowiki
description: Lumina RepoWiki 只读检索已生成的仓库架构 Wiki。需要了解模块划分、技术栈、业务流程或设计文档时使用。不要尝试用 MCP 触发分析；Wiki 由 Git Webhook 更新。
license: MIT
compatibility: Requires Lumina MCP (Streamable HTTP) and network access to the Lumina instance.
metadata:
  author: lumina
  version: "0.1.1"
argument-hint: [ wiki-id | page-path ]
allowed-tools: Read, Write, Edit, Bash, AskUserQuestion, mcp__lumina__repoWiki_list, mcp__lumina__repoWiki_query, mcp__plugin_lumina_lumina__repoWiki_list, mcp__plugin_lumina_lumina__repoWiki_query
---

# Lumina 代码库架构知识库检索指南 (lumina-repowiki)

用于指导 AI Agent 查询 Lumina 为代码仓库生成的结构化 Wiki 文档，快速建立对陌生仓库架构、核心模块与业务逻辑的深度理解。

页面路径、状态约束与 `query` 预留字段见 [`references/query.md`](./references/query.md)。完整 list→query 走查见 [`examples/list-and-query.md`](./examples/list-and-query.md)。

---

## 🎯 核心定位与设计约束

- **只读定位**：RepoWiki 在 MCP 端为**纯只读**知识库检索工具。
- **自动触发**：Wiki 的生成、分析编排（5 角色 SubAgent）与版本迭代完全由 **Git Webhook** 自动触发与 Cron 定时任务驱动，MCP 端不提供手动创建或写入接口。
- **状态要求**：只有状态为 `completed` 的 Wiki 版本才可正常查询。

---

## 🔄 RepoWiki 标准两步检索流

```text
① repoWiki_list()
       │
       ▼ (获取已完成的 Wiki 版本列表与 version_id)
② repoWiki_query(wiki_id, page="...")
       │
       ▼ (读取首页或指定模块架构页面)
   [获得结构化领域知识]
```

---

## 📋 详细检索操作

### 第一步：列出可用 Wiki 版本 (`repoWiki_list`)
调用 `repoWiki_list` 获取系统中已完成分析的 Wiki 版本清单：
```json
{
  "page": 1,
  "size": 20
}
```
返回结果示例：
```text
Wiki 版本列表（共 1 个，第 1/1 页）：

1. [version_id: 1234567890123456789] Lumina-Backend
   分支: master | 语言: Go | commit: 596abd9
   完成时间: 2026-08-26T10:00:00Z
```

---

### 第二步：查询 Wiki 正文与页面 (`repoWiki_query`)

#### 模式 A：查询 Wiki 首页 (Overview)
不传 `page` 参数，直接返回该版本的 Wiki 首页正文与目录导引：
```json
{
  "wiki_id": 1234567890123456789
}
```

#### 模式 B：查询特定架构章节
传入具体页面的相对路径（不带 `.mdx` 或 `.md` 扩展名）：
```json
{
  "wiki_id": 1234567890123456789,
  "page": "content/architecture"
}
```

---

## ⛔ 核心红线 (Hard Rules)

1. **MUST**: 传入正确的数值型 `wiki_id`（对应 `repoWiki_list` 中的 `version_id`）。
2. **MUST**: `page` 路径不要携带扩展名（例如传 `content/architecture`，严禁传 `content/architecture.mdx`）。
3. **NEVER**: 严禁寻找或尝试通过 MCP 触发 Wiki 分析/写入。Wiki 仅由 Git Webhook 自动驱动。
