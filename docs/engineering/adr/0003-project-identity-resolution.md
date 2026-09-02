> 状态：proposed · 依据当前实现补录

## 背景
Pin、Q&A、RepoWiki 和工具客户端都需要把名称或本地路径解析为稳定项目 ID。旧 Wiki 把 `alias_name` 描述为别名数组，当前实现已改成单值别名加路径列表，需要冻结真实身份模型和匹配顺序。

## 决定
1. **Project 是项目身份的唯一记录。** `Name` 全局唯一，`AliasName` 是单个可选别名，`MatchPath` 是路径数组，`Description` 只用于展示；禁止业务模块各自维护第二套项目名称映射；
2. **Project MCP 提供 `project_create`、`project_get`、`project_list`。** `project_get` 按 `project_id → name → match_path` 选择首个非空条件；调用方已有路径时优先使用 `match_path` 解析现有项目；
3. **路径匹配使用项目路径作为查询路径的前缀。** `/workspace/Lumina` 可以匹配 `/workspace/Lumina/internal/logic`；路径必须保留为绝对路径，禁止仅按目录 basename 推断项目；
4. **Pin 的名称解析只接受雪花 ID 或大小写不敏感的完整别名。** `AliasName` 不做旧 Wiki 所述的包含式模糊匹配；需要稳定自动识别时使用 Project 的 `MatchPath`；
5. **PostgreSQL 是 Project 的事实来源，Redis 采用 30 分钟 Cache-Aside。** 缓存维度为 ID、Name、Alias 和每个 MatchPath；创建与更新写入缓存，删除清理所有关联键，缓存缺失时回源数据库；
6. **Project 可以被业务域关联，但不承担跨域业务编排。** 项目删除、关联数据约束和展示逻辑分别留在各自层内，禁止把 Q&A、Pin 或 RepoWiki 流程塞进 ProjectLogic。

## 后果
- 得到：项目 ID、名称、别名和路径的解析规则只有一份，Agent 可用工作目录稳定恢复项目上下文；
- 失去：一个项目不能配置多个自然语言别名；新增该能力需要修改实体、缓存键和歧义处理，而不是把分隔字符串塞进 `AliasName`；
- 违反时暴露方式：同一路径会解析到错误项目，或数据库更新后缓存仍返回旧 Name、Alias、MatchPath。

## 否决项
- **`alias_name` 使用 JSON 数组并做双向包含匹配**：短字符串容易命中多个项目，且与当前单值字段和精确查询不一致；
- **仅按项目名称解析**：Agent 常拿到的是文件路径，重命名项目后名称也不再稳定；
- **把 Redis 当项目事实来源**：TTL 和淘汰会使项目身份消失，缓存只能加速读取；
- **由 PinLogic 调用 ProjectLogic 完成业务编排**：会造成模块 Logic 互相依赖；Pin 只复用项目持久化层的解析能力。
