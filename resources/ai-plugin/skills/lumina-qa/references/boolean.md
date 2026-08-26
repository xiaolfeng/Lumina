# Q&A 题型规范：boolean（布尔确认）

## 📌 定位与适用场景
- **用途**：明确的“是/否”二选一确认。
- **适用场景**：高危操作二度确认（如删库、清空缓存）、开关特性确认、执行/跳过决策等。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "boolean",
  "content": "string (必填，确认提示，支持 Markdown)"
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "boolean",
  "content": "## ⚠️ 确认执行数据库全量重建吗？\n\n该操作将执行 `DROP TABLE` 并重新执行 AutoMigrate，**所有现有数据将被清空**且无法撤销！"
}
```

---

## 📥 后端返回格式

```text
[ANSWER] 是
```
或
```text
[ANSWER] 否
```
