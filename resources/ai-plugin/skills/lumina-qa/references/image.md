# Q&A 题型规范：image（图片上传与 OTP 下载）

## 📌 定位与适用场景
- **用途**：允许用户在浏览器端上传一张或多张图片。
- **适用场景**：UI 截图、架构设计图、报错堆栈图片、设计线框图等。
- **底层机制**：系统自动将 Base64 保存到服务端缓存目录，并生成**一次性临时下载令牌 (OTP Token)**，避免大体积 Base64 塞爆上下文。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "image",
  "content": "string (必填，上传要求描述，支持 Markdown)",
  "config": {
    "maxImages": 3,
    "maxSize": 10485760
  }
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "image",
  "content": "## 请上传出现样式错乱的浏览器全屏截图\n\n支持拖拽上传（最多 3 张，单张 ≤ 10MB）：",
  "config": {
    "maxImages": 3
  }
}
```

---

## 📥 后端返回格式

```text
[ANSWER] 用户已上传内容
---
[FILE_NAME] screenshot-2026-08-26.png
[DOWNLOAD_PATH] .lumina/cache/1234567890123456789/screenshot-2026-08-26.png
[DOWNLOAD_URL]
    - http://localhost:8080/api/v1/qa/download/a1b2c3d4e5f6g7h8...
[IMPORTANT] 下载链接为一次性令牌，使用后即失效。需重新下载请调用 qa_reget_answer。
[NOTE] 此处的 qa_reget_answer 仅用于多媒体重取，不可用于等待用户回答（等待请用 qa_get_answer）。
[TIP] 使用 curl -o <path> <url> 下载后引用路径。
[GIT_TIP] .lumina/cache/ 需加入 .gitignore。
```

---

## 💡 注意事项与避坑指南
1. **下载令牌一次性**：`DOWNLOAD_URL` 携带的 Token 仅可下载一次，有效期 10 分钟。如果需要再次下载，调用 `qa_reget_answer` 获取新 Token。
2. **禁止用 reget 轮询**：再次强调，`qa_reget_answer` 仅用于重取媒体附件，等待回答必须使用 `qa_get_answer`。
