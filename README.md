# bamboo-base-go-template

基于 `bamboo-base-go` 的业务端脚手架模板，采用 `handler -> logic -> repository` 分层架构。

## 技术栈

- **Go 1.25.0**
- **Gin** - Web 框架
- **GORM** - ORM（PostgreSQL 驱动）
- **go-redis/v9** - Redis 客户端
- **swaggo** - Swagger/OpenAPI 文档生成
- **snowflake** - 分布式 ID 生成（基因策略）

## 前置条件

- Go 1.25.0+
- PostgreSQL 14+
- Redis 6+
- `swag` CLI（`go install github.com/swaggo/swag/cmd/swag@latest`）

## 快速开始

1. 复制环境变量模板

```bash
cp .env.example .env
```

2. 安装依赖

```bash
go mod tidy
```

3. 启动 PostgreSQL 和 Redis，确保 `.env` 中的连接配置正确。

4. 运行程序

```bash
make dev    # 生成 Swagger 文档并运行（推荐）
# 或
make swag   # 仅生成文档
make run    # 仅运行
```

5. 验证服务

```bash
curl http://localhost:8080/api/v1/health/ping
```

6. 访问 API 文档（仅在 `XLF_DEBUG=true` 时可用）

```
http://localhost:8080/swagger/index.html
```

## 环境变量

| 变量 | 说明 | 默认值 |
|---|---|---|
| `XLF_DEBUG` | 调试模式（启用 Swagger UI） | `true` |
| `XLF_HOST` | 服务监听地址 | `0.0.0.0` |
| `XLF_PORT` | 服务端口 | `8080` |
| `APP_NAME` | 应用名称 | `bamboo-base-go-template` |
| `APP_VERSION` | 应用版本 | `v0.1.0` |
| `DATABASE_HOST` | PostgreSQL 主机 | `localhost` |
| `DATABASE_PORT` | PostgreSQL 端口 | `5432` |
| `DATABASE_USER` | PostgreSQL 用户名 | `bamboo_user` |
| `DATABASE_PASS` | PostgreSQL 密码 | `bamboo_pass` |
| `DATABASE_NAME` | PostgreSQL 数据库名 | `bamboo_template` |
| `DATABASE_PREFIX` | 表前缀 | `tpl_` |
| `DATABASE_TIMEZONE` | 数据库时区 | `Asia/Shanghai` |
| `NOSQL_HOST` | Redis 主机 | `localhost` |
| `NOSQL_PORT` | Redis 端口 | `6379` |
| `NOSQL_PASS` | Redis 密码 | （空） |
| `NOSQL_DATABASE` | Redis 数据库索引 | `1` |
| `NOSQL_POOL_SIZE` | Redis 连接池大小 | `100` |
| `SNOWFLAKE_DATACENTER_ID` | 雪花算法数据中心 ID | `1` |
| `SNOWFLAKE_NODE_ID` | 雪花算法节点 ID | `1` |

## 目录结构

```text
.
├── api/                        # 请求/响应 DTO（按业务域分包）
├── docs/                       # Swagger 文档（自动生成，请勿手动修改）
├── internal/
│   ├── app/
│   │   ├── route/              # 路由注册与中间件绑定
│   │   └── startup/            # 基础设施初始化与启动节点
│   │       └── prepare/        # 默认种子数据（幂等）
│   ├── constant/               # 业务常量（如基因编号）
│   ├── entity/                 # GORM 实体（必须实现 GetGene()）
│   ├── handler/                # HTTP 处理器（薄控制器层）
│   ├── logic/                   # 业务编排层
│   └── repository/             # 数据访问层
├── main.go                     # 程序入口
├── Makefile                    # 常用命令
└── .env.example                # 环境变量模板
```

## 开发工作流

### 添加新接口

1. 在 `api/<domain>/` 下定义请求/响应 DTO
2. 在 `internal/entity/` 下定义数据实体（如需新表）
3. 在 `internal/repository/` 下实现数据访问
4. 在 `internal/logic/` 下实现业务编排
5. 在 `internal/handler/` 下实现 HTTP 处理器
6. 在 `internal/app/route/` 下注册路由
7. 运行 `make swag` 生成 Swagger 文档

### 添加新实体

1. 在 `internal/entity/` 下创建实体文件
2. 实现 `GetGene() xSnowflake.Gene` 方法
3. 在 `internal/constant/gene_number.go` 中定义基因常量（如需要）
4. 在 `internal/app/startup/startup_database.go` 的 `migrateTables` 中注册
5. 所有字段必须追加行尾中文注释（`// 字段说明`）

### 常用命令

```bash
make dev          # 生成 Swagger 文档并运行（推荐）
make swag         # 仅生成 Swagger 文档
make run          # 直接运行（不重新生成文档）
make tidy         # 整理 Go 模块
make fmt          # 格式化代码
make test         # 运行测试
```

## 架构说明

本项目采用严格的分层架构：

```text
HTTP Request
    -> Route (internal/app/route/)
    -> Handler (internal/handler/)
    -> Logic (internal/logic/)
    -> Repository (internal/repository/)
    -> DB / Redis
```

- **Handler** 只负责请求绑定、调用 Logic、映射响应
- **Logic** 负责业务编排、校验、事务边界
- **Repository** 只负责数据持久化和查询
- 禁止跨层调用（如 Handler 直接调用 Repository）

## 注意事项

- `docs/` 目录下的文件由 `swag init` 自动生成，请勿手动编辑
- 启动时会自动执行数据库迁移（`AutoMigrate`）和种子数据初始化
- 种子数据逻辑必须保证幂等性（可重复执行）
- 所有环境变量读取应使用 `xEnv.GetEnv*` 并提供默认值，禁止直接使用 `os.Getenv`

## 相关技能

项目根目录 `.agent/skills/` 下包含针对本项目的 OpenCode 技能：

- **swagger-writer** - 为 Handler 编写/补全 Swagger 注释
- **entity-build** - 根据描述生成符合规范的 Entity 代码
- **project-style** - 规范分层架构代码风格
