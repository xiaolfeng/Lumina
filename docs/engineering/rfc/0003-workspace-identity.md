> 状态：draft · 承接 [调研 0001](../research/0001-workspace-contents.md) · 工作方式见 [设计 0001](../design/0001-workspace-identity.md)

## 问题
Lumina 的身份边界是「一台单用户实例 + 一平 Project」。`Project.Name` 全局唯一，Q&A / Pin / RepoWiki / Preview 都挂 `project_id`，没有组织这一层。站主要把生活仓库和公司仓库拆开管时，路径前缀会串、Pin 能打进任意项目、控制台列表无法按场景切开。

仓库里「工作区」已经同时指 pnpm 包布局、Preview 沙盒和 Agent 的 `pwd`。再拿这个词顶租户层，后读代码的人会对不上。调研 [0001](../research/0001-workspace-contents.md) 和 Q&A `414082952174052352` 已把方向收成租户层；本提案写出建议采用什么。怎么接线写在设计 0001，不在这里重复 SQL 与 PR 切分。

## 方案
1. **Workspace 是贴在 Project 旁边的身份层，不是第六业务域。**
   结构是 `Workspace 1 — N Project`。Project 仍是一个 git / 代码身份（ADR-0002 继续有效）。`WorkspaceLogic` 禁止 import 或调用 `QaLogic`、`PinLogic`、`RepoWikiLogic`、`PreviewLogic`。跨域组合仍由 Agent 走 MCP。删除空间时批量改 `projects.workspace_id` 只碰 `ProjectRepo`，不经 ProjectLogic 做业务编排。

2. **每个 Project 必属一个 Workspace。**
   字段 `Project.workspace_id` 应用层禁止雪花零值。存量行在实例初始化或 prepare 回填到默认空间。v1 不允许把项目从 A 空间改挂到 B 空间；唯一搬家路径是删除非默认空间时整批移到默认空间。禁止给 Wiki 版本再加一层 `workspace_id`——隔离随 Project 走。

3. **`Project.Name` 保持实例内全局唯一。**
   不改成 `(workspace_id, name)`。两个空间不能注册同名项目。`project_get(name=)` 的现有契约仍可用。`AliasName` 不加唯一索引；已知空间时只在该空间内解析别名，未知且命中多行则业务错误「别名不唯一，请使用项目 ID」。`MatchPath` 前缀匹配（ADR-0002 第 3 条）在带空间过滤之后执行，禁止跨空间用路径命中项目。

4. **实体字段冻结为名称、slug、描述、图标、`IsDefault`。**
   slug 实例内 `uniqueIndex`，正则 `^[a-z]([a-z0-9-]*[a-z0-9])?$`，最长 63。图标禁止含 `://` 或 `/`。基因号 `GeneWorkspace = 48`，只写在 `internal/constant/gene_number.go`。默认空间不另写 Info 键。

5. **实例必须常驻一条默认空间。**
   初始化写入数据库一行：展示名初始为「默认空间」，slug 固定 `default` 且不可改，`IsDefault = true`。展示名可改。禁止删除默认空间，因此正常路径不会出现零空间。`GetDefault` 必须 `WHERE is_default = true ORDER BY created_at ASC`。Create/Update API 不得把客户端传入的 `is_default` 写进库。

6. **非默认空间由站主增删改查。**
   写操作走控制台 REST（`middleware.Auth`），不走 MCP。删除非默认空间时：先把该空间下全部项目的 `workspace_id` 改成默认空间 ID，再删空间行。数据库不建 ON DELETE CASCADE。禁止级联删除项目，禁止删完变成「未分组」。

7. **MCP 首版只读：`workspace_list`、`workspace_get`。**
   禁止 `workspace_create` / `update` / `delete`。Agent 的「当前空间」不存在服务端：凡是靠 `name` 或 `match_path` 解析项目、且未传 `project_id` 的工具，必须带 `workspace_id` 或 workspace slug。`project_id` 仍可全局查询，响应带上所属空间。MCP 与 Handler 禁止直访 `WorkspaceRepo`，只走 Logic。

8. **控制台「当前空间」只存在浏览器 `localStorage` 键 `lumina.currentWorkspaceId`。**
   API 的 query/body 仍显式传 `workspace_id`。禁止用 Cookie 当租户上下文，禁止账号级「当前空间」设置表。读取必须在 `window` 可用之后，避免 TanStack Start 服务端 `beforeLoad` 抛错。缺省或失效时回退到 `is_default` 那一行。

9. **认证保持单用户实例级。**
   不建成员表、角色、邀请。Owner、Token、WebAuthn 固定用户 ID 都不挂空间。SSH Key、API Key、OAuth 客户端、Info `site.*` 品牌名保持实例级。控制台选中空间后，chrome 用该空间的 name/icon/description，不拆 `site.name`。

10. **词表与口语。**
    业务域登记为 空间 / 工作空间 / `workspace`。Preview 文案改称「预览会话」；技能里的 cwd 改称「项目路径」。`pnpm-workspace` 仍是构建术语。Pin 同空间规则见 [RFC-0002](./0002-pin-workspace-isolation.md)。Memory 挂空间见 [RFC-0001](./0001-memory-decision-memory.md) 第 2 条。LLM Provider 与 RepoWiki 全局 prompt 升格不进本提案。

## 不做什么
- 不把 Workspace 做成 VS Code 多根目录容器——Project 已经是仓库身份，再做窗口会和 `MatchPath`、pnpm、Preview 第四次撞名；
- 不把实例本身当成唯一隐藏空间——站主要自己建「生活 / 工作」，默认行只保证永不空仓；
- 不把 `Project.Name` 改成空间内唯一——Q&A 已否决；代价是两空间不能同名，接受；
- 不在 MCP 连接上存「当前空间」——Streamable HTTP 没有与控制台 localStorage 对齐的稳定会话，多 Agent 会写坏；
- 不做成员、标签、通用 settings bag——单用户拆场景用不到，做了等于提前开多用户；
- 不在本提案实现 Memory / LLM 升格代码——契约分别留在 RFC-0001 与后续提案；
- 不给 Wiki 实体加 `workspace_id`——查询过滤走 `Project.workspace_id`。

## 备选

| 方案 | 否决原因 |
| --- | --- |
| 编辑器多根工作区（VS Code `.code-workspace`） | 违反方案第 1 条；解不了生活/工作隔离 |
| 实例即唯一隐藏空间，不暴露 CRUD | 违反方案第 6 条；Q&A 要求自己创建空间 |
| `Project.Name` 改为 `(workspace_id, name)` 唯一 | 违反方案第 3 条；改 `project_get(name)` 契约 |
| MCP `workspace_select` 把选择存进 Redis | 违反方案第 7 条；无可靠会话租户 |
| 给 `pin_push` 加 `workspace_id` 且 from 可空 | 交给 RFC-0002；本提案不规定 Pin 字段 |
| Kaneo 式成员 + 空间级标签 | 违反方案第 9 条；认证仍是单用户 |
| 默认空间只写 Info 键、不入库 | 违反方案第 5 条；改名、列举、MCP get 都没有行可查 |

## 待定
- LLM Provider `Name` 升格后唯一索引是否改为 `(workspace_id, name)`——不阻塞本提案；身份落地后再开配置升格提案。
- RepoWiki「全局 prompt」改成空间级的具体插入位置——不阻塞；现网仍是实例全局 + 项目 `CustomPrompt`。
- 删除非默认空间后，该空间下 Memory 卡片如何处理——Memory 尚未实现；实现时跟 RFC-0001，不得默认扫进别的空间。
