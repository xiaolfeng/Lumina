# Memory 详细设计

## 核心定位

Memory 是 AI Agent 的长期决策记忆系统 — 记录需要跨会话保留、后续持续遵循的重要决策和约定。由 MCP 端主动推送构建，非自动采集。

与 RepoWiki 的代码知识不同，Memory 存储的是**决策性知识**：架构决策、编码规范、团队约定、技术选型理由等。

## 数据结构

### 记忆卡片（逻辑模型）

```json
{
  "id": "雪花 ID",
  "title": "决策标题",
  "content": "## 详细内容\n\n支持 Markdown 渲染（纯文本，不含图片/视频/PDF）",
  "reason": "决策原因说明",
  "scope": "适用范围",
  "priority": "high | medium | low",
  "tags": ["架构", "规范", "安全"],
  "source_project": "关联项目 ID（可选）",
  "status": "active | deprecated | archived",
  "created_at": "创建时间",
  "updated_at": "更新时间"
}
```

### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | 雪花 ID | 自动 | 主键 |
| title | string | 是 | 决策标题，简洁概括决策内容 |
| content | text | 是 | 详细内容，支持 Markdown 格式。MVP 阶段仅支持纯文本（不含图片/视频/PDF） |
| reason | text | 否 | 决策原因，解释为什么做出这个决策 |
| scope | string | 否 | 适用范围，描述此决策适用的上下文边界 |
| priority | enum | 是 | 优先级：high（必须遵循）、medium（建议遵循）、low（参考） |
| tags | string[] | 否 | 标签列表，用于分类和检索 |
| source_project | string | 否 | 关联的项目 ID，表明决策来源 |
| status | enum | 是 | 状态：active（生效中）、deprecated（已废弃）、archived（已归档） |

## 生命周期

```
Agent 调用 memory_create
        │
        ▼
   [active] ◀─── memory_update（可随时修改内容/标签/优先级）
        │
        │   默认查询只返回 active 状态的记忆
        │
        ├── Agent 主动 memory_update(status=deprecated)
        │       │
        │       ▼
        │   [deprecated]（已废弃，仍可查询但默认不返回）
        │
        └── Agent 主动 memory_update(status=archived)
                │
                ▼
            [archived]（归档，历史参考，默认不返回）
```

### 状态说明

| 状态 | 说明 | 默认查询是否返回 |
|------|------|-----------------|
| active | 当前生效的决策 | ✅ 是 |
| deprecated | 已被替代但保留参考 | ❌ 否（需显式指定） |
| archived | 历史归档，仅作记录 | ❌ 否（需显式指定） |

## 检索策略

### 查询维度

- **关键词搜索**：标题 + 内容 + 原因的全文检索
- **标签过滤**：按一个或多个标签筛选
- **优先级过滤**：按 high / medium / low 筛选
- **状态过滤**：默认只返回 active，可指定 deprecated 或 archived
- **来源项目**：按关联项目 ID 筛选

### 排序策略（参考值）

1. 优先级降序（high > medium > low）
2. 更新时间降序（最近更新的在前）

### 缓存策略

- 热点记忆（频繁查询的 active 卡片）可缓存到 Redis
- 标签索引缓存，加速按标签检索
- 缓存更新策略：写入时主动失效

## 数据库模型（逻辑模型）

### 记忆表（参考值）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | 雪花 ID | 主键 |
| title | varchar(255) | 决策标题 |
| content | text | 详细内容（Markdown） |
| reason | text | 决策原因 |
| scope | varchar(512) | 适用范围 |
| priority | varchar(16) | 优先级（high / medium / low） |
| status | varchar(16) | 状态（active / deprecated / archived） |
| source_project | varchar(64) | 关联项目 ID（可选） |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 标签表（参考值）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | 雪花 ID | 主键 |
| name | varchar(64) | 标签名称（唯一） |
| created_at | timestamp | 创建时间 |

### 记忆-标签关联表（参考值）

| 字段 | 类型 | 说明 |
|------|------|------|
| memory_id | 雪花 ID | 关联记忆 |
| tag_id | 雪花 ID | 关联标签 |

## 使用场景

### 场景一：架构决策记录

```
Agent 完成技术选型后：
  → memory_create(
      title="采用严格分层架构",
      content="本项目采用 route→handler→logic→repository 四层架构...",
      reason="确保代码职责清晰，便于维护和测试",
      scope="所有新增接口必须遵循",
      priority="high",
      tags=["架构", "规范"]
    )
```

### 场景二：上下文恢复

```
Agent 新会话开始：
  → memory_query(tags=["架构"], status="active")
  ← 返回所有活跃的架构相关决策
  → 基于这些决策指导后续开发
```

### 场景三：决策更新

```
Agent 发现旧决策不再适用：
  → memory_update(
      id="12345",
      status="deprecated",
      reason="已迁移到微服务架构，原分层约定不再适用"
    )
```
