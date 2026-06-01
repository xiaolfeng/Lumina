# 项目知识库

**生成日期:** 2026-06-01
**提交:** d40697d
**分支:** master

## 概述

`Lumina · 微明` — 赋予 AI 深度代码认知与长期记忆的知识中枢。

基于 `bamboo-base-go` 构建的后端服务 + TanStack Start 前端，包含三大核心功能模块（设计阶段）：

- **RepoWiki**：克隆项目并通过 LLM 分析生成结构化 Wiki 文档
- **Memory**：AI 的长期决策记忆，MCP 端主动推送构建
- **Q&A**：Agent 与用户的富交互式问答通道（SSE 实时推送）

后端通过 Streamable MCP 协议 + HTTP REST API 双通道对外暴露能力；前端通过 REST API + SSE 与后端通信。

> ⚠️ 本文档中涉及的所有模块设计（MCP 工具名称、REST API 路径、数据结构、存储策略等）均为**设计参考方案，并非最终决策**。实际实现时可根据技术约束和开发决策进行调整。详见 `docs/wiki/` 目录。

## 目录结构

```text
./
├── main.go                     # 入口；将生命周期委托给 xMain.Runner
├── go.mod                      # Go 1.25.0；依赖 bamboo-base-go 模块
├── Makefile                    # 开发/测试/格式化命令
├── .env.example                # 必需的环境变量模板
├── api/                        # 请求/响应 DTO（按业务域分包）
│   ├── auth/                   # 认证模块 DTO
│   ├── common/                 # 通用响应结构
│   └── health/                 # 健康检查 DTO
├── docs/
│   ├── swagger.json            # Swagger 规范（自动生成；请勿手动编辑）
│   ├── swagger.yaml            # Swagger 规范（自动生成；请勿手动编辑）
│   └── wiki/                   # 项目设计文档（手动维护）
│       ├── README.md           # 文档总览与导航
│       ├── architecture.md     # 整体架构设计
│       ├── infrastructure.md   # 基础设施层说明
│       ├── repowiki/           # RepoWiki 模块文档
│       ├── memory/             # Memory 模块文档
│       └── qa/                 # Q&A 模块文档
├── internal/
│   ├── app/
│   │   ├── middleware/         # Gin 中间件（认证拦截等）
│   │   ├── route/              # 路由注册与中间件绑定
│   │   └── startup/            # 基础设施初始化与启动节点注册
│   │       └── prepare/        # 幂等种子数据
│   ├── handler/                # HTTP 处理器（薄控制器层）
│   ├── logic/                  # 业务编排层
│   ├── repository/             # 数据库/Redis 访问层
│   │   └── cache/              # Redis 缓存操作
│   ├── entity/                 # GORM 实体（需实现 GetGene() 绑定）
│   └── constant/               # 共享业务常量（基因编号等）
├── web/                        # TanStack Start 前端（pnpm + Vite）
│   ├── package.json            # React 19 + TanStack Start + Tailwind CSS 4
│   ├── vite.config.ts          # Vite 插件链
│   ├── components.json         # shadcn/ui（new-york、zinc、lucide）
│   └── src/                    # 前端源码
│       ├── routes/             # 基于文件的路由
│       ├── components/         # 组件（含 ui/ 子目录）
│       ├── lib/                # 工具函数
│       ├── styles.css          # 全局样式 + Tailwind 主题
│       └── router.tsx          # TanStack Router 入口
└── .agent/skills/              # 项目专属技能（swagger-writer、entity-build、project-style）
```

## 导航指南

| 任务 | 位置 | 备注 |
|---|---|---|
| 新增 API 接口 | `internal/app/route/`、`internal/handler/` | 先注册路由，再写处理器 |
| 新增业务逻辑 | `internal/logic/` | 保持 handler 精简 |
| 新增持久化逻辑 | `internal/repository/` | 仓库方法返回 `*xError.Error` |
| 新增/修改实体 | `internal/entity/`、`internal/app/startup/startup_database.go` | 按依赖顺序更新 `migrateTables` |
| 新增启动能力 | `internal/app/startup/startup.go` + `startup_*.go` | 以 `xRegNode.RegNodeList` 形式注册 |
| 填充默认数据 | `internal/app/startup/prepare/` | 必须保证幂等性 |
| 调整配置/环境变量 | `.env.example`、`internal/app/startup/*.go` | 始终在 `xEnv.GetEnv*` 中提供默认值 |
| 编写 Swagger 文档 | `internal/handler/*.go` 的 godoc + `make swag` | 使用 `swaggo/swag` 注解 |
| 新增请求/响应 DTO | `api/<domain>/` | 按业务域保持子包结构 |
| 新增中间件 | `internal/app/middleware/` | 返回 `gin.HandlerFunc` |
| 查看 RepoWiki 设计 | `docs/wiki/repowiki/` | 模块概述、详细设计、MCP 工具定义 |
| 查看 Memory 设计 | `docs/wiki/memory/` | 模块概述、详细设计、MCP 工具定义 |
| 查看 Q&A 设计 | `docs/wiki/qa/` | 模块概述、详细设计、MCP 工具定义 |
| 新增前端页面 | `web/src/routes/` | 基于文件的路由，文件名即路由路径 |
| 新增前端组件 | `web/src/components/` | 通用组件放此处，页面级组件放 routes |
| 前端样式调整 | `web/src/styles.css` | 仅限全局 CSS 变量和 Tailwind 主题 |
| 前端工具函数 | `web/src/lib/` | 通用工具（如 `cn()`） |
| shadcn/ui 组件 | `web/src/components/ui/` | 通过 `pnpm dlx shadcn@latest add <component>` 添加 |

## 代码地图

| 符号 | 类型 | 位置 | 作用 |
|---|---|---|---|
| `main` | 函数 | `main.go` | 注册启动节点并运行应用 |
| `Init` | 函数 | `internal/app/startup/startup.go` | 启动节点列表工厂 |
| `NewRoute` | 函数 | `internal/app/route/route.go` | 全局中间件 + 路由组 |
| `NewHandler[T]` | 泛型函数 | `internal/handler/handler.go` | Handler 泛型构造模式 |
| `Auth` | 中间件 | `internal/app/middleware/auth.go` | Bearer Token 认证拦截 |
| `AuthLogic` | 结构体 | `internal/logic/auth.go` | 认证业务编排 |
| `HealthLogic.Ping` | 方法 | `internal/logic/health.go` | 服务健康检查编排 |
| `HealthRepo.DatabaseReady` | 方法 | `internal/repository/health.go` | 数据库就绪检查 |
| `getRouter` | 函数 | `web/src/router.tsx` | 前端路由入口 |

## 模块架构

### 三模块独立领域 + 统一基础设施层

```
对外接入层：MCP Server + HTTP REST API + SSE
                    │
模块层：RepoWiki ○ Memory ○ Q&A（互不调用）
                    │
基础设施层：PostgreSQL + Redis + LLM Provider
```

- 三个模块完全独立，不直接交互
- Agent 通过 MCP 自行编排组合调用
- 每个模块同时提供 REST API（前端用）和 MCP Tool（Agent 用）
- 详细架构设计见 `docs/wiki/architecture.md`

### 模块职责

| 模块 | 职责 | 存储策略 |
|------|------|----------|
| RepoWiki | Git 克隆 → LLM 分析 → Wiki 生成 | PG 元数据 + 文件系统 Markdown |
| Memory | 决策卡片 CRUD、标签分类、条件检索 | PG 全量存储 + Redis 可选缓存 |
| Q&A | Session 管理、问题推送（SSE）、答案收集 | PG 全量存储 + Redis 状态缓存 |

### 前后端关系

```
前端（web/）──REST API + SSE──▶ 后端（Go Gin）
                                  │
                                  ├── 认证模块（已实现）
                                  ├── 健康检查（已实现）
                                  ├── RepoWiki（设计中）
                                  ├── Memory（设计中）
                                  └── Q&A（设计中）
```

## 约定

- **框架文档优先**：本项目基于 `bamboo-base-go` 框架构建，使用任何框架组件（`xError`、`xResult`、`xLog`、`xEnv`、`xCtxUtil` 等）前，**必须先通过 `bamboo-document` MCP 查阅官方文档**（板块标识：`bamboo-base-go`），确认 API 签名、参数语义和推荐用法后再编码。禁止凭记忆或猜测使用框架 API。
- **导入别名**：bamboo-base-go 包使用 `x*` 别名（`xLog`、`xEnv`、`xError`、`xResult`、`xReg` 等）。
- **严格分层**：route -> handler -> logic -> repository；禁止跳层调用。
- **上下文依赖注入**：启动阶段将基础设施注册到 context；逻辑层通过 `xCtxUtil.MustGetDB/MustGetRDB` 获取。
- **响应模式**：handler 通过 `xResult.SuccessHasData` 返回成功；错误通过 `ctx.Error` 传递。
- **错误类型**：仓库/逻辑层使用 `*xError.Error` 表示业务/基础设施故障。
- **环境变量族**：`XLF_*`、`APP_*`、`DATABASE_*`、`NOSQL_*`、`SNOWFLAKE_*`、`LLM_*`、`QA_*`。
- **实体 ID 策略**：采用雪花算法基因策略；每个实体必须实现 `GetGene() xSnowflake.Gene`。
- **字段注释**：实体字段必须追加行尾中文注释（`// 字段说明`），且与 `gorm comment` 保持一致。
- **Swagger 注册**：仅在 `XLF_DEBUG=true` 时注册 Swagger UI。
- **SSE 实时推送**：前后端统一使用 SSE（非 WebSocket）进行实时通信。
- **模块独立性**：三个核心模块（RepoWiki、Memory、Q&A）互不调用，Agent 通过 MCP 自行编排。
- **子模块约定**：后端分层详情见 [internal/](./internal/AGENTS.md)，前端专属约定见 [web/](./web/AGENTS.md)。

## 反模式

- 禁止凭记忆或猜测使用 `bamboo-base-go` 框架 API；使用前必须通过 `bamboo-document` MCP（板块 `bamboo-base-go`）查阅官方文档。
- 禁止直接使用 `os.Getenv`；应使用带默认值的 `xEnv.GetEnv*`。
- 禁止手动编辑 `docs/swagger*` 文件；它们由 `swag init` 自动生成。
- 禁止三个模块之间直接调用；Agent 通过 MCP 自行编排。
- 禁止在 Q&A 模块使用 WebSocket；统一使用 SSE。
- 子模块反模式详见各子模块 AGENTS.md。

## 独特风格

- **日志命名**：遵循模块标签（`NamedMAIN`、`NamedINIT`、`NamedCONT`、`NamedLOGC`、`NamedREPO`、`NamedMIDE`）。
- **启动种子阶段**：显式通过 `xCtx.Exec` 节点执行，并隔离在 `prepare/` 目录中。
- **实体 ID 基因策略**：实体级别需绑定基因类型。
- **项目技能**：`.agent/skills/` 包含项目专属技能：`swagger-writer`、`entity-build`、`project-style`。
- **双通道暴露**：每个模块同时提供 REST API 和 MCP Tool。
- **MCP 编排**：Lumina 不做跨模块编排，由 Agent 端自行决定调用顺序和组合。
- **泛型 Handler 构造**：`NewHandler[T]` 统一注入所有 logic 实例。

## 常用命令

```bash
# ── 后端 ──

# 初始化
cp .env.example .env
go mod tidy

# 开发
make dev          # 生成 Swagger 文档并运行（推荐）
make swag         # 仅生成 Swagger 文档
make run          # 直接运行（不重新生成文档）
make tidy         # 整理 Go 模块

# 质量
make fmt          # 格式化代码
make test         # 运行测试（目前尚未脚手架化测试文件）

# 验证
curl http://localhost:8080/api/v1/health/ping

# ── 前端 (web/) ──

# 初始化
cd web && pnpm install

# 开发
pnpm dev          # 启动 Vite 开发服务器（端口 3000）
pnpm build        # 生产构建
pnpm preview      # 预览生产构建

# 质量
pnpm lint         # ESLint 检查
pnpm format       # Prettier 格式化 + ESLint 自动修复
pnpm check        # Prettier 格式检查
pnpm test         # 运行 Vitest 测试

# shadcn/ui 组件
pnpm dlx shadcn@latest add <component>  # 添加 UI 组件
```

## 备注

- 尚未配置 CI 工作流（`.github/workflows` 不存在）。
- `make test` 命令存在，但测试文件尚未脚手架化。
- `docs/wiki/` 为手动维护的设计文档，与 `docs/swagger*` 自动生成文件不同。
- `docs/` 下的 `swagger.json`、`swagger.yaml`、`docs.go` 由 `swag init -g main.go --parseDependency` 自动生成；切勿提交手动编辑。
- `.env` 与 `.env.*` 已被 gitignore；本地开发时从 `.env.example` 复制。
- 雪花算法数据中心/节点 ID 默认为 1/1；可通过 `SNOWFLAKE_DATACENTER_ID` 和 `SNOWFLAKE_NODE_ID` 覆盖。
- Q&A Session 默认最大存活 7 天；可通过 `QA_SESSION_MAX_DURATION`（单位秒）配置。
- RepoWiki 模块需要独立的 LLM Provider 配置；通过 `LLM_*` 环境变量设置。
- 前端位于 `web/` 目录，通过 REST API + SSE 与后端通信；开发服务器端口 3000。
- 认证模块已实现（登录、注册、Token 刷新、Bearer 中间件）；RepoWiki/Memory/Q&A 尚在设计阶段。

## 引用

- [internal/](./internal/AGENTS.md) — 后端业务层详细文档
- [web/](./web/AGENTS.md) — 前端应用详细文档
