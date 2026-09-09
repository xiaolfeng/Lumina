> 状态：proposed · 承接 [RFC-0002](../rfc/0002-pin-workspace-isolation.md)
>
> 本文件只冻投递边界。[ADR-0003](./0003-pin-constraint-delivery.md) 的 FIFO、状态机、行锁、五工具仍然有效，不被取代。

## 背景
Workspace 把实例拆成多个空间之后，ADR-0003 的「任意项目打到任意项目」会让约束越过生活/工作边界。现网 `from` 可空、别名全局 `First()`，必须把允许与禁止冻在投递入口，而不是改 0003 的消费语义。

## 决定
1. **Pin 只允许同一 Workspace 内投递。禁止跨空间 Push。**
   `PinLogic.Push` 比较两端 `entity.Project.WorkspaceID`，不同则 `BusinessError`「不能跨空间推送 Pin」。Pin 表禁止增加 `workspace_id` 列。比较禁止下放到 Repository 当业务判断。

2. **`from_project_id` 与目标都必填。禁止无来源 Pin。**
   REST 必须 `binding:"required"`，Logic 必须拒绝雪花零值，MCP `pin_push` 必须把 `from_project_id` 列入 `required`。控制台创建对话框必须与 REST 同一发布收紧。禁止给 `pin_push` 另加 `workspace_id` 来代替 from。

3. **写入时必须先解析 from（不带空间过滤），再在 `from.WorkspaceID` 内解析 to。**
   MCP `handlePinPush` 禁止维持现网「先 to 后 from」的顺序。`ResolveProject` 必须接受空间参数：雪花 ID 全局查再校验所属空间；别名在已知空间内查，未知空间且多行命中必须报「别名不唯一，请使用项目 ID」，禁止 `First()`。`to_project_name` 禁止解析 `Project.Name`。

4. **`pin_consume` / `pin_list` / `pin_peek` 不因历史行的空间或空 from 而改变 ADR-0003 的消费契约。**
   它们按目标项目解析，别名冲突时改用项目 ID。FIFO、`pending → consumed`、`FOR UPDATE SKIP LOCKED`、五工具集合，一律仍按 ADR-0003。

5. **PinLogic 只通过 `ProjectRepo` 解析项目。禁止调用 `WorkspaceLogic` 或 `ProjectLogic` 做编排。**
   与 ADR-0002 第 6 条一致。

## 后果
- 得到：生活空间的约束进不了工作空间的队列；来源项目可审计；别名不再靠 `First()` 送到错误项目；
- 失去：不能再推「无来源」通知；不能跨空间点对点投递（那不是 Pin 的职责）；
- 违反时暴露方式：另一空间的项目队列出现来源不符的 pending；别名相同的两个项目被随机命中；MCP 不传 from 仍能 Push 成功。

## 否决项
- **from 可空、只在两端都有时比较空间**：零 from 跳过比较，全局 `to` 雪花 ID 仍是跨空间洞；
- **Pin 行加 `workspace_id`**：与项目归属双写，删空间要改两张表；
- **别名加全局唯一索引**：强迫两个空间不能用同一个口语别名；ADR-0002 已把别名定为非唯一单值；
- **handler 先解析 to 再交给 Push 纠正**：别名 COUNT 会在错误的空间范围触发；
- **用本 ADR 修改 ADR-0003 的 FIFO 或插队规则**：投递边界和队列语义分开冻；要改 FIFO 必须新 RFC 取代 0003。
  注意：控制台按 `workspace_id` 过滤 Pin 列表是查询增量，不是投递入口，不受「禁止跨空间 Push」误伤——零值不过滤仍然合法。
