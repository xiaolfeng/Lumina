# Q&A 模块

## 概述

Q&A 是 Lumina 的 Agent 与用户之间的富交互式问答通道。Agent 通过 MCP 创建会话并推送问题，用户在浏览器中通过 SSE 实时接收问题并进行回答。

支持选项选择、文本输入、分批推送、高级面板（Markdown 渲染、HTML/JS/CSS 即时预览）等富交互模式。

## 核心能力

- **会话管理**：创建持久会话，返回浏览器连接 URL，支持超时自动归档
- **问题推送**：支持选项题、文本题、混合题，可分批推送
- **富交互**：问题可附带高级面板（Markdown / HTML+JS+CSS 即时预览）
- **答案收集**：用户提交回答后，Agent 通过 MCP 获取
- **SSE 实时推送**：基于 Server-Sent Events 向浏览器推送问题

## 存储策略

- **PostgreSQL**：Session 元数据、问题完整数据（含选项、面板、答案）
- **Redis**：Session 状态缓存、超时倒计时

## 相关文档

| 文档 | 说明 |
|------|------|
| [design.md](design.md) | 详细设计（Session 生命周期、问题类型、通信架构、数据结构） |
| [mcp-tools.md](mcp-tools.md) | MCP 工具定义与 REST API 接口列表 |

> ⚠️ 本文档为设计参考方案，并非最终决策。实际实现可能调整。

## 通信方式

Q&A 模块采用 **SSE（Server-Sent Events）** 进行后端到浏览器的实时推送：

- 后端通过 SSE 向浏览器推送新问题
- 用户通过 POST 接口提交回答
- Agent 通过 MCP 获取用户回答

选择 SSE 而非 WebSocket 的原因：单向推送场景足够，原生 HTTP 协议，浏览器自动断线重连，实现更轻量。
