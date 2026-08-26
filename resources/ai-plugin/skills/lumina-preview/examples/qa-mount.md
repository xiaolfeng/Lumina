# Example: 把 Preview 挂到 Q&A 选项

契约正文只在 [`../../_shared/preview-qa-contract.md`](../../_shared/preview-qa-contract.md)。这里只演示一次正确调用。

Input: `select` 题的「方案 A」需要给用户看布局稿。

## 顺序

1. 按 [`standalone-review.md`](./standalone-review.md) 把文件传完并 `preview_file_list`。
2. 从返回里取出 `qa_supplement.content`，原样作为字符串。
3. 调用 `qa_push_supplement`：

```json
{
  "session_id": "<QA会话ID>",
  "question_id": "<问题ID>",
  "option_id": "<方案A的option_id>",
  "content_type": "preview",
  "content": "{\"session_id\":\"3333333333333333333\",\"file_id\":\"4444444444444444444\"}"
}
```

4. 再 `qa_get_answer`。

## 反例（都会让详情面板白屏或解析失败）

| content | 错在哪 |
|---|---|
| `http://127.0.0.1:8080/preview?session=abc` | URL 不是引用 |
| `abc1234567890123` | Hash 只给网页用 |
| ` ```json\n{"session_id":"..."}\n``` ` | 多了围栏 |
| `预览：{"session_id":"..."}` | 多了说明文字 |
