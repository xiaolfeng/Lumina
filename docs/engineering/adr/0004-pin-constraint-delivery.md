> 状态：proposed · 依据当前实现补录

## 背景
Pin 已实现跨项目约束推送和 PostgreSQL FIFO 消费。旧 Wiki 还描述了超时天数、过期提示、`force` 重读和可手工改状态等未落地语义；继续把这些参考值当成契约会让 Agent 调用不存在的能力。

## 决定
1. **Pin 只传递明确影响目标项目的约束。** 每条记录保存来源项目、目标项目、标题、Markdown 内容、分类、优先级和消费状态；项目内临时待办和长期决策记忆不得写入 Pin；
2. **生命周期只有 `pending → consumed`，且状态只能由消费动作推进。** `pin_update` 仅允许修改 `priority` 与 `category`；禁止通过更新接口把记录恢复为 `pending`；
3. **`pin_consume` 不传 ID 时消费目标项目最早创建的 `pending` 记录。** Repository 在事务中使用 `FOR UPDATE SKIP LOCKED` 锁定队首并写入 `consumed_at`，保证并发消费者不会重复取得同一条记录；
4. **`pin_consume` 传 ID 时只消费归属目标项目且仍为 `pending` 的指定记录。** 已消费、不属于该项目或不存在的记录都不得再次消费；
5. **`pin_peek` 和 `pin_list` 是只读入口。** 它们可以查看已消费历史但不改变状态；列表按 `created_at ASC` 展示，并可按状态、分类、优先级和来源项目过滤；
6. **Pin 全量存储在 PostgreSQL，不建立 Redis 消费队列。** 数据库事务和行锁是消费原子性的唯一来源；
7. **MCP 固定暴露 `pin_push`、`pin_consume`、`pin_list`、`pin_update`、`pin_peek`。** REST API 服务管理界面；工具和 API 都必须经过同一 Logic 与 Repository 语义。

## 后果
- 得到：每条约束最多被正常消费一次；队列顺序、历史查询和状态写入共享同一事实来源；
- 失去：高优先级记录不会插队，已消费记录不能通过 `force` 再次进入消费结果，当前也没有自动过期与提醒；
- 违反时暴露方式：绕过 `Consume` 修改状态会缺失 `consumed_at`；脱离事务查询再更新会在并发下重复消费；Redis 队列与数据库状态会产生分叉。

## 否决项
- **通过 `pin_update` 修改状态**：消费入口不再唯一，并发与审计语义都会失效；
- **用 Redis List 实现 FIFO**：数据库仍需保存详情和状态，两套队列无法原子提交；
- **`force=true` 重读最近记录**：`pin_peek` 已提供无副作用的精确历史读取，混入消费工具会模糊语义；
- **自动过期后跳过队首**：当前没有被确认的过期策略；在引入前不得默默丢弃待处理约束；
- **按优先级插队**：Pin 当前承诺严格 FIFO；如需改变顺序，应以新 RFC 和 ADR 取代本决定。
