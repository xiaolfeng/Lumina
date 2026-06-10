# Project 模块

## 概述

Project 是 Lumina 的共享聚合根模块，为 Pin 和 Q&A 模块提供项目标识能力。每个项目通过 `alias_name` JSON 数组提供灵活的标识方式，支持 Pin 的模糊匹配和 Q&A 的 Session 挂靠。

## 核心能力

- **项目 CRUD 管理**：创建、列表、详情、更新、删除项目
- **alias_name 别名标识**：通过 JSON 数组提供多个别名，支持模糊匹配
- **REST API + MCP 双通道**：前端通过 REST API 管理，Agent 通过 MCP Tool 查询项目
- **Redis 缓存加速**：项目数据和 alias_name 映射缓存，减少数据库压力

## 存储策略

- **PostgreSQL**：Project 实体完整数据
- **Redis**：项目详情缓存（ID 键）、名称映射缓存（Name 键）、别名映射缓存（Alias 键），TTL 30 分钟

## 相关文档

| 文档 | 说明 |
|------|------|
| [mcp-tools.md](mcp-tools.md) | MCP 工具定义与 REST API 接口列表 |

> ⚠️ 本文档为设计参考方案，并非最终决策。实际实现可能调整。

## 与其他模块的关系

Project 模块作为共享聚合根，为 Pin 和 Q&A 模块提供项目标识：

- **Pin 模块**：通过 `alias_name` 模糊匹配目标项目，支持点对点约束推送
- **Q&A 模块**：Session 通过 `project_id` 挂靠到特定项目

Project 模块不主动调用其他模块，属于纯数据服务。
