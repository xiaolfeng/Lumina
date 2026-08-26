# Q&A 题型规范：file（文件上传与 OTP 下载）

## 📌 定位与适用场景
- **用途**：允许用户上传任意格式的文件（支持通过 `accept` 后缀过滤）。
- **适用场景**：用户本地配置文件（YAML/JSON/ENV）、日志包、抓包数据、导出文档等。
- **底层机制**：同 `image` 题型，服务端存储并通过一次性下载令牌 (OTP Token) 提供访问。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "file",
  "content": "string (必填，上传要求描述，支持 Markdown)",
  "config": {
    "accept": [".json", ".yaml", ".yml", ".env"],
    "maxFiles": 3,
    "maxSize": 5242880
  }
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "file",
  "content": "## 请上传生产环境的脱敏配置文件\n\n仅支持 `.json` 与 `.yaml` 文件：",
  "config": {
    "accept": [".json", ".yaml", ".yml"],
    "maxFiles": 2
  }
}
```

---

## 📥 后端返回格式

```text
[ANSWER] 用户已上传内容
---
[FILE_NAME] config.prod.yaml
[DOWNLOAD_PATH] .lumina/cache/1234567890123456789/config.prod.yaml
[DOWNLOAD_URL]
    - http://localhost:8080/api/v1/qa/download/token123456...
[IMPORTANT] 下载链接为一次性令牌，使用后即失效。需重新下载请调用 qa_reget_answer。
```
