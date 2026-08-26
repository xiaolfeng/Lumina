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

进入前端仓库后先看队列：

```json
{
  "project_name": "Lumina-Frontend",
  "status": "pending",
  "page": 1,
  "size": 10
}
```

不确定内容时 `pin_peek`：

```json
{ "id": "1234567890123456789" }
```

代码改完（上报带上 `trace_id`）再消费。只处理了这一条就精确消费：

```json
{
  "project_name": "Lumina-Frontend",
  "id": "1234567890123456789"
}
```

按时间挨个清队列时不传 `id`：

```json
{ "project_name": "Lumina-Frontend" }
```

## 反例

- 未改代码就 FIFO 连消三条 → 约束被归档，下游会漏适配
- `pin_update` 把状态改成 consumed → 工具不允许，状态只能 `pin_consume`
- 把「记得跑一下测试」推进 Pin → 这不是跨项目契约
