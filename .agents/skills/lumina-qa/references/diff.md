# Q&A 题型规范：diff（代码修改差异对比）

## 📌 定位与适用场景
- **用途**：在双栏或内联视图中展示代码修改前后的 Diff 对比，供用户审查、拒绝或编辑。
- **适用场景**：重构方案审批、Bug 修复代码提议、SQL 索引优化、配置更新等。
- **决策分支**：用户可选择 **Approve（批准）**、**Reject（拒绝）** 或 **Edit（直接在线编辑后提交）**。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "diff",
  "content": "string (必填，修改目的说明，支持 Markdown)",
  "config": {
    "before": "string (必填，修改前原始内容)",
    "after": "string (必填，修改后提议内容)",
    "language": "string (必填，语言标识，如 go / typescript / sql 等)"
  }
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "diff",
  "content": "## 优化用户查询索引提示\n\n添加索引 Hint 避免全表扫描：",
  "config": {
    "before": "SELECT * FROM users WHERE status = 1 AND created_at > ?",
    "after": "SELECT /*+ INDEX(users idx_status_created) */ * FROM users WHERE status = 1 AND created_at > ?",
    "language": "sql"
  }
}
```

---

## 📥 后端返回格式

### 分支 1：用户批准 (Approve)
```text
[ANSWER] 用户已批准该修改
[FINAL]
SELECT /*+ INDEX(users idx_status_created) */ * FROM users WHERE status = 1 AND created_at > ?
```

### 分支 2：用户拒绝 (Reject)
```text
[ANSWER] 用户已拒绝该修改
[FEEDBACK] 线上数据库已默认命中组合索引，无需硬编码 Hint。
```

### 分支 3：用户在线修改后提交 (Edit)
```text
[ANSWER] 用户修改后提交
[FINAL]
SELECT id, username, status FROM users WHERE status = 1 AND created_at > ?
```

---

## 💡 注意事项与避坑指南
- 只有在 `approve` 和 `edit` 时会返回 `[FINAL]` 标记，Agent 可直接提取 `[FINAL]` 下方的代码作为最终应用版本。
