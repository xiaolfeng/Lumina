---
name: lumina-pin
description: Lumina Pin 在项目之间定向推送约束并按 FIFO 消费。发现会影响其他仓库的接口变更、依赖升级或破坏性约定时使用；进入目标仓库开工前先查待办队列。不要拿来记本项目临时待办。
license: MIT
compatibility: Requires Lumina MCP (Streamable HTTP) and network access to the Lumina instance.
metadata:
  author: lumina
  version: "0.1.1"
argument-hint: [ project-name | pin-id ]
allowed-tools: Read, Write, Edit, Bash, AskUserQuestion, mcp__lumina__project_get, mcp__lumina__project_list, mcp__lumina__pin_push, mcp__lumina__pin_consume, mcp__lumina__pin_list, mcp__lumina__pin_peek, mcp__lumina__pin_update, mcp__plugin_lumina_lumina__project_get, mcp__plugin_lumina_lumina__project_list, mcp__plugin_lumina_lumina__pin_push, mcp__plugin_lumina_lumina__pin_consume, mcp__plugin_lumina_lumina__pin_list, mcp__plugin_lumina_lumina__pin_peek, mcp__plugin_lumina_lumina__pin_update
---

# Lumina 跨项目约束与队列管理指南 (lumina-pin)

用于指导 AI Agent 在多代码库、微服务或单体多包架构下，进行跨项目依赖约束的**定向推送**与**先进先出 (FIFO) 消费**，确保上下游契约与技术决策闭环。

目标项目解析与 Pin 相同：名称、别名或雪花 ID 字符串均可，后端 `ResolveProject` 会识别。字段枚举与空队列语义见 [`references/fields.md`](./references/fields.md)。完整发布→消费走查见 [`examples/push-and-consume.md`](./examples/push-and-consume.md)。

---

## 🎯 核心定位与适用场景

- **约束发布方 (Producer)**：你在开发项目 A 时，修改了公共模块或 API 契约，这将直接影响项目 B、C。此时通过 `pin_push` 将约束推送到目标项目。
- **约束消费方 (Consumer)**：你在进入项目 B 开发时，通过 `pin_list` 和 `pin_consume` 按创建时间升序依序消费未处理的约束，确认适配并闭环。

---

## 🔄 Pin 标准两端生命周期

```text
【发布端 A】                                  【消费端 B】
     │                                            │
pin_push(to="B", priority, content)               │
     │                                            │
     ▼ (落库进入 B 的 FIFO 队列)                    │
  [Pending] ──────────────────────────────▶ pin_list(project="B", status="pending")
                                                  │
                                                  ▼
                                            pin_consume(project="B")
                                                  │
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

### 2. 查阅项目约束队列 (`pin_list`)
进入目标项目工作区后，首先查看当前项目有哪些待处理约束：
```json
{
  "project_name": "Lumina-Frontend",
  "status": "pending",
  "page": 1,
  "size": 10
}
```
- 结果按 `createdAt` 升序排列（天然呈现 FIFO 顺序）。

---

### 3. 消费与闭环约束 (`pin_consume`)
确认已在代码中适配或处理了该约束后，将其标记为已消费：

#### 模式 A：FIFO 队首消费（最常用）
不传 `id`，自动取出并消费最旧的一条 pending 约束：
```json
{
  "project_name": "Lumina-Frontend"
}
```

#### 模式 B：精确 ID 消费
当针对性处理了某一条特定约束时，传入其雪花 ID：
```json
{
  "project_name": "Lumina-Frontend",
  "id": "1234567890123456789"
}
```

---

### 4. 只读回查与元数据调整

- **只读预览 (`pin_peek`)**：查看指定 Pin 的完整内容（不改变 pending / consumed 状态）：
  ```json
  {"id": "1234567890123456789"}
  ```
- **调整元数据 (`pin_update`)**：调整约束的优先级或分类（注意：状态不能通过此工具修改）：
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
2. **MUST**: 消费前核对内容。消费代表“该约束已在当前项目中得到处理/知晓”，不可在未做任何处理时盲目批量清空队列。
3. **NEVER**: 严禁将本地临时待办当作跨项目约束。Pin 专用于**跨仓库/跨项目**的契约传递。
