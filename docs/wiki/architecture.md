# 整体架构设计

## 架构总览

Lumina 采用**四模块独立领域 + 统一基础设施层**的架构：

```
┌──────────────────────────────────────────────────────────┐
│                    对外接入层                              │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐ │
│  │ Streamable   │  │  HTTP REST   │  │     SSE        │ │
│  │ MCP Server   │  │  API (Gin)   │  │ (Q&A 实时推送)  │ │
│  └──────┬───────┘  └──────┬───────┘  └───────┬────────┘ │
│         │                 │                  │           │
├─────────┴─────────────────┴──────────────────┴───────────┤
│                    模块层（四大独立领域）                    │
│                                                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌──────────────────┐ │
│  │  RepoWiki   │  │     Pin     │  │   Memory    │  │      Q&A         │ │
│  │             │  │             │  │             │  │                  │ │
│  │ · Git Clone │  │ · 定向推送  │  │ · 决策卡片  │  │ · Session 管理   │ │
│  │ · LLM 分析  │  │ · FIFO 消费 │  │ · 标签分类  │  │ · 问题推送       │ │
│  │ · Wiki 生成 │  │ · 超时提醒  │  │ · Markdown  │  │ · 富交互渲染     │ │
│  │ · 增量同步  │  │ · 项目标识  │  │ · 条件检索  │  │ · 答案收集       │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └────────┬─────────┘ │
│         │                │                │                  │            │
├─────────┴────────────────┴────────────────┴──────────────────┴────────────┤
│                    基础设施层                              │
│  ┌───────────┐  ┌──────────┐  ┌───────────┐             │
│  │ PostgreSQL │  │  Redis   │  │ LLM SDK   │             │
│  │ 元数据/关联 │  │ 缓存/会话│  │ Provider  │             │
│  └───────────┘  └──────────┘  └───────────┘             │
└──────────────────────────────────────────────────────────┘
```

## 核心设计原则

| 原则 | 说明 |
|------|------|
| **四模块独立** | RepoWiki、Pin、Memory、Q&A 各自完整的 handler→logic→repository 分层，互不调用 |
| **MCP 编排** | Agent 通过 MCP 自行组合调用四个模块，Lumina 不做跨模块编排 |
| **双通道暴露** | 每个模块同时提供 REST API（前端用）和 MCP Tool（Agent 用） |
| **统一基础设施** | PostgreSQL、Redis、LLM Provider 三者共享，通过启动阶段注入 context |
| **SSE 实时推送** | Q&A 模块使用 SSE（Server-Sent Events）向后端推送问题到浏览器，用户通过 POST 提交回答 |

## 内部分层架构

每个模块严格遵循项目的分层规范：

```
HTTP Request / MCP Request
    → Route (internal/app/route/)
    → Handler (internal/handler/)
    → Logic (internal/logic/)
    → Repository (internal/repository/)
    → DB / Redis / 文件系统
```

- **Handler** 只负责请求绑定、调用 Logic、映射响应
- **Logic** 负责业务编排、校验、事务边界
- **Repository** 只负责数据持久化和查询
- 禁止跨层调用（如 Handler 直接调用 Repository）

## 对外暴露方式

### MCP Server（Streamable MCP）

通过 Streamable MCP 协议向所有 AI Agent 开放服务能力。每个模块注册独立的工具集：

| 模块 | MCP 工具前缀 | 说明 |
|------|-------------|------|
| RepoWiki | `repoWiki_` | 代码库分析与管理 |
| Pin | `pin_` | 跨项目约束传递 |
| Memory | `memory_` | 决策记忆管理 |
| Q&A | `qa_` | 问答会话管理 |

### HTTP REST API（Gin）

前端界面通过 REST API 与后端通信，路由按模块分组：

| 模块 | REST 前缀 | 说明 |
|------|----------|------|
| RepoWiki | `/api/v1/repowiki/` | 项目分析与 Wiki 浏览 |
| Pin | `/api/v1/pin/` | 跨项目约束管理 |
| Memory | `/api/v1/memory/` | 记忆管理界面 |
| Q&A | `/api/v1/qa/` | 问答界面 + SSE 流 |
| Health | `/api/v1/health/` | 健康检查 |

### SSE 实时通道

Q&A 模块使用 SSE 向浏览器实时推送问题，用户通过 POST 接口提交回答。选择 SSE 而非 WebSocket 的原因：

- Q&A 场景是单向推送（后端→浏览器），不需要双向通信
- 原生 HTTP 协议，无需升级，Gin 直接支持
- 浏览器原生断线自动重连
- 更轻量，无需维护 WebSocket 连接状态映射

## 模块间关系

四个模块完全独立，不直接交互：

```
Pin ──────○ 无直接调用 ○────── RepoWiki
  │                                │
  │                                │
  ├──── 无直接调用 ○─────── Memory
  │                                │
  │                                │
  └──── 无直接调用 ○─────── Q&A

Agent 通过 MCP 自行编排：
  例：Agent 调用 repoWiki_analyze 后，自行决定是否 memory_create 记录决策

Lumina 只负责各模块能力的暴露，编排逻辑完全在 Agent 端。

### Project 聚合根

Pin 和 Q&A 模块共同挂靠在 Project 实体下。Project 作为共享聚合根，通过 `alias_name` 提供项目标识能力：

- **Project 实体**：存储项目基本信息（name、alias_name JSON 数组、description），为 Pin 和 Q&A 提供项目维度的数据归属
- **Pin 关联 Project**：每个 Pin 通过 `from_project_id` 和 `to_project_name` 关联来源和目标项目
- **Q&A 关联 Project**：每个 QA Session 挂靠在特定 Project 下
- **项目标识匹配**：Pin 的 `to_project_name` 通过 `alias_name` JSON 数组的模糊包含匹配（大小写不敏感）定位目标项目，`project_id` 精确匹配作为兜底
