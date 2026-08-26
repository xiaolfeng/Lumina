# Q&A 题型规范：slider（滑块数值选择）

## 📌 定位与适用场景
- **用途**：允许用户在连续的数值范围内拖动滑块选择一个数值。
- **适用场景**：打分（0-10）、百分比配置（0%-100%）、权重分配、并发度设置等。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "slider",
  "content": "string (必填，数值选择说明，支持 Markdown)",
  "config": {
    "min": 0,
    "max": 100,
    "step": 5,
    "defaultValue": 50
  }
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "slider",
  "content": "## 请设定批处理任务的并发协程数 (Worker Pool Size)",
  "config": {
    "min": 1,
    "max": 64,
    "step": 1,
    "defaultValue": 8
  }
}
```

---

## 📥 后端返回格式

```text
[ANSWER] 16
```
