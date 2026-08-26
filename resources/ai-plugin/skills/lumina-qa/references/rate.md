# Q&A 题型规范：rate（多维度独立星级评分）

## 📌 定位与适用场景
- **用途**：对提供的每一个选项维度进行独立的星级/分数评价（1-N 星）。
- **适用场景**：多维度代码评审打分（可读性、性能、健壮性）、用户满意度评价、UI 质感评估等。
- **重要约束**：**必须提供 `options` 列表**，每个 option 将在界面上独立渲染一套星级打分控件。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "rate",
  "content": "string (必填，打分说明，支持 Markdown)",
  "options": [
    {
      "label": "string (必填，维度名称)",
      "description": "string (可选，维度说明)"
    }
  ],
  "config": {
    "max": 5,
    "step": 1
  }
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "rate",
  "content": "## 请对当前重构方案进行多维度打分评估 (1-5 星)",
  "options": [
    {"label": "代码可维护性", "description": "模块解耦程度与可读性"},
    {"label": "性能与并发表现", "description": "吞吐量与内存占用估算"},
    {"label": "向前兼容性", "description": "旧版本客户端契约是否受损"}
  ],
  "config": {
    "max": 5
  }
}
```

---

## 📥 后端返回格式

```text
[ANSWER] 代码可维护性: 5, 性能与并发表现: 4, 向前兼容性: 5
```
