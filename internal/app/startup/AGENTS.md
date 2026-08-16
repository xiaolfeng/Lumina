# STARTUP 启动模块知识库

## 概述
`internal/app/startup/` 负责应用启动时的**业务节点**初始化（RepoWiki Logic、MCP Server）与幂等种子数据填充，并提供 WebSocket Hub 与 RepoWiki 定时重试的 Runner 工厂。数据库与缓存的初始化已移至 `main.go` 的 `xOption.WithDatabase / WithCache` 声明式装配，不再由本模块负责。

## 目录结构
```text
startup/
├── startup.go              # 启动节点列表工厂（仅注册 3 个业务节点）
├── startup_repowiki.go     # RepoWiki Logic 初始化（存储目录 + Logic 构造 + LlmResolver 注入 context）
├── startup_mcp.go          # MCP Server 初始化（创建 QA/Project/Pin/Preview/RepoWiki Logic 并注入）
├── startup_prepare.go      # 种子数据编排入口（调用 prepare.New().Prepare()）
├── startup_websocket.go    # WebSocket Runner 工厂（NewWebSocketRunner → hub.Run(ctx)）
├── startup_cron.go         # Cron Runner 工厂（RepoWiki 超时任务重试，@every 5m）
└── prepare/                # 幂等种子数据
    ├── prepare.go          # Prepare 编排器（prepareInfo → Project → QaHash → Llm → RepoWiki → Settings）
    ├── prepare_info.go     # Info 表种子数据（站点/Q&A/安全/认证配置键）
    ├── prepare_project.go  # 项目缓存清理（字段类型变更时清除旧缓存）
    ├── prepare_qa_hash.go  # Q&A 会话 Hash 缓存修复（历史数据格式迁移）
    ├── prepare_llm.go      # LLM Provider/Model 种子数据（默认 Provider 配置）
    ├── prepare_repowiki.go # RepoWiki 存储目录、默认配置种子 + Webhook 凭证回填 + 版本 ID 迁移
    └── prepare_settings.go # 系统设置种子数据（安全/Q&A/RepoWiki 配置项）
```

## 导航指南
| 任务 | 文件 | 说明 |
|---|---|---|
| 注册新启动节点 | `startup.go` | 在 `Init()` 中追加 `xRegNode.RegNodeList` |
| 修改数据库初始化 | `main.go` | `xOption.WithDatabase(xOptDatabase.FromEnv() + WithAutoMigrate)` |
| 修改缓存初始化 | `main.go` | `xOption.WithCache(xOptCache.FromEnv())` |
| 修改 RepoWiki 初始化 | `startup_repowiki.go` | 存储目录创建、Logic 构造与 LlmResolver 注入 |
| 修改 MCP 初始化 | `startup_mcp.go` | 注入 Logic、注册 MCP 工具 |
| 修改 WebSocket 启动 | `startup_websocket.go` | `NewWebSocketRunner()` 返回的函数由 `main.go` 传入 `xMain.Runner` |
| 修改 Cron 定时任务 | `startup_cron.go` | `NewCronRunner()` 返回的函数由 `main.go` 传入 `xMain.Runner` |
| 新增种子数据 | `prepare/` | 创建 `prepare_<domain>.go`，在 `prepare.go` 的 `Prepare()` 中调用 |
| 新增可迁移实体 | `main.go` | 追加到 `WithAutoMigrate(...)` 列表，注意 FK 依赖顺序 |

## 约定
- **启动函数签名**：`(ctx context.Context) (any, error)`，由 `xRegNode` 框架调用。
- **日志标签**：统一使用 `xLog.NamedINIT`；Cron Runner 使用 `xLog.NamedCRON`；WebSocket Runner 使用 `xLog.NamedCONT`。
- **数据库/缓存装配**：由 `main.go` 的 `xOption.WithDatabase / WithCache` 声明式装配，框架在业务节点之前注册 `DatabaseKey / CacheManagerKey / RedisClientKey`；`startup.go` 仅注册依赖 db/rdb 的业务节点。
- **种子数据幂等性**：`prepare` 中的方法必须可重复执行（推荐 `FirstOrCreate` + `Assign`）。
- **种子数据隔离**：每个业务域一个 `prepare_<domain>.go`，由 `prepare.go` 统一编排。
- **RepoWiki Logic 注入**：`startup_repowiki.go` 构造 `RepoWikiLogic` 并注册到 context 的 `RepoWikiLogicKey`，供 MCP/Handler/Cron 通过 `logic.GetRepoWikiLogicFromContext` 获取。
- **MCP 启动**：`startup_mcp.go` 创建 QA/Project/Pin/Preview/RepoWiki Logic 实例并注入 MCP 包，然后调用 `mcp.InitMCPServer` 生成 HTTP Handler，注册到 context 的 `MCPHandlerKey`。
- **Cron Runner**：`startup_cron.go` 的 `NewCronRunner()` 返回一个由 `main.go` 传入 `xMain.Runner` 的 goroutine 函数，内含 RepoWiki 超时任务重试（默认每 5 分钟，`RetryStaleTask`）。
- **WebSocket Runner**：`startup_websocket.go` 的 `NewWebSocketRunner()` 返回闭包，调用 `websocket.GetGlobalHub().Run(ctx)` 启动 Hub 主循环；Hub 单例在 `route_ws.go` 的 Register 阶段已创建。
- **项目缓存清理**：`prepare_project.go` 在启动时扫描并清除旧格式的项目缓存键，确保字段类型变更后缓存一致性。
- **QA Hash 修复**：`prepare_qa_hash.go` 修复历史会话 Hash 缓存格式，仅在升级时需要。
- **LLM 种子**：`prepare_llm.go` 写入默认 LLM Provider/Model 配置，仅在首次部署时生效（幂等）。

## 反模式
- 禁止在启动节点之外初始化 DB/Redis/MCP/RepoWiki Logic 客户端。
- 禁止从 `main.go` 或路由/处理器直接调用种子逻辑。
- 禁止随意重排 `WithAutoMigrate` 中的实体顺序；必须遵循 FK 依赖。
- 禁止在启动代码中使用无默认值的环境变量读取。
- 禁止在 Cron/WebSocket Runner 之外通过裸 goroutine 调度定时任务；统一走 `xCronRunner` / `xMain.Runner`。

## 调试路径
1. 数据库连接失败 → 检查 `main.go` 的 `WithDatabase(FromEnv())` 与 `DATABASE_*` 环境变量。
2. 迁移失败 → 检查 `main.go` 的 `WithAutoMigrate` 顺序和实体定义（`GetGene()` 是否实现）。
3. 缓存连接失败 → 检查 `main.go` 的 `WithCache(FromEnv())` 与 `NOSQL_*` 环境变量。
4. RepoWiki Logic 缺失 → 检查 `startup_repowiki.go` 是否注册 `RepoWikiLogicKey`，存储目录权限是否正确。
5. MCP 路由缺失 → 检查 `startup_mcp.go` 是否正确注册 `MCPHandlerKey` 及 `SetPreviewLogic` / `SetRepoWikiLogic` 调用。
6. 种子数据异常 → 检查 `prepare/` 下对应文件，确认幂等逻辑。
7. RepoWiki 定时重试未执行 → 检查 `startup_cron.go` 是否被 `main.go` 传入 `xMain.Runner`，Cron 日志（`NamedCRON`）是否有 panic recover 记录。
8. WebSocket 连接异常 → 检查 `startup_websocket.go` 是否传入 `xMain.Runner`，`websocket.GetGlobalHub()` 是否返回 nil。

## 执行顺序（不可更改）
1. `repoWikiInit`（注册为 `bConst.RepoWikiLogicKey`）
2. `mcpInit`（注册为 `MCPHandlerKey`）
3. `businessDataPrepare`（注册为 `xCtx.Exec`）

数据库/缓存由 `main.go` 的 `xOption.WithDatabase / WithCache` 在框架层装配，先于业务节点执行。`repoWikiInit`、`mcpInit` 和 `businessDataPrepare` 依赖 context 中的 DB/Redis 实例（`xCtxUtil.MustGetDB/MustGetRDB`）。WebSocket Runner 与 Cron Runner 由 `main.go` 在 `xMain.Runner` 启动后异步执行，不阻塞启动节点链。

## 迁移顺序
实体迁移由 `main.go` 的 `xOption.WithDatabase(WithAutoMigrate(...))` 声明式配置，当前顺序：
`Info` → `Apikey` → `Project` → `Pin` → `QaSession` → `QaQuestion` → `QaSupplement` → `BiometricCredential` → `SshKey` → `RepoWikiConfig` → `WikiVersion` → `LlmProvider` → `LlmModel` → `WebhookEvent` → `PreviewSession` → `PreviewFile`。

新增实体时根据 FK 依赖关系追加到正确位置。
