> 状态：draft · 承接 [设计 0001](../design/0001-workspace-identity.md) · 修订 [ADR-0004](../adr/0004-pin-constraint-delivery.md) 的边界，不取代其 FIFO / 消费契约

## 问题
Workspace 把实例拆成多个空间之后，Pin 仍按「整台实例里任意两个项目」投递。现网 `from_project_id` 可空（`api/pin/create.go`、MCP `pin_push`），`ResolveProject` 按 ID 或别名全局 `First()`（`internal/logic/pin.go`）。生活空间的别名可以命中工作空间的项目，无来源 Push 可以打进任意 `to`。

ADR-0004 按当前实现补录了 FIFO 与消费，没有组织边界。不能靠改 ADR 正文来补——单向流动，另写本提案。

## 方案
1. **Pin 只在同一 Workspace 内投递。** `from` 与 `to` 解析后的 `Project.workspace_id` 必须相同，否则 `BusinessError`「不能跨空间推送 Pin」。比较发生在 `PinLogic.Push`，读 `ProjectRepo`，不调用 `WorkspaceLogic`（沿用 ADR-0003：Pin 复用项目持久化，不做跨 Logic 编排）。Pin 表不增加 `workspace_id` 列，隔离走项目归属。

2. **`from_project_id` 必填，禁止无来源 Pin。** REST `binding:"required"`，MCP `pin_push` 的 `required` 纳入该字段，Logic 拒绝雪花零值。不给 `pin_push` 另加 `workspace_id`——来源项目已经带空间。控制台创建对话框与 REST 同一发布收紧，避免 `go:embed` 旧 UI 打 400。

3. **解析顺序：先 `from`（无空间过滤），再在 `from.WorkspaceID` 内解析 `to`。** MCP `handlePinPush` 现网先解析 `to` 再可选解析 `from`（`internal/mcp/pin_tools.go`），必须改成这个顺序，不能只改 JSON Schema。`to_project_name` 仍只接受雪花 ID 或别名，不解析 `Project.Name`（与现网 `ResolveProject`、ADR-0003 第 4 条一致）。别名：已知空间则只在该空间查；未知且命中多行则「别名不唯一，请使用项目 ID」。`pin_consume` / `pin_list` / `pin_peek` 仍按目标项目解析，不二次校验历史行的空间。

4. **ADR-0004 第 2–7 条保持。** 生命周期仍是 `pending → consumed`；消费仍是 FIFO + `FOR UPDATE SKIP LOCKED`；存储仍是 PostgreSQL；MCP 仍是那五个工具。本提案不改队列语义。

## 不做什么
- 不给 Pin 行加 `workspace_id`——会与 `Project.workspace_id` 双写，搬项目时还要改 Pin 表；
- 不把 `AliasName` 改成唯一索引——现状就允许重复，冲突用业务错误，不改表；
- 不让无来源 Pin 靠可选 `workspace_id` 继续存在——比收紧已有 `from` 字段更大，且失去来源审计；
- 不回头改 ADR-0004 正文里的 FIFO / 消费条款——接受后另写 ADR 或给 0004 加 superseded 指向；
- 不在本提案里规定控制台列表如何按空间过滤——那是 Workspace 设计 PR2 的接口增量，不是 Pin 契约。

## 备选

| 方案 | 否决原因 |
| --- | --- |
| 给 `pin_push` 增加 `workspace_id`，`from` 仍可空 | 违反方案第 2 条；无来源写入仍是跨空间洞 |
| Pin 表加 `workspace_id` | 违反方案第 1 条的「隔离走项目」；删除空间搬项目时要双表更新 |
| 别名全局 `uniqueIndex` | 违反方案第 3 条；强迫两个空间不能共用口语别名 |
| 维持 from 可空，只在两端都有时比较空间 | 无来源 Push 用全局 `project_id` 仍能打进另一空间 |
| 直接改 ADR-0004 第 1 条 | 单向流动禁止回头改旧文；FIFO 条款会被误伤 |

## 待定
存量 `from_project_id = 0` 的 Pin 行：消费时是否拒绝、是否一次性回填，还是继续可消费——不阻塞本提案评审。Workspace 落地顺序保证第二条空间出现前守卫已在二进制里，新写入不会再产生这类行。
