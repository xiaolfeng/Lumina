# RepoWiki 模块

## 概述

RepoWiki 是 Lumina 的代码库结构化知识沉淀模块。它通过 Git 克隆整个项目，利用 5 个 Agent 角色协作分析代码库，最终生成结构化的 Wiki 文档。

Agent 可通过 MCP 只读工具查询已生成的 Wiki 内容，但 Wiki 的生成和更新由 Lumina 内部流水线与 Git Webhook 自动触发，Agent 无需也无法通过 MCP 触发更新。

## 核心能力

- **Git 克隆**：支持任意 Git 仓库 URL，克隆到本地进行分析；支持 SSH 密钥和私有仓库
- **5 角色 SubAgent 编排**：
  - **Coordinator**：项目整体概要分析
  - **Explore**：并发代码探索
  - **Architect**：构建 Wiki 目录大纲
  - **Writer**：并发文档撰写
  - **Validator**：完整性校验与失败重试
- **版本隔离存储**：每个 Wiki 版本拥有独立的文件系统目录，支持多版本共存和切换
- **Webhook 自动更新**：Git Provider 推送事件自动触发新版本的 Wiki 生成
- **多语言支持**：支持中英文等多种语言的 Wiki 生成
- **MCP 只读访问**：Agent 可通过 `repoWiki_query` / `repoWiki_list` 读取已完成的 Wiki

## 存储策略

- **PostgreSQL**：项目元数据、RepoWikiConfig、WikiVersion 状态与版本信息
- **文件系统**：版本隔离的 Wiki 文档内容（`versions/{version_id}/wiki/`）
- **版本切换**：通过 `RepoWikiConfig.SelectedVersionID` 指定当前对外服务的版本

## 相关文档

| 文档 | 说明 |
|------|------|
| [design.md](design.md) | 详细设计（核心流程、5 角色编排、存储结构、增量更新策略） |
| [multi-agent-design.md](multi-agent-design.md) | 5 角色 SubAgent 编排设计细节 |
| [webhook-design.md](webhook-design.md) | Webhook 更新机制设计（Git Provider 接入、签名验证、处理流程） |
| [mcp-tools.md](mcp-tools.md) | MCP 工具定义与 REST API 接口列表 |

> ⚠️ 本文档为设计参考方案，并非最终决策。实际实现可能调整。
