# 架构总览

Lumina（微明）是给 AI Agent 用的代码知识与人机协作中枢：把仓库结构认知、跨项目约束、富交互问答和可视化预览做成可组合的服务，经 Streamable MCP、HTTP REST 与 WebSocket 对外暴露。人类通过控制台管理，Agent 通过 MCP 自行编排；本仓库不做跨模块编排。

运行形态是**单二进制**：控制台（`web`）与 Wiki Reader（`web-wiki`）的构建产物经 `go:embed` 嵌入 Go 进程。共享视觉语言在 `@lumina/components`。后端基于 `bamboo-base-go`，PostgreSQL 存业务数据，Redis 存会话/缓存/OAuth 令牌元数据。

本文只写**不常变的边界**。日常怎么改代码见 `AGENTS.md`；工程提案与架构决策统一登记在 `docs/README.md`。

## 解决什么问题

- Agent 需要一份可检索的仓库 Wiki，而不是每次从零扫代码。
- Agent 需要向人提问并阻塞等待裁决，而不是猜。
- 跨仓库的破坏性约定需要点对点投递并按 FIFO 消费。
- HTML/CSS/JS 原型需要沙盒预览，不能代替真实仓库交付。
- MCP 客户端需要标准 OAuth 2.1 或 API Key 接入，而不是私有协议。

外部依赖：PostgreSQL、Redis、Git 远端（RepoWiki）、LLM Provider（热配置，密钥加密存储）。Memory 模块仍在设计，未进入运行时。

## Codemap

| 模块 | 职责 | 关键符号 |
| --- | --- | --- |
| 接入 | MCP Streamable HTTP、REST、Webhook、插件市场 | `InitMCPServer`、`NewRoute`、`McpAuth` |
| 认证 / 用户 / WebAuthn | 单用户登录、资料、生物识别 | `AuthLogic`、`BiometricLogic` |
| API Key | `lumi_` 密钥，MCP 回退凭证 | `ApikeyLogic` |
| MCP OAuth 2.1 | 授权服务器 + 资源服务器，PKCE 公共客户端 | `OAuthLogic`、`OAuthClient`、`OAuthStore` |
| 项目 | 名称/别名/路径解析 | `ProjectLogic`、`ProjectRepo` |
| Q&A | 会话、题型推送、回答队列 | `QaLogic`、`Hub`（kind=qa）、`QueueManager` |
| Pin | 跨项目约束，数据库 FIFO | `PinLogic` |
| RepoWiki | 克隆 → 五角色 SubAgent → `.mdx` Wiki | `RepoWikiLogic`、`SubAgentOrchestrator`、`AnalysisPipeline` |
| Preview | 会话文件沙盒预览与分享 | `PreviewLogic`、`Hub`（kind=preview） |
| LLM 热配置 | Provider/Model 与 Agent 角色绑定 | `LlmProviderLogic`、`LLMResolver` |
| SSH / Webhook / 设置 | 克隆凭证、Git Push 触发、分组配置 | `SshKeyLogic`、`WebhookSigner`、`SettingsLogic` |
| AI 插件 | 运行时打包 ZIP 与 marketplace | `AIPluginService`、`AIPluginFS` |
| 前端 | 控制台、Wiki Reader、共享组件 | `web`、`web-wiki`、`@lumina/components` |

## 边界与不变量

这些规则在代码里几乎看不见「禁止」二字，但打破它们会让分层和模块独立性一起塌掉。

- **五领域互不调用。** RepoWiki、Memory、Q&A、Pin、Preview 不得互相 import 业务 Logic。组合只发生在 Agent 的 MCP 调用序列里。
- **分层单向。** `route` → `handler` → `logic` → `repository`。Handler 不得碰 DB/Redis；Logic 不得拼 GORM/Redis 命令，也不得自己持有 `db`/`rdb` 字段；缓存只进 `repository/cache`。
- **`service` 不是第六业务模块。** 只放跨域基础设施（Git、加解密、Webhook 签名、插件打包、Wiki 存储）。业务裁决仍回 Logic。
- **Q&A 实时通道是 WebSocket，不是 SSE。** 旧设计文档若仍写 SSE，视为过时。
- **MCP 认证走 `McpAuth`。** OAuth 访问令牌（`lum_at_`）优先，其余 Bearer 回退 API Key。不得给 MCP 只挂 `ApikeyAuth`：缺少 `WWW-Authenticate` 资源元数据，客户端发现不了授权服务器。
- **OAuth 公开端点与 MCP 一样，必须在 `engine.Use()` 之前注册**，输出裸 JSON，绕开统一响应中间件。
- **OAuth 只做 PKCE 公共客户端。** 动态注册不签发 `client_secret`；令牌原文不进 Redis，只存 SHA-256 摘要。
- **Webhook 必须校验 HMAC。** 密钥来自 `REPOWIKI_HMAC_SECRET`，没有「内网可跳过」的分支。
- **密钥不得明文落库。** LLM API Key 与 SSH 私钥经 AES-256-GCM；API Key 经 bcrypt，仅创建/重置时回发明文。
- **Info 键名只有一处定义。** 业务代码禁止写死 `"qa.session.ttl"` 这类字符串，走 `constant/info_key.go`。
- **RepoWiki prompt 不进 Logic 源码。** 五角色模板在 `resources/prompts/`，经 `PromptLoader` 读取。
- **Wiki 产物只认 `.mdx`。** 没有 `.md` fallback。
- **共享 UI 只在 `@lumina/components`。** `web` / `web-wiki` 不得再拷一份 shadcn、Markdown 原语或主题 CSS。
- **实体必须带基因。** `GetGene()` 使用 `GeneProject=32` 至 `GeneOAuthClient=47`，新增实体还要进 `main.go` 的 `WithAutoMigrate`（顺序跟 FK）。
- **定时任务不走裸 goroutine。** Cron 经 `xCronRunner`，Hub 主循环经 `xMain.Runner`。

## 横切关注点

**错误与响应。** 仓库/逻辑层返回 `*xError.Error`；Handler 成功走 `xResult.SuccessHasData`，失败 `ctx.Error`。禁止在 Handler 手写一套 JSON 信封。

**配置。** 环境变量用带默认值的 `xEnv.GetEnv*`，禁止 `os.Getenv`。数据库/缓存由 `main.go` 的 `WithDatabase` / `WithCache` 声明式装配，启动节点只注册依赖它们的业务（RepoWiki、MCP、种子）。

**日志。** 按层打模块名：`NamedMAIN` / `NamedINIT` / `NamedCONT` / `NamedLOGC` / `NamedREPO` / `NamedMIDE` / `NamedCRON`。

**并发模型。** HTTP 请求在 Gin 协程内同步编排；WebSocket Hub 自管读写与心跳；RepoWiki 分析在流水线内跑，超时由 Cron 扫 `RetryStaleTask`。不要在 Handler 里 `go func` 做业务。

**前端嵌入。** 生产路径是 `pnpm build` 写出 `resources/web/dist` 与 `resources/web-wiki/dist`，再 `go build`。开发时 Vite 独立端口（控制台 3000，Wiki Reader 3001），不改上述边界。
