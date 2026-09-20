# Example: 独立视觉评审

Input: 用户要看一个购物车数量加减的交互稿，不改仓库。

## 1. 会话

`preview_session_list` 没有可复用会话时：

```json
{
  "project_id": "1234567890123456789",
  "title": "购物车数量加减"
}
```

不要打开此时返回的空 `preview_url`。

## 2. 逐文件上传

先 HTML（可从 [`../assets/index.html`](../assets/index.html) 改），再 CSS / JS。

```json
{
  "session_id": "3333333333333333333",
  "filename": "index.html",
  "content": "<!DOCTYPE html>\n<html lang=\"zh-CN\">\n<head>\n  <meta charset=\"utf-8\">\n  <title>Cart</title>\n  <link rel=\"stylesheet\" href=\"style.css\">\n</head>\n<body>\n  <div id=\"app\"></div>\n  <script src=\"app.js\"></script>\n</body>\n</html>"
}
```

```json
{
  "session_id": "3333333333333333333",
  "filename": "style.css",
  "content": "#app{font:16px/1.4 sans-serif}"
}
```

```json
{
  "session_id": "3333333333333333333",
  "filename": "app.js",
  "content": "document.getElementById('app').textContent='qty: 1'"
}
```

## 3. 核对

```json
{ "session_id": "3333333333333333333" }
```

确认 `entry_file` 为 `index.html`，`workflow.state` 为 `ready_for_review`，并拿到绝对 `preview_url`。

## 4. 打开

```bash
open "http://127.0.0.1:8080/preview/<hash>/index.html"
```

无 GUI 时才在对话里给出可点击 URL。

## 反例

- 创建会话后立刻打开 URL → 白屏
- `filename: "assets/app.js"` → 上传拒绝
- `href="/style.css"` → 沙盒里加载失败
