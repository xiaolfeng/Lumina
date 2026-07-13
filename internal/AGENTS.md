# INTERNAL 业务层知识库

## 概述
`internal/` 实现了 Lumina 的业务运行时管道：route -> middleware -> handler -> logic -> repository -> entity，严格分层，禁止跨层调用。同时包含 MCP Server 工具注册、WebSocket 实时通信层和共享服务层。

## 目录结构
```text
internal/
├── app/
│   ├── middleware/           # Gin 中间件
│   │   ├── auth.go           # Bearer Token 验证 → 注入用户到 context
│   │   └── apikey.go         # API Key 认证（`lumi_` 前缀 + bcrypt 校验）
│   ├── route/                # 路由注册与中间件绑定
│   │   ├── route.go          # 全局中间件 + 路由组入口 + 前端 SPA 集成
│   │   ├── route_auth.go     # 认证路由（公开 + 受保护）
│   │   ├── route_apikey.go   # API Key 路由（CRUD + 重置，受 Auth 保护）
│   │   ├── route_project.go  # 项目路由（CRUD，受 Auth 保护）
│   │   ├── route_pin.go      # Pin 路由（CRUD，受 Auth 保护）
│   │   ├── route_user.go     # 用户路由（个人资料 + 密码修改）
│   │   ├── route_qa.go       # Q&A REST API 路由（会话/问题/配置管理）
│   │   ├── route_qa_download.go # Q&A 文件下载（Token 校验 + 文件流）
│   │   ├── route_ws.go       # Q&A WebSocket 端点（Hub 初始化 + 认证）
│   │   ├── route_mcp.go      # MCP Streamable HTTP 端点（API Key 认证）
│   │   ├── route_health.go   # 健康检查路由
│   │   ├── route_frontend.go # 前端 SPA 静态资源 + fallback
│   │   ├── route_swagger.go  # Swagger UI（仅 XLF_DEBUG=true）
│   │   └── route_repowiki.go # RepoWiki 路由（配置/版本/Webhook）
│   └── startup/              # 基础设施初始化与种子数据（详见子模块文档）
├── handler/                  # HTTP 处理器（薄控制器层）
│   ├── handler.go            # NewHandler[T] 泛型构造器 + service 注入
│   ├── auth.go               # 认证处理器（登录、刷新、初始化、登出、状态）
│   ├── user.go               # 用户处理器（个人资料、密码修改）
│   ├── biometric.go          # WebAuthn 生物认证处理器（注册/登录/凭证管理）
│   ├── apikey.go             # API Key 处理器（CRUD + 重置 + 分页）
│   ├── project.go            # 项目处理器（CRUD + 分页）
│   ├── pin.go                # Pin 处理器（CRUD + 分页）
│   ├── qa.go                 # Q&A 处理器（会话 CRUD、问题详情、配置管理）
│   ├── qa_download.go        # Q&A 文件下载处理器（Token 校验 + 文件流输出）
│   ├── repowiki.go           # RepoWiki 配置/版本管理处理器
│   ├── wiki_reader.go        # Wiki 内容读取处理器（公开/密码保护）
│   └── health.go             # 健康检查处理器
├── logic/                    # 业务编排层
│   ├── logic.go              # logic 基础结构（db/rdb/log）
│   ├── auth.go               # 认证逻辑（Token 验证、密码校验、资料更新）
│   ├── webauthn_user.go      # WebAuthn 用户适配器（实现 webauthn.User 接口）
│   ├── biometric.go          # WebAuthn 生物认证逻辑（注册/登录/凭证 CRUD）
│   ├── apikey.go             # API Key 逻辑（密钥生成/哈希/脱敏/CRUD/校验）
│   ├── project.go            # 项目逻辑（CRUD、名称唯一校验、别名解析）
│   ├── pin.go                # Pin 逻辑（Push/Consume/Peek/List/项目解析）
│   ├── qa_logic.go           # Q&A 核心业务逻辑（Session/Question/Supplement 编排）
│   ├── qa_format.go          # Q&A 题型格式化（15+ 题型的 Markdown 格式化）
│   ├── qa_helper.go          # Q&A 辅助函数（选项处理、类型判断等）
│   ├── qa_mcp.go             # Q&A MCP 工具实现（10+ 工具的 handler 逻辑）
│   ├── qa_mcp_helpers.go     # Q&A MCP 辅助函数（工具参数解析、结果组装）
│   ├── qa_download.go        # Q&A 文件下载逻辑（Token 校验 + 文件流）
│   ├── repowiki_logic.go     # RepoWiki 核心逻辑（分析入口、配置/版本 CRUD）
│   ├── repowiki_pipeline.go  # RepoWiki 分析管道（Git 准备 + 状态机驱动）
│   ├── repowiki_orchestrator.go # 5 角色 SubAgent 编排引擎
│   ├── repowiki_subagent_prompts.go # 5 角色 system/user prompt
│   ├── repowiki_types.go     # RepoWiki 内部类型（WikiEntry/ValidationError/ExploreOutput）
│   ├── repowiki_cron.go      # RepoWiki 定时清理任务
│   ├── repowiki_webhook.go   # RepoWiki Webhook 处理逻辑
│   └── health.go             # 健康检查逻辑
├── repository/               # 数据访问层
│   ├── info.go               # Info 键值配置持久化（GetByKey/UpdateValue/UpdateValuesInTx）
│   ├── token.go              # Token 持久化
│   ├── apikey.go             # API Key 持久化（CRUD + 分页 + 校验）
│   ├── project.go            # 项目持久化（CRUD + 分页 + 别名查询；缓存委托 cache.ProjectCache）
│   ├── pin.go                # Pin 持久化（CRUD + FIFO 消费 + 分页）
│   ├── biometric_credential.go # WebAuthn 凭证持久化（CRUD + 按 CredentialID 查询）
│   ├── qa_session.go         # Q&A Session 持久化（CRUD + 分页 + 状态/类型过滤）
│   ├── qa_question.go        # Q&A Question 持久化（CRUD + 批量创建）
│   ├── qa_supplement.go      # Q&A Supplement 持久化（创建 + 按 Session 查询）
│   ├── repowiki_config.go    # RepoWikiConfig 持久化
│   ├── wiki_version.go       # WikiVersion 持久化
│   ├── health.go             # 数据库就绪检查
│   └── cache/                # Redis 缓存操作（Cache-Aside 策略子层）
│       ├── token.go          # Token 缓存（AT/RT 存储，实现 KeyCache 接口）
│       ├── project.go        # 项目多维度缓存（ID/Name/Alias/MatchPath）
│       ├── biometric_credential.go # WebAuthn 凭证缓存 + Challenge 会话存储
│       ├── qa_session.go     # QA 会话缓存（ID→详情 + Hash→ID）
│       └── qa_retry.go       # qa_get_answer 重试计数器（INCR/Reset）
├── service/                  # 共享服务层（跨模块复用的基础设施）
│   ├── download_token.go     # 文件下载 Token 生成与校验（短时效签名）
│   ├── file_cache.go         # 文件缓存管理（上传文件本地暂存 + 清理）
│   ├── media_answer.go       # 媒体回答处理（图片/文件附件的回答格式化）
│   ├── wiki_storage.go       # RepoWiki 文件系统存储与路径管理
│   └── wiki_auth_token.go    # Wiki 访问密码 Token 生成与校验
├── entity/                   # GORM 实体
│   ├── info.go               # 站点配置实体（单用户模式）
│   ├── apikey.go             # API Key 实体（密钥哈希/前缀/后缀/过期时间）
│   ├── project.go            # 项目实体（名称/别名/描述）
│   ├── pin.go                # Pin 实体（标题/内容/分类/优先级/来源/目标项目）
│   ├── biometric_credential.go # WebAuthn 凭证实体（CredentialID/公钥/签名计数）
│   ├── qa_session.go         # Q&A Session 实体（状态/类型/TTL）
│   ├── qa_question.go        # Q&A Question 实体（题型/标题/选项/回答）
│   ├── qa_supplement.go      # Q&A Supplement 实体（补充内容/附件）
│   ├── repowiki_config.go    # RepoWiki 配置实体（仓库地址/LLM 参数/当前选中版本）
│   └── wiki_version.go       # Wiki 版本实体（版本号/状态/文件路径）
├── mcp/                      # MCP Server 工具注册
│   ├── server.go             # MCP Server 初始化 + StreamableHTTPHandler 创建 + Logic 注入入口
│   ├── qa_tools.go           # Q&A MCP 工具注册（10+ 工具定义 + schema）
│   ├── qa_handlers.go        # Q&A MCP 工具 handler 实现（工具执行逻辑）
│   ├── qa_type_details.go    # Q&A MCP 题型详情定义（15+ 题型 schema 细节）
│   ├── project_tools.go      # Project MCP 工具（CRUD + 别名解析 + match_path 数组）
│   ├── pin_tools.go          # Pin MCP 工具（Push/Consume/List/Update/Peek）
│   └── repowiki_tools.go     # RepoWiki MCP 工具（只读：query/list）
├── websocket/                # WebSocket 实时通信层
│   ├── hub.go                # 连接管理器（sessionID → deviceID 二级索引 + 心跳检测）
│   ├── handler.go            # WebSocket 升级处理器 + 业务消息分发
│   ├── connection.go         # 单个连接封装（读写 goroutine + 优雅关闭）
│   └── message.go            # 消息类型定义（15 种消息类型）
├── qa/                       # Q&A 回答队列
│   └── queue.go              # 会话级 FIFO 回答队列（Enqueue/Consume/WaitAndConsume）
└── constant/                 # 共享业务常量
    ├── cache.go              # Redis Key 前缀/过期时间（带环境前缀格式化）
    ├── context.go            # Context Key（如 CtxOwnerKey）
    ├── gene_number.go        # 雪花算法基因编号（GeneProject=32 ~ GeneWikiVersion=40）
    ├── biometric.go          # WebAuthn 相关常量（RP ID/Origin/超时）
    └── pin.go                # Pin 模块常量（分类/优先级枚举）
```

## 导航指南
| 任务 | 文件/目录 | 说明 |
|---|---|---|
| 新增路由组 | `app/route/route.go` + `route_*.go` | 在 `NewRoute` 中调用路由注册函数 |
| 新增中间件 | `app/middleware/` | 返回 `gin.HandlerFunc`，在路由注册中绑定 |
| 新增处理器 | `handler/handler.go` 定义类型，`handler/*.go` 实现 | 通过 `NewHandler[T]` 构造 |
| 新增业务逻辑 | `logic/*.go` | Logic 通过 context 注入获取 db/rdb；QA 逻辑按职责拆分到 `qa_*.go` |
| 新增数据访问 | `repository/*.go` | 返回 `(data, *xError.Error)` |
| 新增 Redis 缓存 | `repository/cache/*.go` | 实现 KeyCache 接口或独立缓存操作 |
| 新增共享服务 | `service/*.go` | 跨模块复用的基础设施（如下载 Token、媒体回答处理） |
| 新增实体 | `entity/*.go` + `startup/startup_database.go` | 实现并追加到 `migrateTables` |
| 新增基因编号 | `constant/gene_number.go` | 定义 `GeneXxx` 常量供实体 `GetGene()` 使用 |
| 新增种子数据 | `startup/prepare/` | 创建 `prepare_<domain>.go` |
| 新增请求/响应 DTO | `api/<domain>/` | 按业务域保持子包结构 |
| 新增业务常量 | `constant/*.go` | 基因编号、缓存 Key、Context Key、模块枚举 |
| 新增 MCP 工具 | `mcp/*.go` | 在 `server.go` 注册，在 `startup_mcp.go` 注入 Logic |
| 新增 WebSocket 消息类型 | `websocket/message.go` | 定义 MessageType 常量和 Message 结构 |
| 新增 Q&A 回答队列 | `qa/queue.go` | 会话级 FIFO 队列，由 `QaLogic` 调用 |
| 新增 RepoWiki 分析入口 | `logic/repowiki_logic.go` | 配置/版本 CRUD 与分析启动 |
| 新增 RepoWiki 子 Agent 编排 | `logic/repowiki_orchestrator.go` | 5 阶段预定义编排，不持有 db/rdb |
| 新增 RepoWiki 分析管道 | `logic/repowiki_pipeline.go` | Git 准备 + 状态机驱动 |
| 新增 RepoWiki 文件存储 | `service/wiki_storage.go` | 版本隔离路径管理与文件 I/O |
| 新增 RepoWiki MCP 工具 | `mcp/repowiki_tools.go` | 只读工具：repoWiki_query / repoWiki_list |

## 代码地图

| 符号 | 类型 | 位置 | 作用 |
|---|---|---|---|
| `NewHandler[T]` | 泛型函数 | `handler/handler.go` | Handler 泛型构造模式，注入全部 Logic |
| `SubAgentOrchestrator` | 结构体 | `logic/repowiki_orchestrator.go` | 5 角色 SubAgent 编排引擎（overview → explore → architect → writer → validator） |
| `AnalysisPipeline` | 结构体 | `logic/repowiki_pipeline.go` | RepoWiki 分析管道（Git 准备 + 状态机驱动） |
| `RepoWikiLogic` | 结构体 | `logic/repowiki_logic.go` | RepoWiki 业务编排（配置/版本/分析入口） |
| `WikiStorageService` | 结构体 | `service/wiki_storage.go` | RepoWiki 文件系统存储与路径管理 |
| `RepoWikiConfig` | 结构体 | `entity/repowiki_config.go` | RepoWiki 配置实体（仓库地址/当前选中版本） |
| `WikiVersion` | 结构体 | `entity/wiki_version.go` | Wiki 版本实体（版本号/状态/文件路径） |
| `RepoWikiConfigRepo` | 结构体 | `repository/repowiki_config.go` | RepoWikiConfig 持久化 |
| `WikiVersionRepo` | 结构体 | `repository/wiki_version.go` | WikiVersion 持久化 |
| `CoordinatorSystemPrompt` | 常量 | `logic/repowiki_subagent_prompts.go` | Coordinator 角色 system prompt |
| `BuildOverviewUserPrompt` | 函数 | `logic/repowiki_subagent_prompts.go` | 构建 Coordinator 概要阶段 user prompt |
| `WikiEntry` | 结构体 | `logic/repowiki_types.go` | Architect 输出的 Wiki 目录条目 |
| `ValidationError` | 结构体 | `logic/repowiki_types.go` | Validator 输出的校验错误项 |
| `ExploreOutput` | 结构体 | `logic/repowiki_types.go` | Explore Agent 的单个产出项 |
| `ModelRunConfig` | 结构体 | `logic/repowiki_types.go` | Agent 运行时的模型配置 |
| `RegisterRepoWikiTools` | 函数 | `mcp/repowiki_tools.go` | 注册 RepoWiki MCP 只读工具 |
| `SetRepoWikiLogic` | 函数 | `mcp/repowiki_tools.go` | 注入 RepoWikiLogic 到 MCP 工具 |

## 约定
- **严格分层**：route → middleware → handler → logic → repository；禁止跳层调用。`service/` 层为跨模块共享基础设施，可被 logic 层调用。
- **Handler 精简**：仅做请求绑定、调用 logic、映射结果到响应；禁止在 handler 中直接操作 DB/Redis。
- **Logic 编排**：业务编排层，持久化和 SQL 归 repository 层；共享服务（如下载 Token）通过 `service/` 调用。
- **Repository 返回值**：统一 `(data, *xError.Error)` 风格，不用裸 `error`。
- **日志命名**：按层使用 `xLog.WithName` — `NamedCONT`（handler）、`NamedLOGC`（logic）、`NamedREPO`（repository）、`NamedINIT`（startup）、`NamedMIDE`（middleware）。
- **上下文传递**：使用 `ctx.Request.Context()` 或注入的 context 下发调用。
- **认证中间件**：通过 `middleware.Auth(ctx)` 创建，注入认证标记到 context（`CtxOwnerKey`）。
- **API Key 中间件**：`middleware.ApikeyAuth` 验证 `lumi_` 前缀的 API Key，用于 MCP 端点认证。
- **泛型 Handler 构造**：`NewHandler[T]` 统一注入所有 logic 实例到 `service` 结构体（含 health/auth/apikey/project/qa/biometric/pin 共 7 个 Logic）。
- **实体 ID 策略**：雪花算法基因策略；每个实体必须实现 `GetGene() xSnowflake.Gene`。
- **字段注释**：实体字段必须追加行尾中文注释（`// 字段说明`），且与 `gorm comment` 一致。
- **缓存键前缀**：通过 `xEnv.NoSqlPrefix` 环境变量自动拼接前缀，使用 `RedisKey.Get(args...)` 格式化。
- **分页规范**：使用 `xModels.PageRequest.Normalize()` 规范化分页参数，`xModels.NewPageFromRequest` 构建分页响应。
- **API Key 安全**：密钥使用 bcrypt 哈希存储，仅创建/重置时返回完整密钥；查询和列表使用 `maskKey` 脱敏。
- **Project 缓存策略**：采用 Cache-Aside 模式（ID→详情、Name→ID、Alias→ID 三层映射，TTL 30 分钟）。
- **MCP 路由**：必须在 `engine.Use()` 之前注册（绕开 `ResponseMiddleware`），使用 `gin.WrapH` 包装 `http.Handler`。
- **MCP Logic 注入**：`startup_mcp.go` 中通过 `mcp.SetQaLogic/SetProjectLogic/SetPinLogic` 注入 Logic 实例。
- **WebSocket 管理**：`Hub` 按 sessionID → deviceID 二级索引管理连接；心跳检测间隔 5s，超时 15s。
- **Q&A 推送回调**：`logic.OnQuestionPushed` / `logic.OnSupplementPushed` / `logic.OnQuestionCancelled` / `logic.OnSessionArchived` 函数变量在 `route_ws.go` 中设置，解耦 Logic 层和 WebSocket 层。
- **回答队列**：每个 Session 独立的 FIFO 队列，支持 `WaitAndConsume` 阻塞等待新回答。
- **Pin FIFO 消费**：Pin 模块的 FIFO 消费基于数据库实现（`ConsumeOldestPending` + `ConsumeByID`），不依赖 Redis 队列。
- **WebAuthn 凭证**：CredentialID 全局唯一，使用 bcrypt 存储公钥；Challenge 通过 Redis 缓存短暂会话（`cache/biometric_credential.go`）。
- **文件下载 Token**：`service/download_token.go` 生成短时效签名 Token，用于 Q&A 文件附件下载鉴权。
- **QA 逻辑拆分**：`qa.go` 已按职责拆分为 `qa_logic.go`（核心编排）、`qa_format.go`（题型格式化）、`qa_helper.go`（辅助函数）、`qa_mcp.go`（MCP 工具）、`qa_mcp_helpers.go`（MCP 辅助）、`qa_download.go`（文件下载）；新增 QA 逻辑时按职责归入对应文件。
- **MCP 工具拆分**：`mcp/qa_tools.go`（工具注册）、`qa_handlers.go`（handler 实现）、`qa_type_details.go`（题型 schema）三文件分工；新增 MCP 工具时在 `qa_tools.go` 注册、`qa_handlers.go` 实现。
- **RepoWiki 子 Agent 编排**：`SubAgentOrchestrator` 按预定义 5 阶段（Coordinator → Explore → Architect → Writer → Validator）生成 Wiki，`repowiki_subagent_prompts.go` 定义 5 角色 system/user prompt，`repowiki_types.go` 定义内部类型，`repowiki_pipeline.go` 负责 Git 准备与状态机驱动。
- **RepoWiki 版本隔离**：每个 Wiki 版本存储在 `versions/{vid}/` 下，`RepoWikiConfig.SelectedVersionID` 指定当前对外服务版本；旧版 config 级目录已废弃，新版本完成后清理。
- **RepoWiki MCP 只读**：MCP 端仅暴露 `repoWiki_query` / `repoWiki_list` 两个只读工具，Wiki 更新由 Git Webhook 自动触发。

## 反模式
- 禁止从路由直接调用 repository 或绕过 logic 层。
- 禁止在 handler 中手写原始 Gin JSON 响应；使用 `xResult` 辅助函数。
- 禁止在 logic/repository 构造函数内部创建 DB/Redis 客户端；应从 context 获取注入的依赖。
- 禁止绕过 `NewHandler[T]` 模式手动构造 handler。
- 禁止将业务常量写在 handler/logic 文件中；统一放 `constant/`。
- 禁止核心业务模块（RepoWiki、Memory、Q&A、Pin）之间直接调用。
- 禁止直接使用 `os.Getenv`；应使用带默认值的 `xEnv.GetEnv*`。
- 禁止新增实体后不追加到 `migrateTables`。
- 禁止在 repository 外部直接操作 Redis；缓存逻辑封装在 repository/cache 子层。
- 禁止 logic 结构体持有 db/rdb 字段；所有数据访问必须经由 repository（+cache 子层）。
- 禁止在 logic 层拼接 GORM/Redis 命令（含 entity.Info 配置读取）；统一走 InfoRepo。
- 禁止在 repository 内定义私有 cacheKey() 方法；缓存键统一用 `constant.RedisKey.Get()`。
- 禁止在 Q&A 模块使用 SSE 进行问题推送；统一使用 WebSocket。
- 禁止 WebAuthn Challenge 存储在内存中；必须通过 Redis 缓存以支持多实例部署。
## 调试路径

1. 请求未路由 → 检查 `app/route/route.go` 路由组和 `route_*.go` 注册。
2. 路由正确但响应异常 → 检查 `handler/*.go` 绑定，然后 `logic/*.go` 编排。
3. 认证失败 → 检查 `middleware/auth.go` → `logic/auth.go` → `repository/cache/token.go`。
4. API Key 认证失败 → 检查 `middleware/apikey.go` → `logic/apikey.go` 的 `ValidateAPIKey`。
5. WebAuthn 注册/登录失败 → 检查 `handler/biometric.go` → `logic/biometric.go` → `repository/biometric_credential.go` + `cache/biometric_credential.go`（Challenge 是否过期）。
6. MCP 工具调用失败 → 检查 `mcp/server.go` 注册 + `startup_mcp.go` Logic 注入（SetXxxLogic）。
7. Pin 消费顺序异常 → 检查 `repository/pin.go` 的 `ConsumeOldestPending` 排序逻辑（FIFO 按 createdAt 升序）。
8. 数据库操作失败 → 检查 `repository/*.go` 和启动阶段的迁移状态。
9. Redis 缓存异常 → 检查 `repository/cache/*.go` 和 `startup_redis.go` 连接配置。
10. WebSocket 连接问题 → 检查 `websocket/hub.go` 连接管理 + `route_ws.go` Hub 初始化。
11. Q&A 问题推送不达 → 检查 `logic.OnQuestionPushed` 回调是否在 `route_ws.go` 中正确设置。
12. 回答队列阻塞 → 检查 `qa/queue.go` 的 `WaitAndConsume` 和消费者 goroutine。
13. 文件下载失败 → 检查 `service/download_token.go` Token 校验 + `service/file_cache.go` 文件是否存在。
14. RepoWiki 分析失败 → 检查 `logic/repowiki_orchestrator.go` 5 阶段执行日志 + `logic/repowiki_pipeline.go` 状态机驱动；Git 问题先看 `service/repo_git.go` 克隆/拉取日志。

## 引用
- [startup/](./app/startup/AGENTS.md) — 启动模块详细文档
