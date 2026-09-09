> 状态：draft · 承接 [RFC-0003](./0003-workspace-identity.md) · 修订 [ADR-0003](../adr/0003-pin-constraint-delivery.md) 的投递边界，不取代其 FIFO / 消费契约

## 问题
ADR-0003 按现网补录了 FIFO 与消费：Pin 在整台实例里从任意项目打到任意项目。`CreatePinRequest.FromProjectID` 没有 `binding:"required"`，MCP `pin_push` 的 `from_project_id` 不在 `required` 里，Logic 把零值当成「无来源」写入（`internal/logic/pin.go` 的 `Push`）。`ResolveProject` 按雪花 ID 或别名全局查询，别名走 `FindByAliasName` + `First()`，没有空间谓词。

Workspace 落地后，生活空间和工作空间可以各有一个别名 `blog`。无来源 Push 只要知道目标项目的雪花 ID，就能打进另一个空间。ADR-0003 第 2–7 条（状态机、FIFO、行锁、五工具）仍然对；缺的是投递边界。单向流动，不改 0003 正文，用本提案补。

## 方案
1. **Pin 只允许同一 Workspace 内投递。**
   `PinLogic.Push` 在两端项目都解析成功之后比较 `from.WorkspaceID` 与 `to.WorkspaceID`，不同则 `BusinessError`「不能跨空间推送 Pin」。Pin 表不增加 `workspace_id` 列。隔离读 `entity.Project.WorkspaceID`。禁止把比较下放到 Repository 当「是否允许写入」的业务判断。

2. **`from_project_id` 与目标都必填。禁止无来源 Pin。**
   REST：`api/pin/create.go` 的 `FromProjectID` 加 `binding:"required"`，Logic 对 `IsZero()` 再挡「缺少来源项目」。MCP：`pin_push` 的 `required` 改为 `title`、`content`、`priority`、`to_project_name`、`from_project_id`。控制台 `create-dialog.tsx` 与 `web/src/lib/models/request/pin.ts` 必须在同一发布里把 from 改成必填——二进制 `go:embed` 旧对话框会把今日合法的空 from 打成 HTTP 400。不给 `pin_push` 另加 `workspace_id` 字段。

3. **`ResolveProject` 增加空间参数，调用方必须按场景传入。**
   签名改为 `ResolveProject(ctx, nameOrID string, workspaceID SnowflakeID)`。
   输入能解析为雪花 ID：按 ID 全局 `GetByID`；若 `workspaceID` 非零且项目所属空间不匹配，返回 NotFound「项目不存在」，不泄露「在别的空间」。
   否则当别名（大小写不敏感完整匹配，ADR-0002 第 4 条仍有效）：`workspaceID` 非零则只在该空间查；为零则先 COUNT，大于 1 返回 `BusinessError`「别名不唯一，请使用项目 ID」，等于 1 返回该行，等于 0 返回 NotFound。
   `to_project_name` 仍只接受雪花 ID 或别名，禁止解析 `Project.Name`。

4. **写入解析顺序：先 from，再在 from 的空间内解析 to。**
   `PinLogic.Push` 先 `ResolveProject(from, 0)`，再 `ResolveProject(to, from.WorkspaceID)`，再做方案第 1 条的比较。
   MCP `handlePinPush`（`internal/mcp/pin_tools.go`）现网先解析 to 再可选解析 from。必须改成与 Push 相同的顺序，不能只改 JSON Schema。禁止 handler 先 `ResolveProject(to, 0)` 再把 ID 塞给 Push——两空间同别名会在 handler 层误报 COUNT>1，或绑到错误空间的 ID。

5. **消费与只读入口不二次校验历史行的空间。**
   `pin_consume`、`pin_list`、`pin_peek` 仍按目标项目解析，`ResolveProject(project_name, 0)`。别名冲突时要求改用项目 ID。它们不是跨空间写入路径。ADR-0003 的 FIFO、`FOR UPDATE SKIP LOCKED`、按 ID 消费、peek/list 只读，全部保持。

6. **PinLogic 继续持有 `*repository.ProjectRepo`，禁止改成调用 `WorkspaceLogic` 或 `ProjectLogic`。**
   与 ADR-0002 第 6 条、`internal/logic/pin.go` 的 `pinRepo` 注释一致：解析是数据访问，不是跨域编排。

7. **控制台列表按空间过滤是 Workspace 的接口增量，不是本提案的投递契约。**
   `PinListRequest` 增加可选 `workspace_id`、SQL 用 `to_project_id IN (SELECT id FROM projects WHERE workspace_id = ?)`，写在设计 0001。本提案不把它当成 Pin 写入规则。零值表示不过滤，供看板名称映射。

## 不做什么
- 不给 Pin 行加 `workspace_id`——与 `Project.workspace_id` 双写；删空间搬项目时要改两张表；
- 不给 `AliasName` 加唯一索引——现网就允许重复，冲突用方案第 3 条的业务错误；
- 不保留「from 为空则跳过空间比较」——`project_id` 全局可查，无来源 Push 就是跨空间洞；
- 不回头改 ADR-0003 第 2–7 条的 FIFO / 状态机 / 五工具——接受后另写 ADR 冻投递边界，0003 只加相关链接；
- 不在 consume 时因为历史 `from_project_id = 0` 而拒绝消费——存量处理见待定，写入路径先堵住。

## 备选

| 方案 | 否决原因 |
| --- | --- |
| `pin_push` 增加可选 `workspace_id`，from 仍可空 | 违反方案第 2 条；无来源写入仍能打进任意 `to` |
| Pin 表加 `workspace_id` | 违反方案第 1 条；删空间要双表更新 |
| 别名全局 `uniqueIndex` | 违反方案第 3 条；强迫两个空间不能共用口语别名 |
| 只在 from、to 都有时比较空间，from 可空 | 违反方案第 2 条；零 from 的比较被跳过 |
| 直接改 ADR-0003 第 1 条塞进空间语义 | 单向流动；FIFO 条款会被误伤 |
| handler 继续先解析 to，Push 里再纠正 | 违反方案第 4 条；别名 COUNT 发生在错误的空间范围 |

## 待定
存量 `from_project_id = 0` 的 Pin 行（现网合法）：消费时拒绝、一次性回填 from、还是继续可消费——不阻塞本提案评审。Workspace 落地顺序要求写入守卫先于第二条空间出现，新行不会再是 0。选定之前，consume / peek / list 按方案第 5 条不因 from=0 失败。
