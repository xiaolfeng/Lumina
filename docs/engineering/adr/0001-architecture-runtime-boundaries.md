> 状态：proposed · 依据当前实现补录
>
> 相关：[ADR-0005](./0005-workspace-identity.md) 把 Workspace 定为 Project 旁的身份层，不是第六业务域。本文件第 2 条的五域互不编排仍然有效。

## 背景
旧 Wiki 把设计阶段的四模块、SSE 和环境变量式 LLM 配置写在同一份架构说明中。当前代码已经采用五个独立业务域、WebSocket、数据库热配置和声明式基础设施装配，需要按实现冻结边界，防止后续改动回到旧结构。

## 决定
1. **HTTP 请求遵循 `route → handler → logic → repository`。** Route 注册路径和中间件；Handler 绑定请求并映射响应；Logic 承担校验、状态流转与业务编排；Repository 只处理数据库和缓存。跨模块复用的基础能力放在 `internal/service/`，禁止 Handler 直接访问 Repository；
2. **RepoWiki、Memory、Q&A、Pin、Preview 是互不直接编排的业务域。** 已实现域各自暴露能力；Memory 在 [0001 RFC](../rfc/0001-memory-decision-memory.md) 接受并实现前不注册占位接口。跨域流程由 Agent 通过 MCP 组合；
3. **REST API 服务前端，Streamable HTTP MCP 服务 Agent。** Q&A 与 Preview 的实时状态使用 WebSocket；禁止恢复旧 Wiki 中的 SSE 方案；
4. **数据库和缓存由 `main.go` 使用 `WithDatabase`、`WithCache` 声明式装配。** 需要数据库或缓存的业务节点从 context 获取依赖；迁移实体统一登记在 `WithAutoMigrate`；
5. **WebSocket Hub 与 RepoWiki 定时重试作为 Runner 启动。** 它们接收可取消 context 并独立于启动节点链运行；禁止用裸 goroutine 替代 Runner 生命周期；
6. **控制台与 Wiki Reader 的构建产物通过 `go:embed` 进入同一 Go 二进制。** 两个前端继续共享 `@lumina/components`；生产构建必须先产出前端资源再编译后端；
7. **LLM Provider、模型和 RepoWiki 角色分配存入数据库并支持热配置。** API Key 使用 AES-256-GCM 加密；禁止恢复为 `LLM_PROVIDER`、`LLM_API_KEY`、`LLM_MODEL` 这组运行时业务配置；
8. **MCP、Webhook 和 OAuth 公共端点在全局响应中间件之前注册。** MCP 使用 OAuth 2.1 优先、API Key 回退的认证中间件，保证协议发现与裸响应不被通用包装破坏。

## 后果
- 得到：接入通道、业务编排、持久化和运行时生命周期各有唯一落点；双前端可随单二进制交付；
- 失去：业务域不能通过内部 Logic 直接拼装快捷流程，跨域需求必须放到 Agent 或明确新增的接入层；
- 违反时暴露方式：跨层调用会出现在导入关系中；错误路由顺序会改变 MCP、Webhook 或 OAuth 响应；前端未先构建会导致嵌入资源缺失；绕过 Runner 会留下无法随 context 停止的后台任务。

## 否决项
- **Q&A 使用 SSE**：当前回答、补充、恢复和多设备状态需要双向消息，WebSocket 已承担该职责；
- **启动节点内直接初始化数据库和 Redis**：框架声明式装配已经提供依赖与迁移顺序，再保留手工初始化会出现双重所有权；
- **每个前端复制一套 UI、Markdown 和主题代码**：两端会产生视觉和渲染差异，公共实现必须留在 `@lumina/components`；
- **LLM 密钥以环境变量明文作为业务配置**：无法热更新，也绕过现有加密存储；环境变量只保留加密密钥等部署级秘密；
- **MCP 只挂 API Key 认证**：OAuth 客户端无法获得资源元数据和授权入口。
