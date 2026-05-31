# Lumina · 微明 — 项目设计文档

> **烛照幽微，知常曰明**
> 赋予 AI 深度代码认知与长期记忆的知识中枢

## 文档结构

本目录包含 Lumina 项目的完整设计文档，按模块分层组织：

```
docs/wiki/
├── README.md              ← 你在这里（文档总览与导航）
├── architecture.md        ← 整体架构设计
├── infrastructure.md      ← 基础设施层说明
├── repowiki/              ← RepoWiki 模块文档
│   ├── README.md          ← 模块概述
│   ├── design.md          ← 详细设计
│   └── mcp-tools.md       ← MCP 工具与 REST API 定义
├── memory/                ← Memory 模块文档
│   ├── README.md          ← 模块概述
│   ├── design.md          ← 详细设计
│   └── mcp-tools.md       ← MCP 工具与 REST API 定义
└── qa/                    ← Q&A 模块文档
    ├── README.md          ← 模块概述
    ├── design.md          ← 详细设计
    └── mcp-tools.md       ← MCP 工具与 REST API 定义
```

## 核心模块

| 模块 | 说明 | 文档入口 |
|------|------|----------|
| 🔍 **RepoWiki** | 代码库结构化知识沉淀，克隆项目并通过 LLM 分析生成 Wiki | [repowiki/README.md](repowiki/README.md) |
| 🧠 **Memory** | AI 的长期决策记忆，跨会话保留重要约定与决策 | [memory/README.md](memory/README.md) |
| 💬 **Q&A** | Agent 与用户的富交互式问答通道，支持选项、文本、高级面板 | [qa/README.md](qa/README.md) |

## 全局文档

| 文档 | 说明 |
|------|------|
| [architecture.md](architecture.md) | 整体架构设计（模块划分、分层原则、对外暴露方式） |
| [infrastructure.md](infrastructure.md) | 基础设施层（PostgreSQL、Redis、LLM Provider、环境变量） |

## ⚠️ 文档性质声明

**本文档目录下的所有内容均为设计阶段的参考方案，并非最终决策。**

- MCP 工具名称、参数、REST API 路径、数据结构等均为**设计参考值**
- 实际实现时可根据技术约束和开发决策进行调整
- 文档中的逻辑模型不等同于最终物理表结构
- 三个模块的设计方案可能在实现过程中被修改或替代

本文档的价值在于**对齐设计意图和方向**，而非锁定实现细节。

## 设计约定

- 所有文档使用**简体中文**
- 三个模块完全独立，不直接交互，Agent 通过 MCP 自行编排
- 每个模块按 README（概述）→ design（详细设计）→ mcp-tools（接口定义）三层组织
