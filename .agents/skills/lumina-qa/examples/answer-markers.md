# Example: `qa_get_answer` 返回标记

`qa_get_answer` 单次阻塞约 25s。按返回文本标记分支，不要引入外部等待。

## `[ANSWERED]` / `[ANSWER]`

```text
[ANSWER] 用户选择：开发环境
```

解析答案，进入下一步。不要再调 `qa_get_answer` 等同一题。

## `[NEED_SUPPLEMENT]` + `[USER_NOTE]`

```text
[NEED_SUPPLEMENT]
[USER_NOTE] 请补一张当前表结构的 ER 图，并说明回滚步骤
```

按用户笔记调用 `qa_push_supplement`，然后**再次** `qa_get_answer`。

## `[PENDING]`

```text
[PENDING]
```

用户还没提交。立刻再调 `qa_get_answer`。不要 `bash sleep`，不要改用 `qa_reget_answer` 轮询。

## `[STOPPED]`

```text
[STOPPED]
```

等待过久，进入休眠保护。停掉工具调用，告诉用户：「如果已在浏览器回答，回复『继续』，我再取答案。」用户回复后再调 `qa_get_answer`。

## `[SKIPPED]`

```text
[SKIPPED]
```

用户跳过本题。记下跳过，继续后续逻辑，不要把跳过当成默认选项。

## `qa_reget_answer` 用在哪

只用于已经回答过的内容：重读文本、刷新图片/文件的一次性下载令牌。不能代替 `qa_get_answer` 等待。
