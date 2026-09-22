---
name: lumina-pin
description: Lumina Pin 在项目之间定向推送约束并按 FIFO 消费。发现会影响其他仓库的接口变更、依赖升级或破坏性约定时使用；进入目标仓库开工前先查待办队列。不要拿来记本项目临时待办。
license: MIT
compatibility: Requires Lumina MCP (Streamable HTTP) and network access to the Lumina instance.
metadata:
  author: lumina
  version: "0.1.0"
argument-hint: [ project-name | pin-id ]
allowed-tools: Read, Write, Edit, Bash, AskUserQuestion, mcp__lumina__project_get, mcp__lumina__project_list, mcp__lumina__pin_push, mcp__lumina__pin_consume, mcp__lumina__pin_list, mcp__lumina__pin_peek, mcp__lumina__pin_update
---

# Lumina 跨项目约束与队列管理指南 (lumina-pin)

用于指导 AI Agent 在多代码库、微服务或单体多包架构下，进行跨项目依赖约束的**定向推送**与**先进先出 (FIFO) 消费**，确保上下游契约与技术决策闭环。

目标项目解析与 Pin 相同：名称、别名或雪花 ID 字符串均可，后端 `ResolveProject` 会识别。字段枚举与空队列语义见 [`references/fields.md`](./references/fields.md)。完整发布→消费走查见 [`examples/push-and-consume.md`](./examples/push-and-consume.md)。

---

## 🎯 核心定位与适用场景

- **约束发布方 (Producer)**：你在开发项目 A 时，修改了公共模块或 API 契约，这将直接影响项目 B、C。此时通过 `pin_push` 将约束推送到目标项目。
- **约束消费方 (Consumer)**：你在进入项目 B 开发时，先通过 `pin_peek` 或 `pin_list` 只读读取待处理约束正文，评估影响并做出技术决策；完成代码适配或确认知晓后，再调用 `pin_consume` 显式确认消费并闭环。

---

## 🔄 Pin 标准两端生命周期

```text
【发布端 A】                                  【消费端 B】
     │                                            │
pin_push(to="B", priority, content)               │
     │                                            │
     ▼ (落库进入 B 的 FIFO 队列)                    │
  [Pending] ──────────────────────────────▶ pin_peek(project="B") 或 pin_list
                                                  │ (只读读取正文，状态仍为 Pending，开展本地决策与适配)
                                                  ▼
                                            pin_consume(project="B", id="...")
                                                  │ (代码适配完成，显式确认消费闭环)
                                                  ▼
                                            [Consumed] (状态单向归档)
```

---

## 📋 详细操作流程

### 1. 发布跨项目约束 (`pin_push`)
当发现需要通知其他项目的约束时调用：
```json
{
  "to_project_name": "Lumina-Frontend",
  "title": "WebSocket 消息体新增 trace_id 字段",
  "priority": "high",
  "category": "api_change",
  "content": "## 接口变更说明\n\n后端 `qa_push_question` 广播的 WebSocket 消息体中已新增 `trace_id` 顶层字段，前端在渲染 interact 页面时需在错误上报中携带该 ID。"
}
```
- **`to_project_name`**：支持目标项目的**名称**、**别名**或**雪花 ID 字符串**（后端自动解析）。
- **`priority`**：`high`（高）/ `medium`（中）/ `low`（低）。
- **`category`**：`notice`（普通通知）/ `dependency`（依赖升级）/ `api_change`（接口变动）/ `other`（其他）。

---

### 2. 只读读取约束详情与决策评估 (`pin_peek` / `pin_list`) —— 只读不改状态
进入目标项目后，首先读取待处理约束并做出技术决策，**绝不改变任何状态（保持 pending）**：

#### 模式 A：只读预览队首待处理约束 (`pin_peek`)
传入 `project_name`（不传 `id`），直接只读获取队首待处理约束的完整标题与 Markdown 正文：
```json
{
  "project_name": "Lumina-Frontend"
}
```

#### 模式 B：全览待处理列表及内容 (`pin_list`)
查看当前项目积压的待处理约束列表（每条均包含标题、正文、分类与优先级）：
```json
{
  "project_name": "Lumina-Frontend",
  "status": "pending",
  "page": 1,
  "size": 10
}
```

#### 模式 C：精确查看指定 ID 约束 (`pin_peek`)
```json
{
  "id": "1234567890123456789"
}
```

---

### 3. 本地代码适配与显式消费闭环 (`pin_consume`)
在本地审阅正文、完成决策并落实代码修改（或确认知晓）后，调用 `pin_consume` 显式将约束状态标记为已消费：

#### 模式 A：精确 ID 消费（推荐）
针对性消费刚刚已处理完毕的特定约束：
```json
{
  "project_name": "Lumina-Frontend",
  "id": "1234567890123456789"
}
```

#### 模式 B：FIFO 队首消费
不传 `id`，消费当前队首约束：
```json
{
  "project_name": "Lumina-Frontend"
}
```

---

### 4. 约束元数据调整 (`pin_update`)
调整约束的优先级或分类（状态只能通过 `pin_consume` 推进到 `consumed`）：
```json
{
  "id": "1234567890123456789",
  "priority": "medium",
  "category": "notice"
}
```

---

## ⛔ 核心红线 (Hard Rules)

1. **MUST**: 状态流转单向原子化。`pending` 转换为 `consumed` 只能通过 `pin_consume`，严禁通过任何其他方式绕过。
2. **MUST**: 读取与消费分离。读取操作（`pin_peek` / `pin_list`）绝不修改状态，Agent 必须先读取正文、评估决策并完成本地处理，随后显式调用 `pin_consume` 确认闭环。
3. **NEVER**: 严禁将本地临时待办当作跨项目约束。Pin 专用于**跨仓库/跨项目**的契约传递。
