# Example: 晋升后再 Fork 发新版本

Input: 购物车交互稿已经在 Preview 里核对通过，用户要对外分享，并可能再改一版。

## 1. 核对草稿

`preview_file_list(session_id)` 确认 `entry_file` 为 `index.html`。

## 2. 看 slug

```json
{ "project_id": "1234567890123456789", "page": 1, "size": 20 }
```

若 `design-cart` 未被占用，继续晋升。

## 3. 晋升

```json
{
  "session_id": "3333333333333333333",
  "slug": "design-cart",
  "title": "购物车交互",
  "set_as_active": true
}
```

打开返回文本 `page` 区段下的 `page_url` 标签，形如 `http://127.0.0.1:8080/pages/<project>/design-cart/index.html`。

```bash
open "<page_url>"
```

## 4. 继续改

```json
{ "page_id": "<pages_promote 返回的 page 区段下的 id>" }
```

用返回的 `preview_url` 打开登录工作台，`preview_file_upload` 覆写文件后再晋升。Fork 会话不必再传 slug。

## 反例

- 没有 HTML 就 `pages_promote` → 拒绝
- 在 MCP 里传 password → 无此字段
- 线上已更新却不带 `confirm_conflict` → 只返回冲突，指针不变
