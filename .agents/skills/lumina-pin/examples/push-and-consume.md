# Example: 发布一条 API 变更并在下游消费

## 发布端（项目 A · 后端）

Input: 你刚给 WebSocket 消息加了 `trace_id`，前端不改就会丢字段。

```json
{
  "from_project_id": "Lumina",
  "to_project_name": "Lumina-Frontend",
  "title": "WebSocket 消息体新增 trace_id",
  "priority": "high",
  "category": "api_change",
  "content": "## 接口变更\n\n`qa_push_question` 广播体新增顶层 `trace_id`。interact 页错误上报需要带上该 ID。\n\n兼容：旧前端忽略未知字段不会崩，但排障对不上。"
}
```

Output: 得到 Pin 的雪花 ID，状态 `pending`，落在目标项目 FIFO 队尾。

## 消费端（项目 B · 前端）

### 阶段 1：只读读取约束并评估决策（不改变 pending 状态）

进入前端仓库后，只读预览队首待处理约束详情与正文：

```json
{
  "project_name": "Lumina-Frontend"
}
```
（调用 `pin_peek`）

Output: 只读获得完整正文，状态仍为 `pending`：
```text
Pin 详情：

ID: 1234567890123456789
标题: WebSocket 消息体新增 trace_id
内容:
## 接口变更
`qa_push_question` 广播体新增顶层 `trace_id`。interact 页错误上报需要带上该 ID。
分类: api_change
状态: pending
优先级: high
...
```

或者调用 `pin_list` 查看全部待处理约束及其内容正文：

```json
{
  "project_name": "Lumina-Frontend",
  "status": "pending",
  "page": 1,
  "size": 10
}
```

### 阶段 2：展开本地代码适配与修复

根据读取到的正文要求，在前端代码中进行适配（例如为 WebSocket 错误上报带上 `trace_id`），并运行测试确认无误。

### 阶段 3：显式消费闭环约束 (`pin_consume`)

代码修改完成并验证通过后，调用 `pin_consume` 显式确认闭环：

```json
{
  "project_name": "Lumina-Frontend",
  "id": "1234567890123456789"
}
```

## 反例

- 未看内容或未改代码就盲目批量调 `pin_consume` → 导致约束漏适配
- `pin_update` 把状态改成 consumed → 工具不允许，状态只能 `pin_consume`
- 把「记得跑一下测试」推进 Pin → 这不是跨项目契约
