# INTERNAL 业务层知识库

## 概述
`internal/` 实现了 Lumina 的业务运行时管道：route -> middleware -> handler -> logic -> repository -> entity，严格分层，禁止跨层调用。

## 目录结构
```text
internal/
├── app/
│   ├── middleware/        # Gin 中间件（认证拦截等）
│   │   └── auth.go       # Bearer Token 验证 → 注入用户到 context
│   ├── route/             # 路由注册与中间件绑定
│   │   ├── route.go       # 全局中间件 + 路由组入口 + 前端 SPA 集成
│   │   ├── route_auth.go  # 认证路由（公开 + 受保护）
│   │   ├── route_apikey.go# API Key 路由（CRUD + 重置，受 Auth 保护）
│   │   ├── route_project.go # 项目路由（CRUD，受 Auth 保护）
│   │   ├── route_health.go# 健康检查路由
│   │   ├── route_frontend.go # 前端 SPA 静态资源 + fallback
│   │   └── route_swagger.go # Swagger UI（仅 XLF_DEBUG=true）
│   └── startup/           # 基础设施初始化与种子数据
│       ├── startup.go     # 启动节点列表工厂
│       ├── startup_database.go  # GORM + PostgreSQL + AutoMigrate
│       ├── startup_redis.go     # go-redis 初始化
│       ├── startup_prepare.go   # 种子数据编排
│       └── prepare/       # 幂等种子数据
├── handler/               # HTTP 处理器（薄控制器层）
│   ├── handler.go         # NewHandler[T] 泛型构造器 + service 注入
│   ├── auth.go            # 认证处理器（登录、刷新、初始化）
│   ├── apikey.go          # API Key 处理器（CRUD + 重置 + 分页）
│   ├── project.go         # 项目处理器（CRUD + 分页）
│   └── health.go          # 健康检查处理器
├── logic/                 # 业务编排层
│   ├── logic.go           # logic 基础结构（db/rdb/log）
│   ├── auth.go            # 认证逻辑（Token 验证、密码校验）
│   ├── apikey.go          # API Key 逻辑（密钥生成/哈希/脱敏/CRUD）
│   ├── project.go         # 项目逻辑（CRUD、名称唯一校验、别名解析）
│   └── health.go          # 健康检查逻辑
├── repository/            # 数据访问层
│   ├── token.go           # Token 持久化
│   ├── apikey.go          # API Key 持久化（CRUD + 分页）
│   ├── project.go         # 项目持久化（CRUD + 分页 + Redis 缓存 + 别名查询）
│   ├── health.go          # 数据库就绪检查
│   └── cache/             # Redis 缓存操作
│       └── token.go       # Token 缓存（AT/RT 存储）
├── entity/                # GORM 实体
│   ├── info.go            # 站点配置实体（单用户模式）
│   ├── apikey.go          # API Key 实体（密钥哈希/前缀/后缀/过期时间）
│   └── project.go         # 项目实体（名称/别名/描述）
└── constant/              # 共享业务常量
    ├── cache.go           # Redis Key 前缀/过期时间（带环境前缀格式化）
    ├── context.go         # Context Key（如 CtxOwnerKey）
    └── gene_number.go     # 雪花算法基因编号（GeneProject = 32）
```

## 导航指南
| 任务 | 文件/目录 | 说明 |
|---|---|---|
| 新增路由组 | `app/route/route.go` + `route_*.go` | 在 `NewRoute` 中调用路由注册函数 |
| 新增中间件 | `app/middleware/` | 返回 `gin.HandlerFunc`，在路由注册中绑定 |
| 新增处理器 | `handler/handler.go` 定义类型，`handler/*.go` 实现 | 通过 `NewHandler[T]` 构造 |
| 新增业务逻辑 | `logic/*.go` | Logic 通过 context 注入获取 db/rdb |
| 新增数据访问 | `repository/*.go` | 返回 `(data, *xError.Error)` |
| 新增 Redis 缓存 | `repository/cache/*.go` | Token 等缓存操作 |
| 新增实体 | `entity/*.go` + `startup/startup_database.go` | 实现并追加到 `migrateTables` |
| 新增基因编号 | `constant/gene_number.go` | 定义 `GeneXxx` 常量供实体 `GetGene()` 使用 |
| 新增种子数据 | `startup/prepare/` | 创建 `prepare_<domain>.go` |
| 新增请求/响应 DTO | `api/<domain>/` | 按业务域保持子包结构 |
| 新增业务常量 | `constant/*.go` | 基因编号、缓存 Key、Context Key |

## 约定
- **严格分层**：route → middleware → handler → logic → repository；禁止跳层调用。
- **Handler 精简**：仅做请求绑定、调用 logic、映射结果到响应；禁止在 handler 中直接操作 DB/Redis。
- **Logic 编排**：业务编排层，持久化和 SQL 归 repository 层。
- **Repository 返回值**：统一 `(data, *xError.Error)` 风格，不用裸 `error`。
- **日志命名**：按层使用 `xLog.WithName` — `NamedCONT`（handler）、`NamedLOGC`（logic）、`NamedREPO`（repository）、`NamedINIT`（startup）。
- **上下文传递**：使用 `ctx.Request.Context()` 或注入的 context 下发调用。
- **认证中间件**：通过 `middleware.Auth(ctx)` 创建，注入认证标记到 context（`CtxOwnerKey`）。
- **泛型 Handler 构造**：`NewHandler[T]` 统一注入所有 logic 实例到 `service` 结构体。
- **实体 ID 策略**：雪花算法基因策略；每个实体必须实现 `GetGene() xSnowflake.Gene`。
- **字段注释**：实体字段必须追加行尾中文注释（`// 字段说明`），且与 `gorm comment` 一致。
- **缓存键前缀**：通过 `xEnv.NoSqlPrefix` 环境变量自动拼接前缀，使用 `RedisKey.Get(args...)` 格式化。
- **分页规范**：使用 `xModels.PageRequest.Normalize()` 规范化分页参数，`xModels.NewPageFromRequest` 构建分页响应。
- **API Key 安全**：密钥使用 bcrypt 哈希存储，仅创建/重置时返回完整密钥；查询和列表使用 `maskKey` 脱敏。
- **Project 缓存策略**：采用 Cache-Aside 模式（ID→详情、Name→ID、Alias→ID 三层映射，TTL 30 分钟）。

## 反模式
- 禁止从路由直接调用 repository 或绕过 logic 层。
- 禁止在 handler 中手写原始 Gin JSON 响应；使用 `xResult` 辅助函数。
- 禁止在 logic/repository 构造函数内部创建 DB/Redis 客户端；应从 context 获取注入的依赖。
- 禁止绕过 `NewHandler[T]` 模式手动构造 handler。
- 禁止将业务常量写在 handler/logic 文件中；统一放 `constant/`。
- 禁止核心业务模块（RepoWiki、Memory、Q&A、Pin）之间直接调用。
- 禁止直接使用 `os.Getenv`；应使用带默认值的 `xEnv.GetEnv*`。
- 禁止新增实体后不追加到 `migrateTables`。
- 禁止在 repository 外部直接操作 Redis；缓存逻辑封装在 repository 层内部。

## 调试路径
1. 请求未路由 → 检查 `app/route/route.go` 路由组和 `route_*.go` 注册。
2. 路由正确但响应异常 → 检查 `handler/*.go` 绑定，然后 `logic/*.go` 编排。
3. 认证失败 → 检查 `middleware/auth.go` → `logic/auth.go` → `repository/cache/token.go`。
4. 数据库操作失败 → 检查 `repository/*.go` 和启动阶段的迁移状态。
5. Redis 缓存异常 → 检查 `repository/cache/*.go` 和 `startup_redis.go` 连接配置。
6. API Key 创建失败 → 检查 `logic/apikey.go` 密钥生成（`crypto/rand`）和 bcrypt 哈希。
7. 项目名称冲突 → 检查 `logic/project.go` 的 `GetByName` 唯一性校验逻辑。

## 引用
- [startup/](./app/startup/AGENTS.md) — 启动模块详细文档
