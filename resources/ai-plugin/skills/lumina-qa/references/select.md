# Q&A 题型规范：select（单选）

## 📌 定位与适用场景
- **用途**：用户从选项列表中选择一个选项。最基础、最常用的交互类型。
- **适用场景**：环境选择（开发/测试/生产）、部署模式、单方案确认、主干分支选择等二选一或多选一场景。
- **特殊能力**：支持 Supplement 机制。设置 `supplement: true` 后，可使用 `qa_push_supplement` 为每个选项单独推送详细技术文档。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "select",
  "content": "string (必填，问题标题与描述，支持 Markdown)",
  "supplement": true,
  "options": [
    {
      "label": "string (必填，选项简要标签，建议 1-5 词)",
      "description": "string (可选，选项简要说明)"
    }
  ]
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "select",
  "content": "## 请选择数据库迁移目标环境\n\n本次迁移涉及表结构变更，请确认执行环境。",
  "supplement": true,
  "options": [
    {
      "label": "开发环境",
      "description": "本地 Docker 容器，无外部依赖"
    },
    {
      "label": "测试环境",
      "description": "预发集成环境，已挂载测试数据"
    },
    {
      "label": "生产环境",
      "description": "线上环境，需遵循变更窗口期约束"
    }
  ]
}
```

---

## 📥 后端返回格式

```text
[ANSWER] 用户选择：开发环境
[DESCRIPTION] 本次迁移涉及表结构变更，请确认执行环境。
[OPTION_DESCRIPTION] 本地 Docker 容器，无外部依赖
[SUPPLEMENT] （若 Agent 为该选项推送过 supplement，此处输出内容）
```

---

## 💡 注意事项与避坑指南
1. **选项 Label 简洁**：`label` 应简短有力，详细的架构说明或代码示例应通过 `qa_push_supplement(option_id=...)` 注入。
2. **Supplement 时序**：若传了 `supplement: true`，必须在调用 `qa_get_answer` 之前为每个选项推送 supplement，否则前端将一直等待详情加载。
