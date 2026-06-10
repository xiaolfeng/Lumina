# STARTUP 启动模块知识库

## 概述
`internal/app/startup/` 负责应用启动时的基础设施初始化（数据库、Redis）和幂等种子数据填充。

## 目录结构
```text
startup/
├── startup.go              # 启动节点列表工厂（注册顺序）
├── startup_database.go     # GORM + PostgreSQL 初始化 + AutoMigrate
├── startup_redis.go        # go-redis 初始化
├── startup_prepare.go      # 种子数据编排入口
└── prepare/                # 幂等种子数据
    ├── prepare.go          # Prepare 编排器（聚合所有 prepare 方法）
    └── prepare_info.go     # Info 表种子数据（站点配置）
```

## 导航指南
| 任务 | 文件 | 说明 |
|---|---|---|
| 注册新启动节点 | `startup.go` | 在 `Init()` 中追加 `xRegNode.RegNodeList` |
| 修改数据库初始化 | `startup_database.go` | DSN 构建、GORM 配置、表前缀 |
| 修改 Redis 初始化 | `startup_redis.go` | 连接配置、池大小 |
| 新增种子数据 | `prepare/` | 创建 `prepare_<domain>.go`，在 `prepare.go` 的 `Prepare()` 中调用 |
| 新增可迁移实体 | `startup_database.go` | 追加到 `migrateTables` 切片，注意 FK 依赖顺序 |

## 约定
- **启动函数签名**：`(ctx context.Context) (any, error)`，由 `xRegNode` 框架调用。
- **日志标签**：统一使用 `xLog.NamedINIT`。
- **数据库表前缀**：通过 `DATABASE_PREFIX` 环境变量驱动（默认 `tpl_`），使用 `SingularTable: true`。
- **种子数据幂等性**：`prepare` 中的方法必须可重复执行（推荐 `FirstOrCreate` + `Assign`）。
- **种子数据隔离**：每个业务域一个 `prepare_<domain>.go`，由 `prepare.go` 统一编排。

## 反模式
- 禁止在启动节点之外初始化 DB/Redis 客户端。
- 禁止从 `main.go` 或路由/处理器直接调用种子逻辑。
- 禁止随意重排 `migrateTables` 中的实体顺序；必须遵循 FK 依赖。
- 禁止在启动代码中使用无默认值的环境变量读取。

## 调试路径
1. 数据库连接失败 → 检查 `startup_database.go` 的 DSN 构建（`DATABASE_*` 环境变量）。
2. 迁移失败 → 检查 `migrateTables` 顺序和实体定义（`GetGene()` 是否实现）。
3. Redis 连接失败 → 检查 `startup_redis.go` 的 `NOSQL_*` 环境变量。
4. 种子数据异常 → 检查 `prepare/` 下对应文件，确认幂等逻辑。

## 执行顺序（不可更改）
1. `databaseInit`（注册为 `xCtx.DatabaseKey`）
2. `nosqlInit`（注册为 `xCtx.RedisClientKey`）
3. `businessDataPrepare`（注册为 `xCtx.Exec`）

`prepare` 依赖 context 中的 DB 实例（`xCtxUtil.MustGetDB`），因此 DB 节点必须先执行。

## 迁移顺序
`migrateTables` 当前：`Info` → `Apikey` → `Project`。

新增实体时根据 FK 依赖关系追加到正确位置。
