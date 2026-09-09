> 状态：proposed · 承接 [RFC-0003](../rfc/0003-workspace-identity.md)

## 背景
身份边界若继续写在调研和设计稿里，后续 PR 会各写一套「空间」含义。必须把 Workspace 相对于 Project、MCP、认证的允许与禁止冻住。一次性的迁移步骤、缓存键和 PR 切分留在 [设计 0001](../design/0001-workspace-identity.md)，本文件不收录。

## 决定
1. **Workspace 是身份层，贴在 Project 旁边。禁止把它做成与 Q&A / Pin / RepoWiki / Preview 平级、互相调用的第六业务域。**
   `WorkspaceLogic` 禁止 import 上述域的 Logic。删除空间时只通过 `ProjectRepo` 改 `projects.workspace_id`。跨域流程仍由 Agent 走 MCP。

2. **每个 Project 必须属于恰好一个 Workspace。禁止出现 `workspace_id` 为零的在役项目。**
   v1 禁止提供「把项目改挂到另一空间」的接口。唯一搬家路径：删除非默认空间时，该空间下全部项目改挂默认空间。Wiki 版本禁止再加 `workspace_id`，隔离随 Project。

3. **`Project.Name` 在实例内全局唯一。禁止改成空间内唯一。**
   `AliasName` 禁止加唯一索引。带空间解析别名时只在该空间查；不带空间且命中多行，必须报「别名不唯一，请使用项目 ID」，禁止 `First()` 随便挑一行。`MatchPath` 前缀匹配必须在空间过滤之后执行。

4. **默认空间必须入库，slug 固定为 `default` 且禁止修改，`IsDefault` 为 true 且禁止删除。**
   展示名初始为「默认空间」，允许改名。禁止用 Info 键代替这一行。`GetDefault` 必须按 `created_at ASC` 取最早的 `is_default = true`。Create/Update API 禁止接受客户端写入 `is_default`。实例禁止零空间。

5. **非默认空间允许站主经认证 REST 增删改查。禁止 MCP 暴露 create / update / delete。**
   删除非默认空间必须先把项目改挂默认空间再删行。禁止 ON DELETE CASCADE 删掉项目。禁止删除后变成未分组。MCP 只允许 `workspace_list` 与 `workspace_get`。

6. **靠 `name` 或 `match_path` 解析项目且未传 `project_id` 时，必须带 `workspace_id` 或 workspace slug。禁止在服务端为 MCP 连接保存「当前空间」。**
   `project_id` 允许全局查询，响应必须带所属空间。MCP 与 Handler 禁止直访 Workspace 的 Repository。

7. **控制台当前空间只允许存在 `localStorage` 键 `lumina.currentWorkspaceId`。禁止用 Cookie 或账号级设置表充当租户上下文。**
   HTTP API 必须在 query 或 body 显式传 `workspace_id`。服务端渲染阶段禁止读 `localStorage`。键缺失或指向已删空间时，回退到默认空间。

8. **认证保持单用户实例级。禁止为 Workspace 建成员表、角色或邀请。**
   SSH Key、API Key、OAuth 客户端、Info `site.*` 品牌名保持实例级。控制台 chrome 用空间自身的 name/icon/description，禁止为此拆分 `site.name`。

9. **基因号 `GeneWorkspace = 48` 只允许写在 `internal/constant/gene_number.go`。**
   禁止在其他 constant 文件重复定义。slug 正则与图标字符限制按 RFC-0003 第 4 条执行。

10. **口语：Preview 称预览会话，Agent cwd 称项目路径。禁止再用「工作区」指这两处。**
    `pnpm-workspace` 仍指 monorepo 包布局。Pin 投递边界见 [ADR-0006](./0006-pin-workspace-isolation.md)。Memory 隔离见 [RFC-0001](../rfc/0001-memory-decision-memory.md) 第 2–3 条，在 Memory 实现前不注册占位接口。

## 后果
- 得到：生活/工作可以分成两条空间；项目、MCP 路径解析、控制台列表有同一套归属规则；默认空间保证升级后旧项目还能被解析；
- 失去：两个空间不能注册同名项目；不能在 MCP 里建空间；不能把项目随便挪到另一个空间；没有成员协作；
- 违反时暴露方式：`match_path` 命中另一空间的仓库；Pin 或列表出现另一空间的对象；`workspace_id = 0` 的项目让空间比较变成空操作；MCP 工具不带空间仍能按路径命中。

## 否决项
- **VS Code 式多根目录 Workspace**：和 Project / MatchPath 重叠，解不了场景隔离；
- **实例即唯一隐藏空间、不暴露 CRUD**：站主无法拆生活/工作；
- **MCP `workspace_select` 把当前空间存进 Redis**：多 Agent 并行会串空间，调试看不到请求里的空间 ID；
- **`Project.Name` 改为空间内唯一**：破坏现网 `project_get(name)`，Q&A 已否决；
- **成员表或空间级 ACL**：认证仍是单用户，做了等于提前开多用户；
- **删除空间时级联删除项目**：把 RepoWiki 版本和 Q&A 会话一起带走，破坏面大于搬家到默认空间；
- **给 WikiVersion 加 `workspace_id`**：与 Project 归属双写。注意：按 `Project.workspace_id` 过滤 Wiki 查询不受此禁令限制。
