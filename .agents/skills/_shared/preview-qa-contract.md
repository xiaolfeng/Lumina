# Preview 与 Q&A 跨模块联动数据契约 (Preview & Q&A Contract)

Lumina 支持将前端 Preview 原型无缝挂载到 Q&A 问题或特定选项的右侧详情面板中。为了保证前端沙盒隔离渲染能够精确加载入口文件，必须严格遵守以下数据契约。

---

## 🔄 标准跨模块联动流程

```text
① 创建 Preview 会话并上传前端文件
       │
       ▼
② preview_file_list(session_id) 核对清单
       │
       ▼ 获取返回中的 qa_supplement 对象：
       │ {
       │   "content_type": "preview",
       │   "content": "{\"session_id\":\"123456\",\"file_id\":\"789012\"}"
       │ }
       │
       ▼
③ qa_push_supplement(
       session_id = <QA会话ID>,
       question_id = <问题ID>,
       option_id = <选项ID（可选）>,
       content_type = "preview",
       content = <原样传入上面返回的 JSON 字符串>
   )
       │
       ▼
④ qa_get_answer(session_id = <QA会话ID>)
```

---

## 📜 严格参数规范

### 1. `content_type` 字段
- **必须为** `"preview"`。

### 2. `content` 字段格式
- **必须是且只能是**合法的 JSON 字符串：
  ```json
  {"session_id":"1234567890123456789","file_id":"9876543210987654321"}
  ```
- **字段含义**：
  - `session_id`：目标 Preview 会话的雪花 ID 字符串。
  - `file_id`：作为渲染入口的 HTML 文件的雪花 ID 字符串（由 `preview_file_list` 自动计算的 `entry_file` 对应 ID）。

---

## 🚫 典型错误反例清单

| 错误写法（反例） | 导致后果 | 正确做法 |
|---|---|---|
| ❌ 传递网页 URL：<br>`"content": "http://localhost:8080/preview?session=abc123"` | 前端无法解析会话/文件 ID，渲染白屏 | 使用 `{"session_id":"...","file_id":"..."}` |
| ❌ 传递 Hash：<br>`"content": "abc1234567890123"` | Hash 仅用于独立浏览器访问，非 API 引用 | 使用 `{"session_id":"...","file_id":"..."}` |
| ❌ 添加 Markdown 围栏：<br>`"content": "```json\n{\"session_id\":\"...\"}\n```"` | JSON 解析崩溃 | 直接传递纯 JSON 字符串，不加任何 Markdown 围栏 |
| ❌ 附加额外说明文字：<br>`"content": "这是预览原型：{\"session_id\":\"...\"}"` | 字符串无法反序列化 | 纯 JSON 文本，额外说明写在 Markdown supplement 中 |
