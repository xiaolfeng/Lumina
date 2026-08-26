# Example: select + 选项级 Supplement

Input: 用户要在 PostgreSQL 与 SQLite 之间选本地开发库，需要看到迁移成本再拍板。

## 1. 解析项目并开会话

```json
{ "match_path": "/Users/xiaolfeng/personal-project/Lumina" }
```

`project_get` 命中后：

```json
{
  "project_id": "1234567890123456789",
  "session_type": "temporary",
  "title": "本地开发库选型",
  "agent_name": "lumina-agent"
}
```

`qa_session_create` 返回含 `[URL] http://127.0.0.1:8080/interact?session=...`。立刻：

```bash
open "http://127.0.0.1:8080/interact?session=..."
```

## 2. 先核对题型

调用 `qa_what_question(question_type="select")`，或读 [`../references/select.md`](../references/select.md)。不要凭记忆填字段。

## 3. 推送问题

```json
{
  "session_id": "1111111111111111111",
  "question_type": "select",
  "content": "## 本地开发数据库怎么选？\n\n本次只影响本机与 CI 的 sqlite 回退策略，不改生产。",
  "supplement": true,
  "options": [
    { "label": "PostgreSQL", "description": "与生产一致，需要本机 Docker" },
    { "label": "SQLite", "description": "零依赖，部分 JSON 查询行为不同" }
  ]
}
```

## 4. 立刻注入 Supplement

`qa_push_question` 返回各选项 `option_id` 后，**在 `qa_get_answer` 之前**分别推送：

```json
{
  "session_id": "1111111111111111111",
  "question_id": "2222222222222222222",
  "option_id": "<postgresql-option-id>",
  "content_type": "markdown",
  "content": "## PostgreSQL\n\n- 与 `docker-compose.yaml` 现有服务对齐\n- 迁移脚本无需分叉\n- 本机需要 Docker Desktop"
}
```

```json
{
  "session_id": "1111111111111111111",
  "question_id": "2222222222222222222",
  "option_id": "<sqlite-option-id>",
  "content_type": "markdown",
  "content": "## SQLite\n\n- `make test` 可无 Docker 跑通\n- JSON 运算符与 Postgres 不完全一致\n- 不适合验证并发锁"
}
```

## 5. 阻塞等待

反复调用 `qa_get_answer(session_id)`，直到出现：

```text
[ANSWER] 用户选择：PostgreSQL
[DESCRIPTION] 本次只影响本机与 CI 的 sqlite 回退策略，不改生产。
[OPTION_DESCRIPTION] 与生产一致，需要本机 Docker
```

收到 `[PENDING]` 就立刻再调 `qa_get_answer`，不要 `sleep`。状态机样例见 [`answer-markers.md`](./answer-markers.md)。
