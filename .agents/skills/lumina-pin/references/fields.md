# Pin 字段与队列语义

`pin_push` / `pin_list` / `pin_consume` 的项目标识都走同一套解析：项目名称、别名或雪花 ID 字符串。不必先 `project_get`，但来源项目建议带上 `from_project_id`，方便消费方回溯。

## `priority`

| 值 | 何时用 |
|---|---|
| `high` | 破坏性接口、不改就会编译/运行失败 |
| `medium` | 需要跟进，但有兼容窗口 |
| `low` | 提醒类，可延后 |

## `category`

| 值 | 何时用 |
|---|---|
| `api_change` | 请求/响应/事件字段变更 |
| `dependency` | 共享库或运行时版本升级 |
| `notice` | 约定、文档、流程（默认） |
| `other` | 以上都不贴切时 |

## 队列

- `pin_list(status="pending")` 按 `createdAt` 升序，FIFO 可见。
- 不传 `id` 的 `pin_consume` 取最旧 pending；队列空时返回「暂无待处理约束」，不是错误。
- `pin_peek` 只读，不改状态。
- `pin_update` 只能改 `priority` / `category`，不能把状态改成 `consumed`。

## 不要当成 Pin 的东西

本仓库里的 TODO、某次 PR 的自检清单、仅影响当前工作树的提醒。那些走本地任务或对话即可。
