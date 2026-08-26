# Q&A 题型规范：code（代码编辑器输入）

## 📌 定位与适用场景
- **用途**：提供带语法高亮的代码编辑器，供用户直接输入或粘贴代码片段。
- **适用场景**：正则表达式、SQL 查询片段、复杂 JSON/YAML 配置、自定义算法实现等。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "code",
  "content": "string (必填，输入要求描述，支持 Markdown)",
  "config": {
    "language": "string (必填，语言标识，如 go / typescript / sql / json / regex 等)",
    "placeholder": "string (可选，占位代码片段)"
  }
}
```

支持的语言标识包括：`javascript`/`typescript`/`go`/`json`/`python`/`sql`/`yaml`/`regex`/`shell`/`css`/`html`/`rust`/`java` 等。

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "code",
  "content": "## 请提供用于校验邮箱域名的正则表达式",
  "config": {
    "language": "regex",
    "placeholder": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
  }
}
```

---

## 📥 后端返回格式

```text
[ANSWER] ^[a-zA-Z0-9_.+-]+@(?:(?:[a-zA-Z0-9-]+\.)?[a-zA-Z]+\.)?(example|mycorp)\.com$
[LANGUAGE] regex
```
