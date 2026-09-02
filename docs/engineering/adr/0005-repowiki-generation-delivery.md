> 状态：proposed · 依据当前实现补录

## 背景
RepoWiki 已从旧版串行 Pass 和三角色 ReAct 方案演进为五角色预定义流水线，并完成版本隔离、Webhook 更新、定时重试和只读 MCP。旧 Wiki 同时保留多代路径、角色名和 Markdown 产物描述，需要按当前实现冻结生成与发布边界。

## 决定
1. **一份 `RepoWikiConfig` 以 `ProjectID` 唯一关联一个 Project。** 配置保存 Git 地址、默认分支、语言、SSH Key、访问密码、项目级提示词、Webhook 凭证与监听分支；当前对外版本由 `SelectedVersionID` 指向；
2. **每次分析创建独立 `WikiVersion`。** 版本记录 commit、分支、语言、模型、状态、阶段、消耗、重试次数和各产物路径；版本既是生成任务记录，也是发布与回滚单位；
3. **生成顺序固定为 Coordinator → Explore → Architect → Writer → Validator。** Coordinator 生成概要；Explore 按 scope 并发阅读；Architect 生成目录结构；Writer 分批并发写页；Validator 校验完整性并触发有限补写。业务代码不得恢复由 Coordinator 自主决定流程的 ReAct 编排；
4. **五个角色分别解析数据库中的模型分配。** 角色名统一使用 `repowiki.coordinator`、`repowiki.explore`、`repowiki.architect`、`repowiki.write`、`repowiki.validator`；系统提示词只从 `resources/prompts/*.md` 内嵌资源加载，项目级 `CustomPrompt` 只能追加约束；
5. **产物按 `versions/{versionID}/` 隔离。** 概要、扫描结果、探索产物、架构输出、manifest、会话和最终 Wiki 都属于该版本；最终页面使用 `.mdx` 与 frontmatter，禁止恢复旧版 config 级目录或 `.md` 页面契约；
6. **版本完成后更新 `SelectedVersionID`。** Wiki Reader 与查询逻辑读取已选中的完成版本；版本清理必须保留当前选中版本，并在指针失效时回退到最近完成版本；
7. **RepoWiki MCP 只暴露 `repoWiki_list` 与 `repoWiki_query`。** MCP 只能列出已完成版本并读取页面；创建配置、首次分析、版本管理和删除走受认证的 REST API；
8. **Git push 通过 Webhook 触发更新。** 接收端点先按 URL token 找配置，再检测 Provider、验证 HMAC 或静态 token、解析 push、校验仓库 URL、监听分支和重复 commit；有效事件创建版本，处理结果记录到 `WebhookEvent`；
9. **RepoWiki 超时任务由 Cron Runner 定期重试。** 定时任务从 context 取得同一 RepoWikiLogic，禁止另建不受应用生命周期管理的后台调度；
10. **LLM Provider API Key 与 SSH 私钥必须加密存储。** 两者使用 AES-256-GCM，运行时通过配置解析服务取得明文；日志、响应和 Wiki 产物不得包含密钥。

## 后果
- 得到：每个版本可独立生成、校验、选择和清理；Agent 读取知识时不会意外触发昂贵分析；不同角色可使用不同模型；
- 失去：MCP 客户端不能主动创建或更新 Wiki；一次完整生成受五阶段顺序约束；版本隔离会增加磁盘占用；
- 违反时暴露方式：绕过版本目录会覆盖已发布 Wiki；恢复写型 MCP 工具会破坏触发权边界；角色名或提示词路径漂移会导致模型或 prompt 解析失败；未验证 Webhook 会允许伪造更新。

## 否决项
- **MCP 暴露 analyze、update、delete**：Agent 是 Wiki 消费者，更新触发权属于管理端与 Git push；
- **Coordinator 通过 ReAct 自主选择阶段**：执行顺序、并发量和 token 消耗不可预测；
- **所有角色共用一个模型键**：无法按阅读、结构化输出、写作和校验职责配置模型；
- **不同版本共写 `wiki/{configID}/`**：新版本会覆盖旧版本，无法回滚或保留失败现场；
- **在 Logic 中硬编码系统 prompt**：提示词无法作为内嵌资源统一审查与测试；
- **Webhook 跳过签名或分支校验**：外部请求可未经授权触发分析并消耗模型额度。
