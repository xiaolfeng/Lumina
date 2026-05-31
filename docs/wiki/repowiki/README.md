# RepoWiki 模块

## 概述

RepoWiki 是 Lumina 的代码库结构化知识沉淀模块。它通过 Git 克隆整个项目，利用 LLM 分阶段分析代码库，最终生成结构化的 Wiki 文档。

Agent 通过 MCP 自主调用 RepoWiki 工具，完成从克隆到 Wiki 生成的完整流程。

## 核心能力

- **Git 克隆**：支持任意 Git 仓库 URL，克隆到本地进行分析
- **LLM 分阶段分析**：概览 → 模块 → 架构 → 阅读指南，逐步深入
- **Wiki 文档生成**：产出 Markdown 文档 + Mermaid 图表 + 元数据 JSON
- **增量更新**：基于 commit hash 差异，仅重新分析变更部分
- **多语言支持**：支持中英文等多种语言的 Wiki 生成

## 存储策略

- **PostgreSQL**：项目元数据（名称、Git URL、commit hash、分析状态、版本记录）
- **文件系统**：Markdown 文档内容（`.lumina/repowiki/{project_id}/{lang}/`）

## 相关文档

| 文档 | 说明 |
|------|------|
| [design.md](design.md) | 详细设计（核心流程、存储结构、增量更新策略、LLM 分析管道） |
| [mcp-tools.md](mcp-tools.md) | MCP 工具定义与 REST API 接口列表 |

> ⚠️ 本文档为设计参考方案，并非最终决策。实际实现可能调整。
