# 项目知识库

**生成日期:** 2026-07-15
**提交:** 79f3e61
**分支:** master

## 概述

`Lumina · 微明` — 赋予 AI 深度代码认知与长期记忆的知识中枢。

基于 `bamboo-base-go` 构建的后端服务 + 双前端（TanStack Start 控制台 + Wiki Reader），包含四大核心功能模块：

- **RepoWiki**：克隆项目并通过 5 角色 SubAgent 编排生成结构化 Wiki 文档（已实现，含 Webhook 自动触发 + Cron 定时清理）
- **Memory**：AI 的长期决策记忆，MCP 端主动推送构建（设计中）
- **Q&A**：Agent 与用户的富交互式问答通道（WebSocket 实时推送）— ✅ 已实现
- **Pin**：跨项目依赖约束传递，点对点定向推送与 FIFO 队列消费 — ✅ 已实现

后端通过 Streamable MCP 协议 + HTTP REST API + WebSocket 三通道对外暴露能力；前端通过 REST API + WebSocket 与后端通信。控制台前端（`web/`）与 Wiki Reader 前端（`web-wiki/`）构建产物分别通过 `go:embed` 嵌入 Go 二进制，支持单文件部署。两前端通过 `@lumina/components` workspace 包共享 shadcn/ui 组件、Markdown 渲染原语、motion 动画变体和微明主题 CSS。

> ⚠️ 本文档中涉及的所有模块设计（MCP 工具名称、REST API 路径、数据结构、存储策略等）均为**设计参考方案，并非最终决策**。实际实现时可根据技术约束和开发决策进行调整。详见 `docs/wiki/` 目录。

## 目录结构

```text
./
├── main.go                     # 入口；嵌入双前端 dist → xMain.Runner + Cron Runner
├── embed_frontend.go           # go:embed 声明（web/dist → frontendDist）
├── embed_frontend_wiki.go      # go:embed 声明（web-wiki/dist → wikiFrontendDist）
├── go.mod                      # Go 1.25.0；依赖 bamboo-base-go 模块
├── Makefile                    # 开发/测试/格式化/一键构建命令
├── .env.example                # 必需的环境变量模板
├── resources/                  # 项目级内嵌资源（prompts、模板等）
│   ├── embed.go                # go:embed 暴露 PromptFiles
│   └── prompts/                # RepoWiki 5 角色 system prompt 文件
│       ├── coordinator.md      # Coordinator 角色 prompt
│       ├── explore.md          # Explore 角色 prompt
│       ├── architect.md        # Architect 角色 prompt
│       ├── write.md            # Writer 角色 prompt
│       └── validator.md        # Validator 角色 prompt
├── components/                 # @lumina/components 共享组件包（shadcn ui + markdown 原语 + motion + 主题 CSS）
├── api/                        # 请求/响应 DTO（按业务域分包）
│   ├── auth/                   # 认证模块 DTO
│   ├── user/                   # 用户模块 DTO（资料/密码）
│   ├── biometric/              # WebAuthn 生物认证 DTO
│   ├── apikey/                 # API Key 模块 DTO（CRUD + 重置）
│   ├── project/                # 项目模块 DTO（CRUD）
│   ├── pin/                    # Pin 模块 DTO（CRUD + 筛选）
│   ├── repowiki/               # RepoWiki 模块 DTO（配置/版本/Wiki）
│   ├── qa/                     # Q&A 模块 DTO（会话/问题/配置）
│   ├── llm/                    # LLM 模块 DTO（Provider/Model/Agent 分配）
│   ├── ssh/                    # SSH Key 模块 DTO（CRUD）
│   ├── webhook/                # Webhook 模块 DTO（事件接收）
│   ├── settings/               # 系统设置 DTO（分组配置读写）
│   ├── common/                 # 通用响应结构
│   └── health/                 # 健康检查 DTO
├── docs/
│   ├── swagger.json            # Swagger 规范（自动生成；请勿手动编辑）
│   ├── swagger.yaml            # Swagger 规范（自动生成；请勿手动编辑）
│   └── wiki/                   # 项目设计文档（手动维护）
│       ├── README.md           # 文档总览与导航
│       ├── architecture.md     # 整体架构设计
│       ├── infrastructure.md   # 基础设施层说明
│       ├── repowiki/           # RepoWiki 模块文档（含 webhook-design、multi-agent-design）
│       ├── memory/             # Memory 模块文档
│       └── pin/                # Pin 模块文档
├── internal/
│   ├── app/
│   │   ├── middleware/         # Gin 中间件（认证拦截、API Key 校验、Wiki Auth、MCP 兼容）
│   │   ├── route/              # 路由注册与中间件绑定（含双前端 SPA fallback + MCP + WebSocket）
│   │   │   ├── route_repowiki.go  # RepoWiki REST API 路由
│   │   │   ├── route_llm.go       # LLM REST API 路由
│   │   │   ├── route_ssh.go       # SSH Key REST API 路由
│   │   │   ├── route_webhook.go   # Webhook 接收路由
│   │   │   └── route_settings.go  # 系统设置路由
│   │   └── startup/            # 基础设施初始化与启动节点注册（含 Cron Runner）
│   │       └── prepare/        # 幂等种子数据（含 llm/repowiki/settings 种子）
│   ├── handler/                # HTTP 处理器（薄控制器层，含 bind.go 通用绑定辅助）
│   │   ├── repowiki.go         # RepoWiki 配置/版本/分析触发处理器
│   │   ├── wiki_reader.go      # Wiki 内容读取处理器（公开/密码保护/manifest）
│   │   ├── llm.go              # LLM Provider/Model CRUD + Agent 模型分配
│   │   ├── ssh_key.go          # SSH Key CRUD + 密钥生成/公钥导出
│   │   ├── webhook.go          # Webhook 事件接收处理器
│   │   └── settings.go         # 系统设置处理器
│   ├── logic/                  # 业务编排层（QA 逻辑按职责拆分为 qa_*.go 多文件）
│   │   ├── repowiki_logic.go   # RepoWiki 核心逻辑（分析入口、配置/版本 CRUD）
│   │   ├── repowiki_pipeline.go # RepoWiki 分析管道（Git 准备 + 状态机驱动）
│   │   ├── repowiki_orchestrator.go # 5 角色 SubAgent 编排引擎
│   │   ├── repowiki_subagent_prompts.go # 5 角色 system/user prompt 构建
│   │   ├── repowiki_types.go   # RepoWiki 内部类型（WikiEntry / ValidationError / ExploreOutput）
│   │   ├── repowiki_cron.go    # RepoWiki 定时清理任务逻辑
│   │   ├── repowiki_webhook.go # RepoWiki Webhook 处理逻辑
│   │   ├── llm_provider.go     # LLM Provider 逻辑（CRUD + API Key 加密）
│   │   ├── llm_model.go        # LLM Model 逻辑（CRUD + Agent 角色分配）
│   │   ├── ssh_key.go          # SSH Key 逻辑（CRUD + 密钥对生成）
│   │   └── settings.go         # 系统设置逻辑（分组配置读写）
│   ├── repository/             # 数据库/Redis 访问层
│   │   ├── repowiki_config.go  # RepoWikiConfig 持久化
│   │   ├── wiki_version.go     # WikiVersion 持久化
│   │   ├── llm_provider.go     # LlmProvider 持久化
│   │   ├── llm_model.go        # LlmModel 持久化
│   │   ├── ssh_key.go          # SshKey 持久化
│   │   ├── webhook_event.go    # WebhookEvent 持久化
│   │   └── cache/              # Redis 缓存子层（含 repowiki、ssh_key 缓存）
│   ├── service/                # 共享服务层（Git 服务、加密辅助、Webhook 解析、LLM Provider 等）
│   │   ├── wiki_storage.go     # RepoWiki 文件系统存储与路径管理
│   │   ├── wiki_auth_token.go  # Wiki 访问密码 Token 生成与校验
│   │   ├── git_service.go      # Git 仓库克隆/拉取服务（go-git）
│   │   ├── agent_factory.go    # LLM Agent 工厂（创建 SubAgent 运行实例）
│   │   ├── crypto_helper.go    # AES-256-GCM 加解密辅助
│   │   ├── dependency_extractor.go # 依赖关系提取器
│   │   ├── file_scanner.go     # 仓库文件扫描器
│   │   ├── llm_resolver.go     # LLM Resolver（按 Agent 角色解析模型配置）
│   │   ├── prompt_loader.go    # Prompt 加载器（从 resources/prompts 读取）
│   │   ├── repo_tools.go       # RepoWiki 子 Agent 工具集
│   │   ├── ssh_key_gen.go      # SSH 密钥对生成
│   │   ├── webhook_parser.go   # Webhook Payload 解析器
│   │   └── webhook_signer.go   # Webhook HMAC 签名生成与校验
│   ├── entity/                 # GORM 实体（需实现 GetGene() 绑定，Gene 32~44）
│   ├── mcp/                    # MCP Server 工具注册（QA/Project/Pin/RepoWiki 工具拆分）
│   ├── websocket/              # WebSocket 连接管理 + 消息分发
│   ├── qa/                     # Q&A 回答队列（会话级 FIFO）
│   └── constant/               # 共享业务常量（基因编号、LLM、RepoWiki、Settings 等）
├── web/                        # TanStack Start 控制台前端（pnpm + Vite）
│   ├── package.json            # React 19 + TanStack Start + Tailwind CSS 4
│   ├── vite.config.ts          # Vite 插件链（含代码拆分配置）
│   ├── components.json         # shadcn/ui（new-york、zinc、lucide）
│   └── src/                    # 前端源码
│       ├── routes/             # 基于文件的路由（公开页 + 认证页 + 控制台 + Interact 交互）
│       ├── components/         # 组件（含 ui/、landing/、通用组件和业务子目录 apikey/、project/、pin/、profile/、qa/、interact/、llm/、ssh/、repowiki/、settings/）
│       ├── hooks/              # React Hooks（15 个，含 LLM/SSH/RepoWiki/Webhook/Settings）
│       ├── lib/                # 工具函数 + API 客户端 + 类型定义 + WebAuthn 辅助 + Cookie 工具
│       ├── styles.css          # 全局样式 + Tailwind 主题
│       └── router.tsx          # TanStack Router 入口
├── web-wiki/                   # TanStack Start Wiki Reader 前端（独立 SPA，部署在 /wiki/）
│   ├── package.json            # React 19 + TanStack Router + Tailwind CSS 4
│   ├── vite.config.ts          # Vite + TanStack Router 插件（base: /wiki/）
│   └── src/                    # 前端源码（只读 Wiki 渲染 + 密码门）
└── .agent/skills/              # 项目专属技能（swagger-writer、entity-build、project-style、mcp-qa-test）
```

## 导航指南

| 任务 | 位置 | 备注 |
|---|---|---|
| 新增 API 接口 | `internal/app/route/`、`internal/handler/` | 先注册路由，再写处理器 |
| 新增业务逻辑 | `internal/logic/` | 保持 handler 精简 |
| 新增持久化逻辑 | `internal/repository/` | 仓库方法返回 `*xError.Error` |
| 新增共享服务 | `internal/service/` | 跨模块复用的基础设施（如 Git 服务、加密辅助） |
| 新增/修改实体 | `internal/entity/`、`internal/app/startup/startup_database.go` | 按依赖顺序更新 `migrateTables` |
| 新增启动能力 | `internal/app/startup/startup.go` + `startup_*.go` | 以 `xRegNode.RegNodeList` 形式注册 |
| 新增定时任务 | `internal/app/startup/startup_cron.go` | 通过 `xCronRunner` 注册，由 `main.go` 传入 `xMain.Runner` |
| 填充默认数据 | `internal/app/startup/prepare/` | 必须保证幂等性 |
| 调整配置/环境变量 | `.env.example`、`internal/app/startup/*.go` | 始终在 `xEnv.GetEnv*` 中提供默认值 |
| 编写 Swagger 文档 | `internal/handler/*.go` 的 godoc + `make swag` | 使用 `swaggo/swag` 注解 |
| 新增请求/响应 DTO | `api/<domain>/` | 按业务域保持子包结构 |
| 新增中间件 | `internal/app/middleware/` | 返回 `gin.HandlerFunc` |
| 查看 RepoWiki 设计 | `docs/wiki/repowiki/` | 模块概述、详细设计、MCP 工具、Webhook 设计、多 Agent 设计 |
| 查看 Memory 设计 | `docs/wiki/memory/` | 模块概述、详细设计、MCP 工具定义 |
| 查看 Pin 设计 | `docs/wiki/pin/` | 模块概述、详细设计、MCP 工具定义 |
| 新增 MCP 工具 | `internal/mcp/` | 注册到 `server.go`，Logic 注入到 `startup_mcp.go` |
| 新增 WebSocket 消息 | `internal/websocket/message.go` | 定义 MessageType 常量 |
| 新增 RepoWiki prompt | `resources/prompts/*.md` | 通过 `service/prompt_loader.go` 加载，禁止硬编码 |
| 新增 LLM Provider/Model | `internal/entity/llm_*.go` + `logic/llm_*.go` + `handler/llm.go` | API Key 经 AES-256-GCM 加密 |
| 新增 SSH Key | `internal/entity/ssh_key.go` + `logic/ssh_key.go` + `service/ssh_key_gen.go` | 私钥加密存储 |
| 新增 Webhook 事件 | `handler/webhook.go` + `logic/repowiki_webhook.go` + `service/webhook_parser.go` | HMAC 签名校验必经 |
| 新增前端页面 | `web/src/routes/` | 基于文件的路由，文件名即路由路径 |
| 新增前端业务组件 | `web/src/components/<domain>/` | 按业务域组织 |
| 新增前端 API 封装 | `web/src/lib/apis/` | 使用 apiClient 封装 |
| 新增前端数据 Hook | `web/src/hooks/` | 基于 TanStack Query |
| 前端样式调整 | `web/src/styles.css` | 仅限全局 CSS 变量和 Tailwind 主题 |
| shadcn/ui 组件 | `web/src/components/ui/` 或 `components/` | 通过 `pnpm dlx shadcn@latest add <component>` 添加 |
| Wiki Reader 前端 | `web-wiki/` | 独立 SPA，部署在 `/wiki/`，只读渲染 |

## 代码地图

| 符号 | 类型 | 位置 | 作用 |
|---|---|---|---|
| `main` | 函数 | `main.go` | 嵌入双前端资源 → 注册启动节点 → 运行应用 + Cron Runner |
| `frontendDist` | 变量 | `embed_frontend.go` | `go:embed all:web/dist` 嵌入控制台前端构建产物 |
| `wikiFrontendDist` | 变量 | `embed_frontend_wiki.go` | `go:embed all:web-wiki/dist` 嵌入 Wiki Reader 前端构建产物 |
| `PromptFiles` | 变量 | `resources/embed.go` | `go:embed prompts/*.md` 内嵌 RepoWiki 5 角色 prompt |
| `Init` | 函数 | `internal/app/startup/startup.go` | 启动节点列表工厂（DB → Redis → RepoWiki → MCP → Prepare） |
| `NewCronRunner` | 函数 | `internal/app/startup/startup_cron.go` | Cron Runner 工厂（RepoWiki 定时清理，传入 `xMain.Runner`） |
| `NewRouteWithFrontend` | 函数 | `internal/app/route/route.go` | 全局中间件 + API 路由 + MCP + WebSocket + 双前端 SPA fallback |
| `NewHandler[T]` | 泛型函数 | `internal/handler/handler.go` | Handler 泛型构造模式，注入全部 Logic（12 个） |
| `Auth` | 中间件 | `internal/app/middleware/auth.go` | Bearer Token 认证拦截（单用户模式，注入认证标记） |
| `ApikeyAuth` | 中间件 | `internal/app/middleware/apikey.go` | API Key 认证（`lumi_` 前缀 + bcrypt 校验，MCP 端点使用） |
| `WikiAuth` | 中间件 | `internal/app/middleware/wiki_auth.go` | Wiki Reader 访问认证（密码 Token / Cookie 会话） |
| `McpCompat` | 中间件 | `internal/app/middleware/mcp_compat.go` | MCP 端点 Streamable HTTP 兼容性处理 |
| `AuthLogic` | 结构体 | `internal/logic/auth.go` | 单用户认证业务编排（Info 表 + Token 管理 + 资料更新） |
| `BiometricLogic` | 结构体 | `internal/logic/biometric.go` | WebAuthn 生物认证编排（注册/登录/凭证 CRUD） |
| `ApikeyLogic` | 结构体 | `internal/logic/apikey.go` | API Key 业务编排（创建/列表/更新/删除/重置/密钥生成/哈希/脱敏/校验） |
| `ProjectLogic` | 结构体 | `internal/logic/project.go` | 项目业务编排（CRUD + 名称唯一校验 + 别名解析） |
| `PinLogic` | 结构体 | `internal/logic/pin.go` | Pin 约束编排（Push/Consume/Peek/List + 项目解析） |
| `QaLogic` | 结构体 | `internal/logic/qa_logic.go` | Q&A 业务编排（Session/Question/Supplement + 队列消费） |
| `LlmProviderLogic` | 结构体 | `internal/logic/llm_provider.go` | LLM Provider 编排（CRUD + API Key 加密存储） |
| `LlmModelLogic` | 结构体 | `internal/logic/llm_model.go` | LLM Model 编排（CRUD + Agent 角色模型分配） |
| `SshKeyLogic` | 结构体 | `internal/logic/ssh_key.go` | SSH Key 编排（CRUD + 密钥对生成/公钥导出） |
| `SettingsLogic` | 结构体 | `internal/logic/settings.go` | 系统设置编排（分组配置读写 + Info 表编排） |
| `ProjectRepo` | 结构体 | `internal/repository/project.go` | 项目持久化（CRUD + Redis Cache-Aside 缓存） |
| `InitMCPServer` | 函数 | `internal/mcp/server.go` | 创建 MCP Server + 注册 QA/Project/Pin/RepoWiki 工具 + 返回 StreamableHTTPHandler |
| `Hub` | 结构体 | `internal/websocket/hub.go` | WebSocket 连接管理器（sessionID → deviceID 二级索引 + 心跳检测） |
| `QueueManager` | 结构体 | `internal/qa/queue.go` | Q&A 回答队列管理器（会话级 FIFO 队列 + 阻塞消费） |
| `DownloadToken` | 结构体 | `internal/service/download_token.go` | 文件下载 Token 生成与校验（短时效签名） |
| `MediaAnswerService` | 结构体 | `internal/service/media_answer.go` | 媒体回答处理（图片/文件附件格式化） |
| `CryptoHelper` | 结构体 | `internal/service/crypto_helper.go` | AES-256-GCM 加解密（LLM API Key / SSH 私钥加密存储） |
| `GitService` | 结构体 | `internal/service/git_service.go` | Git 仓库克隆/拉取服务（go-git 封装） |
| `AgentFactory` | 结构体 | `internal/service/agent_factory.go` | LLM Agent 工厂（创建 SubAgent 运行实例） |
| `LLMResolver` | 结构体 | `internal/service/llm_resolver.go` | 按 Agent 角色解析运行时模型配置 |
| `PromptLoader` | 结构体 | `internal/service/prompt_loader.go` | 从 `resources/prompts` 读取内嵌 prompt 文件 |
| `RepoTools` | 结构体 | `internal/service/repo_tools.go` | RepoWiki 子 Agent 工具集（文件读取/目录树/搜索） |
| `DependencyExtractor` | 结构体 | `internal/service/dependency_extractor.go` | 依赖关系提取（解析 import/require 构建模块依赖图） |
| `FileScanner` | 结构体 | `internal/service/file_scanner.go` | 仓库文件扫描（清单 + 忽略规则 + 大小限制） |
| `SshKeyGen` | 结构体 | `internal/service/ssh_key_gen.go` | SSH 密钥对生成（ed25519/rsa） |
| `WebhookParser` | 结构体 | `internal/service/webhook_parser.go` | Webhook Payload 解析（GitHub/GitLab 事件格式） |
| `WebhookSigner` | 结构体 | `internal/service/webhook_signer.go` | Webhook HMAC 签名生成与校验 |
| `RepoWikiConfig` | 结构体 | `internal/entity/repowiki_config.go` | RepoWiki 配置实体（Gene=39，仓库地址/Webhook 配置/当前选中版本） |
| `WikiVersion` | 结构体 | `internal/entity/wiki_version.go` | Wiki 版本实体（Gene=40，版本号/状态/文件路径/token 统计） |
| `LlmProvider` | 结构体 | `internal/entity/llm_provider.go` | LLM Provider 实体（Gene=41，名称/BaseURL/加密 API Key） |
| `LlmModel` | 结构体 | `internal/entity/llm_model.go` | LLM Model 实体（Gene=42，Provider 关联/Agent 角色分配） |
| `WebhookEvent` | 结构体 | `internal/entity/webhook_event.go` | Webhook 事件实体（Gene=43，事件 ID/分支/状态） |
| `SshKey` | 结构体 | `internal/entity/ssh_key.go` | SSH Key 实体（Gene=44，名称/指纹/公钥/加密私钥） |
| `RepoWikiLogic` | 结构体 | `internal/logic/repowiki_logic.go` | RepoWiki 业务编排（配置/版本/分析入口） |
| `GetRepoWikiLogicFromContext` | 函数 | `internal/logic/repowiki_logic.go` | 从 context 获取 RepoWikiLogic（MCP/Cron/Handler 共用） |
| `SubAgentOrchestrator` | 结构体 | `internal/logic/repowiki_orchestrator.go` | 5 角色 SubAgent 编排引擎（overview → explore → architect → writer → validator） |
| `AnalysisPipeline` | 结构体 | `internal/logic/repowiki_pipeline.go` | RepoWiki 分析管道（Git 准备 + 状态机驱动） |
| `HealthLogic.Ping` | 方法 | `internal/logic/health.go` | 服务健康检查编排 |
| `HealthRepo.DatabaseReady` | 方法 | `internal/repository/health.go` | 数据库就绪检查 |
| `getRouter` | 函数 | `web/src/router.tsx` | 控制台前端路由入口 |
| `apiClient` | 变量 | `web/src/lib/apis/client.ts` | axios 实例（Token 注入 + 401 自动清理） |
| `useAuth` | Hook | `web/src/hooks/useAuth.ts` | 认证组合 Hook（登录/登出/刷新/初始化/自动续期/WebAuthn） |
| `useBiometric` | Hook | `web/src/hooks/useBiometric.ts` | WebAuthn 生物认证 Hook（注册/登录/凭证管理） |
| `useQaWebSocket` | Hook | `web/src/hooks/useQaWebSocket.ts` | Q&A WebSocket 连接管理（自动重连 + 消息回调 + 会话恢复） |
| `useQaSession` | Hook | `web/src/hooks/useQaSession.ts` | Q&A 会话状态管理（问题推送 + 回答提交 + 文件上传） |
| `useQaAdmin` | Hook | `web/src/hooks/useQaAdmin.ts` | Q&A 管理端数据 Hook（会话列表/详情/删除/配置） |
| `useLlmConfig` | Hook | `web/src/hooks/useLlmConfig.ts` | LLM 配置 Hook（Provider/Model CRUD + Agent 模型分配） |
| `useSshKey` | Hook | `web/src/hooks/useSshKey.ts` | SSH Key 数据 Hook（CRUD + 分页） |
| `useRepoWiki` | Hook | `web/src/hooks/useRepoWiki.ts` | RepoWiki Hook（配置/版本/分析触发/Webhook 事件） |
| `useSettings` | Hook | `web/src/hooks/useSettings.ts` | 系统设置 Hook（分组配置读写） |

## 模块架构

### 四模块独立领域 + 统一基础设施层

```
对外接入层：MCP Server + HTTP REST API + WebSocket + Webhook
                    │
模块层：RepoWiki ○ Memory ○ Q&A ✅ Pin ✅（互不调用）
                    │
基础设施层：PostgreSQL + Redis + LLM Provider（热配置）+ Git Service + 文件系统
```

- 四个模块完全独立，不直接交互
- Agent 通过 MCP 自行编排组合调用
- 每个模块同时提供 REST API（前端用）和 MCP Tool（Agent 用）
- Q&A 模块额外提供 WebSocket 通道用于实时问题推送
- RepoWiki 模块额外提供 Webhook 接收端点用于 Git Push 自动触发
- LLM Provider 配置已从环境变量迁移到数据库热配置，通过前端管理页面配置
- 详细架构设计见 `docs/wiki/architecture.md`

### 已实现模块

| 模块 | 职责 | 存储策略 | 状态 |
|------|------|----------|------|
| 认证 | 单用户登录/初始化/Token 管理/资料更新 | Info 表 + Redis Token 缓存 | ✅ 已实现 |
| 用户 | 个人资料编辑/密码修改 | Info 表 | ✅ 已实现 |
| WebAuthn | 生物特征认证（注册/登录/凭证管理） | BiometricCredential 表 + Redis Challenge 缓存 | ✅ 已实现 |
| API Key | 密钥 CRUD/重置/脱敏/校验 | Apikey 表（bcrypt 哈希） | ✅ 已实现 |
| 项目 | 项目 CRUD/别名管理 | Project 表 + Redis 缓存（Cache-Aside） | ✅ 已实现 |
| Q&A | Session 管理、问题推送（WebSocket）、回答队列、MCP 工具、文件上传下载 | QaSession + QaQuestion + QaSupplement 表 + Redis 队列 | ✅ 已实现 |
| Pin | 跨项目约束传递、定向推送、FIFO 消费 | Pin 表（数据库 FIFO） | ✅ 已实现 |
| RepoWiki | Git 克隆 → 5 角色 SubAgent 编排 → Wiki 生成 + Webhook 自动触发 + Cron 清理 | RepoWikiConfig + WikiVersion 表 + 文件系统版本隔离 | ✅ 已实现 |
| LLM 热配置 | Provider/Model CRUD + Agent 角色模型分配 + API Key 加密存储 | LlmProvider + LlmModel 表 + AES-256-GCM 加密 | ✅ 已实现 |
| SSH Key | 密钥对生成/CRUD/公钥导出 + 私钥加密存储 | SshKey 表 + Redis 缓存 | ✅ 已实现 |
| Webhook | Git Push 事件接收 + HMAC 校验 + RepoWiki 触发 | WebhookEvent 表 | ✅ 已实现 |
| 系统设置 | 站点/安全/Q&A/RepoWiki 分组配置读写 | Info 表 | ✅ 已实现 |
| MCP Server | Streamable HTTP 端点、QA/Project/Pin/RepoWiki 工具注册（20+ 个工具） | 复用各模块 Logic 层 | ✅ 已实现 |
| WebSocket | 连接管理、消息分发、心跳检测、会话恢复 | Hub 内存管理（sessionID → deviceID 索引） | ✅ 已实现 |
| Cron Runner | RepoWiki 定时清理（超时任务清理 + 失败重试） | 无持久化（基于 xCronRunner） | ✅ 已实现 |
| 健康检查 | 服务/数据库就绪检查 | 无持久化 | ✅ 已实现 |
| Wiki Reader 前端 | 只读 Wiki 渲染 SPA + 密码门 | 无持久化（go:embed 嵌入） | ✅ 已实现 |

### 设计中模块

| 模块 | 职责 | 存储策略 |
|------|------|----------|
| Memory | 决策卡片 CRUD、标签分类、条件检索 | PG 全量存储 + Redis 可选缓存 |

### 前后端关系

```
控制台前端（web/）──REST API + WebSocket──▶ 后端（Go Gin）
                                              │
Wiki Reader 前端（web-wiki/）──REST API───▶ 后端（/wiki/* + WikiAuth）
                                              │
                                              ├── 认证模块（已实现）
                                              ├── 用户模块（已实现）
                                              ├── WebAuthn 生物认证（已实现）
                                              ├── API Key 模块（已实现）
                                              ├── 项目模块（已实现）
                                              ├── Q&A 模块（已实现）
                                              ├── Pin 模块（已实现）
                                              ├── RepoWiki 模块（已实现，含 Webhook + Cron）
                                              ├── LLM 热配置（已实现）
                                              ├── SSH Key 模块（已实现）
                                              ├── Webhook 模块（已实现）
                                              ├── 系统设置（已实现）
                                              ├── MCP Server（已实现）
                                              ├── WebSocket Hub（已实现）
                                              ├── Cron Runner（已实现）
                                              ├── 健康检查（已实现）
                                              ├── 双前端 SPA 服务（已实现，go:embed）
                                              └── Memory（设计中）
```

## 约定

- **框架文档优先**：本项目基于 `bamboo-base-go` 框架构建，使用任何框架组件（`xError`、`xResult`、`xLog`、`xEnv`、`xCtxUtil` 等）前，**必须先通过 `bamboo-document` MCP 查阅官方文档**（板块标识：`bamboo-base-go`），确认 API 签名、参数语义和推荐用法后再编码。禁止凭记忆或猜测使用框架 API。
- **导入别名**：bamboo-base-go 包使用 `x*` 别名（`xLog`、`xEnv`、`xError`、`xResult`、`xReg`、`xCron`、`xCronRunner` 等）。
- **严格分层**：route -> handler -> logic -> repository；禁止跳层调用。`service/` 为跨模块共享服务层，可被 logic 调用。
- **上下文依赖注入**：启动阶段将基础设施注册到 context；逻辑层通过 `xCtxUtil.MustGetDB/MustGetRDB` 获取；RepoWikiLogic 通过 `logic.GetRepoWikiLogicFromContext` 获取。
- **响应模式**：handler 通过 `xResult.SuccessHasData` 返回成功；错误通过 `ctx.Error` 传递。
- **错误类型**：仓库/逻辑层使用 `*xError.Error` 表示业务/基础设施故障。
- **环境变量族**：`XLF_*`、`APP_*`、`DATABASE_*`、`NOSQL_*`、`SNOWFLAKE_*`、`LLM_*`、`QA_*`、`REPOWIKI_*`、`XLF_BIOMETRIC_*`。
- **实体 ID 策略**：采用雪花算法基因策略；每个实体必须实现 `GetGene() xSnowflake.Gene`，基因编号定义在 `constant/gene_number.go`（GeneProject=32 ~ GeneSSHKey=44）。
- **字段注释**：实体字段必须追加行尾中文注释（`// 字段说明`），且与 `gorm comment` 保持一致。
- **Swagger 注册**：仅在 `XLF_DEBUG=true` 时注册 Swagger UI。
- **WebSocket 实时推送**：Q&A 模块使用 WebSocket 进行实时问题推送（非 SSE）。
- **模块独立性**：四个核心模块（RepoWiki、Memory、Q&A、Pin）互不调用，Agent 通过 MCP 自行编排。
- **双前端嵌入部署**：通过 `go:embed` 将 `web/dist` 和 `web-wiki/dist` 分别嵌入 Go 二进制，构建命令 `make generate` 完成前端打包 → Swagger → Go 编译全流程。
- **MCP 路由注册**：MCP 端点必须在 `engine.Use()` 之前注册以绕开 `ResponseMiddleware`。
- **共享组件包**：shadcn ui 组件、Markdown 渲染原语、motion 动画变体、微明主题 CSS 统一在 `@lumina/components` workspace 包管理，被 `web` 和 `web-wiki` 共同消费。
- **资源内嵌**：RepoWiki 5 角色 system prompt 通过 `go:embed` 内嵌在 `resources/prompts/`，由 `service/prompt_loader.go` 读取；禁止在 logic 中硬编码 prompt 文本。
- **LLM 热配置**：LLM Provider/Model 配置存储在数据库（非环境变量），通过前端管理页面配置；API Key 经 `service/crypto_helper.go` AES-256-GCM 加密，密钥由 `LLM_ENCRYPT_SECRET` 环境变量提供。
- **加密存储**：LLM API Key 和 SSH 私钥必须经 AES-256-GCM 加密后存储，禁止明文落库。
- **Webhook 签名校验**：所有 Webhook 请求必须经 `service/webhook_signer.go` 校验 HMAC 签名，密钥由 `REPOWIKI_HMAC_SECRET` 环境变量提供。
- **Cron 任务**：定时任务通过 `xCronRunner` 注册（`startup_cron.go`），由 `main.go` 传入 `xMain.Runner` 异步执行，不阻塞启动节点链。
- **子模块约定**：后端分层详情见 [internal/](./internal/AGENTS.md)，控制台前端专属约定见 [web/](./web/AGENTS.md)，Wiki Reader 前端约定见 [web-wiki/](./web-wiki/AGENTS.md)。

## 反模式

- 禁止凭记忆或猜测使用 `bamboo-base-go` 框架 API；使用前必须通过 `bamboo-document` MCP（板块 `bamboo-base-go`）查阅官方文档。
- 禁止直接使用 `os.Getenv`；应使用带默认值的 `xEnv.GetEnv*`。
- 禁止手动编辑 `docs/swagger*` 文件；它们由 `swag init` 自动生成。
- 禁止四个模块之间直接调用；Agent 通过 MCP 自行编排。
- 禁止在 Q&A 模块使用 SSE 进行问题推送；统一使用 WebSocket。
- 禁止明文存储 LLM API Key 或 SSH 私钥；必须经 `crypto_helper.go` AES-256-GCM 加密。
- 禁止在 logic 中硬编码 RepoWiki prompt 文本；统一放 `resources/prompts/` 通过 `prompt_loader.go` 加载。
- 禁止在 Webhook 处理中跳过 HMAC 签名校验。
- 禁止在 `resources/prompts/` 外散落 prompt 文件；所有内嵌静态资源集中在 `resources/` 目录。
- 禁止通过裸 goroutine 调度定时任务；统一走 `xCronRunner`。
- 子模块反模式详见各子模块 AGENTS.md。

## 独特风格

- **日志命名**：遵循模块标签（`NamedMAIN`、`NamedINIT`、`NamedCONT`、`NamedLOGC`、`NamedREPO`、`NamedMIDE`、`NamedCRON`）。
- **启动种子阶段**：显式通过 `xCtx.Exec` 节点执行，并隔离在 `prepare/` 目录中。
- **实体 ID 基因策略**：实体级别需绑定基因类型（`GeneProject = 32` ~ `GeneSSHKey = 44`），定义在 `constant/gene_number.go`。
- **项目技能**：`.agent/skills/` 包含项目专属技能：`swagger-writer`、`entity-build`、`project-style`、`mcp-qa-test`。
- **双通道暴露**：每个模块同时提供 REST API 和 MCP Tool。
- **MCP 编排**：Lumina 不做跨模块编排，由 Agent 端自行决定调用顺序和组合。
- **泛型 Handler 构造**：`NewHandler[T]` 统一注入所有 logic 实例（12 个 Logic）。
- **双前端嵌入**：`embed_frontend.go` + `embed_frontend_wiki.go` + `route_frontend.go` 实现双 SPA fallback，单二进制部署。
- **API Key 安全**：`lumi_` 前缀 + base64 RawURL 编码 + bcrypt 哈希，仅创建/重置时返回完整密钥。
- **API Key 认证中间件**：`middleware.ApikeyAuth` 专门用于 MCP 端点认证，与 `middleware.Auth`（Bearer Token）分离。
- **Wiki Auth 中间件**：`middleware.WikiAuth` 处理 Wiki Reader 的密码 Token / Cookie 会话认证，与控制台认证分离。
- **Project 缓存**：三层映射（ID→详情、Name→ID、Alias→ID），Cache-Aside 模式，30 分钟 TTL。
- **WebSocket Hub**：按 sessionID → deviceID 二级索引管理连接，心跳检测间隔 5s / 超时 15s，支持会话恢复。
- **Q&A 回答队列**：会话级 FIFO 队列，支持 `WaitAndConsume` 阻塞等待新回答（MCP 工具消费）。
- **Q&A 推送回调**：`logic.OnQuestionPushed` / `logic.OnSupplementPushed` / `logic.OnQuestionCancelled` / `logic.OnSessionArchived` 函数变量在 `route_ws.go` 中设置，解耦 Logic 层和 WebSocket 层。
- **Q&A 逻辑拆分**：原 `qa.go`（1778 行）按职责拆分为 `qa_logic.go`（核心编排）、`qa_format.go`（题型格式化）、`qa_helper.go`（辅助函数）、`qa_mcp.go`（MCP 工具）、`qa_mcp_helpers.go`（MCP 辅助）、`qa_download.go`（文件下载）。
- **MCP 工具拆分**：`mcp/qa_tools.go` 拆分为 `qa_tools.go`（注册）、`qa_handlers.go`（handler 实现）、`qa_type_details.go`（题型 schema 细节）。
- **Q&A 题型格式化**：Logic 层内置 15+ 题型格式化函数（select/multi-select/text/boolean/code/image/file/slider/rank/rate/plan/options/diff/review）。
- **Pin FIFO 消费**：基于数据库实现 FIFO（`ConsumeOldestPending` 按 createdAt 升序 + `ConsumeByID` 精确消费），不依赖 Redis 队列。
- **WebAuthn 集成**：后端 `logic/biometric.go` + `webauthn_user.go` 适配器，前端 `lib/webauthn/helpers.ts` 处理浏览器端编解码。
- **Interact 前端页面**：独立布局（非 Console），支持 15+ 种题型组件 + 交互原语（primitives/），WebSocket 实时交互，断线重连。
- **前端通用组件**：`confirm-delete-dialog.tsx`、`page-header.tsx`、`skeleton-table.tsx` 跨模块复用，替代各域重复的删除对话框和页面头部。
- **前端首页拆分**：`routes/_public/index.tsx` 拆分为 `components/landing/` 下 hero/features/tech 区块组件。
- **RepoWiki 子 Agent 编排**：`SubAgentOrchestrator` 按预定义 5 阶段（Coordinator → Explore → Architect → Writer → Validator）生成 Wiki，prompt 模板内嵌在 `resources/prompts/*.md` 通过 `service/prompt_loader.go` 加载，`repowiki_subagent_prompts.go` 负责动态构建 user prompt，`repowiki_types.go` 定义内部类型，`repowiki_pipeline.go` 负责 Git 准备与状态机驱动。
- **RepoWiki 版本隔离**：每个 Wiki 版本存储在 `versions/{vid}/` 下，`RepoWikiConfig.SelectedVersionID` 指定当前对外服务版本；旧版 config 级目录已废弃，新版本完成后清理。
- **RepoWiki MCP 只读**：MCP 端仅暴露 `repoWiki_query` / `repoWiki_list` 两个只读工具，Wiki 更新由 Git Webhook 自动触发。
- **RepoWiki Webhook 自动触发**：Git Push 事件经 `service/webhook_signer.go` HMAC 校验 + `service/webhook_parser.go` 解析后，由 `logic/repowiki_webhook.go` 匹配分支规则并触发分析。
- **RepoWiki Cron 清理**：`startup_cron.go` 注册定时任务（默认每 5 分钟），通过 `logic/repowiki_cron.go` 清理超时任务和失败重试。
- **LLM 热配置**：Provider/Model 配置存储在数据库，支持 Agent 角色（Coordinator/Explore/Architect/Writer/Validator）分别分配不同模型；API Key 经 AES-256-GCM 加密存储。
- **文件下载 Token**：`service/download_token.go` 生成短时效签名 Token，用于 Q&A 文件附件下载鉴权。
- **媒体回答处理**：`service/media_answer.go` 处理图片/文件附件的回答格式化，供 Q&A MCP 工具调用。
- **双前端共享**：`@lumina/components` workspace 包统一管理 shadcn ui 组件、Markdown 渲染原语、motion 动画变体、微明主题 CSS，被 web 和 web-wiki 共同消费。
- **内嵌资源集中管理**：`resources/` 目录集中管理项目级内嵌静态资源（如 RepoWiki prompt），通过 `go:embed` 暴露给各业务包引用，避免资源文件散落在业务包内部。
- **Cron Runner 模式**：`startup.NewCronRunner()` 返回的函数由 `main.go` 传入 `xMain.Runner` 的 goroutineFunc 参数，与启动节点链解耦异步执行。

## 常用命令

```bash
# ── 后端 ──

# 初始化
cp .env.example .env
go mod tidy

# 开发
make dev-backend   # 生成 Swagger 文档并运行（推荐）
make dev-frontend  # 启动控制台前端开发服务器（端口 3000）
make swag          # 仅生成 Swagger 文档
make run           # 运行已编译二进制

# 一键构建（双前端打包 → Swagger → Go 编译）
make generate      # 或 make build

# 质量
make tidy          # 整理 Go 模块
make fmt           # 格式化代码
make test          # 运行测试

# 验证
curl http://localhost:8080/api/v1/health/ping

# ── 控制台前端 (web/) ──

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

# ── Wiki Reader 前端 (web-wiki/) ──

cd web-wiki
pnpm install      # 安装依赖
pnpm dev          # 开发服务器（端口 3001）
pnpm build        # 类型检查 + 生产构建
pnpm lint         # ESLint 检查
pnpm format       # Prettier 格式化 + ESLint 自动修复
```

## 备注

- 双前端通过 `go:embed` 嵌入 Go 二进制，构建顺序：先 `pnpm build`（产出 `web/dist` + `web-wiki/dist`），再 `go build`。使用 `make generate` 一键完成。
- 前端独立开发时使用 `make dev-frontend`（控制台 Vite dev server 端口 3000）和 `cd web-wiki && pnpm dev`（Wiki Reader 端口 3001），但生产部署时前后端合一。
- 尚未配置 CI 工作流（`.github/workflows` 不存在）。
- `make test` 命令存在，已有测试用例覆盖 project、llm_model、llm_provider、ssh_key_gen、webhook_parser、webhook_signer、wiki_auth_token、wiki_storage、git_service、file_scanner、dependency_extractor、repowiki_orchestrator 等模块。
- `docs/wiki/` 为手动维护的设计文档，与 `docs/swagger*` 自动生成文件不同。
- `docs/wiki/qa/` 设计文档已删除（Q&A 模块已从设计进入实现阶段）。
- `docs/wiki/repowiki/` 新增 `webhook-design.md` 和 `multi-agent-design.md` 设计文档。
- `docs/` 下的 `swagger.json`、`swagger.yaml`、`docs.go` 由 `swag init -g main.go --parseDependency` 自动生成；切勿提交手动编辑。
- `.env` 与 `.env.*` 已被 gitignore；本地开发时从 `.env.example` 复制。
- 雪花算法数据中心/节点 ID 默认为 1/1；可通过 `SNOWFLAKE_DATACENTER_ID` 和 `SNOWFLAKE_NODE_ID` 覆盖。
- Q&A Session 默认最大存活 7 天；可通过 `QA_SESSION_MAX_DURATION`（单位秒）配置。
- RepoWiki 模块需要独立的 LLM Provider 配置；**LLM 配置已从环境变量迁移到数据库热配置**，通过前端「系统设置」页面管理 Provider 和 Model，并为 Agent 角色分配模型。`LLM_ENCRYPT_SECRET` 环境变量（必填）用于加密 API Key。
- RepoWiki 存储路径默认 `./.lumina/repowiki`，可通过 `REPOWIKI_STORAGE_PATH` 配置；支持并发数、超时、配额、重试等精细控制（`REPOWIKI_*` 环境变量族）。
- RepoWiki Webhook 需配置 `REPOWIKI_HMAC_SECRET` 用于 HMAC 签名校验。
- 认证模块已实现（登录、初始化、Token 刷新、Bearer 中间件、资料更新、密码修改）。
- WebAuthn 生物认证已实现（注册/登录/凭证 CRUD/Challenge 缓存），前端集成 `useBiometric` Hook + `lib/webauthn/helpers.ts`。
- API Key 和项目模块已完整实现（后端 CRUD + 前端管理页面）。
- Q&A 模块已完整实现（后端 CRUD + WebSocket 推送 + MCP 工具 + 文件上传下载 + 前端管理页 + Interact 交互页 + 感谢页）。
- Pin 模块已完整实现（后端 Push/Consume/Peek + MCP 工具 + 前端管理页），基于数据库 FIFO 消费。
- RepoWiki 已实现完整 5 角色 SubAgent 编排（Coordinator/Explore/Architect/Writer/Validator）和版本隔离存储，prompt 模板内嵌在 `resources/prompts/`。支持 Webhook 自动触发和 Cron 定时清理。MCP 端只提供 `repoWiki_query` / `repoWiki_list` 只读工具。
- LLM 热配置已实现（Provider/Model CRUD + Agent 角色模型分配 + API Key AES-256-GCM 加密存储 + 前端管理页）。
- SSH Key 模块已实现（密钥对生成/CRUD/公钥导出 + 私钥加密存储 + 前端管理页）。
- Webhook 模块已实现（Git Push 事件接收 + HMAC 校验 + RepoWiki 触发 + 事件历史查询）。
- 系统设置已实现（站点/安全/Q&A/RepoWiki 分组配置读写 + 前端多标签页设置页）。
- MCP Server 已实现，注册了 QA（10 工具）、Project（3 工具）、Pin（5 工具）、RepoWiki（2 只读工具）四套工具共 20 个，通过 API Key 认证。
- WebSocket Hub 已实现，支持 sessionID → deviceID 二级索引，心跳检测，优雅关闭，断线重连和会话恢复。
- Cron Runner 已实现，注册 RepoWiki 定时清理任务（默认每 5 分钟），由 `main.go` 传入 `xMain.Runner` 异步执行。
- Wiki Reader 前端（`web-wiki/`）已实现，独立 SPA 部署在 `/wiki/`，支持密码门认证和只读 Wiki 渲染。

## 引用

- [internal/](./internal/AGENTS.md) — 后端业务层详细文档
- [web/](./web/AGENTS.md) — 控制台前端应用详细文档
- [web-wiki/](./web-wiki/AGENTS.md) — Wiki Reader 前端详细文档
