# Q&A 题型规范：review（内容逐段审阅）

## 📌 定位与适用场景
- **用途**：展示设计文档、规范或 API 定义，供用户逐段审阅并反馈意见。
- **适用场景**：API 契约评审、架构文档审查、PR 概述审阅等。
- **决策分支**：用户可选择 **Approve（批准）** 或 **Revise（提出修改意见）**（无 Reject 分支）。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "review",
  "content": "string (必填，审阅提示，支持 Markdown)",
  "config": {
    "sections": [
      {
        "id": "string (必填，段落 ID)",
        "title": "string (必填，段落标题)",
        "content": "string (必填，段落 Markdown 内容)"
      }
    ]
  }
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "review",
  "content": "## 请审阅新建的 Pin 模块对外接口契约",
  "config": {
    "sections": [
      {
        "id": "sec-push",
        "title": "1. 约束推送接口 (POST /pin/push)",
        "content": "接收 `title`, `content`, `priority`, `to_project_name` 参数，返回雪花 ID。"
      },
      {
        "id": "sec-consume",
        "title": "2. 约束消费接口 (POST /pin/consume)",
        "content": "默认先进先出 (FIFO) 消费队首 pending 记录；可传 `id` 精确消费。"
      }
    ]
  }
}
```

---

## 📥 后端返回格式

### 分支 1：用户批准 (Approve)
```text
[ANSWER] 用户批准了该修改
```

### 分支 2：用户提出修改 (Revise)
```text
[ANSWER] 用户要求修改
[FEEDBACK] 消费接口需要增加按 category 过滤的能力。
```
