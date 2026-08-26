# Q&A 题型规范：text（自由文本输入）

## 📌 定位与适用场景
- **用途**：允许用户自由输入单行或多行文本。最通用的输入交互类型。
- **适用场景**：用户反馈收集、自定义需求描述、API 路径输入、错误信息补充等。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "text",
  "content": "string (必填，问题说明，支持 Markdown)",
  "config": {
    "multiline": true,
    "placeholder": "string (可选，输入框占位提示)",
    "maxLength": 500
  }
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "text",
  "content": "## 请描述该 Bug 的复现步骤与触发条件\n\n请尽可能详细提供触发路径或异常入参：",
  "config": {
    "multiline": true,
    "placeholder": "1. 打开控制台\n2. 点击某按钮...\n3. 触发 500 错误",
    "maxLength": 1000
  }
}
```

---

## 📥 后端返回格式

```text
[ANSWER] 1. 进入控制台 /projects 页面；2. 快速连击删除按钮导致防抖失效。
```
