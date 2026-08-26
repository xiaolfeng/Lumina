# Q&A 题型规范：rank（拖拽优先级排序）

## 📌 定位与适用场景
- **用途**：用户在界面中通过上下拖拽或箭头调整选项列表的相对顺序，以确定优先级。
- **适用场景**：需求实施优先级排期、功能重要性排序、待办事项排序等。
- **重要约束**：**必须提供 `options` 列表**。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "rank",
  "content": "string (必填，排序引导说明，支持 Markdown)",
  "options": [
    {"label": "string (必填，选项 1)"},
    {"label": "string (必填，选项 2)"},
    {"label": "string (必填，选项 3)"}
  ]
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "rank",
  "content": "## 请按紧急程度为以下技术债排序\n\n拖拽调整顺序，最上方代表最高优先级：",
  "options": [
    {"label": "修复内存泄漏隐患"},
    {"label": "升级 Go 1.25 到最新补丁版本"},
    {"label": "补充 Preview 模块自动化集成测试"},
    {"label": "重构 Handler 层参数绑定逻辑"}
  ]
}
```

---

## 📥 后端返回格式

```text
[ANSWER] 1. 修复内存泄漏隐患 → 2. 补充 Preview 模块自动化集成测试 → 3. 升级 Go 1.25 到最新补丁版本 → 4. 重构 Handler 层参数绑定逻辑
```
