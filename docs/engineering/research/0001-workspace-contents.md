> 调研日期：2026-09-09

## 背景
Lumina 要把 Workspace 做成产品概念。仓库里目前没有对应实体，但「工作区」一词已经同时指 pnpm 包布局、Preview 沙盒和 Agent 本机路径。核心问题：Workspace 在现有单用户 + Project 身份模型里是哪一层、首版往里面装什么。

## 发现

### 仓库现状：没有 Workspace 这一层

`internal/constant/gene_number.go` 的基因号从 `GeneProject=32` 到 `GeneOAuthClient=47`，没有 Workspace。`internal/entity/project.go` 只有 `Name`、`AliasName`、`MatchPath`、`Description`；`Name` 带全局 `uniqueIndex`。

业务记录一律挂项目，不挂组织：

| 实体 | 关联 |
| --- | --- |
| `QaSession` | `project_id` |
| `PreviewSession` | `project_id` |
| `RepoWikiConfig` | `project_id`（unique） |
| `Pin` | `from_project_id` / `to_project_id` |

认证与配置是实例级。`entity/info.go` 写明「个人项目不需要多用户体系」；WebAuthn 使用固定用户 ID（`internal/logic/webauthn_user.go`）；Token 仓储注释写明单用户无「按用户」索引。站点名、Logo、域名在 `constant/info_key.go` 的 `site.*` 键上，主键是键名本身，没有空间维度。`LlmProvider.Name` 同样是全局 `uniqueIndex`。

ADR-0002 把 Project 定为「项目身份的唯一记录」，路径前缀匹配例子是 `/workspace/Lumina`——这里的 workspace 是路径片段，不是实体。ADR-0001 的业务域清单是 RepoWiki / Memory / Q&A / Pin / Preview，没有 Workspace。`docs/scope-manage.md` 在本稿登记前也没有该词。

「工作区」在代码里的三种用法：

| 说法 | 实际对象 | 出处 |
| --- | --- | --- |
| pnpm workspace | monorepo 包 | `pnpm-workspace.yaml` |
| 预览工作区 | `PreviewSession` 沙盒 | `entity/preview_session.go` 注释、`mcp/preview_tools.go` |
| Agent 当前工作区 | `pwd` 绝对路径 | `resources/ai-plugin/skills/_shared/project-resolver.md` |

控制台前端 `web/src` 没有 Workspace 路由或组件。

### 业界把 Workspace 用成两种东西

**租户 / 组织。** Linear 把 workspace 定义为组织里所有 issue 的家，建议一个组织只用一个空间；设置包含名称与 URL、成员、标签、模板、集成；多空间则成员名单和账单各自独立（[Linear Workspaces](https://linear.app/docs/workspaces)）。Kaneo MCP 在 2026-09-09 实测：空间有 `id` / `name` / `slug` / `logo` / `description` / `metadata`；`list_projects` 必须带 `workspaceId`；标签和成员挂在空间上。当前账号两个空间（「筱锋の个人」「筱锋の时光川行」），前者下面已有 Lumina / Zephyr / NewAPI 三个项目，成员仅 owner 一人。

**编辑器窗口。** VS Code 的 workspace 是一个窗口打开的一份或多份文件夹，用来存 settings / tasks / launch；多根时写 `.code-workspace`。文档明确说 VS Code 没有 Visual Studio 那种 project / solution（[VS Code Workspaces](https://code.visualstudio.com/docs/editor/workspaces)）。

两套词同名不同层。Lumina 若做成租户层，就不能再拿 Preview 会话或 `pwd` 去顶这个名字。

### Q&A 对齐（会话 `414082952174052352`，2026-09-09）

交互页：https://lumina.x-lf.com/interact?session=8578f3c7c2b9f302913be32860f4a1f2

| 题目 | 回答 |
| --- | --- |
| 定位 | 组织租户层：`Workspace 1 — N Project`。Project 仍表示一个代码仓库 |
| 用途 | 单用户把自己的生活 / 工作拆开管，不是给别人用，也不照搬 Kaneo |
| 成员 | 继续单用户，空间不建成员表 |
| 词表 | 2 字「空间」/ 4 字「工作空间」/ 英文 `workspace` |
| 首版装什么 | 标识字段（名称、slug、描述、图标）；`Project.workspace_id`；默认空间；MCP `workspace_list` / `workspace_get` |
| 首版不装 | 空间级标签、通用设置袋、成员表 |
| 基数 | 用户自己创建，支持增删查改 |
| 默认空间 | 创建实例时入库一条，名字就叫「默认空间」，可改名，不可删除 |
| 删空间 | 里面的项目移到默认空间 |
| 现有项目 | 跟着默认空间走（默认空间在初始化时就有） |
| `Project.Name` | 仍全局唯一，空间只分组 |
| Pin | 只允许同一空间内；`from` / `to` 必须同 Workspace |
| 升到空间的配置 | LLM Provider/Model、站点外观、RepoWiki 全局 prompt |
| 留在实例 | SSH 密钥、API Key、OAuth 客户端 |
| MCP 解析 | 先选空间，再在该空间内做 `match_path` |
| Memory | 只挂空间，不绑具体仓库 |
| Wiki | 不另做空间级 Wiki；Project 归入空间后随项目隔离 |
| 落地顺序 | 实体与归属 → 控制台切换 → MCP 工具 → 名称唯一性 → 配置升格 → 成员 |

补充原话：默认空间「不可以删除，但是可以改名字，创建时候自带的……也需要数据库入库」；删除其他空间后「移动到默认空间」；「依然是单用户……方便区分空间不同管理。例如生活和工作」。

### 对现有模块的冲击（按现状推断，不是方案）

身份。Project 要多一个必填 `workspace_id`。现有项目在初始化/升级时挂到默认空间，否则「先选空间再解析」会让旧数据从 MCP 里消失。`Name` 保持全局唯一，则 `project_get(name=)` 的现有契约还能用；但 `match_path` 必须加空间范围，否则生活/工作里若出现路径前缀重叠会串空间。

Pin。ADR-0003 没有组织边界。要满足「只允许空间内」，消费和推送都得校验两端项目的 `workspace_id` 相同。

Memory。RFC-0001 仍是 draft，卡片含「来源项目」，待定项写了来源项目是否必填。用户要求记忆只挂空间：生活空间的决策不能在工作空间被检索到。这和「来源项目」不是同一条轴。

RepoWiki。配置已经按项目 unique。空间隔离可以只靠 `Project.workspace_id` 做查询过滤，不必给 Wiki 版本再加一层外键。全局 prompt 若升到空间，要在现有「全局 prompt + 项目 CustomPrompt」上面再插一层，或把「全局」的含义从实例改成空间。

LLM 与外观。`LlmProvider.Name` 全局唯一、Info `site.*` 以键名为主键：升到空间不是加个字段那么简单，唯一索引和 Info 表形状都要改。用户把这项排在实体、控制台切换、MCP 之后，可以作为第二批。

认证。用户明确不做成员。Owner、Token、WebAuthn 可以继续实例级；空间没有 ACL。

MCP。现在 `project_get` / `project_list` 没有 `workspace_id` 参数。用户要 Agent 先选空间，等于工具签名和技能 `project-resolver.md` 都要改。勾进首版的 MCP 工具是 `list` / `get`；增删查改是产品能力，写接口走控制台 REST 还是也进 MCP，Q&A 没钉死。

## 结论

事实结论：Lumina 运行时没有 Workspace；身份边界是实例 + Project。业界同名概念分成租户层（Linear / Kaneo）和编辑器窗口（VS Code）两支。本仓库里的三种「工作区」口语都不等于租户。Q&A 意向见上表，尚未拍板。

倾向：把 `workspace` 登记为新业务域，下一篇 RFC 按「身份层扩展」写，不要做成与 Q&A/Pin 平级、互相调用的第六业务域。首版骨架是实体 + `Project.workspace_id` + 默认空间 + 控制台切换 + MCP `list`/`get`。标签、成员、Name 改空间内唯一、SSH/API Key/OAuth 升格，不进首版。Preview 改称「预览会话」，Agent cwd 称「项目路径」，pnpm workspace 留在构建词汇。

开放问题：

1. 默认空间的初始 slug 怎么生成；slug 是实例内唯一还是全局唯一。
2. MCP 是否暴露 `workspace_create` / `update` / `delete`，还是写操作只走控制台 REST。
3. Agent 的「当前空间」是每个工具带 `workspace_id`，还是另有会话级状态。
4. LLM 升格后 `LlmProvider.Name` 的唯一索引改成 `(workspace_id, name)` 还是继续全局。
5. 站点外观升格后，Info `site.name` / `site.logo-url` 是加空间前缀、拆表，还是空间实体自己持有展示字段（名称/图标已在标识字段里）。
6. Memory RFC 的「来源项目」是删掉、改成可选，还是记忆挂空间、来源项目只做引用。
7. 基因号占用：下一个空位是 48，是否给 Workspace 用。
8. 控制台「当前空间」存在哪：Cookie、localStorage，还是账号级设置。

## 参考

- Linear Workspaces — https://linear.app/docs/workspaces
- VS Code: What is a VS Code workspace? — https://code.visualstudio.com/docs/editor/workspaces
- Kaneo MCP `list_workspaces` / `list_projects` / `list_workspace_members` 实测（2026-09-09）
- ADR-0002 项目标识 — [../adr/0002-project-identity-resolution.md](../adr/0002-project-identity-resolution.md)
- ADR-0003 Pin — [../adr/0003-pin-constraint-delivery.md](../adr/0003-pin-constraint-delivery.md)
- RFC-0001 Memory — [../rfc/0001-memory-decision-memory.md](../rfc/0001-memory-decision-memory.md)
- `internal/entity/project.go`、`qa_session.go`、`pin.go`、`preview_session.go`、`info.go`、`llm_provider.go`
- `internal/constant/gene_number.go`、`info_key.go`
- Q&A 会话 `414082952174052352`（项目 Lumina `404337268617126912`）
