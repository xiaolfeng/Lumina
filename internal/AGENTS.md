<!-- deep-init:synced@fa47a98 -->

# INTERNAL 业务层知识库

## 概述
`internal/` 实现了 Lumina 的业务运行时管道：route -> middleware -> handler -> logic -> repository -> entity，严格分层，禁止跨层调用。同时包含 MCP Server 工具注册、MCP OAuth 2.1、AI 插件动态分发、WebSocket 实时通信层、RepoWiki 编排引擎、Preview 预览与 LPW 渲染引擎、Pages 持久快照模块和跨模块共享服务层。

## 目录结构
```text
internal/
├── app/
│   ├── middleware/           # Gin 中间件
│   │   ├── auth.go           # Bearer Token 验证 → 注入用户到 context
│   │   ├── apikey.go         # 纯 API Key 认证（`lumi_` 前缀 + bcrypt；MCP 端点已改走 mcp_auth）
│   │   ├── mcp_auth.go       # MCP 端点认证（OAuth 2.1 `lum_at_` 优先，API Key 回退 + WWW-Authenticate）
│   │   ├── wiki_auth.go      # Wiki Reader 访问认证（密码 Token / Cookie 会话）
│   │   ├── pages_auth.go     # Pages 密码门中间件（HMAC Cookie 校验；顶层 document 放行给 SPA，拦截子资源）
│   │   ├── sandbox_subresource.go # Preview/Pages 沙盒隔离子资源安全处理
│   │   ├── mcp_compat.go     # MCP 端点兼容性中间件（Streamable HTTP 请求处理）
│   │   ├── cors.go           # 白名单 CORS（`XLF_ALLOWED_ORIGINS`，替代全局 ACAO:*）
│   │   ├── security.go       # 安全响应头（nosniff / X-Frame-Options / CSP frame-ancestors）
│   │   └── webauthn.go       # WebAuthn Origin 解析（按请求注入 context，供 RPID 动态推导）
│   ├── route/                # 路由注册与中间件绑定
│   │   ├── route.go          # 全局中间件 + 路由组入口 + 双前端 SPA 集成
│   │   ├── route_auth.go     # 认证路由（公开 + 受保护）
│   │   ├── route_apikey.go   # API Key 路由（CRUD + 重置，受 Auth 保护）
│   │   ├── route_workspace.go # 工作空间路由（CRUD + 默认空间，受 Auth 保护）
│   │   ├── route_project.go  # 项目路由（CRUD，受 Auth 保护）
│   │   ├── route_pin.go      # Pin 路由（CRUD，受 Auth 保护）
│   │   ├── route_user.go     # 用户路由（个人资料 + 密码修改）
│   │   ├── route_qa.go       # Q&A REST API 路由（会话/问题/配置管理）
│   │   ├── route_qa_download.go # Q&A 文件下载（Token 校验 + 文件流）
│   │   ├── route_ws.go       # WebSocket 端点（Hub 初始化 + 认证，Q&A + Preview 共用）
│   │   ├── route_mcp.go      # MCP Streamable HTTP 端点（McpAuth 认证）
│   │   ├── route_health.go   # 健康检查路由
│   │   ├── route_frontend.go # 双前端 SPA 静态资源 + fallback（web/ + web-wiki/）
│   │   ├── route_swagger.go  # Swagger UI（仅 XLF_DEBUG=true）
│   │   ├── route_repowiki.go # RepoWiki 路由（配置/版本管理，受 Auth 保护）
│   │   ├── route_llm.go      # LLM 路由（Provider/Model CRUD + Agent 模型分配，受 Auth 保护）
│   │   ├── route_ssh.go      # SSH Key 路由（CRUD，受 Auth 保护）
│   │   ├── route_webhook.go  # Webhook 路由（RepoWiki Git Webhook 接收，HMAC 签名校验）
│   │   ├── route_settings.go # 系统设置路由（站点/安全/Q&A/RepoWiki/Preview 配置读写）
│   │   ├── route_preview.go  # Preview 路由（登录 Cookie/Bearer + 管理端 + 引擎级路径直出）
│   │   ├── route_pages.go    # Pages 路由（公开解锁/元信息 + 管理端 + 引擎级路径直出）
│   │   ├── route_dashboard.go # Dashboard 路由（GET /dashboard/overview，受 Auth 保护）
│   │   ├── route_oauth.go    # MCP OAuth 2.1（well-known / authorize / register / token + 登录态 consent）
│   │   └── route_plugin.go   # AI 插件公开分发（marketplace.json / lumina.zip / .well-known/skills）
│   └── startup/              # 业务节点初始化与种子数据（详见子模块文档）
├── handler/                  # HTTP 处理器（薄控制器层）
│   ├── handler.go            # NewHandler[T] 泛型构造器 + service 注入（注入 18 个 Logic）
│   ├── bind.go               # 通用请求绑定辅助（ShouldBindJSON/PageRequest 规范化）
│   ├── auth.go               # 认证处理器（登录、刷新、初始化、登出、状态）
│   ├── user.go               # 用户处理器（个人资料、密码修改）
│   ├── biometric.go          # WebAuthn 生物认证处理器（注册/登录/凭证管理）
│   ├── apikey.go             # API Key 处理器（CRUD + 重置 + 分页）
│   ├── workspace.go          # 工作空间处理器（CRUD + 默认空间标记）
│   ├── project.go            # 项目处理器（CRUD + 分页 + 所属工作空间）
│   ├── pin.go                # Pin 处理器（CRUD + 分页）
│   ├── qa.go                 # Q&A 处理器（会话 CRUD、问题详情、配置管理）
│   ├── qa_download.go        # Q&A 文件下载处理器（Token 校验 + 文件流输出）
│   ├── repowiki.go           # RepoWiki 处理器（配置/版本/分析触发/Webhook 配置）
│   ├── wiki_reader.go        # Wiki 内容读取处理器（公开/密码保护/manifest/.mdx 内容流 + computeNav 前后导航）
│   ├── llm.go                # LLM 处理器（Provider/Model CRUD + Agent 角色模型分配）
│   ├── ssh_key.go            # SSH Key 处理器（CRUD + 密钥生成/公钥导出）
│   ├── webhook.go            # Webhook 处理器（Git Push 事件接收 + HMAC 校验）
│   ├── settings.go           # 系统设置处理器（分组配置读写 + 环境信息）
│   ├── preview.go            # Preview 处理器（会话/文件 CRUD + 晋升 + 路径直出）
│   ├── pages.go              # Pages 处理器（列表/版本/Fork/密码门/直出）
│   ├── serve.go              # Preview / Pages 静态资源流式直出与 MIME 分发处理器
│   ├── dashboard.go          # Dashboard 处理器（概览统计）
│   ├── ai_plugin.go          # AI 插件分发（市场清单 / ZIP / well-known 技能）
│   ├── oauth.go              # MCP OAuth 2.1（元数据 / DCR / 授权码 / 令牌 / 同意页）
│   └── health.go             # 健康检查处理器
├── logic/                    # 业务编排层
│   ├── logic.go              # logic 基础结构（db/rdb/log）+ context 获取 Logic 辅助
│   ├── auth.go               # 认证逻辑（Token 验证、密码校验、资料更新）
│   ├── webauthn_user.go      # WebAuthn 用户适配器（实现 webauthn.User 接口）
│   ├── biometric.go          # WebAuthn 生物认证逻辑（注册/登录/凭证 CRUD + RPID 动态推导）
│   ├── apikey.go             # API Key 逻辑（密钥生成/哈希/脱敏/CRUD/校验）
│   ├── workspace.go          # 工作空间逻辑（CRUD、默认空间防删/唯一、项目计数聚合）
│   ├── project.go            # 项目逻辑（CRUD、空间校验、名称唯一校验、别名解析）
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
│   ├── repowiki_subagent_prompts.go # 5 角色 system/user prompt 构建
│   ├── repowiki_types.go     # RepoWiki 内部类型（WikiEntry/ValidationError/ExploreOutput/ModelRunConfig）
│   ├── repowiki_cron.go      # RepoWiki 定时重试任务逻辑（超时任务扫描 + 失败重试）
│   ├── repowiki_webhook.go   # RepoWiki Webhook 处理逻辑（分支解析 + 触发分析）
│   ├── llm_provider.go       # LLM Provider 逻辑（CRUD + API Key 加密存储）
│   ├── llm_model.go          # LLM Model 逻辑（CRUD + Agent 角色模型分配）
│   ├── ssh_key.go            # SSH Key 逻辑（CRUD + 密钥对生成/公钥导出）
│   ├── settings.go           # 系统设置逻辑（分组配置读写 + Info 表编排）
│   ├── preview_logic.go      # Preview 逻辑（会话/文件管理 + 会话过期 + WebSocket 同步回调）
│   ├── preview_lines.go      # Preview 行级编辑纯函数（拆行/合并/行变换/行号格式化）
│   ├── preview_entry.go      # Preview 根入口文件解析与分流
│   ├── preview_lpw_logic.go  # Preview LPW 1.1 节点文档编排（kind 层级校验 + 节点写入）
│   ├── preview_lpw_contract.go # Container variant / Layout pattern / Annotation 契约表（与前端 contract.ts 快照对齐）
│   ├── preview_lpw_tree.go   # LPW 1.1 节点树纯函数
│   ├── pages_logic.go        # Pages 逻辑（晋升快照、Fork、版本指针、密码门）
│   ├── dashboard.go          # Dashboard 逻辑（六类指标聚合）
│   ├── runtime_url.go        # 运行时域名解析 + Preview/Pages 路径式深链构建
│   ├── oauth_logic.go        # MCP OAuth 2.1 编排（DCR / 授权码+PKCE / 令牌 / RFC 8707 资源绑定）
│   ├── ai_plugin.go          # AI 插件编排（域名收敛后交给 service 打包清单与 ZIP）
│   └── health.go             # 健康检查逻辑
├── repository/               # 数据访问层
│   ├── info.go               # Info 键值配置持久化（GetByKey/UpdateValue/UpdateValuesInTx）
│   ├── token.go              # Token 持久化
│   ├── apikey.go             # API Key 持久化（CRUD + 分页 + 校验）
│   ├── workspace.go          # 工作空间持久化（CRUD + 默认标记部分唯一索引 + 项目计数）
│   ├── project.go            # 项目持久化（CRUD + 分页 + 空间过滤；缓存委托 cache.ProjectCache）
│   ├── pin.go                # Pin 持久化（CRUD + FIFO 消费 + 分页）
│   ├── biometric_credential.go # WebAuthn 凭证持久化（CRUD + 按 CredentialID 查询）
│   ├── qa_session.go         # Q&A Session 持久化（CRUD + 分页 + 状态/类型过滤）
│   ├── qa_question.go        # Q&A Question 持久化（CRUD + 批量创建）
│   ├── qa_supplement.go      # Q&A Supplement 持久化（创建 + 按 Session 查询）
│   ├── repowiki_config.go    # RepoWikiConfig 持久化
│   ├── wiki_version.go       # WikiVersion 持久化
│   ├── llm_provider.go       # LlmProvider 持久化（CRUD + 分页）
│   ├── llm_model.go          # LlmModel 持久化（CRUD + 分页 + 按 Agent 角色查询）
│   ├── ssh_key.go            # SshKey 持久化（CRUD + 分页 + 指纹查询）
│   ├── webhook_event.go      # WebhookEvent 持久化（CRUD + 状态/分支过滤）
│   ├── preview_session.go    # PreviewSession 持久化（CRUD + 按 Hash 查询 + Fork 溯源 + 过期时间）
│   ├── preview_file.go       # PreviewFile 持久化（CRUD + 按 Session 查询）
│   ├── page.go               # Page 持久化（项目内 slug、生效指针、访问策略、晋升事务）
│   ├── page_version.go       # PageVersion 持久化（不可变快照版本）
│   ├── page_file.go          # PageFile 持久化（版本内扁平文件）
│   ├── dashboard.go          # Dashboard 统计持久化（原生 SQL 聚合六类指标）
│   ├── oauth_client.go       # OAuth 动态客户端持久化（RFC 7591，无 client_secret）
│   ├── health.go             # 数据库就绪检查
│   └── cache/                # Redis 缓存操作（Cache-Aside 策略子层）
│       ├── base.go           # 缓存基础依赖（RDB + TTL，替代旧 xCache.Cache）
│       ├── token.go          # Token 缓存（AT/RT 存储，实现 KeyCache 接口）
│       ├── workspace.go      # 工作空间缓存
│       ├── project.go        # 项目多维度缓存（ID/Name/Alias/MatchPath + 空间维度）
│       ├── biometric_credential.go # WebAuthn 凭证缓存 + Challenge 会话存储
│       ├── qa_session.go     # QA 会话缓存（ID→详情 + Hash→ID）
│       ├── qa_retry.go       # qa_get_answer 重试计数器（INCR/Reset）
│       ├── repowiki.go       # RepoWiki 配置缓存（Config + Wiki 版本列表）
│       ├── ssh_key.go        # SSH Key 缓存（指纹→详情 + 私钥临时缓存）
│       └── oauth_store.go    # OAuth 运行时缓存（授权请求 / 授权码 / 令牌元数据，短 TTL 不落库）
├── service/                  # 共享服务层（跨模块复用的基础设施）
│   ├── download_token.go     # 文件下载 Token 生成与校验（短时效签名）
│   ├── file_cache.go         # 文件缓存管理（上传文件本地暂存 + 清理 + 路径穿越防护）
│   ├── media_answer.go       # 媒体回答处理（图片/文件附件的回答格式化）
│   ├── wiki_storage.go       # RepoWiki 文件系统存储与路径管理（ReadPage 读取 .mdx + frontmatter / ReadMetaJSON 读取 per-folder meta.json）
│   ├── wiki_auth_token.go    # Wiki 访问密码 Token 生成与校验
│   ├── page_auth_token.go    # Pages 密码门 HMAC Cookie 签发与校验
│   ├── lpw_schema.go         # LPW JSON Schema 声明与校验服务
│   ├── git_service.go        # Git 仓库克隆/拉取服务（go-git 封装）
│   ├── agent_factory.go      # LLM Agent 工厂（创建 SubAgent 运行实例）
│   ├── crypto_helper.go      # AES-256-GCM 加解密辅助（LLM API Key / SSH 私钥加密存储）
│   ├── dependency_extractor.go # 依赖关系提取器（解析 import/require 提取模块依赖图）
│   ├── file_scanner.go       # 文件扫描器（仓库文件清单 + 忽略规则 + 大小限制）
│   ├── llm_provider.go       # LLM Provider 服务接口（调用外部 LLM API）
│   ├── llm_provider_stub.go  # LLM Provider Stub（测试用桩实现）
│   ├── llm_resolver.go       # LLM Resolver（按 Agent 角色解析运行时模型配置）
│   ├── prompt_loader.go      # Prompt 加载器（从 resources/prompts 读取内嵌 prompt）
│   ├── repo_tools.go         # 仓库工具集（文件读取/目录树/搜索等 RepoWiki 子 Agent 工具）
│   ├── ssh_key_gen.go        # SSH 密钥对生成（ed25519/rsa）
│   ├── webhook_parser.go     # Webhook Payload 解析器（GitHub/GitLab 事件格式）
│   ├── webhook_signer.go     # Webhook HMAC 签名生成与校验
│   ├── oauth_crypto.go       # OAuth 随机令牌 / SHA-256 摘要 / PKCE S256 校验
│   └── ai_plugin.go          # AI 插件打包（embed 或调试热重载 → ZIP + SHA-256 + 市场/技能清单）
├── entity/                   # GORM 实体
│   ├── info.go               # 站点配置实体（单用户模式）
│   ├── apikey.go             # API Key 实体（密钥哈希/前缀/后缀/过期时间）
│   ├── workspace.go          # 工作空间实体（Gene=48，名称/描述/图标/默认标记）
│   ├── project.go            # 项目实体（Gene=32，所属工作空间/名称/别名/描述）
│   ├── pin.go                # Pin 实体（Gene=34，标题/内容/分类/优先级/来源/目标项目）
│   ├── biometric_credential.go # WebAuthn 凭证实体（Gene=37，CredentialID/公钥/签名计数）
│   ├── qa_session.go         # Q&A Session 实体（Gene=35，状态/类型/TTL）
│   ├── qa_question.go        # Q&A Question 实体（Gene=36，题型/标题/选项/回答）
│   ├── qa_supplement.go      # Q&A Supplement 实体（Gene=38，补充内容/附件）
│   ├── repowiki_config.go    # RepoWiki 配置实体（Gene=39，仓库地址/LLM 参数/当前选中版本/Webhook 配置）
│   ├── wiki_version.go       # Wiki 版本实体（Gene=40，版本号/状态/文件路径/token 统计）
│   ├── llm_provider.go       # LLM Provider 实体（Gene=41，名称/BaseURL/加密 API Key）
│   ├── llm_model.go          # LLM Model 实体（Gene=42，Provider 关联/模型名/参数/Agent 角色分配）
│   ├── ssh_key.go            # SSH Key 实体（Gene=44，名称/指纹/公钥/加密私钥）
│   ├── webhook_event.go      # Webhook 事件实体（Gene=43，事件 ID/分支/状态/Payload 摘要）
│   ├── preview_session.go    # Preview 会话实体（Gene=45，Hash/标题/状态/Fork 溯源/过期时间）
│   ├── preview_file.go       # Preview 文件实体（Gene=46，会话关联/文件名/MIME/内容）
│   ├── oauth_client.go       # OAuth 动态客户端（Gene=47，名称/回调 JSON 数组/scope=mcp）
│   ├── page.go               # Page 实体（Gene=49，项目内 slug/生效指针/访问策略）
│   ├── page_version.go       # PageVersion 实体（Gene=50，不可变快照版本）
│   └── page_file.go          # PageFile 实体（Gene=51，版本内扁平文件）
├── mcp/                      # MCP Server 工具注册
│   ├── server.go             # MCP Server 初始化 + StreamableHTTPHandler 创建 + 38 个业务工具注册
│   ├── workspace_tools.go    # Workspace MCP 工具（只读：workspace_list / workspace_get）
│   ├── project_tools.go      # Project MCP 工具（project_list / project_create / project_get）
│   ├── pin_tools.go          # Pin MCP 工具（pin_push / pin_consume / pin_peek / pin_list / pin_update）
│   ├── repowiki_tools.go     # RepoWiki MCP 工具（只读：repoWiki_query / repoWiki_list）
│   ├── preview_tools.go      # Preview MCP 工具注册（7 个基础文件与会话工具）
│   ├── preview_handlers.go   # Preview MCP 工具 handler 实现（快照、单文件与行级编辑）
│   ├── preview_schemas.go    # Preview MCP 输出 Schema 构建器
│   ├── preview_lpw_tools.go  # Preview LPW 节点语义工具族注册（7 个工具）
│   ├── preview_lpw_handlers.go # Preview LPW 节点语义工具 handler 实现
│   ├── pages_tools.go        # Pages MCP 工具（3 个：pages_list / pages_promote / pages_fork）
│   ├── qa_tools.go           # Q&A MCP 工具注册（10 个工具定义与 schema）
│   ├── qa_handlers.go        # Q&A MCP 工具 handler 实现（工具执行逻辑）
│   └── qa_type_details.go    # Q&A MCP 题型详情定义（15+ 题型 schema 细节）
├── websocket/                # WebSocket 实时通信层
│   ├── hub.go                # 连接管理器（sessionID → deviceID 二级索引 + 心跳检测）
│   ├── handler.go            # WebSocket 升级处理器 + 业务消息分发
│   ├── connection.go         # 单个连接封装（Kind 区分 qa/preview + 读写 goroutine + 优雅关闭）
│   ├── preview_handler.go    # Preview 消息处理（preview_sync 推送）
│   └── message.go            # 消息类型定义（16 种消息类型）
├── qa/                       # Q&A 回答队列
│   └── queue.go              # 会话级 FIFO 回答队列（Enqueue/Consume/WaitAndConsume）
└── constant/                 # 共享业务常量
    ├── cache.go              # Redis Key 前缀/过期时间（带环境前缀格式化）
    ├── context.go            # Context Key（如 CtxOwnerKey、RepoWikiLogicKey）
    ├── gene_number.go        # 雪花算法基因编号（GeneProject=32 ~ GenePageFile=51）
    ├── info_key.go           # Info 表配置键常量（键名规范：层级 . 分隔、同层多词 - 连接）
    ├── biometric.go          # WebAuthn 相关常量（RP ID/Origin/超时）
    ├── workspace.go          # 工作空间常量（默认空间名称与标识）
    ├── project.go            # 项目相关常量
    ├── pin.go                # Pin 模块常量（分类/优先级枚举）
    ├── llm.go                # LLM 模块常量（Agent 角色/模型参数默认值）
    ├── oauth.go              # MCP OAuth 2.1 常量（令牌前缀 lum_at_/lum_rt_、端点路径、PKCE S256）
    ├── ai_plugin.go          # AI 插件分发常量（市场路径、ZIP 文件名、embed 根目录）
    ├── repowiki.go           # RepoWiki 模块常量（状态/角色/环境变量键）
    ├── preview.go            # Preview 模块常量（MIME 类型/大小上限/过期时间/状态）
    ├── page.go               # Pages 模块常量（访问策略枚举/Cookie 键/Token 有效期）
    └── settings.go           # 系统设置常量（配置分组键/默认值）
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
| 新增共享服务 | `service/*.go` | 跨模块复用的基础设施（如 Git 服务、加密辅助、Webhook 解析、LPW Schema） |
| 新增实体 | `entity/*.go` + `main.go` | 实现并追加到 `WithAutoMigrate(...)` 列表 |
| 新增基因编号 | `constant/gene_number.go` | 定义 `GeneXxx` 常量供实体 `GetGene()` 使用 |
| 新增种子数据 | `startup/prepare/` | 创建 `prepare_<domain>.go` |
| 新增请求/响应 DTO | `api/<domain>/` | 按业务域保持子包结构 |
| 新增业务常量 | `constant/*.go` | 基因编号、缓存 Key、Context Key、模块枚举 |
| 新增 Info 配置键 | `constant/info_key.go` | 键名规范：层级 `.` 分隔、同层多词 `-` 连接、禁止 `_` |
| 新增 MCP 工具 | `mcp/*.go` | 在 `server.go` 注册，在 `startup_mcp.go` 注入 Logic |
| 新增 WebSocket 消息类型 | `websocket/message.go` | 定义 MessageType 常量和 Message 结构 |
| 新增 Q&A 回答队列 | `qa/queue.go` | 会话级 FIFO 队列，由 `QaLogic` 调用 |
| 新增 Preview 会话/文件 | `entity/preview_*.go` + `logic/preview_logic.go` + `handler/preview.go` + `mcp/preview_tools.go` | 文件上限 256KB，MIME 推断，`preview_sync` 实时同步 |
| 新增 LPW 分块写入操作 | `logic/preview_lpw_logic.go` + `preview_lpw_contract.go` + `mcp/preview_lpw_tools.go` | Schema 1.1、kind 层级与节点增删改查 |
| 新增 Pages 页面/版本 | `entity/page*.go` + `logic/pages_logic.go` + `handler/pages.go` + `mcp/pages_tools.go` | 不可变快照、Fork 派生、路径直出、密码门 Cookie 鉴权 |
| 新增 Dashboard 统计 | `logic/dashboard.go` + `repository/dashboard.go` + `handler/dashboard.go` | 六类指标聚合，原生 SQL |
| 新增 RepoWiki 分析入口 | `logic/repowiki_logic.go` | 配置/版本 CRUD 与分析启动 |
| 新增 RepoWiki 子 Agent 编排 | `logic/repowiki_orchestrator.go` | 5 阶段预定义编排，不持有 db/rdb |
| 新增 RepoWiki 分析管道 | `logic/repowiki_pipeline.go` | Git 准备 + 状态机驱动 |
| 新增 RepoWiki 文件存储 | `service/wiki_storage.go` | 版本隔离路径管理与文件 I/O |
| 新增 RepoWiki MCP 工具 | `mcp/repowiki_tools.go` | 只读工具：repoWiki_query / repoWiki_list |
| 新增 LLM Provider/Model | `entity/llm_*.go` + `repository/llm_*.go` + `logic/llm_*.go` + `handler/llm.go` | 热配置，API Key 经 `crypto_helper.go` AES-256-GCM 加密 |
| 新增 SSH Key 管理 | `entity/ssh_key.go` + `repository/ssh_key.go` + `logic/ssh_key.go` + `service/ssh_key_gen.go` | 密钥对生成走 `ssh_key_gen.go`，私钥加密存储 |
| 新增 Webhook 接收 | `handler/webhook.go` + `logic/repowiki_webhook.go` + `service/webhook_parser.go` + `webhook_signer.go` | HMAC 签名校验在 `webhook_signer.go` |
| 新增系统设置 | `logic/settings.go` + `handler/settings.go` + `constant/settings.go` | 分组配置读写，持久化到 Info 表 |
| 新增 MCP OAuth 端点 | `handler/oauth.go` + `logic/oauth_logic.go` + `route_oauth.go` | 公开端点必须在 `engine.Use()` 之前注册（裸 JSON）；consent 走登录态 |
| 改 MCP 认证 | `app/middleware/mcp_auth.go` | OAuth `lum_at_` 优先、API Key 回退；401 带 `WWW-Authenticate` 资源元数据 |
| 新增 AI 插件分发 | `handler/ai_plugin.go` + `logic/ai_plugin.go` + `service/ai_plugin.go` | 源码在 `resources/ai-plugin`，运行时打包 ZIP；ZCode 用独立 marketplace.zcode.json |

## 代码地图

| 符号 | 类型 | 位置 | 作用 |
|---|---|---|---|
| `NewHandler[T]` | 泛型函数 | `handler/handler.go` | Handler 泛型构造模式，注入全部 Logic（18 个，含 OAuth / AI Plugin / Workspace / Pages） |
| `BindJSON` | 辅助函数 | `handler/bind.go` | 统一请求绑定 + 分页参数规范化 |
| `computeNav` | 函数 | `handler/wiki_reader.go` | 根据 manifest 计算当前 Wiki 页的 prev/next/breadcrumb |
| `Cors` | 中间件 | `app/middleware/cors.go` | 白名单 CORS（`XLF_ALLOWED_ORIGINS`），命中才反射 `Access-Control-Allow-Origin` |
| `SecurityHeaders` | 中间件 | `app/middleware/security.go` | 安全响应头（nosniff / X-Frame-Options SAMEORIGIN / CSP frame-ancestors 'self'） |
| `WebAuthnOrigin` | 中间件 | `app/middleware/webauthn.go` | 按请求解析浏览器 Origin 注入 context，供 RPID 动态推导 |
| `McpAuth` | 中间件 | `app/middleware/mcp_auth.go` | MCP 认证：`lum_at_` OAuth 优先，其余 Bearer 走 API Key；失败返回 RFC 资源元数据 |
| `PagesAuth` | 中间件 | `app/middleware/pages_auth.go` | Pages 密码门 HMAC Cookie 校验中间件 |
| `WorkspaceLogic` | 结构体 | `logic/workspace.go` | 工作空间 CRUD、默认空间保证与项目计数 |
| `PagesLogic` | 结构体 | `logic/pages_logic.go` | Pages 晋升快照、Fork、版本指针与密码门认证 |
| `PreviewLogic` | 结构体 | `logic/preview_logic.go` | Preview 会话/文件编排 + 行级编辑 + 会话过期 + WebSocket 同步回调 |
| `PreviewLpwLogic` | 结构体 | `logic/preview_lpw_logic.go` | LPW 1.1 节点文档编排（kind 层级校验 + 节点写入） |
| `DashboardLogic` | 结构体 | `logic/dashboard.go` | 看板六类指标聚合 |
| `resolveRuntimeDomain` | 函数 | `logic/runtime_url.go` | 解析站点运行时域名（Info site.domain → env 回退） |
| `buildPreviewURL` | 函数 | `logic/runtime_url.go` | 构建路径式 `/preview/<hash>/<filename>` 深链 |
| `buildPageURL` | 函数 | `logic/runtime_url.go` | 构建路径式 `/pages/<projectName>/<slug>/<filename>` 深链 |
| `SubAgentOrchestrator` | 结构体 | `logic/repowiki_orchestrator.go` | 5 角色 SubAgent 编排引擎（overview → explore → architect → writer → validator） |
| `AnalysisPipeline` | 结构体 | `logic/repowiki_pipeline.go` | RepoWiki 分析管道（Git 准备 + 状态机驱动） |
| `RepoWikiLogic` | 结构体 | `logic/repowiki_logic.go` | RepoWiki 业务编排（配置/版本/分析入口） |
| `GetRepoWikiLogicFromContext` | 函数 | `logic/repowiki_logic.go` | 从 context 获取 RepoWikiLogic（MCP/Cron/Handler 共用） |
| `WikiStorageService` | 结构体 | `service/wiki_storage.go` | RepoWiki 文件系统存储与路径管理 |
| `ReadPage` | 方法 | `service/wiki_storage.go` | 读取 .mdx 页面文件并解析 YAML frontmatter |
| `ReadMetaJSON` | 方法 | `service/wiki_storage.go` | 读取 per-folder meta.json（路径不存在时返回 nil, nil） |
| `GitService` | 结构体 | `service/git_service.go` | Git 仓库克隆/拉取（go-git 封装） |
| `AgentFactory` | 结构体 | `service/agent_factory.go` | LLM Agent 工厂（创建 SubAgent 运行实例） |
| `CryptoHelper` | 结构体 | `service/crypto_helper.go` | AES-256-GCM 加解密（LLM API Key / SSH 私钥加密存储） |
| `FileCacheService` | 结构体 | `service/file_cache.go` | Q&A 会话文件缓存（base64 落盘 + 路径穿越防护 + 会话清理） |
| `DependencyExtractor` | 结构体 | `service/dependency_extractor.go` | 依赖关系提取（解析 import/require 构建模块依赖图） |
| `FileScanner` | 结构体 | `service/file_scanner.go` | 仓库文件扫描（清单 + 忽略规则 + 大小限制） |
| `LLMResolver` | 结构体 | `service/llm_resolver.go` | 按 Agent 角色解析运行时模型配置 |
| `PromptLoader` | 结构体 | `service/prompt_loader.go` | 从 `resources/prompts` 读取内嵌 prompt 文件 |
| `RepoTools` | 结构体 | `service/repo_tools.go` | RepoWiki 子 Agent 工具集（文件读取/目录树/搜索） |
| `SshKeyGen` | 结构体 | `service/ssh_key_gen.go` | SSH 密钥对生成（ed25519/rsa） |
| `WebhookParser` | 结构体 | `service/webhook_parser.go` | Webhook Payload 解析（GitHub/GitLab 事件格式） |
| `WebhookSigner` | 结构体 | `service/webhook_signer.go` | Webhook HMAC 签名生成与校验 |
| `PageAuthToken` | 结构体 | `service/page_auth_token.go` | Pages 密码门 HMAC Cookie 签发与校验 |
| `LpwSchemaLoader` | 结构体 | `service/lpw_schema.go` | LPW JSON Schema 加载与静态规范校验 |
| `RepoWikiConfig` | 结构体 | `entity/repowiki_config.go` | RepoWiki 配置实体（仓库地址/Webhook 配置/当前选中版本） |
| `WikiVersion` | 结构体 | `entity/wiki_version.go` | Wiki 版本实体（版本号/状态/文件路径/token 统计） |
| `LlmProvider` | 结构体 | `entity/llm_provider.go` | LLM Provider 实体（名称/BaseURL/加密 API Key） |
| `LlmModel` | 结构体 | `entity/llm_model.go` | LLM Model 实体（Provider 关联/Agent 角色分配） |
| `SshKey` | 结构体 | `entity/ssh_key.go` | SSH Key 实体（名称/指纹/公钥/加密私钥） |
| `WebhookEvent` | 结构体 | `entity/webhook_event.go` | Webhook 事件实体（事件 ID/分支/状态） |
| `Workspace` | 结构体 | `entity/workspace.go` | 工作空间实体（Gene=48，名称/描述/图标/默认标记） |
| `PreviewSession` | 结构体 | `entity/preview_session.go` | Preview 会话实体（Gene=45，Hash 16 位 + 标题/状态/过期时间） |
| `PreviewFile` | 结构体 | `entity/preview_file.go` | Preview 文件实体（Gene=46，会话关联 + 文件名/MIME/内容） |
| `Page` | 结构体 | `entity/page.go` | Page 实体（Gene=49，项目内 slug/生效版本指针/访问策略） |
| `PageVersion` | 结构体 | `entity/page_version.go` | PageVersion 实体（Gene=50，不可变快照版本/SemVer 语义） |
| `PageFile` | 结构体 | `entity/page_file.go` | PageFile 实体（Gene=51，版本内扁平文件） |
| `OAuthClient` | 结构体 | `entity/oauth_client.go` | OAuth 动态客户端（Gene=47，公共客户端，无 secret，仅 PKCE） |
| `OAuthLogic` | 结构体 | `logic/oauth_logic.go` | MCP OAuth 2.1 编排（元数据 / DCR / 授权码+PKCE / 令牌刷新 / RFC 8707） |
| `OAuthStore` | 结构体 | `repository/cache/oauth_store.go` | 授权请求、授权码、令牌元数据（短 TTL，仅缓存） |
| `OAuthRandomToken` | 函数 | `service/oauth_crypto.go` | 256 位随机令牌 + SHA-256 摘要 + PKCE S256 校验 |
| `AIPluginLogic` | 结构体 | `logic/ai_plugin.go` | 按请求域名渲染 marketplace / ZIP / well-known |
| `AIPluginService` | 结构体 | `service/ai_plugin.go` | 从 embed（或 `LUMINA_AI_PLUGIN_DIR` / 调试热重载）打包插件 |
| `RepoWikiConfigRepo` | 结构体 | `repository/repowiki_config.go` | RepoWikiConfig 持久化 |
| `WikiVersionRepo` | 结构体 | `repository/wiki_version.go` | WikiVersion 持久化 |
| `LlmProviderRepo` | 结构体 | `repository/llm_provider.go` | LlmProvider 持久化 |
| `LlmModelRepo` | 结构体 | `repository/llm_model.go` | LlmModel 持久化 |
| `SshKeyRepo` | 结构体 | `repository/ssh_key.go` | SshKey 持久化 |
| `WebhookEventRepo` | 结构体 | `repository/webhook_event.go` | WebhookEvent 持久化 |
| `WikiAuthToken` | 结构体 | `service/wiki_auth_token.go` | Wiki 访问密码 Token 生成与校验 |
| `CoordinatorSystemPrompt` | 常量 | `logic/repowiki_subagent_prompts.go` | Coordinator 角色 system prompt |
| `BuildOverviewUserPrompt` | 函数 | `logic/repowiki_subagent_prompts.go` | 构建 Coordinator 概要阶段 user prompt |
| `WikiEntry` | 结构体 | `logic/repowiki_types.go` | Architect 输出的 Wiki 目录条目 |
| `ValidationError` | 结构体 | `logic/repowiki_types.go` | Validator 输出的校验错误项 |
| `ExploreOutput` | 结构体 | `logic/repowiki_types.go` | Explore Agent 的单个产出项 |
| `ModelRunConfig` | 结构体 | `logic/repowiki_types.go` | Agent 运行时的模型配置 |
| `RegisterRepoWikiTools` | 函数 | `mcp/repowiki_tools.go` | 注册 RepoWiki MCP 只读工具 |
| `SetRepoWikiLogic` | 函数 | `mcp/repowiki_tools.go` | 注入 RepoWikiLogic 到 MCP 工具 |
| `RegisterPreviewTools` | 函数 | `mcp/preview_tools.go` | 注册 Preview 基础 MCP 工具（7 个） |
| `RegisterPreviewLpwTools` | 函数 | `mcp/preview_lpw_tools.go` | 注册 Preview LPW 节点语义 MCP 工具族（7 个） |
| `RegisterPagesTools` | 函数 | `mcp/pages_tools.go` | 注册 Pages MCP 工具（3 个） |

## 约定
- **严格分层**：route → middleware → handler → logic → repository；禁止跳层调用。`service/` 层为跨模块共享基础设施，可被 logic 层调用。
- **Handler 精简**：仅做请求绑定、调用 logic、映射结果到响应；禁止在 handler 中直接操作 DB/Redis。请求绑定走 `handler/bind.go` 的统一辅助函数。
- **Logic 编排**：业务编排层，持久化和 SQL 归 repository 层；共享服务（如下载 Token、Git 克隆、加密辅助、LPW Schema）通过 `service/` 调用。
- **Repository 返回值**：统一 `(data, *xError.Error)` 风格，不用裸 `error`。
- **日志命名**：按层使用 `xLog.WithName` — `NamedCONT`（handler）、`NamedLOGC`（logic）、`NamedREPO`（repository）、`NamedINIT`（startup）、`NamedMIDE`（middleware）、`NamedCRON`（cron）。
- **上下文传递**：使用 `ctx.Request.Context()` 或注入的 context 下发调用。
- **认证中间件**：通过 `middleware.Auth(ctx)` 创建，注入认证标记到 context（`CtxOwnerKey`）。
- **API Key 中间件**：`middleware.ApikeyAuth` 验证 `lumi_` 前缀的 API Key；MCP 端点不再挂它，改走 `McpAuth`。
- **MCP 认证中间件**：`middleware.McpAuth` 在 MCP 路由上使用——Bearer 带 `lum_at_` 前缀走 OAuth（含 RFC 8707 资源绑定），其余走 API Key；401 必须带 `WWW-Authenticate: Bearer resource_metadata=...`，否则客户端无法发现授权服务器。
- **Wiki Auth 中间件**：`middleware.WikiAuth` 处理 Wiki Reader 的密码 Token / Cookie 会话认证，保护 `/wiki/*` 路由。
- **Pages 密码门**：`middleware.PagesAuth` 拦截 Pages 子资源并校验 HMAC 签名 Cookie；顶层 document 导航放行给前端 SPA 渲染密码表单。
- **沙盒隔离中间件**：`middleware.SandboxSubresource` 保障 Preview 与 Pages 的静态子资源隔离安全，严格防御跨路径逃逸。
- **MCP 兼容中间件**：`middleware.McpCompat` 处理 Streamable HTTP 请求的兼容性（如 SSE 响应头处理）。
- **安全中间件**：`middleware.SecurityHeaders` 设置安全响应头（nosniff / X-Frame-Options / CSP）；`middleware.Cors` 按 `XLF_ALLOWED_ORIGINS` 白名单反射 CORS；`middleware.WebAuthnOrigin` 按请求解析 Origin 注入 context。三者均在 `route.go` 全局注册。
- **泛型 Handler 构造**：`NewHandler[T]` 统一注入所有 logic 实例到 `service` 结构体（health/auth/apikey/project/qa/biometric/pin/repowiki/ssh/llmProvider/llmModel/settings/preview/dashboard/aiPlugin/oauth/workspace/pages 共 18 个）。
- **实体 ID 策略**：雪花算法基因策略；每个实体必须实现 `GetGene() xSnowflake.Gene`，基因编号定义在 `constant/gene_number.go`（GeneProject=32 ~ GenePageFile=51）。
- **字段注释**：实体字段必须追加行尾中文注释（`// 字段说明`），且与 `gorm comment` 一致。
- **Info 配置键统一**：所有 Info 表键名在 `constant/info_key.go` 集中定义，禁止在业务代码写死键名字符串；键名规范为层级 `.` 分隔、同层多词 `-` 连接、禁止 `_`（如 `qa.session.ttl`）。
- **缓存键前缀**：通过 `xEnv.NoSqlPrefix` 环境变量自动拼接前缀，使用 `RedisKey.Get(args...)` 格式化。
- **分页规范**：使用 `xModels.PageRequest.Normalize()` 规范化分页参数，`xModels.NewPageFromRequest` 构建分页响应。
- **API Key 安全**：密钥使用 bcrypt 哈希存储，仅创建/重置时返回完整密钥；查询和列表使用 `maskKey` 脱敏。
- **LLM API Key 加密**：LLM Provider 的 API Key 使用 AES-256-GCM 加密存储（`service/crypto_helper.go`），密钥由 `LLM_ENCRYPT_SECRET` 环境变量提供；禁止明文存储。
- **SSH 私钥加密**：SSH Key 的私钥使用 AES-256-GCM 加密存储（`crypto_helper.go` 的 `EncryptSSHPrivateKey`/`DecryptSSHPrivateKey`，解密向后兼容明文 PEM），禁止明文落库。
- **工作空间隔离**：Workspace 是顶层身份层，每个项目必归属于一个空间（`project.workspace_id` NOT NULL）；系统包含唯一默认空间。
- **Project 缓存策略**：采用 Cache-Aside 模式（ID→详情、Name→ID、Alias→ID 三层映射，TTL 30 分钟），支持按 Workspace 批量更新与失效。
- **RepoWiki Logic 注入**：通过 `logic.GetRepoWikiLogicFromContext(ctx)` 获取，由 `startup_repowiki.go` 在启动阶段注入到 context。
- **MCP 路由**：必须在 `engine.Use()` 之前注册（绕开 `ResponseMiddleware`），使用 `gin.WrapH` 包装 `http.Handler`。OAuth well-known / authorize / register / token 同理，输出裸 JSON。
- **MCP OAuth 2.1**：Lumina 同时当授权服务器与资源服务器；仅 `authorization_code` + PKCE S256 + `refresh_token`；公共客户端无 `client_secret`；访问令牌前缀 `lum_at_`、刷新令牌 `lum_rt_`；令牌原文不进缓存，只存 SHA-256 摘要。TTL 可由 `LUMINA_OAUTH_ACCESS_TTL` / `LUMINA_OAUTH_REFRESH_TTL`（秒）覆盖。
- **AI 插件分发**：源码在 `resources/ai-plugin`（`go:embed all:ai-plugin`），运行时由 `AIPluginService` 打 ZIP 并渲染带真实 SHA-256 的 marketplace.json。调试态（`XLF_DEBUG=true`）每次从磁盘重读；`LUMINA_AI_PLUGIN_DIR` 可覆盖源目录。ZCode 不支持 archive 源，必须走 `/api/v1/plugins/marketplace.zcode.json`（url+zip）。
- **MCP Logic 注入**：`startup_mcp.go` 中通过 `mcp.SetQaLogic/SetProjectLogic/SetPinLogic/SetPreviewLogic/SetPreviewLpwLogic/SetPagesLogic/SetWorkspaceLogic/SetRepoWikiLogic` 注入 Logic 实例。
- **MCP 工具集结构**：MCP 端点共注册 39 个工具，分属八个领域：Workspace（2）、QA（10）、Project（3）、Pin（5）、RepoWiki（2）、Preview（7）、Preview LPW（7）、Pages（3）。
- **Pages 快照与密码门**：Pages 承接 Preview 的晋升快照，深拷贝文件并冻结为不可变版本；支持 Fork 派生回 Preview 会话。访问策略在控制台配置，密码门走 HMAC Cookie 认证；MCP 工具严禁在晋升时设置密码。
- **Preview 行级编辑与过期**：Preview 模块支持单文件 256KB 上限、行级精准替换与读取（`preview_lines.go`），历史会话支持自动过期与 Cron 定时清理（`ExpireStaleSessions`）。
- **LPW 文档渲染引擎**：LPW（Lumina Paper Workshop）支持基于 Schema 校验的结构化分块写入（追加/插入/修改/删除/巡检），前端优先使用 React 原生渲染与 ECharts 懒加载。
- **WebSocket 管理**：`Hub` 按 sessionID → deviceID 二级索引管理连接；心跳检测间隔 5s，超时 15s；连接 `Kind` 区分 `qa`/`preview`，单帧上限 10MB。
- **Q&A 推送回调**：`logic.OnQuestionPushed` / `logic.OnSupplementPushed` / `logic.OnQuestionCancelled` / `logic.OnSessionArchived` 函数变量在 `route_ws.go` 中设置，解耦 Logic 层和 WebSocket 层。
- **回答队列**：每个 Session 独立的 FIFO 队列，支持 `WaitAndConsume` 阻塞等待新回答。
- **Pin FIFO 消费**：Pin 模块的 FIFO 消费基于数据库实现（`ConsumeOldestPending` + `ConsumeByID`），不依赖 Redis 队列。
- **WebAuthn 凭证**：CredentialID 全局唯一，使用 bcrypt 存储公钥；Challenge 通过 Redis 缓存短暂会话（`cache/biometric_credential.go`）。
- **WebAuthn RPID 动态推导**：按请求 Origin（`middleware.WebAuthnOrigin` 双写 Request.Context 与 Keys 注入）动态推导 RPID/RPOrigins，支持注册域后缀共享凭证（经 publicsuffix 校验）；`XLF_BIOMETRIC_ALLOWED_ORIGINS` 按完整 Origin（scheme://host[:port]）严格匹配阻止 DNS-rebinding。白名单外或配置与访问域名不匹配时直接返回显式错误，禁止静默回退启动期静态配置。
- **文件下载 Token**：`service/download_token.go` 生成短时效签名 Token，用于 Q&A 文件附件下载鉴权。
- **文件缓存防护**：`service/file_cache.go` 的 `IsWithinCacheDir` 用绝对路径 + `EvalSymlinks` + `filepath.Rel` 前缀校验，防路径穿越与符号链接逃逸。
- **QA 逻辑拆分**：`qa.go` 已按职责拆分为 `qa_logic.go`（核心编排）、`qa_format.go`（题型格式化）、`qa_helper.go`（辅助函数）、`qa_mcp.go`（MCP 工具）、`qa_mcp_helpers.go`（MCP 辅助）、`qa_download.go`（文件下载）；新增 QA 逻辑时按职责归入对应文件。
- **MCP 工具拆分**：`mcp/qa_tools.go`（工具注册）、`qa_handlers.go`（handler 实现）、`qa_type_details.go`（题型 schema）三文件分工；新增 MCP 工具时在 `qa_tools.go` 注册、`qa_handlers.go` 实现。
- **RepoWiki 子 Agent 编排**：`SubAgentOrchestrator` 按预定义 5 阶段（Coordinator → Explore → Architect → Writer → Validator）生成 Wiki，prompt 模板内嵌在 `resources/prompts/*.md` 通过 `service/prompt_loader.go` 加载，`repowiki_subagent_prompts.go` 负责动态构建 user prompt，`repowiki_types.go` 定义内部类型，`repowiki_pipeline.go` 负责 Git 准备与状态机驱动。
- **RepoWiki 版本隔离**：每个 Wiki 版本存储在 `versions/{vid}/` 下，`RepoWikiConfig.SelectedVersionID` 指定当前对外服务版本；旧版 config 级目录已废弃，新版本完成后清理。
- **RepoWiki MCP 只读**：MCP 端仅暴露 `repoWiki_query` / `repoWiki_list` 两个只读工具，Wiki 更新由 Git Webhook 自动触发。
- **RepoWiki 资源内嵌**：5 角色 system prompt 通过 `go:embed` 内嵌在 `resources/prompts/`，由 `service/prompt_loader.go` 读取，禁止在 logic 中硬编码 prompt 文本。
- **Webhook 签名校验**：所有 Webhook 请求必须经 `service/webhook_signer.go` 校验 HMAC 签名，签名密钥由 `REPOWIKI_HMAC_SECRET` 环境变量提供。
- **系统设置分组**：设置项按分组（站点/安全/Q&A/RepoWiki/Preview）组织，持久化到 Info 表，通过 `logic/settings.go` 统一读写。
- **双前端嵌入**：`web/dist`（主控制台）和 `web-wiki/dist`（Wiki Reader）分别通过 `go:embed` 嵌入，由 `route_frontend.go` 统一提供 SPA fallback。

## 反模式
- 禁止从路由直接调用 repository 或绕过 logic 层。
- 禁止在 handler 中手写原始 Gin JSON 响应；使用 `xResult` 辅助函数。
- 禁止在 logic/repository 构造函数内部创建 DB/Redis 客户端；应从 context 获取注入的依赖。
- 禁止绕过 `NewHandler[T]` 模式手动构造 handler。
- 禁止将业务常量写在 handler/logic 文件中；统一放 `constant/`。
- 禁止核心业务模块（RepoWiki、Memory、Q&A、Pin、Preview、Pages）之间直接调用。
- 禁止直接使用 `os.Getenv`；应使用带默认值的 `xEnv.GetEnv*`。
- 禁止新增实体后不追加到 `main.go` 的 `WithAutoMigrate(...)`。
- 禁止在业务代码中写死 Info 键名字符串；统一用 `constant/info_key.go` 定义的常量。
- 禁止在 repository 外部直接操作 Redis；缓存逻辑封装在 repository/cache 子层。
- 禁止 logic 结构体持有 db/rdb 字段；所有数据访问必须经由 repository（+cache 子层）。
- 禁止在 logic 层拼接 GORM/Redis 命令（含 entity.Info 配置读取）；统一走 InfoRepo。
- 禁止在 repository 内定义私有 cacheKey() 方法；缓存键统一用 `constant.RedisKey.Get()`。
- 禁止在 Q&A 模块使用 SSE 进行问题推送；统一使用 WebSocket。
- 禁止 WebAuthn Challenge 存储在内存中；必须通过 Redis 缓存以支持多实例部署。
- 禁止明文存储 LLM API Key 或 SSH 私钥；必须经 `crypto_helper.go` AES-256-GCM 加密。
- 禁止在 logic 中硬编码 RepoWiki prompt 文本；统一放 `resources/prompts/` 通过 `prompt_loader.go` 加载。
- 禁止在 Webhook 处理中跳过 HMAC 签名校验。
- 禁止在 `resources/prompts/` 外散落 prompt 文件；所有内嵌静态资源集中在 `resources/` 目录。
- 禁止给 MCP 端点重新挂 `ApikeyAuth` 而丢掉 `McpAuth`——没有 `WWW-Authenticate` 资源元数据，OAuth 客户端无法发现授权服务器。
- 禁止把 OAuth 访问/刷新令牌原文写入 Redis；缓存键必须是 SHA-256 摘要（`service.OAuthHashToken`）。
- 禁止给 OAuth 动态客户端签发 `client_secret`；本实现只支持 PKCE 公共客户端。
- 禁止手写 marketplace.json 当作运行时产物；清单必须由 `AIPluginService` 按当前域名和 ZIP SHA-256 渲染。
- 禁止在插件技能里硬编码实例 MCP 地址；打包时由服务写入 `.mcp.json`。
- 禁止在 Pages 晋升时通过 MCP 设置密码门密码；密码门访问策略只能在控制台设置。

## 调试路径
1. 请求未路由 → 检查 `app/route/route.go` 路由组和 `route_*.go` 注册。
2. 路由正确但响应异常 → 检查 `handler/*.go` 绑定，然后 `logic/*.go` 编排。
3. 认证失败 → 检查 `middleware/auth.go` → `logic/auth.go` → `repository/cache/token.go`。
4. API Key 认证失败 → 检查 `middleware/apikey.go`（或 MCP 上的 `mcp_auth.go` 回退分支）→ `logic/apikey.go` 的 `ValidateAPIKey`。
4b. MCP OAuth 401 → 检查令牌是否 `lum_at_` 前缀、`ValidateAccessToken` 的 RFC 8707 资源是否等于当前 MCP URL、Redis `oauth:at:*` 是否过期；客户端发现失败则看 `/.well-known/oauth-protected-resource` 是否在 `engine.Use()` 之前注册。
4c. 插件 ZIP / 清单 500 → 检查 `resources/ai-plugin` 是否被 embed、`service/ai_plugin.go` 热重载目录、以及 `LUMINA_AI_PLUGIN_DIR`。
5. WebAuthn 注册/登录失败 → 检查 `handler/biometric.go` → `logic/biometric.go` → `repository/biometric_credential.go` + `cache/biometric_credential.go`（Challenge 是否过期）。
6. MCP 工具调用失败 → 检查 `mcp/server.go` 注册 + `startup_mcp.go` Logic 注入（SetXxxLogic）。
7. Pin 消费顺序异常 → 检查 `repository/pin.go` 的 `ConsumeOldestPending` 排序逻辑（FIFO 按 createdAt 升序）。
8. 数据库操作失败 → 检查 `repository/*.go` 和启动阶段的迁移状态（`main.go` 的 `WithAutoMigrate`）。
9. Redis 缓存异常 → 检查 `repository/cache/*.go` 和 `main.go` 的 `WithCache(FromEnv())` 连接配置。
10. WebSocket 连接问题 → 检查 `websocket/hub.go` 连接管理 + `route_ws.go` Hub 初始化。
11. Q&A 问题推送不达 → 检查 `logic.OnQuestionPushed` 回调是否在 `route_ws.go` 中正确设置。
12. 回答队列阻塞 → 检查 `qa/queue.go` 的 `WaitAndConsume` 和消费者 goroutine。
13. 文件下载失败 → 检查 `service/download_token.go` Token 校验 + `service/file_cache.go` 文件是否存在。
14. RepoWiki 分析失败 → 检查 `logic/repowiki_orchestrator.go` 5 阶段执行日志 + `logic/repowiki_pipeline.go` 状态机驱动；Git 问题先看 `service/git_service.go` 克隆/拉取日志。
15. RepoWiki prompt 缺失 → 检查 `resources/prompts/*.md` 是否存在 + `service/prompt_loader.go` 读取日志。
16. LLM Provider 调用失败 → 检查 `logic/llm_provider.go` 配置 + `service/crypto_helper.go` 解密 + `service/llm_resolver.go` 角色解析。
17. SSH Key 生成/导出异常 → 检查 `service/ssh_key_gen.go` + `logic/ssh_key.go` + `repository/ssh_key.go`（指纹/私钥加密）。
18. Webhook 触发失败 → 检查 `service/webhook_signer.go` HMAC 校验 + `service/webhook_parser.go` Payload 解析 + `logic/repowiki_webhook.go` 分支匹配。
19. Wiki Reader 401 → 检查 `middleware/wiki_auth.go` Token/Cookie 校验 + `service/wiki_auth_token.go`。
20. 系统设置读写异常 → 检查 `logic/settings.go` + `repository/info.go`（Info 表分组键）。
21. Preview 文件上传或行级编辑失败 → 检查 `logic/preview_logic.go` 的 `validateFilename`/`inferMimeType`、`logic/preview_lines.go` 的行号范围、文件大小是否超过 256KB。
22. Preview 实时同步不达 → 检查 `logic.OnPreviewChanged` 回调 + `websocket/preview_handler.go` 的 `preview_sync` 推送。
23. LPW 校验失败 → 检查 `resources/lpw/schema/v1.1.json` + `logic/preview_lpw_contract.go` 契约表 + 前端 `components/src/lpw/contract.ts`（三方快照对齐）。
24. Pages 访问或直出失败 → 检查 `middleware/pages_auth.go` Cookie 签名 + `service/page_auth_token.go` + `handler/serve.go` 文件存在性。
25. Workspace 切换或归属异常 → 检查 `repository/workspace.go` 的默认空间约束 + `logic/workspace.go` 的项目关联。
26. Dashboard 统计异常 → 检查 `repository/dashboard.go` 原生 SQL + `logic/dashboard.go` 聚合逻辑。

## 引用
- [startup/](./app/startup/AGENTS.md) — 启动模块详细文档
- [components/](../components/AGENTS.md) — 共享 UI / Markdown / 主题包（前端消费，后端不引用）
