# Q&A 题型规范：plan（分段方案展示与审批）

## 📌 定位与适用场景
- **用途**：分段展示一个结构化计划或方案，用户可逐段审阅并做出整体决策。
- **适用场景**：大型重构计划、数据迁移步骤、分阶段发布流程等。
- **决策分支**：用户可选择 **Approve（批准）**、**Reject（拒绝）** 或 **Revise（提出分段修改意见）**。

---

## ⚙️ 参数格式

```json
{
  "session_id": "string (必填，目标会话 ID)",
  "question_type": "plan",
  "content": "string (必填，方案总览，支持 Markdown)",
  "config": {
    "sections": [
      {
        "id": "string (必填，阶段唯一标识)",
        "title": "string (必填，阶段标题)",
        "content": "string (必填，阶段具体内容，支持 Markdown)"
      }
    ]
  }
}
```

---

## 📝 实战 JSON 示例

```json
{
  "session_id": "1234567890123456789",
  "question_type": "plan",
  "content": "## Redis 集群平滑迁移三阶段实施计划\n\n请审阅各阶段执行步骤：",
  "config": {
    "sections": [
      {
        "id": "phase-1",
        "title": "阶段一：双写与数据同步",
        "content": "开启业务层双写，启动增量同步工具校验 key 一致性。"
      },
      {
        "id": "phase-2",
        "title": "阶段二：切读与观测",
        "content": "将只读流量切换至新集群，观测延迟与命中率 30 分钟。"
      },
      {
        "id": "phase-3",
        "title": "阶段三：切写与下线旧集群",
        "content": "全量流量切至新集群，停用双写并归档旧集群配置。"
      }
    ]
  }
}
```

---

## 📥 后端返回格式

### 分支 1：用户批准 (Approve)
```text
[ANSWER] 用户已批准该计划
[PLAN_DETAIL]
1. 阶段一：双写与数据同步
   开启业务层双写，启动增量同步工具校验 key 一致性。
2. 阶段二：切读与观测
   将只读流量切换至新集群，观测延迟与命中率 30 分钟。
3. 阶段三：切写与下线旧集群
   全量流量切至新集群，停用双写并归档旧集群配置。
```

### 分支 2：用户拒绝 (Reject)
```text
[ANSWER] 用户已拒绝该计划
[FEEDBACK] 目前处于大促封网期，禁止一切集群迁移操作。
```

### 分支 3：用户要求修订 (Revise)
```text
[ANSWER] 用户要求修改该计划
[REVISIONS]
1. [phase-2] 观测时间延长至 2 小时，并加入 CPU 与网络 IO 告警看板。
[FEEDBACK] 整体可行，但阶段二的稳定性观测需强化。
```
