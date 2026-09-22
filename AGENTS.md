<!-- deep-init:synced@fa47a98 -->

# 项目知识库

## 概述

`Lumina · 微明` — 赋予 AI 深度代码认知与长期记忆的知识中枢。

基于 `bamboo-base-go` 构建的后端服务 + 双前端（TanStack Start 控制台 + Wiki Reader），包含六大核心能力与协作维度：

- **Workspace**：工作空间身份隔离层，项目必须隶属特定空间，支持多空间切换与唯一默认空间 — ✅ 已实现
- **RepoWiki**：克隆项目并通过 5 角色 SubAgent 编排生成结构化 Wiki 文档（Writer 输出 .mdx + frontmatter，manifest 支持 per-folder meta.json + separator/icon，含 Webhook 自动触发 + Cron 定时重试）— ✅ 已实现
- **Memory**：AI 的长期决策记忆，MCP 端主动推送构建（设计中）
- **Q&A**：Agent 与用户的富交互式问答通道（WebSocket 实时推送）— ✅ 已实现
- **Pin**：跨项目依赖约束传递，点对点定向推送与 FIFO 队列消费 — ✅ 已实现
- **Preview / Pages**：前端可视化预览工作台与项目级不可变即时页面（路径式寻址、行级精准编辑、LPW 文档渲染引擎、快照晋升、可选密码门）— ✅ 已实现

后端通过 Streamable MCP 协议 + HTTP REST API + WebSocket 三通道对外暴露能力；MCP 端点认证为 OAuth 2.1（`lum_at_`）优先、API Key 回退，注册了八套共 40 个业务工具，并提供 AI 插件动态打包分发。前端通过 REST API + WebSocket 与后端通信。控制台前端（`web/`）与 Wiki Reader 前端（`web-wiki/`）构建产物分别通过 `go:embed` 嵌入 Go 二进制，支持单文件部署。两前端通过 `@lumina/components` workspace 包共享 shadcn/ui 组件、Markdown 渲染原语、motion 动画变体和微明主题 CSS。项目采用 pnpm monorepo 管理双前端与共享组件包，并提供 Dockerfile + docker-compose 及 GitHub Actions 发布流水线支持容器化部署。数据库与缓存初始化由 `main.go` 的 `xOption.WithDatabase / WithCache` 声明式装配（bamboo-base-go v1.2.3）。

> 工程提案与架构决策按生命周期登记在 `docs/README.md`；当前架构地图见 `ARCHITECTURE.md`。

## 目录结构

```text
./
├── main.go                     # 入口；嵌入双前端资源 → WithDatabase/WithCache 声明式装配 → 运行应用 + WebSocket/Cron Runner
├── go.mod                      # Go 1.25.11；依赖 bamboo-base-go v1.2.3 模块
├── Makefile                    # 开发/测试/格式化/构建/容器化发布命令
├── .env.example                # 必需的环境变量模板
├── Dockerfile                  # 多阶段构建镜像（前端 pnpm build → Go 编译 → 精简运行镜像，EXPOSE 8800）
├── .goreleaser.yaml            # Release 跨平台原始二进制（amd64/arm64，Linux/Windows/macOS）与校验和
├── .dockerignore               # 镜像构建忽略清单
├── docker-compose.yaml         # 精简编排（app + db + redis）
├── docker-compose.full.yaml    # 完整编排（含可选服务）
├── .github/workflows/          # CI/CD 流水线
│   ├── ci.yml                  # PR 自动跑打包校验（gofmt / vet / race）
│   └── docker-publish.yml      # 版本 tag 或手动发版；校验通过后才构建镜像
├── scripts/
│   ├── check_release_version.go # 发布版本连续性预检（正式版逐级、预发布序号连续）
│   └── check_release_version_test.go # 版本转换规则单元测试
├── pnpm-workspace.yaml         # pnpm monorepo 工作区（web + web-wiki + components）
├── pnpm-lock.yaml              # monorepo 根锁文件（由 web/pnpm-lock.yaml 迁移而来）
├── components/                 # @lumina/components 共享 workspace 包
│   ├── package.json            # 包名 @lumina/components，被 web / web-wiki 共同消费
│   ├── components.json         # shadcn/ui 配置（new-york、zinc、lucide）
│   ├── vitest.config.ts        # Vitest 测试配置
│   └── src/                    # 详见 [components/](./components/AGENTS.md)
│       ├── index.ts            # 包导出入口（ui/markdown/motion/styles/hooks/lib）
│       ├── ui/                 # shadcn/ui 组件（含 splitter 等，从 web 迁移而来）
│       ├── markdown/           # Markdown 渲染原语（markdown/markdown-lite/markdown-mermaid/remark-fenced-blocks/table-of-contents/prose/fenced-components）
│       ├── motion/             # motion 动画变体与缓动函数（自 web/src/lib/motion.ts 迁移）
│       ├── styles/             # 微明主题 CSS（theme.css，静烛 v1 设计语言）
│       ├── hooks/              # 共享 Hooks（use-mobile 等）
│       └── lib/                # 共享工具（cn() 等）
├── resources/                  # 项目级内嵌资源（prompts、前端构建产物等）
│   ├── embed.go                # go:embed 暴露 PromptFiles / FrontendDist / WikiFrontendDist / AIPluginFS
│   ├── ai-plugin/              # AI 插件源（.claude-plugin + skills），运行时动态打 ZIP
│   ├── prompts/                # RepoWiki 5 角色 system prompt 文件
│   │   ├── coordinator.md      # Coordinator 角色 prompt
│   │   ├── explore.md          # Explore 角色 prompt
│   │   ├── architect.md        # Architect 角色 prompt
│   │   ├── write.md            # Writer 角色 prompt
│   │   └── validator.md        # Validator 角色 prompt
│   ├── web/dist                # 控制台前端构建产物（pnpm build 产出，go:embed 嵌入）
│   └── web-wiki/dist           # Wiki Reader 前端构建产物（pnpm build 产出，go:embed 嵌入）
├── api/                        # 请求/响应 DTO（按业务域分包，域内按操作类型拆文件）
│   ├── auth/                   # 认证 DTO（initialize/login/refresh/status）
│   ├── user/                   # 用户 DTO（info/password/profile/credential）
│   ├── biometric/              # WebAuthn DTO（availability/login/register）
│   ├── apikey/                 # API Key DTO（create/detail/list/reset/update）
│   ├── workspace/              # 工作空间 DTO（create/detail/list/update）
│   ├── project/                # 项目 DTO（create/detail/update）
│   ├── pin/                    # Pin DTO（create/detail/list/update）
│   ├── repowiki/               # RepoWiki DTO（create_config/detail_config/update_config/version/webhook/wiki）
│   ├── qa/                     # Q&A DTO（session/question/supplement/config）
│   ├── llm/                    # LLM DTO（provider/model/agent）
│   ├── ssh/                    # SSH Key DTO（create/detail/list/update）
│   ├── webhook/                # Webhook DTO（event/response）
│   ├── settings/               # 系统设置 DTO（setting/update）
│   ├── preview/                # Preview DTO（create/detail/list）
│   ├── pages/                  # Pages DTO（list/version/fork/auth）
│   ├── dashboard/              # Dashboard DTO（overview）
│   ├── oauth/                  # MCP OAuth 2.1 DTO（consent/metadata/register/token）
│   ├── plugin/                 # AI 插件市场与 well-known DTO
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
│   │   ├── middleware/         # Gin 中间件（认证拦截、API Key、MCP OAuth+Key、Wiki Auth、Pages 密码门、沙盒隔离、MCP 兼容、CORS、安全头、WebAuthn Origin）
│   │   ├── route/              # 路由注册与中间件绑定（含双前端 SPA fallback + MCP + WebSocket）
│   │   │   ├── route_workspace.go # 工作空间 REST API 路由
│   │   │   ├── route_project.go   # 项目 REST API 路由
│   │   │   ├── route_repowiki.go  # RepoWiki REST API 路由
│   │   │   ├── route_llm.go       # LLM REST API 路由
│   │   │   ├── route_ssh.go       # SSH Key REST API 路由
│   │   │   ├── route_webhook.go   # Webhook 接收路由
│   │   │   ├── route_settings.go  # 系统设置路由
│   │   │   ├── route_preview.go   # Preview REST API 与直出路由
│   │   │   ├── route_pages.go     # Pages REST API 与直出路由
│   │   │   └── route_dashboard.go # Dashboard REST API 路由
│   │   └── startup/            # 业务节点初始化与种子数据（详见子模块文档）
│   │       └── prepare/        # 幂等种子数据（含 workspace/llm/repowiki/settings/preview 种子）
│   ├── handler/                # HTTP 处理器（薄控制器层，含 bind.go 通用绑定辅助）
│   │   ├── workspace.go        # 工作空间处理器
│   │   ├── project.go          # 项目处理器
│   │   ├── repowiki.go         # RepoWiki 配置/版本/分析触发处理器
│   │   ├── wiki_reader.go      # Wiki 内容读取处理器（公开/密码保护/manifest）
│   │   ├── llm.go              # LLM Provider/Model CRUD + Agent 模型分配
│   │   ├── ssh_key.go          # SSH Key CRUD + 密钥生成/公钥导出
│   │   ├── webhook.go          # Webhook 事件接收处理器
│   │   ├── settings.go         # 系统设置处理器
│   │   ├── preview.go          # Preview 会话/文件处理器
│   │   ├── pages.go            # Pages 页面/快照处理器
│   │   ├── serve.go            # 静态资源流式直出与安全沙盒处理器
│   │   ├── dashboard.go        # Dashboard 概览统计处理器
│   │   ├── oauth.go            # MCP OAuth 2.1 处理器
│   │   └── ai_plugin.go        # AI 插件分发处理器
│   ├── logic/                  # 业务编排层（QA 逻辑按职责拆分为 qa_*.go 多文件）
│   │   ├── workspace.go        # 工作空间逻辑（CRUD + 默认空间保护 + 项目关联）
│   │   ├── project.go          # 项目逻辑（CRUD + 空间校验 + 别名解析）
│   │   ├── repowiki_logic.go   # RepoWiki 核心逻辑（分析入口、配置/版本 CRUD）
│   │   ├── repowiki_pipeline.go # RepoWiki 分析管道（Git 准备 + 状态机驱动）
│   │   ├── repowiki_orchestrator.go # 5 角色 SubAgent 编排引擎
│   │   ├── repowiki_subagent_prompts.go # 5 角色 system/user prompt 构建
│   │   ├── repowiki_types.go   # RepoWiki 内部类型（WikiEntry / ValidationError / ExploreOutput）
│   │   ├── repowiki_cron.go    # RepoWiki 定时重试任务逻辑
│   │   ├── repowiki_webhook.go # RepoWiki Webhook 处理逻辑
│   │   ├── llm_provider.go     # LLM Provider 逻辑（CRUD + API Key 加密）
│   │   ├── llm_model.go        # LLM Model 逻辑（CRUD + Agent 角色分配）
│   │   ├── ssh_key.go          # SSH Key 逻辑（CRUD + 密钥对生成）
│   │   ├── settings.go         # 系统设置逻辑（分组配置读写）
│   │   ├── preview_logic.go    # Preview 逻辑（会话/文件管理 + 行级编辑 + 过期清理 + WebSocket 同步）
│   │   ├── preview_lines.go    # Preview 行级编辑与区间读取纯函数
│   │   ├── preview_lpw_logic.go # LPW 1.1 节点文档编排（kind 层级校验 + 节点写入）
│   │   ├── preview_lpw_contract.go # Container variant / Layout pattern / Annotation 契约表（与前端 contract.ts 快照对齐）
│   │   ├── preview_lpw_tree.go # LPW 1.1 节点树纯函数
│   │   ├── pages_logic.go      # Pages 逻辑（快照晋升 + Fork + 密码门 + 静态复制）
│   │   ├── dashboard.go        # Dashboard 逻辑（六类指标聚合）
│   │   ├── runtime_url.go      # 运行时域名解析 + Preview/Pages 深链构建
│   │   ├── oauth_logic.go      # MCP OAuth 2.1 编排
│   │   └── ai_plugin.go        # AI 插件打包编排
│   ├── repository/             # 数据库/Redis 访问层
│   │   ├── workspace.go        # Workspace 持久化
│   │   ├── project.go          # Project 持久化（带 Workspace 关联）
│   │   ├── repowiki_config.go  # RepoWikiConfig 持久化
│   │   ├── wiki_version.go     # WikiVersion 持久化
│   │   ├── llm_provider.go     # LlmProvider 持久化
│   │   ├── llm_model.go        # LlmModel 持久化
│   │   ├── ssh_key.go          # SshKey 持久化
│   │   ├── webhook_event.go    # WebhookEvent 持久化
│   │   ├── preview_session.go  # PreviewSession 持久化（含过期时间）
│   │   ├── preview_file.go     # PreviewFile 持久化
│   │   ├── page.go             # Page 持久化
│   │   ├── page_version.go     # PageVersion 持久化
│   │   ├── page_file.go        # PageFile 持久化
│   │   ├── dashboard.go        # Dashboard 统计持久化
│   │   └── cache/              # Redis 缓存子层（含 workspace、project、repowiki、ssh_key 缓存）
│   ├── service/                # 共享服务层（Git 服务、加密辅助、Webhook 解析、LLM Provider、LPW Schema 等）
│   │   ├── wiki_storage.go     # RepoWiki 文件系统存储与路径管理
│   │   ├── wiki_auth_token.go  # Wiki 访问密码 Token 生成与校验
│   │   ├── page_auth_token.go  # Pages 密码门 Cookie Token 签发与校验
│   │   ├── lpw_schema.go       # LPW JSON Schema 加载与校验服务
│   │   ├── git_service.go      # Git 仓库克隆/拉取服务（go-git）
│   │   ├── agent_factory.go    # LLM Agent 工厂（创建 SubAgent 运行实例）
│   │   ├── crypto_helper.go    # AES-256-GCM 加解密辅助（LLM API Key / SSH 私钥）
│   │   ├── dependency_extractor.go # 依赖关系提取器
│   │   ├── file_scanner.go     # 仓库文件扫描器
│   │   ├── file_cache.go       # 文件缓存管理（上传文件暂存 + 路径穿越防护）
│   │   ├── llm_resolver.go     # LLM Resolver（按 Agent 角色解析模型配置）
│   │   ├── prompt_loader.go    # Prompt 加载器（从 resources/prompts 读取）
│   │   ├── repo_tools.go       # RepoWiki 子 Agent 工具集
│   │   ├── ssh_key_gen.go      # SSH 密钥对生成
│   │   ├── webhook_parser.go   # Webhook Payload 解析器
│   │   └── webhook_signer.go   # Webhook HMAC 签名生成与校验
│   ├── entity/                 # GORM 实体（需实现 GetGene() 绑定，Gene 32~51）
│   ├── mcp/                    # MCP Server 工具注册（Workspace/Project/Pin/QA/RepoWiki/Preview/LPW/Pages 共 38 工具）
│   ├── websocket/              # WebSocket 连接管理 + 消息分发
│   ├── qa/                     # Q&A 回答队列（会话级 FIFO）
│   └── constant/               # 共享业务常量（基因编号、Info 键、LLM、RepoWiki、Preview、Pages、Settings 等）
├── web/                        # TanStack Start 控制台前端（pnpm + Vite）
│   ├── package.json            # React 19 + TanStack Start + Tailwind CSS 4
│   ├── vite.config.ts          # Vite 插件链（含代码拆分配置）
│   ├── components.json         # shadcn/ui（new-york、zinc、lucide）
│   └── src/                    # 前端源码
│       ├── routes/             # 基于文件的路由（公开页 + 认证页 + 控制台 + Interact + Preview 工作台 + Pages 展示态）
│       ├── components/         # 组件（landing/、通用组件和业务子目录；shadcn/ui 与 markdown/motion 原语由 @lumina/components 提供）
│       ├── hooks/              # React Hooks（含 Workspace/LLM/SSH/RepoWiki/Webhook/Settings/Dashboard/Preview/Pages/MCP）
│       ├── lib/                # 工具函数 + API 客户端 + 类型定义 + WebAuthn 辅助 + Cookie 工具
│       ├── styles.css          # 全局样式 + Tailwind 主题
│       └── router.tsx          # TanStack Router 入口
├── web-wiki/                   # TanStack Start Wiki Reader 前端（独立 SPA，部署在 /wiki/）
│   ├── package.json            # React 19 + TanStack Router + Tailwind CSS 4
│   ├── vite.config.ts          # Vite + TanStack Router 插件（base: /wiki/）
│   └── src/                    # 前端源码（只读 .mdx Wiki 渲染 + 密码门 + 目录搜索）
│       ├── lib/                # 工具函数与 API 客户端
│       │   ├── source.ts       # 运行时页面树构建（buildPageTree）与图标映射
│       │   └── frontmatter.ts  # 前端 frontmatter 解析与 TOC 提取
│       └── components/         # 业务组件
│           ├── page-tree-sidebar.tsx # Wiki 侧边导航树
│           ├── docs-page.tsx       # 三栏文档页面布局
│           ├── markdown-renderer.tsx # Markdown 渲染（共享 Markdown + proseArticle 排版）
│           ├── toc.tsx             # 文章目录（scrollspy）
│           ├── breadcrumb.tsx      # 面包屑导航
│           ├── prev-next.tsx       # 上一页/下一页
│           └── search.tsx          # 客户端全文搜索（Orama）
└── .agents/
    ├── skills/                 # 项目专属技能（lumina-qa、lumina-preview、lumina-preview-lpw、lumina-pages、lumina-pin、lumina-repowiki、mcp-qa-test、swagger-writer、entity-build、project-style 及 _shared/ 规范）
    └── plugins/
        └── marketplace.json    # Codex 仓库市场清单（Git 源，非 HTTP archive）
```

## 导航指南

| 任务 | 位置 | 备注 |
|---|---|---|
| 新增 API 接口 | `internal/app/route/`、`internal/handler/` | 先注册路由，再写处理器 |
| 新增业务逻辑 | `internal/logic/` | 保持 handler 精简 |
| 新增持久化逻辑 | `internal/repository/` | 仓库方法返回 `*xError.Error` |
| 新增共享服务 | `internal/service/` | 跨模块复用的基础设施（如 Git 服务、加密辅助、LPW Schema） |
| 新增/修改实体 | `internal/entity/`、`main.go` | 实现并追加到 `WithAutoMigrate(...)` 列表 |
| 新增启动能力 | `internal/app/startup/startup.go` + `startup_*.go` | 以 `xRegNode.RegNodeList` 形式注册 |
| 新增定时任务 | `internal/app/startup/startup_cron.go` | 通过 `xCronRunner` 注册，由 `main.go` 传入 `xMain.Runner` |
| 新增 WebSocket Runner | `internal/app/startup/startup_websocket.go` | 由 `main.go` 传入 `xMain.Runner` |
| 填充默认数据 | `internal/app/startup/prepare/` | 必须保证幂等性 |
| 调整配置/环境变量 | `.env.example`、`internal/app/startup/*.go`、`main.go` | 始终在 `xEnv.GetEnv*` 中提供默认值 |
| 编写 Swagger 文档 | `internal/handler/*.go` 的 godoc + `make swag` | 使用 `swaggo/swag` 注解 |
| 新增请求/响应 DTO | `api/<domain>/` | 按业务域保持子包结构；按操作类型拆分文件（create.go/update.go/detail.go/list.go 等），同一操作的相关 request+response 放在同一文件 |
| 新增中间件 | `internal/app/middleware/` | 返回 `gin.HandlerFunc` |
| 查看物理架构与边界 | `ARCHITECTURE.md` | 给人读的当前架构地图与稳定边界 |
| 贡献代码 / 开 PR | `CONTRIBUTING.md` | 本地开发、质量门、提交与 PR |
| 上报漏洞 | `SECURITY.md` | 只走 GitHub 私密公告，不要开公开 Issue |
| 查看版本变更 | `CHANGELOG.md` | Keep a Changelog，自 v1.0.0-beta.18 起 |
| 查看工程文档索引 | `docs/README.md` | draft / research / RFC / design / ADR 统一入口 |
| 查看 RepoWiki 决策 | `docs/engineering/adr/0004-repowiki-generation-delivery.md` | 生成流水线、版本隔离、Webhook 与只读 MCP |
| 查看 Memory 提案 | `docs/engineering/rfc/0001-memory-decision-memory.md` | 尚未实现的长期决策记忆方案 |
| 查看 Pin 决策 | `docs/engineering/adr/0003-pin-constraint-delivery.md` | 跨项目约束与 FIFO 消费契约 |
| 新增 MCP 工具 | `internal/mcp/` | 注册到 `server.go`，Logic 注入到 `startup_mcp.go` |
| 新增 WebSocket 消息 | `internal/websocket/message.go` | 定义 MessageType 常量 |
| 新增 RepoWiki prompt | `resources/prompts/*.md` | 通过 `service/prompt_loader.go` 加载，禁止硬编码 |
| 新增 Info 配置键 | `internal/constant/info_key.go` | 键名规范：层级 `.` 分隔、同层多词 `-` 连接、禁止 `_` |
| 新增 LLM Provider/Model | `internal/entity/llm_*.go` + `logic/llm_*.go` + `handler/llm.go` | API Key 经 AES-256-GCM 加密 |
| 新增 SSH Key | `internal/entity/ssh_key.go` + `logic/ssh_key.go` + `service/ssh_key_gen.go` | 私钥加密存储 |
| 新增 Webhook 事件 | `handler/webhook.go` + `logic/repowiki_webhook.go` + `service/webhook_parser.go` | HMAC 签名校验必经 |
| 新增 Preview 会话/文件 | `entity/preview_*.go` + `logic/preview_logic.go` + `handler/preview.go` + `mcp/preview_tools.go` | 文件上限 256KB，`preview_sync` 实时同步 |
| 新增 LPW 分块写入操作 | `logic/preview_lpw_logic.go` + `preview_lpw_contract.go` + `mcp/preview_lpw_tools.go` | Schema 1.1、kind 层级与节点增删改查 |
| 新增 Pages 页面/版本 | `entity/page*.go` + `logic/pages_logic.go` + `handler/pages.go` + `mcp/pages_tools.go` | 不可变快照、Fork 派生、路径直出、密码门 Cookie 鉴权 |
| 新增 Dashboard 统计 | `logic/dashboard.go` + `repository/dashboard.go` + `handler/dashboard.go` | 六类指标聚合 |
| 新增 MCP OAuth | `api/oauth/` + `logic/oauth_logic.go` + `handler/oauth.go` + `route_oauth.go` | 公开端点须在 `engine.Use()` 前注册；consent 走登录态 |
| 新增 AI 插件/技能 | `resources/ai-plugin/` + `service/ai_plugin.go` | 运行时打包，禁止手写 marketplace.json 当产物 |
| 新增前端页面 | `web/src/routes/` | 基于文件的路由，文件名即路由路径 |
| 新增前端业务组件 | `web/src/components/<domain>/` | 按业务域组织 |
| 新增前端 API 封装 | `web/src/lib/apis/` | 使用 apiClient 封装 |
| 新增前端数据 Hook | `web/src/hooks/` | 基于 TanStack Query |
| 前端样式调整 | `components/src/styles/theme.css` | 主题色盘已迁到共享包（静烛 v1） |
| shadcn/ui 组件 | `components/src/ui/` | 通过 `pnpm dlx shadcn@latest add <component>` 添加到共享包 |
| 新增共享组件/原语 | `components/src/` | 被 web / web-wiki 共同消费的 ui/markdown/motion/主题，详见 [components/](./components/AGENTS.md) |
| Wiki Reader 前端 | `web-wiki/` | 独立 SPA，部署在 `/wiki/`，只读渲染 |
| 调整 Docker 镜像 | `Dockerfile`、`.dockerignore` | 多阶段构建，前端产物先构建再嵌入 |
| 调整容器编排 | `docker-compose.yaml` / `docker-compose.full.yaml` | 精简/完整两套编排 |
| 调整 CI/CD | `.github/workflows/ci.yml`、`docker-publish.yml` | PR 只跑 `make check`；tag / 手动发版先校验再打包 |

## 代码地图

| 符号 | 类型 | 位置 | 作用 |
|---|---|---|---|
| `main` | 函数 | `main.go` | 嵌入双前端资源 → WithDatabase/WithCache 声明式装配 → 注册启动节点 → 运行应用 + WebSocket/Cron Runner |
| `frontendDist` | 变量 | `resources/embed.go` | `go:embed all:web/dist` 嵌入控制台前端构建产物（导出为 `resources.FrontendDist`） |
| `wikiFrontendDist` | 变量 | `resources/embed.go` | `go:embed all:web-wiki/dist` 嵌入 Wiki Reader 前端构建产物（导出为 `resources.WikiFrontendDist`） |
| `PromptFiles` | 变量 | `resources/embed.go` | `go:embed prompts/*.md` 内嵌 RepoWiki 5 角色 prompt |
| `Init` | 函数 | `internal/app/startup/startup.go` | 启动节点列表工厂（RepoWiki → MCP → Prepare 三个业务节点） |
| `NewCronRunner` | 函数 | `internal/app/startup/startup_cron.go` | Cron Runner 工厂（RepoWiki 超时任务重试 + Preview 过期会话清理，传入 `xMain.Runner`） |
| `NewWebSocketRunner` | 函数 | `internal/app/startup/startup_websocket.go` | WebSocket Runner 工厂（Hub 主循环，传入 `xMain.Runner`） |
| `NewRoute` | 函数 | `internal/app/route/route.go` | 全局中间件 + API 路由 + MCP + WebSocket + 双前端 SPA fallback |
| `NewHandler[T]` | 泛型函数 | `internal/handler/handler.go` | Handler 泛型构造模式，注入全部 Logic（18 个，含 OAuth / AI Plugin / Workspace / Pages） |
| `Auth` | 中间件 | `internal/app/middleware/auth.go` | Bearer Token 认证拦截（单用户模式，注入认证标记） |
| `ApikeyAuth` | 中间件 | `internal/app/middleware/apikey.go` | 纯 API Key 认证（`lumi_` 前缀 + bcrypt）；MCP 端点已改走 `McpAuth` |
| `McpAuth` | 中间件 | `internal/app/middleware/mcp_auth.go` | MCP 认证：OAuth `lum_at_` 优先，API Key 回退，401 带资源元数据 |
| `WikiAuth` | 中间件 | `internal/app/middleware/wiki_auth.go` | Wiki Reader 访问认证（密码 Token / Cookie 会话） |
| `PagesAuth` | 中间件 | `internal/app/middleware/pages_auth.go` | Pages 密码门 HMAC Cookie 校验中间件 |
| `McpCompat` | 中间件 | `internal/app/middleware/mcp_compat.go` | MCP 端点 Streamable HTTP 兼容性处理 |
| `Cors` | 中间件 | `internal/app/middleware/cors.go` | 白名单 CORS（`XLF_ALLOWED_ORIGINS`） |
| `SecurityHeaders` | 中间件 | `internal/app/middleware/security.go` | 安全响应头（nosniff / X-Frame-Options / CSP） |
| `WebAuthnOrigin` | 中间件 | `internal/app/middleware/webauthn.go` | 按请求解析 Origin 注入 context，供 RPID 动态推导 |
| `AuthLogic` | 结构体 | `internal/logic/auth.go` | 单用户认证业务编排（Info 表 + Token 管理 + 资料更新） |
| `BiometricLogic` | 结构体 | `internal/logic/biometric.go` | WebAuthn 生物认证编排（注册/登录/凭证 CRUD + RPID 动态推导） |
| `ApikeyLogic` | 结构体 | `internal/logic/apikey.go` | API Key 业务编排（创建/列表/更新/删除/重置/密钥生成/哈希/脱敏/校验） |
| `WorkspaceLogic` | 结构体 | `internal/logic/workspace.go` | 工作空间业务编排（CRUD + 默认空间保护 + 项目关联） |
| `ProjectLogic` | 结构体 | `internal/logic/project.go` | 项目业务编排（CRUD + 空间校验 + 名称唯一校验 + 别名解析） |
| `PinLogic` | 结构体 | `internal/logic/pin.go` | Pin 约束编排（Push/Consume/Peek/List + 项目解析） |
| `QaLogic` | 结构体 | `internal/logic/qa_logic.go` | Q&A 业务编排（Session/Question/Supplement + 队列消费） |
| `LlmProviderLogic` | 结构体 | `internal/logic/llm_provider.go` | LLM Provider 编排（CRUD + API Key 加密存储） |
| `LlmModelLogic` | 结构体 | `internal/logic/llm_model.go` | LLM Model 编排（CRUD + Agent 角色模型分配） |
| `SshKeyLogic` | 结构体 | `internal/logic/ssh_key.go` | SSH Key 编排（CRUD + 密钥对生成/公钥导出） |
| `SettingsLogic` | 结构体 | `internal/logic/settings.go` | 系统设置编排（分组配置读写 + Info 表编排） |
| `PreviewLogic` | 结构体 | `internal/logic/preview_logic.go` | Preview 会话/文件编排 + 行级编辑 + 过期清理 + WebSocket 同步 |
| `PreviewLpwLogic` | 结构体 | `internal/logic/preview_lpw_logic.go` | LPW 1.1 节点文档编排（kind 层级校验 + 节点写入） |
| `PagesLogic` | 结构体 | `logic/pages_logic.go` | Pages 业务编排（快照晋升 + Fork + 密码门） |
| `DashboardLogic` | 结构体 | `internal/logic/dashboard.go` | 看板六类指标聚合 |
| `ProjectRepo` | 结构体 | `internal/repository/project.go` | 项目持久化（CRUD + Redis Cache-Aside 缓存 + 工作空间过滤） |
| `InitMCPServer` | 函数 | `internal/mcp/server.go` | 创建 MCP Server + 注册 Workspace/QA/Project/Pin/RepoWiki/Preview/LPW/Pages 共 38 工具 + 返回 StreamableHTTPHandler |
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
| `PageAuthToken` | 结构体 | `internal/service/page_auth_token.go` | Pages 访问凭证 HMAC Cookie 签发与校验 |
| `LpwSchemaLoader` | 结构体 | `internal/service/lpw_schema.go` | LPW JSON Schema 加载与静态规范校验 |
| `Workspace` | 结构体 | `internal/entity/workspace.go` | 工作空间实体（Gene=48，名称/描述/图标/默认标记） |
| `RepoWikiConfig` | 结构体 | `internal/entity/repowiki_config.go` | RepoWiki 配置实体（Gene=39，仓库地址/Webhook 配置/自定义提示词/当前选中版本） |
| `WikiVersion` | 结构体 | `internal/entity/wiki_version.go` | Wiki 版本实体（Gene=40，版本号/状态/文件路径/token 统计） |
| `LlmProvider` | 结构体 | `internal/entity/llm_provider.go` | LLM Provider 实体（Gene=41，名称/BaseURL/加密 API Key） |
| `LlmModel` | 结构体 | `internal/entity/llm_model.go` | LLM Model 实体（Gene=42，Provider 关联/Agent 角色分配） |
| `WebhookEvent` | 结构体 | `internal/entity/webhook_event.go` | Webhook 事件实体（Gene=43，事件 ID/分支/状态） |
| `SshKey` | 结构体 | `internal/entity/ssh_key.go` | SSH Key 实体（Gene=44，名称/指纹/公钥/加密私钥） |
| `PreviewSession` | 结构体 | `internal/entity/preview_session.go` | Preview 会话实体（Gene=45，Hash/标题/状态/过期时间） |
| `PreviewFile` | 结构体 | `internal/entity/preview_file.go` | Preview 文件实体（Gene=46，会话关联/文件名/MIME/内容） |
| `OAuthClient` | 结构体 | `internal/entity/oauth_client.go` | OAuth 动态客户端（Gene=47，公共客户端，仅 PKCE） |
| `Page` | 结构体 | `internal/entity/page.go` | Page 实体（Gene=49，项目内 slug/生效指针/访问策略） |
| `PageVersion` | 结构体 | `internal/entity/page_version.go` | PageVersion 实体（Gene=50，不可变快照版本） |
| `PageFile` | 结构体 | `internal/entity/page_file.go` | PageFile 实体（Gene=51，版本内扁平文件） |
| `OAuthLogic` | 结构体 | `internal/logic/oauth_logic.go` | MCP OAuth 2.1 编排（DCR / 授权码+PKCE / 令牌 / RFC 8707） |
| `AIPluginLogic` | 结构体 | `internal/logic/ai_plugin.go` | AI 插件市场清单 / ZIP / well-known 编排 |
| `AIPluginService` | 结构体 | `internal/service/ai_plugin.go` | 从 embed 或调试目录打包插件 |
| `AIPluginFS` | 变量 | `resources/embed.go` | `go:embed all:ai-plugin` 内嵌插件与技能源 |
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
| `useWorkspace` | Hook | `web/src/hooks/useWorkspace.ts` | 工作空间 Hook（CRUD + 列表 + 切换） |
| `useCurrentWorkspace` | Hook | `web/src/hooks/useCurrentWorkspace.ts` | 当前激活空间状态 Hook（持久化 + 默认回退） |
| `useQaWebSocket` | Hook | `web/src/hooks/useQaWebSocket.ts` | Q&A WebSocket 连接管理（自动重连 + 消息回调 + 会话恢复） |
| `useQaSession` | Hook | `web/src/hooks/useQaSession.ts` | Q&A 会话状态管理（问题推送 + 回答提交 + 文件上传） |
| `useQaAdmin` | Hook | `web/src/hooks/useQaAdmin.ts` | Q&A 管理端数据 Hook（会话列表/详情/删除/配置） |
| `usePages` | Hook | `web/src/hooks/usePages.ts` | Pages 页面 Hook（列表/详情/版本/晋升/密码门/Fork） |
| `useLlmConfig` | Hook | `web/src/hooks/useLlmConfig.ts` | LLM 配置 Hook（Provider/Model CRUD + Agent 模型分配） |
| `useSshKey` | Hook | `web/src/hooks/useSshKey.ts` | SSH Key 数据 Hook（CRUD + 分页） |
| `useRepoWiki` | Hook | `web/src/hooks/useRepoWiki.ts` | RepoWiki Hook（配置/版本/分析触发/Webhook 事件） |
| `useSettings` | Hook | `web/src/hooks/useSettings.ts` | 系统设置 Hook（分组配置读写） |
| `useDashboard` | Hook | `web/src/hooks/useDashboard.ts` | 看板统计 Hook（useDashboardOverview，多页复用 KPI） |
| `usePreviewAdmin` | Hook | `web/src/hooks/usePreviewAdmin.ts` | Preview 管理 Hook（会话/文件 CRUD） |
| `usePreviewWebSocket` | Hook | `web/src/hooks/usePreviewWebSocket.ts` | Preview WebSocket Hook（preview_sync 实时同步 + 重连） |
| `formatAnswer` | 函数 | `web/src/lib/format-answer.ts` | 题型 answer 格式化（各题型 → 可读字符串，跨 QA/interact 复用） |
| `ReadPage` | 方法 | `internal/service/wiki_storage.go` | 读取 .mdx 页面文件并解析 YAML frontmatter |
| `computeNav` | 函数 | `internal/handler/wiki_reader.go` | 根据 manifest 计算当前页的 prev/next/breadcrumb |
| `buildPageTree` | 函数 | `web-wiki/src/lib/source.ts` | 从 manifest 构建运行时页面树（含 parent 指针与 leaves） |
| `PageTreeSidebar` | 组件 | `web-wiki/src/components/page-tree-sidebar.tsx` | Wiki 侧边导航树（递归目录 + separator + 展开状态） |
| `DocsPage` | 组件 | `web-wiki/src/components/docs-page.tsx` | 三栏文档页面布局（Sidebar | Article | TOC） |

## 模块架构

物理边界、缺席性约束与横切约定以 [ARCHITECTURE.md](./ARCHITECTURE.md) 为准（给人读的架构地图）。下面只留 AI 开工时够用的摘要。

```
对外接入：MCP（OAuth 2.1 / API Key）+ REST + WebSocket + Webhook + 插件市场
        │
领域：Workspace ○ RepoWiki ○ Memory（设计中）○ Q&A ○ Pin ○ Preview / Pages —— 互不调用
        │
基础设施：PostgreSQL + Redis + LLM 热配置 + Git + 文件系统
```

- 各领域互不 import；Agent 经 MCP 自行编排。
- 分层单向：`route` → `handler` → `logic` → `repository`；`service` 只放跨域基础设施。
- Q&A / Preview 用 WebSocket，不用 SSE。
- 控制台 `web/`、Wiki Reader `web-wiki/` 经 `go:embed` 进同一二进制；UI 原语在 `@lumina/components`。
- 工程提案与架构决策从 `docs/README.md` 进入；稳定架构边界以 `ARCHITECTURE.md` 为准。

## 约定

- **框架文档优先**：本项目基于 `bamboo-base-go` 框架构建，使用任何框架组件（`xError`、`xResult`、`xLog`、`xEnv`、`xCtxUtil` 等）前，**必须先通过 `bamboo-document` MCP 查阅官方文档**（板块标识：`bamboo-base-go`），确认 API 签名、参数语义和推荐用法后再编码。禁止凭记忆或猜测使用框架 API。
- **导入别名**：bamboo-base-go 包使用 `x*` 别名（`xLog`、`xEnv`、`xError`、`xResult`、`xReg`、`xCron`、`xCronRunner` 等）。
- **严格分层**：route -> handler -> logic -> repository；禁止跳层调用。`service/` 为跨模块共享服务层，可被 logic 调用。
- **声明式装配**：数据库与缓存由 `main.go` 的 `xOption.WithDatabase(xOptDatabase.FromEnv() + WithAutoMigrate)` / `xOption.WithCache(xOptCache.FromEnv())` 声明式装配；`startup.go` 仅注册依赖 db/rdb 的业务节点（RepoWiki/MCP/Prepare）。
- **上下文依赖注入**：启动阶段将基础设施注册到 context；逻辑层通过 `xCtxUtil.MustGetDB/MustGetRDB` 获取；RepoWikiLogic 通过 `logic.GetRepoWikiLogicFromContext` 获取。
- **响应模式**：handler 通过 `xResult.SuccessHasData` 返回成功；错误通过 `ctx.Error` 传递。
- **错误类型**：仓库/逻辑层使用 `*xError.Error` 表示业务/基础设施故障。
- **环境变量族**：`XLF_*`、`APP_*`、`DATABASE_*`、`NOSQL_*`、`LUMINA_*`、`SNOWFLAKE_*`、`LLM_*`、`QA_*`、`REPOWIKI_*`、`XLF_BIOMETRIC_*`。OAuth TTL 用 `LUMINA_OAUTH_ACCESS_TTL` / `LUMINA_OAUTH_REFRESH_TTL`（秒）；插件源目录用 `LUMINA_AI_PLUGIN_DIR`。
- **实体 ID 策略**：采用雪花算法基因策略；每个实体必须实现 `GetGene() xSnowflake.Gene`，基因编号定义在 `constant/gene_number.go`（GeneProject=32 ~ GenePageFile=51）。
- **字段注释**：实体字段必须追加行尾中文注释（`// 字段说明`），且与 `gorm comment` 保持一致。
- **Info 配置键统一**：所有 Info 表键名在 `constant/info_key.go` 集中定义，禁止在业务代码写死键名字符串；键名规范为层级 `.` 分隔、同层多词 `-` 连接、禁止 `_`（如 `qa.session.ttl`）。
- **Swagger 注册**：仅在 `XLF_DEBUG=true` 时注册 Swagger UI。
- **WebSocket 实时推送**：Q&A 与 Preview 模块使用 WebSocket 进行实时推送（非 SSE）。
- **模块独立性**：核心领域模块（Workspace、RepoWiki、Memory、Q&A、Pin、Preview、Pages）互不调用，Agent 通过 MCP 自行编排。
- **双前端嵌入部署**：通过 `go:embed` 将 `resources/web/dist` 和 `resources/web-wiki/dist` 分别嵌入 Go 二进制，构建命令 `make generate` 完成前端打包 → Swagger → Go 编译全流程。
- **MCP 路由注册**：MCP 端点必须在 `engine.Use()` 之前注册以绕开 `ResponseMiddleware`。OAuth well-known / authorize / register / token 同样需要裸 JSON，必须在同一时机注册。
- **MCP 认证**：MCP 路由挂 `middleware.McpAuth`（`lum_at_` OAuth 优先，其余 Bearer 走 API Key）。不要改回只挂 `ApikeyAuth`，否则客户端拿不到 `WWW-Authenticate` 资源元数据。
- **安全中间件**：`middleware.SecurityHeaders` / `middleware.Cors` / `middleware.WebAuthnOrigin` 在 `route.go` 全局注册，提供安全响应头、白名单 CORS 与 WebAuthn Origin 解析。
- **共享组件包**：shadcn ui 组件、Markdown 渲染原语、motion 动画变体、微明主题 CSS 统一在 `@lumina/components` workspace 包管理，被 `web` 和 `web-wiki` 共同消费。
- **pnpm monorepo**：`web`、`web-wiki`、`components` 三个 workspace 包由根 `pnpm-workspace.yaml` 统一管理，锁文件为根目录 `pnpm-lock.yaml`；在根目录执行 `pnpm install` 安装全部依赖。
- **DTO 拆分**：`api/<domain>/` 内按操作类型拆分文件（`create.go`/`update.go`/`detail.go`/`list.go` 等），同一操作的相关 request+response 放在同一文件，禁止将所有 DTO 塞入单文件。
- **资源内嵌**：RepoWiki 5 角色 system prompt 通过 `go:embed` 内嵌在 `resources/prompts/`，由 `service/prompt_loader.go` 读取；前端构建产物与 AI 插件源由 `resources/embed.go` 统一 `go:embed` 暴露（`AIPluginFS`），禁止在 logic 中硬编码 prompt 文本。
- **LLM 热配置**：LLM Provider/Model 配置存储在数据库（非环境变量），通过前端管理页面配置；API Key 经 `service/crypto_helper.go` AES-256-GCM 加密，密钥由 `LLM_ENCRYPT_SECRET` 环境变量提供。
- **加密存储**：LLM API Key 和 SSH 私钥必须经 AES-256-GCM 加密后存储，禁止明文落库。
- **Webhook 签名校验**：所有 Webhook 请求必须经 `service/webhook_signer.go` 校验 HMAC 签名，密钥由 `REPOWIKI_HMAC_SECRET` 环境变量提供。
- **Cron 任务**：定时任务通过 `xCronRunner` 注册（`startup_cron.go`），包含 RepoWiki 超时重试与 Preview 过期会话清理，由 `main.go` 传入 `xMain.Runner` 异步执行，不阻塞启动节点链。
- **容器化构建与发版**：`Dockerfile` 采用多阶段构建；镜像标签由 `Makefile` 的 `validate-version` + `docker-build` + `publish` 目标驱动。版本 tag 或手动发版先通过 `go run ./scripts/check_release_version.go` 校验版本连续性，再执行 `make check`（版本规则测试 / gofmt / vet / race）；AI Action 固定通过 `https://ai-intl.x-lf.com/v1` 的 `glm-5.3-flash` 生成正文，只读取 `AI_API_KEY` Secret；GoReleaser 构建 amd64/arm64 的 Linux、Windows 与 macOS 原始二进制并直接上传 GitHub Release。跨版本预检失败时清理本次 tag 并跳过全部构建。PR 只跑质量门。
- **子模块约定**：后端分层详情见 [internal/](./internal/AGENTS.md)，控制台前端专属约定见 [web/](./web/AGENTS.md)，Wiki Reader 前端约定见 [web-wiki/](./web-wiki/AGENTS.md)，共享组件包见 [components/](./components/AGENTS.md)。

## 反模式

- 禁止凭记忆或猜测使用 `bamboo-base-go` 框架 API；使用前必须通过 `bamboo-document` MCP（板块 `bamboo-base-go`）查阅官方文档。
- 禁止直接使用 `os.Getenv`；应使用带默认值的 `xEnv.GetEnv*`。
- 禁止手动编辑 `docs/swagger*` 文件；它们由 `swag init` 自动生成。
- 禁止核心业务模块之间直接调用；Agent 通过 MCP 自行编排。
- 禁止在 Q&A 模块使用 SSE 进行问题推送；统一使用 WebSocket。
- 禁止明文存储 LLM API Key 或 SSH 私钥；必须经 `crypto_helper.go` AES-256-GCM 加密。
- 禁止在业务代码中写死 Info 键名字符串；统一用 `constant/info_key.go` 定义的常量。
- 禁止在 logic 中硬编码 RepoWiki prompt 文本；统一放 `resources/prompts/` 通过 `prompt_loader.go` 加载。
- 禁止在 Webhook 处理中跳过 HMAC 签名校验。
- 禁止在 `resources/prompts/` 外散落 prompt 文件；所有内嵌静态资源集中在 `resources/` 目录。
- 禁止通过裸 goroutine 调度定时任务；统一走 `xCronRunner`。
- 禁止在 `web` / `web-wiki` 内重新创建已迁入 `@lumina/components` 的 ui/markdown/motion/主题代码；应通过共享包导入。
- 禁止给 MCP 端点只挂 `ApikeyAuth`；必须走 `McpAuth`，否则 OAuth 客户端无法发现授权服务器。
- 禁止把 OAuth 令牌原文写入 Redis；缓存键必须是 SHA-256 摘要。
- 禁止手写运行时 marketplace.json；清单由 `AIPluginService` 按当前域名和 ZIP SHA-256 渲染。ZCode 不支持 archive 源，必须使用 `marketplace.zcode.json`。
- 子模块反模式详见各子模块 AGENTS.md。

## 独特风格

- **日志命名**：遵循模块标签（`NamedMAIN`、`NamedINIT`、`NamedCONT`、`NamedLOGC`、`NamedREPO`、`NamedMIDE`、`NamedCRON`）。
- **启动种子阶段**：显式通过 `xCtx.Exec` 节点执行，并隔离在 `prepare/` 目录中。
- **实体 ID 基因策略**：实体级别需绑定基因类型（`GeneProject = 32` ~ `GenePageFile = 51`），定义在 `constant/gene_number.go`。
- **项目技能**：`.agents/skills/` 包含项目专属技能：`lumina-qa`、`lumina-preview`、`lumina-preview-lpw`、`lumina-pages`、`lumina-pin`、`lumina-repowiki`、`mcp-qa-test`、`swagger-writer`、`entity-build`、`project-style` 及 `_shared/` 共享规范。
- **双通道暴露**：每个模块同时提供 REST API 和 MCP Tool。
- **MCP 编排**：Lumina 不做跨模块编排，由 Agent 端自行决定调用顺序和组合。
- **泛型 Handler 构造**：`NewHandler[T]` 统一注入所有 logic 实例（18 个 Logic，含 OAuth / AI Plugin / Workspace / Pages）。
- **双前端嵌入**：`resources/embed.go` 统一管理 `FrontendDist` + `WikiFrontendDist` + `PromptFiles` + `AIPluginFS`，配合 `route_frontend.go` 实现双 SPA fallback，单二进制部署。
- **pnpm monorepo 共享包**：`@lumina/components` workspace 包集中管理 shadcn/ui、Markdown 渲染原语（含 remark fenced-blocks 插件）、motion 动画变体、微明主题 CSS，被 `web` 和 `web-wiki` 共同消费。
- **DTO 按操作类型拆分**：`api/` 各业务域内以 `create.go`/`update.go`/`detail.go`/`list.go` 等操作命名文件，同一操作 request+response 同文件。
- **API Key 安全**：`lumi_` 前缀 + base64 RawURL 编码 + bcrypt 哈希，仅创建/重置时返回完整密钥。
- **API Key 认证中间件**：`middleware.ApikeyAuth` 仍是纯 API Key 校验实现；MCP 端点改挂 `McpAuth`（OAuth 优先 + API Key 回退），与 `middleware.Auth`（控制台 Bearer Token）分离。
- **MCP OAuth 2.1**：Lumina 同时充当授权服务器与资源服务器；仅 PKCE S256 公共客户端（无 secret）；访问令牌前缀 `lum_at_`；同意页在 `/_public/oauth`。
- **AI 插件动态打包**：`resources/ai-plugin` 经 embed 后按请求域名写入 `.mcp.json` 再打 ZIP；调试态热重载；Codex 市场走仓库内 `.agents/plugins/marketplace.json`（Git 源）。
- **Wiki Auth 中间件**：`middleware.WikiAuth` 处理 Wiki Reader 的密码 Token / Cookie 会话认证，与控制台认证分离。
- **Project 缓存**：三层映射（ID→详情、Name→ID、Alias→ID），Cache-Aside 模式，30 分钟 TTL。
- **WebSocket Hub**：按 sessionID → deviceID 二级索引管理连接，连接 `Kind` 区分 qa/preview，心跳检测间隔 5s / 超时 15s，支持会话恢复。
- **Q&A 回答队列**：会话级 FIFO 队列，支持 `WaitAndConsume` 阻塞等待新回答（MCP 工具消费）。
- **Q&A 推送回调**：`logic.OnQuestionPushed` / `logic.OnSupplementPushed` / `logic.OnQuestionCancelled` / `logic.OnSessionArchived` 函数变量在 `route_ws.go` 中设置，解耦 Logic 层和 WebSocket 层。
- **Q&A 逻辑拆分**：原 `qa.go`（1778 行）按职责拆分为 `qa_logic.go`（核心编排）、`qa_format.go`（题型格式化）、`qa_helper.go`（辅助函数）、`qa_mcp.go`（MCP 工具）、`qa_mcp_helpers.go`（MCP 辅助）、`qa_download.go`（文件下载）。
- **MCP 工具拆分**：`mcp/qa_tools.go` 拆分为 `qa_tools.go`（注册）、`qa_handlers.go`（handler 实现）、`qa_type_details.go`（题型 schema 细节）。
- **Q&A 题型格式化**：Logic 层内置 15+ 题型格式化函数（select/multi-select/text/boolean/code/image/file/slider/rank/rate/plan/options/diff/review）。
- **Pin FIFO 消费**：基于数据库实现 FIFO（`ConsumeOldestPending` 按 createdAt 升序 + `ConsumeByID` 精确消费），不依赖 Redis 队列。
- **WebAuthn 集成**：后端 `logic/biometric.go` + `webauthn_user.go` 适配器，前端 `lib/webauthn/helpers.ts` 处理浏览器端编解码；RPID 按请求 Origin 动态推导，支持注册域后缀共享凭证。
- **Interact 前端页面**：独立布局（非 Console），支持 15+ 种题型组件 + 交互原语（primitives/），WebSocket 实时交互，断线重连。
- **前端通用组件**：`confirm-delete-dialog.tsx`、`page-header.tsx`、`skeleton-table.tsx` 跨模块复用，替代各域重复的删除对话框和页面头部。
- **前端首页拆分**：`routes/_public/index.tsx` 拆分为 `components/landing/` 下 hero/features/tech 区块组件。
- **HTML 沙盒隔离**：`sandbox-frame.tsx` 用 iframe `sandbox="allow-scripts"`（不加 `allow-same-origin`）实现 opaque origin 隔离，替代旧 Shadow DOM + DOMPurify 白名单方案。
- **静烛 v1 设计语言**：共享 `theme.css` 全量重写，色盘为静烛/微明意象（`--sea-ink`/`--lagoon`/`--palm`/`--sand`/`--foam`），`--radius:0px` 全平直角。
- **Preview 实时同步**：`OnPreviewChanged` 回调驱动 `preview_sync` 消息推送，对外分享页与管理端实时同步会话文件变更。
- **Preview 工作台**：路径式寻址 `/preview/<hash>/<file>`，必须登录（Cookie 回退）；iframe 与地址栏同路径，相对 CSS/JS 原生命中。旧 `?session=` 深链由前端重定向。
- **LPW 文档渲染引擎**：支持基于 Schema 1.1 校验的三类同级节点结构化写入（追加/插入/修改/删除/巡检），前端优先使用 React 原生渲染与 ECharts 懒加载，防范 DOM 竞态。
- **Pages 快照与密码门**：不可变即时页面快照（深拷贝 Preview 文件），支持版本指针与 Fork 回退；密码门走 HMAC 签名 Cookie 校验。
- **Dashboard KPI 聚合**：`repository/dashboard.go` 用原生 SQL 聚合六类指标，前端 `useDashboardOverview` 被多个 Console 页复用 KPI。
- **RepoWiki 子 Agent 编排**：`SubAgentOrchestrator` 按预定义 5 阶段（Coordinator → Explore → Architect → Writer → Validator）生成 Wiki，prompt 模板内嵌在 `resources/prompts/*.md` 通过 `service/prompt_loader.go` 加载，`repowiki_subagent_prompts.go` 负责动态构建 user prompt，`repowiki_types.go` 定义内部类型，`repowiki_pipeline.go` 负责 Git 准备与状态机驱动。
- **RepoWiki 版本隔离**：每个 Wiki 版本存储在 `versions/{vid}/` 下，`RepoWikiConfig.SelectedVersionID` 指定当前对外服务版本；旧版 config 级目录已废弃，新版本完成后清理。
- **RepoWiki MCP 只读**：MCP 端仅暴露 `repoWiki_query` / `repoWiki_list` 两个只读工具，Wiki 更新由 Git Webhook 自动触发。
- **RepoWiki Webhook 自动触发**：Git Push 事件经 `service/webhook_signer.go` HMAC 校验 + `service/webhook_parser.go` 解析后，由 `logic/repowiki_webhook.go` 匹配分支规则并触发分析。
- **RepoWiki Cron 重试**：`startup_cron.go` 注册定时任务（默认每 5 分钟），通过 `logic/repowiki_cron.go` 的 `RetryStaleTask` 扫描超时任务并重试或标记失败。
- **LLM 热配置**：Provider/Model 配置存储在数据库，支持 Agent 角色（Coordinator/Explore/Architect/Writer/Validator）分别分配不同模型；API Key 经 AES-256-GCM 加密存储。
- **文件下载 Token**：`service/download_token.go` 生成短时效签名 Token，用于 Q&A 文件附件下载鉴权。
- **媒体回答处理**：`service/media_answer.go` 处理图片/文件附件的回答格式化，供 Q&A MCP 工具调用。
- **文件缓存防护**：`service/file_cache.go` 的 `IsWithinCacheDir` 用绝对路径 + `EvalSymlinks` + `filepath.Rel` 前缀校验，防路径穿越与符号链接逃逸。
- **双前端共享**：`@lumina/components` workspace 包统一管理 shadcn/ui 组件、Markdown 渲染原语、motion 动画变体、微明主题 CSS，被 web 和 web-wiki 共同消费。
- **内嵌资源集中管理**：`resources/` 目录集中管理项目级内嵌静态资源（prompt 文件 + 双前端构建产物 + AI 插件源），通过 `go:embed` 暴露给各业务包引用，避免资源文件散落在业务包内部。
- **Runner 模式**：`startup.NewWebSocketRunner()` 与 `startup.NewCronRunner()` 返回的函数由 `main.go` 传入 `xMain.Runner` 的 goroutineFunc 参数，与启动节点链解耦异步执行。
- **容器化部署**：`Dockerfile` 多阶段构建 + `docker-compose.yaml`（精简）/ `docker-compose.full.yaml`（完整）双编排。`ci.yml` 在 PR 上跑打包校验；`docker-publish.yml` 在版本 tag 或手动发版时，依次执行校验、Docker 镜像发布、AI 正文生成与 GoReleaser 二进制发布。

## 常用命令

```bash
# ── 后端 ──

# 初始化
cp .env.example .env
go mod tidy

# 开发
make dev-backend   # 生成 Swagger 文档并运行（推荐）
make dev-frontend  # 启动控制台前端开发服务器（端口 3000）
make dev-wiki-frontend # 启动 Wiki Reader 前端开发服务器（端口 3001）
make swag          # 仅生成 Swagger 文档
make run           # 运行已编译二进制

# 一键构建（双前端打包 → Swagger → Go 编译）
make generate      # 或 make build

# 质量
make tidy          # 整理 Go 模块
make fmt           # 格式化代码
make test          # 运行测试
make vet           # go vet 静态检查
make lint          # golangci-lint 检查
make test-release-version # 测试发布版本连续性规则
make check         # 打包前校验：版本规则测试 + gofmt + go vet + go test -race

# 验证（端口由 XLF_PORT 决定，默认 8080；容器内默认 8800）
curl http://localhost:8080/api/v1/health/ping

# ── 容器化 / 发布 ──

make validate-version  # 校验镜像 tag 与版本一致性
make docker-build      # 构建 Docker 镜像
make publish VERSION=vX.X.X  # 推送代码并打 Git Tag 推送至远端触发 Release 流水线（别名 make public）

# 或直接使用 docker compose
docker compose up -d                    # 精简编排（app + db + redis）
docker compose -f docker-compose.full.yaml up -d  # 完整编排

# ── monorepo（根目录）──

pnpm install       # 安装 web / web-wiki / components 全部依赖

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
pnpm dlx shadcn@latest add <component>  # 添加 UI 组件（输出到 @lumina/components）

# ── Wiki Reader 前端 (web-wiki/) ──

cd web-wiki
pnpm install      # 安装依赖
pnpm dev          # 开发服务器（端口 3001）
pnpm build        # 类型检查 + 生产构建
pnpm test         # 运行 Vitest 测试
pnpm lint         # ESLint 检查
pnpm format       # Prettier 格式化 + ESLint 自动修复

# ── 共享组件包 (components/) ──

cd components
pnpm build        # 构建 @lumina/components
pnpm test         # 运行 Vitest 测试（markdown/remark-fenced-blocks）
```

## 备注

- 双前端通过 `go:embed` 嵌入 Go 二进制，构建顺序：先 `pnpm build`（产出 `resources/web/dist` + `resources/web-wiki/dist`），再 `go build`。使用 `make generate` 一键完成。
- 前端独立开发时使用 `make dev-frontend`（控制台 Vite dev server 端口 3000）和 `make dev-wiki-frontend` / `cd web-wiki && pnpm dev`（Wiki Reader 端口 3001），但生产部署时前后端合一。
- CI：`.github/workflows/ci.yml` 在 Pull Request 上自动执行 `make check`。`.github/workflows/docker-publish.yml` 在版本 tag 或手动触发时先校验版本连续性；正式版本 major / minor / patch 只能逐级递增，预发布序号必须连续。预检通过后运行 Go 质量门、构建镜像；发布模式再生成 AI Release 正文，并通过 `.goreleaser.yaml` 上传跨平台二进制与 SHA-256 校验和。发版任一步失败时自动删除仍指向本次提交的远端版本 tag，tag 已移动则拒绝误删。
- 项目已 pnpm monorepo 化：锁文件为根目录 `pnpm-lock.yaml`，`web/pnpm-lock.yaml` 与 `web-wiki/pnpm-lock.yaml` 已移除；依赖安装统一在根目录执行 `pnpm install`。
- 依赖已升级：bamboo-base-go 全模块升至 v1.2.3（DB/Cache 改由 `xOption.WithDatabase/WithCache` 声明式装配），Go 1.25.11。
- `make test` 命令存在，已有测试用例覆盖 project、llm_model、llm_provider、ssh_key_gen、webhook_parser、webhook_signer、wiki_auth_token、wiki_storage、git_service、file_scanner、dependency_extractor、repowiki_orchestrator、biometric、runtime_url、prompt_loader、preview_tools、webauthn_user、oauth_logic、ai_plugin、preview MIME 等模块。
- 根位置公约文件：`ARCHITECTURE.md`（架构边界）、`CONTRIBUTING.md`、`CODE_OF_CONDUCT.md`（Contributor Covenant 2.1）、`SECURITY.md`、`SUPPORT.md`、`CHANGELOG.md`（自 v1.0.0-beta.18 起）。`README.md` 与 `LICENSE` 为既有人工文件，未改写。
- 工程文档统一位于 `docs/engineering/`，由 `docs/README.md` 登记；Memory 仍处于 RFC，已实现架构、Project、Pin 与 RepoWiki 约束记录为 proposed ADR。
- `docs/` 下的 `swagger.json`、`swagger.yaml`、`docs.go` 由 `swag init -g main.go --parseDependency` 自动生成；切勿提交手动编辑。
- `.env` 与 `.env.*` 已被 gitignore；本地开发时从 `.env.example` 复制。
- 雪花算法数据中心/节点 ID 默认为 1/1；可通过 `SNOWFLAKE_DATACENTER_ID` 和 `SNOWFLAKE_NODE_ID` 覆盖。
- Q&A Session 默认最大存活 7 天；可通过 `QA_SESSION_MAX_DURATION`（单位秒）配置。
- 数据库/缓存驱动通过 `DATABASE_DRIVER`（postgres/mysql/sqlite/oracle/sqlserver）与 `NOSQL_DRIVER`（redis/memory/none）声明式选择；为空则禁用对应组件。
- 安全环境变量：`XLF_INITIALIZE_TOKEN`（一次性初始化令牌，防公开初始化接口被抢先接管）、`XLF_ALLOWED_ORIGINS`（CORS 白名单）、`XLF_BIOMETRIC_ALLOWED_ORIGINS`（WebAuthn RPID 动态推导白名单，防 DNS-rebinding）。
- RepoWiki 模块需要独立的 LLM Provider 配置；**LLM 配置已从环境变量迁移到数据库热配置**，通过前端「系统设置」页面管理 Provider 和 Model，并为 Agent 角色分配模型。`LLM_ENCRYPT_SECRET` 环境变量（必填）用于加密 LLM API Key 与 SSH 私钥。
- RepoWiki 存储路径默认 `./.lumina/repowiki`，可通过 `REPOWIKI_STORAGE_PATH` 配置；支持并发数、超时、配额、重试等精细控制（`REPOWIKI_*` 环境变量族）。
- RepoWiki Webhook 需配置 `REPOWIKI_HMAC_SECRET` 用于 HMAC 签名校验。
- RepoWiki 生成产物为 `.mdx` 格式（含 YAML frontmatter），**旧版 `.md` 已不再兼容**；`RepoWikiConfig` 已移除冗余 `Name` 字段，新增 `CustomPrompt` 双层提示词机制（全局 prompt + 项目级自定义提示词）。
- 认证模块已实现（登录、初始化、Token 刷新、Bearer 中间件、资料更新、密码修改）。
- WebAuthn 生物认证已实现（注册/登录/凭证 CRUD/Challenge 缓存 + RPID 按请求 Origin 动态推导），前端集成 `useBiometric` Hook + `lib/webauthn/helpers.ts`。
- 工作空间（Workspace）模块已完整实现（后端 CRUD + 默认空间保护 + 项目关联 + 前端切换器与管理页）。
- API Key 和项目模块已完整实现（后端 CRUD + 前端管理页面）。
- Q&A 模块已完整实现（后端 CRUD + WebSocket 推送 + MCP 工具 + 文件上传下载 + 前端管理页 + Interact 交互页 + 感谢页）。
- Pin 模块已完整实现（后端 Push/Consume/Peek + MCP 工具 + 前端管理页），基于数据库 FIFO 消费。
- RepoWiki 已实现完整 5 角色 SubAgent 编排（Coordinator/Explore/Architect/Writer/Validator）和版本隔离存储，prompt 模板内嵌在 `resources/prompts/`。支持 Webhook 自动触发和 Cron 定时重试。MCP 端只提供 `repoWiki_query` / `repoWiki_list` 只读工具。
- Preview 模块已实现（会话/文件 CRUD + 行级精准替换/区间读取 + LPW 渲染引擎 + WebSocket 实时同步 + 路径式工作台 + 对外分享页 + 管理端），单文件上限 256KB。
- Pages 模块已完整实现（Preview 会话一键晋升快照 + 访问策略配置 + 密码门 HMAC Cookie 鉴权 + 路径直出 + Fork 派生回 Preview 会话）。
- Dashboard 看板已实现（`GET /dashboard/overview` 六类指标聚合 + 前端 KPI 分栏 + bento grid）。
- LLM 热配置已实现（Provider/Model CRUD + Agent 角色模型分配 + API Key AES-256-GCM 加密存储 + 前端管理页）。
- SSH Key 模块已实现（密钥对生成/CRUD/公钥导出 + 私钥加密存储 + 前端管理页）。
- Webhook 模块已实现（Git Push 事件接收 + HMAC 校验 + RepoWiki 触发 + 事件历史查询）。
- 系统设置已实现（站点/安全/Q&A/RepoWiki/Preview 分组配置读写 + 前端多标签页设置页）。
- 安全中间件已实现（CORS 白名单 + 安全响应头 + WebAuthn Origin 解析 + Pages 密码门 + 沙盒子资源隔离）。
- MCP Server 已实现，注册了 Workspace（2 只读工具）、QA（10 工具）、Project（3 工具）、Pin（5 工具）、RepoWiki（2 只读工具）、Preview（7 基础工具）、Preview LPW（8 节点语义工具族）、Pages（3 工具）八套工具共 40 个；认证为 OAuth 2.1 优先、API Key 回退。
- MCP OAuth 2.1 已实现（RFC 8414 / 9728 / 7591 / 8707，PKCE S256，consent 页 `/oauth`）。
- AI 插件动态分发已实现（marketplace.json、ZCode 专用清单、lumina.zip、`.well-known/skills`）。
- WebSocket Hub 已实现，支持 sessionID → deviceID 二级索引，连接 `Kind` 区分 qa/preview，心跳检测，优雅关闭，断线重连和会话恢复。
- Cron Runner 已实现，注册 RepoWiki 超时任务重试与 Preview 过期会话清理（默认每 5 分钟），由 `main.go` 传入 `xMain.Runner` 异步执行。
- WebSocket Runner 已实现（`NewWebSocketRunner`），Hub 主循环由 `main.go` 传入 `xMain.Runner` 异步执行。
- Wiki Reader 前端（`web-wiki/`）已实现，独立 SPA 部署在 `/wiki/`，支持密码门认证和只读 .mdx Wiki 渲染。
- Docker 容器化已实现（多阶段 Dockerfile + 双 docker-compose 编排 + GitHub Actions：PR 校验、版本 tag / 手动发版在校验通过后发布镜像；发布模式同时生成 AI 正文并上传 GoReleaser 跨平台二进制）；容器内部端口默认 8800（`EXPOSE 8800`），由 `XLF_PORT` 环境变量驱动。

## 调试路径

1. 后端路由 / 分层 / MCP / OAuth → 见 [internal/](./internal/AGENTS.md) 调试路径。
2. 启动节点、迁移、种子数据 → 见 [internal/app/startup/](./internal/app/startup/AGENTS.md)。
3. 控制台页面、接入指南、Interact、Preview 前端 → 见 [web/](./web/AGENTS.md)。
4. Wiki Reader 渲染与密码门 → 见 [web-wiki/](./web-wiki/AGENTS.md)。
5. 共享组件、主题色、Markdown 排版（代码块黑底） → 见 [components/](./components/AGENTS.md)。

## 引用

- [ARCHITECTURE.md](./ARCHITECTURE.md) — 物理架构地图与长期边界（人类贡献者）
- [CONTRIBUTING.md](./CONTRIBUTING.md) — 本地开发与 PR
- [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md) — 行为准则
- [SECURITY.md](./SECURITY.md) — 漏洞私密上报
- [SUPPORT.md](./SUPPORT.md) — 求助渠道
- [CHANGELOG.md](./CHANGELOG.md) — 版本变更
- [internal/](./internal/AGENTS.md) — 后端业务层详细文档
- [web/](./web/AGENTS.md) — 控制台前端应用详细文档
- [web-wiki/](./web-wiki/AGENTS.md) — Wiki Reader 前端详细文档
- [components/](./components/AGENTS.md) — 共享 UI / Markdown / motion / 主题包
