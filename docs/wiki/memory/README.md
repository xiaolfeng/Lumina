# Memory 模块

## 概述

Memory 是 Lumina 的 AI 长期决策记忆模块。它由 MCP 端主动调用、主动推送构建，用于存储跨会话保留的重要决策、约定和规范。

记忆内容作为**比较重要且后续需要遵循的内容**，与临时上下文不同，Memory 存储的是需要长期遵守的决策性知识。

## 核心能力

- **决策卡片**：结构化的决策记录（标题、内容、原因、适用范围、优先级）
- **Markdown 内容**：详细内容支持 Markdown 渲染（MVP 阶段不支持图片、视频、PDF）
- **标签分类**：多标签体系，支持按标签、优先级、状态条件检索
- **生命周期管理**：active → deprecated → archived 三态流转

## 存储策略

- **PostgreSQL**：记忆卡片完整数据、标签表、多对多关联、全文检索索引
- **Redis（可选）**：热点记忆缓存、标签索引缓存

## 相关文档

| 文档 | 说明 |
|------|------|
| [design.md](design.md) | 详细设计（数据结构、生命周期、检索策略） |
| [mcp-tools.md](mcp-tools.md) | MCP 工具定义与 REST API 接口列表 |

> ⚠️ 本文档为设计参考方案，并非最终决策。实际实现可能调整。

## 与其他模块的关系

Memory 模块完全独立，不与 RepoWiki 或 Q&A 直接交互。Agent 通过 MCP 自行决定何时创建记忆。

典型使用场景：
- Agent 在完成架构决策后，主动调用 `memory_create` 记录决策
- Agent 在新会话开始时，调用 `memory_query` 获取相关历史决策作为上下文
- Agent 发现某决策不再适用，调用 `memory_update` 将状态改为 deprecated
