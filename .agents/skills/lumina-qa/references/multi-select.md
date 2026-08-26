# Q&A 题型规范：multi-select（多选）

## 📌 定位与适用场景
- **用途**：用户从选项列表中勾选一个或多个选项。
- **适用场景**：功能特性启用、中间件勾选、权限角色分配、批量操作目标确认等。
- **特殊能力**：支持 `config.min` 和 `config.max` 限制选择数量；支持用户填写“其他（`__other__`）”自定义输入。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "multi-select",
  "content": "string (必填，问题描述，支持 Markdown)",
  "supplement": true,
  "options": [
    {
      "label": "string (必填，选项标签)",
      "description": "string (可选，选项说明)"
    }
  ],
  "config": {
    "min": 1,
    "max": 3
  }
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "multi-select",
  "content": "## 请勾选本次重构需要纳入的模块\n\n可多选（最少 1 个，最多 3 个）：",
  "supplement": true,
  "options": [
    {"label": "用户鉴权模块", "description": "JWT 改造与 OAuth2 接入"},
    {"label": "缓存加速层", "description": "Redis Cache-Aside 封装"},
    {"label": "审计日志中间件", "description": "操作链路记录"},
    {"label": "OpenAPI 自动生成", "description": "Swag 注释与客户端 SDK"}
  ],
  "config": {
    "min": 1,
    "max": 3
  }
}
```

---

## 📥 后端返回格式

多选结果以 `---` 分隔每一项：

```text
[ANSWER] 用户选择 2 项
[DESCRIPTION] 可多选（最少 1 个，最多 3 个）：
---
[OPTION] 用户鉴权模块
[OPTION_DESCRIPTION] JWT 改造与 OAuth2 接入
[SUPPLEMENT] （该选项的 supplement）
---
[OPTION] 缓存加速层
[OPTION_DESCRIPTION] Redis Cache-Aside 封装
---
[OPTION] __other__
[SUPPLEMENT] （若用户勾选了其他并输入，此处输出自定义内容）
```

---

## 💡 注意事项与避坑指南
- 解析返回文本时，请按 `---` 分割块以提取用户选中的所有 `[OPTION]`。
