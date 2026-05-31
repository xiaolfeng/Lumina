# 基础设施层

## 概述

Lumina 的三大模块共享统一的基础设施层，通过启动阶段注入 context，业务层通过 `xCtxUtil.MustGetDB/MustGetRDB` 获取依赖。

## PostgreSQL

### 用途

| 模块 | 存储内容 |
|------|----------|
| RepoWiki | 项目元数据（名称、Git URL、commit hash、版本记录、状态）、Wiki 文档关联 |
| Memory | 记忆卡片完整数据（标题、内容、元数据）、标签表、多对多关联 |
| Q&A | Session 元数据、问题完整数据（含选项、面板、答案）、会话-问题关联 |

### 约定

- 表前缀通过 `DATABASE_PREFIX` 环境变量配置
- 实体采用雪花算法基因策略，必须实现 `GetGene() xSnowflake.Gene`
- 所有字段追加行尾中文注释（`// 字段说明`），与 `gorm comment` 保持一致
- 新增实体必须追加到 `migrateTables` 中

## Redis

### 用途

| 模块 | 用途 |
|------|------|
| RepoWiki | 可选：分析任务状态缓存 |
| Memory | 可选：热点记忆缓存（频繁查询的 active 状态卡片）、标签索引缓存 |
| Q&A | Session 状态缓存（快速验证 session 是否有效）、超时倒计时 |

### 约定

- Key 前缀通过 `NOSQL_PREFIX` 环境变量配置
- 连接池大小通过 `NOSQL_POOL_SIZE` 配置，默认 100

## LLM Provider

### 概述

RepoWiki 模块需要调用 LLM 进行代码分析，使用独立的 Provider 配置。

### 配置项（参考值）

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `LLM_PROVIDER` | Provider 类型（openai / anthropic / custom） | `openai` |
| `LLM_API_KEY` | API 密钥 | 无（必填） |
| `LLM_MODEL` | 模型名称 | `gpt-4o` |
| `LLM_BASE_URL` | 自定义端点（兼容 OpenAI 协议即可） | 空（使用官方） |
| `LLM_MAX_TOKENS` | 单次请求最大 token 数 | `4096` |
| `LLM_TEMPERATURE` | 生成温度 | `0.3` |

### 设计说明

- LLM Provider 通过 Agent SDK 对接，Lumina 本身不直接管理模型调用
- 配置通过环境变量注入，使用 `xEnv.GetEnv*` 读取并提供默认值
- 未来可扩展支持多 Provider 轮询、降级策略

## 环境变量汇总

### 通用配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `XLF_DEBUG` | 调试模式（启用 Swagger UI） | `true` |
| `XLF_HOST` | 服务监听地址 | `0.0.0.0` |
| `XLF_PORT` | 服务端口 | `8080` |
| `APP_NAME` | 应用名称 | `Lumina` |
| `APP_VERSION` | 应用版本 | `v0.1.0` |

### 数据库配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DATABASE_HOST` | PostgreSQL 主机 | `localhost` |
| `DATABASE_PORT` | PostgreSQL 端口 | `5432` |
| `DATABASE_USER` | PostgreSQL 用户名 | `bamboo_user` |
| `DATABASE_PASS` | PostgreSQL 密码 | `bamboo_pass` |
| `DATABASE_NAME` | PostgreSQL 数据库名 | `lumina` |
| `DATABASE_PREFIX` | 表前缀 | `lum_` |
| `DATABASE_TIMEZONE` | 数据库时区 | `Asia/Shanghai` |

### Redis 配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `NOSQL_HOST` | Redis 主机 | `localhost` |
| `NOSQL_PORT` | Redis 端口 | `6379` |
| `NOSQL_PASS` | Redis 密码 | 空 |
| `NOSQL_DATABASE` | Redis 数据库索引 | `1` |
| `NOSQL_PREFIX` | Key 前缀 | `lum:` |
| `NOSQL_POOL_SIZE` | 连接池大小 | `100` |

### Q&A 配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `QA_SESSION_MAX_DURATION` | Session 最大存活时间（秒） | `604800`（7 天） |

### LLM 配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `LLM_PROVIDER` | Provider 类型 | `openai` |
| `LLM_API_KEY` | API 密钥 | 无（必填） |
| `LLM_MODEL` | 模型名称 | `gpt-4o` |
| `LLM_BASE_URL` | 自定义端点 | 空 |
| `LLM_MAX_TOKENS` | 最大 token 数 | `4096` |
| `LLM_TEMPERATURE` | 生成温度 | `0.3` |

### 雪花算法配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SNOWFLAKE_DATACENTER_ID` | 数据中心 ID | `1` |
| `SNOWFLAKE_NODE_ID` | 节点 ID | `1` |
