# PROJECT KNOWLEDGE BASE

**生成日期:** 2026-05-31
**提交:** e362301
**分支:** master

## OVERVIEW
`bamboo-base-go-template` 是基于 `bamboo-base-go` 构建的后端脚手架模板。
通过启动节点（数据库、Redis、种子数据）串联 Gin 路由与分层业务模块（`handler -> logic -> repository`）。

## STRUCTURE
```text
./
├── main.go                     # 入口；将生命周期委托给 xMain.Runner
├── go.mod                      # Go 1.25.0；依赖 bamboo-base-go 模块
├── Makefile                    # 开发/测试/格式化命令
├── .env.example                # 必需的环境变量模板
├── api/                        # 请求/响应 DTO（按业务域分包）
├── docs/                       # Swagger 规范（自动生成；请勿手动编辑）
├── internal/
│   ├── app/
│   │   ├── route/              # 路由注册与中间件绑定
│   │   │   ├── route.go        # 全局中间件 + 路由组入口
│   │   │   ├── route_health.go # 健康检查路由注册
│   │   │   └── route_swagger.go# Swagger UI 注册（仅调试模式）
│   │   └── startup/            # 基础设施初始化与启动节点注册
│   │       ├── startup.go      # 启动节点列表工厂
│   │       ├── startup_database.go   # GORM + PostgreSQL + 自动迁移
│   │       ├── startup_redis.go      # go-redis 初始化
│   │       ├── startup_prepare.go    # 种子数据编排
│   │       └── prepare/        # 幂等种子数据（角色等）
│   ├── handler/                # HTTP 处理器（薄控制器层）
│   ├── logic/                  # 业务编排层
│   ├── repository/             # 数据库/Redis 访问层
│   ├── entity/                 # GORM 实体（需实现 GetGene() 绑定）
│   └── constant/               # 共享业务常量（基因编号等）
└── .agent/skills/              # 项目专属技能（swagger-writer、entity-build、project-style）
```

## WHERE TO LOOK
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

## CODE MAP
本仓库无 LSP 工作区视图；代码映射通过扫描构建。

| 符号 | 类型 | 位置 | 作用 |
|---|---|---|---|
| `main` | 函数 | `main.go` | 注册启动节点并运行应用 |
| `Init` | 函数 | `internal/app/startup/startup.go` | 启动节点列表工厂 |
| `NewRoute` | 函数 | `internal/app/route/route.go` | 全局中间件 + 路由组 |
| `NewHandler[T]` | 泛型函数 | `internal/handler/handler.go` | Handler 构造模式 |
| `HealthLogic.Ping` | 方法 | `internal/logic/health.go` | 服务健康检查编排 |
| `HealthRepo.DatabaseReady` | 方法 | `internal/repository/health.go` | 数据库就绪检查 |

## CONVENTIONS
- **导入别名**：bamboo-base-go 包使用 `x*` 别名（`xLog`、`xEnv`、`xError`、`xResult`、`xReg` 等）。
- **严格分层**：route -> handler -> logic -> repository；禁止跳层调用。
- **上下文依赖注入**：启动阶段将基础设施注册到 context；逻辑层通过 `xCtxUtil.MustGetDB/MustGetRDB` 获取。
- **响应模式**：handler 通过 `xResult.SuccessHasData` 返回成功；错误通过 `ctx.Error` 传递。
- **错误类型**：仓库/逻辑层使用 `*xError.Error` 表示业务/基础设施故障。
- **环境变量族**：`XLF_*`、`APP_*`、`DATABASE_*`、`NOSQL_*`、`SNOWFLAKE_*`。
- **实体 ID 策略**：采用雪花算法基因策略；每个实体必须实现 `GetGene() xSnowflake.Gene`。
- **字段注释**：实体字段必须追加行尾中文注释（`// 字段说明`），且与 `gorm comment` 保持一致。
- **Swagger 注册**：仅在 `XLF_DEBUG=true` 时注册 Swagger UI。

## ANTI-PATTERNS (THIS PROJECT)
- 禁止从路由直接调用仓库或绕过 logic 层。
- 禁止直接使用 `os.Getenv`；应使用带默认值的 `xEnv.GetEnv*`。
- 禁止在 handler 中手写原始 Gin JSON 响应；优先使用 `xResult` 辅助函数。
- 禁止在 logic/repository 构造函数内部创建 DB/Redis 客户端；应从启动阶段/context 获取注入的依赖。
- 禁止新增实体后不将其追加到 `migrateTables`。
- 禁止手动编辑 `docs/` 下的文件；它们由 `swag init` 自动生成。

## UNIQUE STYLES
- **日志命名**：遵循模块标签（`NamedMAIN`、`NamedINIT`、`NamedCONT`、`NamedLOGC`、`NamedREPO`）。
- **启动种子阶段**：显式通过 `xCtx.Exec` 节点执行，并隔离在 `prepare/` 目录中。
- **实体 ID 基因策略**：实体级别需绑定基因类型。
- **项目技能**：`.agent/skills/` 包含项目专属技能：`swagger-writer`、`entity-build`、`project-style`。

## COMMANDS
```bash
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
```

## NOTES
- 尚未配置 CI 工作流（`.github/workflows` 不存在）。
- `make test` 命令存在，但测试文件尚未脚手架化。
- 保持 AGENTS 层级简洁；领域细节应放入 `internal/AGENTS.md` 及更深作用域的文件中。
- `docs/` 由 `swag init -g main.go --parseDependency` 完全自动生成；切勿提交手动编辑。
- `.env` 与 `.env.*` 已被 gitignore；本地开发时从 `.env.example` 复制。
- `Role` 实体使用 `Name`（字符串主键）而非雪花 ID；其他实体使用 `xModels.BaseEntity`（雪花 ID）。
- 雪花算法数据中心/节点 ID 默认为 1/1；可通过 `SNOWFLAKE_DATACENTER_ID` 和 `SNOWFLAKE_NODE_ID` 覆盖。
