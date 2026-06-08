# Pin 详细设计

## 核心定位

Pin 是 Lumina 的**跨项目依赖约束传递系统**。当 Agent 在项目 A 的工作中发现了项目 B 需要知晓的约束、变更或注意事项时，Agent 可以向项目 B 推送一条 Pin。项目 B 的 Agent 在后续工作中按 FIFO 顺序消费这些 Pins，确保跨项目依赖不会遗漏。

与 Memory 的**决策归档**不同，Pin 关注的是**约束传递**：

- Memory 记录的是"本项目应该遵循什么"
- Pin 传递的是"其他项目需要你注意什么"

Pin 具有明确的定向性和时效性。一条 Pin 从创建到消费完成，生命周期清晰且不可逆。

## 数据结构

### Pin 逻辑模型

```json
{
  "id": "雪花 ID",
  "from_project_id": "来源项目 ID（推送方）",
  "to_project_id": "目标项目 ID（接收方）",
  "title": "约束标题",
  "content": "## 详细内容\n\n支持 Markdown 渲染（纯文本，不含图片/视频/PDF）",
  "category": "注意事项 | 依赖约束 | 接口变更 | 其他",
  "status": "pending | consumed",
  "priority": "high | medium | low",
  "expire_days": "超时天数（默认读取 Info 表 pin_expire_days）",
  "consumed_at": "消费时间",
  "created_at": "创建时间",
  "updated_at": "更新时间"
}
```

### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | 雪花 ID | 自动 | 主键 |
| from_project_id | 雪花 ID | 是 | 来源项目 ID，标识这条 Pin 由哪个项目推送 |
| to_project_id | 雪花 ID | 是 | 目标项目 ID，标识这条 Pin 应该被哪个项目消费。实际匹配时支持 alias_name 模糊匹配 |
| title | string | 是 | 约束标题，简洁概括约束内容 |
| content | text | 是 | 详细内容，支持 Markdown 格式。MVP 阶段仅支持纯文本（不含图片/视频/PDF） |
| category | enum | 是 | 分类：注意事项（一般提醒）、依赖约束（依赖关系变更）、接口变更（API/契约变更）、其他（未归类） |
| status | enum | 是 | 状态：pending（待消费）、consumed（已消费） |
| priority | enum | 是 | 优先级：high（必须处理）、medium（建议处理）、low（参考） |
| expire_days | int | 否 | 超时天数，超过此天数的 Pin 在列表中标记为"已过期"。默认值为 Info 表中 pin_expire_days 配置 |
| consumed_at | timestamp | 自动 | 消费时间，pin_get 成功返回时自动设置 |
| created_at | timestamp | 自动 | 创建时间 |
| updated_at | timestamp | 自动 | 更新时间 |

### Project 逻辑模型

Project 是 Lumina 的**共享聚合根**，同时被 Pin 模块和 Q&A 模块依附。它代表一个被 Lumina 管理的代码项目或知识域。

```json
{
  "id": "雪花 ID",
  "name": "项目名称",
  "alias_name": ["别名1", "别名2", "别名3"],
  "description": "项目描述",
  "created_at": "创建时间",
  "updated_at": "更新时间"
}
```

### Project 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | 雪花 ID | 自动 | 主键 |
| name | string | 是 | 项目名称，正式名称 |
| alias_name | string[] | 否 | 别名列表，JSON 字符串数组。用于项目标识模糊匹配，支持大小写不敏感的部分匹配 |
| description | text | 否 | 项目描述 |
| created_at | timestamp | 自动 | 创建时间 |
| updated_at | timestamp | 自动 | 更新时间 |

### Info 表扩展

复用现有的认证 Info 表（单用户信息表），新增一条系统配置记录：

| key | value | 说明 |
|-----|-------|------|
| pin_expire_days | 30 | Pin 默认超时天数。超过此天数的 Pin 在 pin_list 中显示为"已过期" |

这条记录遵循 Info 表已有的 key-value 模式，在系统初始化时幂等写入。

## 生命周期

```
Agent 调用 pin_push(from_project -> to_project)
        |
        v
   [pending] --- pin_get() 自动消费 ---> [consumed]
        |                                      |
        |  超时提醒（> expire_days）           |  force=true 可重读
        |  Agent 通过 pin_list 可见            |  （不改变状态，仅返回内容）
        |                                      |
        +-- pin_update(status/priority/category) 可修改
```

### 状态说明

| 状态 | 说明 | pin_get 是否返回 | pin_list 是否可见 |
|------|------|-----------------|------------------|
| pending | 待消费，尚未被目标项目 Agent 读取 | ✅ 是（ oldest first ） | ✅ 是 |
| consumed | 已消费，已被目标项目 Agent 读取 | ❌ 否（除非 force=true） | ✅ 是（可查看历史） |

### 关键语义

- **消费不可逆**：pin_get 成功返回后，状态自动变为 consumed，consumed_at 被写入。没有任何操作可以将状态回退到 pending。
- **force 仅重读**：force=true 参数允许 Agent 重新读取一条已消费的 Pin，但**不会修改** status 或 consumed_at。它只返回内容，不改变状态。
- **超时非状态变更**：超过 expire_days 的 Pin 不会自动改变状态，而是在 pin_list 的结果中附加"已过期"提示。过期 Pin 仍然可以被消费（如果状态是 pending），也可以被 force 重读（如果状态是 consumed）。

## 消费模式

### FIFO 队列消费

Pin 采用**先进先出（FIFO）**队列的消费模式：

1. **排序规则**：所有 pending 状态的 Pin 按 created_at ASC（创建时间从早到晚）排序
2. **pin_get 行为**：每次调用 pin_get，返回**最老**的一条 pending Pin，并**自动将其标记为 consumed**
3. **顺序保证**：同一目标项目的 pending Pins 严格按照创建顺序被消费，不允许插队或跳跃
4. **并发安全**：若多个 Agent 同时调用 pin_get，必须保证每条 Pin 只被消费一次（数据库层面保证唯一消费）

### pin_list 的索引浏览

pin_list 支持按索引浏览，用于 Agent 查看待处理或已处理的 Pin 列表：

- 返回结果按 created_at ASC 排序（与消费顺序一致）
- 支持分页或索引偏移，允许 Agent "跳过"查看后面的 Pins
- pin_list 不改变任何 Pin 的状态，它是纯读取操作
- 返回中包含每条 Pin 的"是否已过期"提示

## 项目标识匹配逻辑

Pin 的目标项目匹配采用**两步匹配策略**：

### 第一步：alias_name 模糊匹配

当 Agent 推送 Pin 时，to_project_id 参数可以是一个项目标识字符串。系统首先尝试 alias_name 匹配：

1. 将输入字符串统一转为小写，去除首尾空格
2. 查询 Project 表，遍历所有项目的 alias_name 数组
3. 对 alias_name 中的每个别名，同样转为小写后进行**包含匹配**（即输入字符串包含于别名，或别名包含于输入字符串）
4. 如果匹配到唯一项目，则使用该项目的真实 ID 作为 to_project_id
5. 如果匹配到多个项目，返回匹配歧义提示，要求 Agent 使用精确的项目 ID

### 第二步：精确 ID 回退

如果 alias_name 模糊匹配失败（无匹配或歧义匹配）：

1. 尝试将输入字符串作为精确的 project_id（雪花 ID）进行查询
2. 如果精确匹配成功，直接使用该项目
3. 如果精确匹配也失败，返回"目标项目不存在"错误

### 匹配示例

| 输入 | Project.alias_name | 匹配结果 | 说明 |
|------|-------------------|----------|------|
| "lumina" | ["lumina", "微明"] | ✅ 匹配 | 完全匹配别名 |
| "Lumina" | ["lumina"] | ✅ 匹配 | 大小写不敏感 |
| "微明" | ["lumina", "微明"] | ✅ 匹配 | 中文别名匹配 |
| "lum" | ["lumina"] | ✅ 匹配 | 输入包含于别名（部分匹配） |
| "lumina-web" | ["lumina"] | ✅ 匹配 | 别名包含于输入（反向包含） |
| "backend" | ["lumina", "微明"] | ❌ 不匹配 | 无关联别名 |
| "base" | ["bamboo-base", "base-go"] | ⚠️ 歧义 | 同时匹配两个别名，需精确 ID |

### 边界说明

- alias_name 中的空格和特殊字符保留原样存入，匹配时仅做大小写转换和首尾空格去除
- 极短输入（如单字符）可能产生大量匹配，建议 Agent 使用至少 3 个字符的项目标识
- 项目创建时应填充合理的 alias_name，至少包含项目名称的小写形式和常见缩写

## 超时提醒与 force 重读

### 超时提醒机制

当 Pin 的存活时间（当前时间 - created_at）超过 expire_days 时：

- pin_list 返回中，该 Pin 附加 `expired: true` 提示
- 过期提示文本为"已过期"，供 Agent 判断此约束是否仍然 relevant
- **状态不改变**：过期不会自动将 pending 转为 consumed，也不会删除 Pin
- Agent 可以自行决定如何处理过期 Pin：忽略、手动更新内容、或通过 pin_get 正常消费

### force 重读机制

已消费的 Pin（status=consumed）默认不会被 pin_get 返回。但 Agent 可以通过 force=true 参数重读：

1. Agent 调用 pin_get(id="xxx", force=true)
2. 系统返回该 Pin 的完整内容
3. **状态保持不变**：status 仍为 consumed，consumed_at 不会被修改
4. 重复 force 调用不会产生副作用，结果幂等

force 的典型使用场景：

- Agent 需要回顾之前消费的约束细节
- Agent 在长时间工作后想确认已处理约束的上下文
- Agent 发现某条已消费约束可能与当前任务相关，需要重新查看

## 数据库模型（逻辑模型）

### Pin 表（参考值）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | 雪花 ID | 主键 |
| from_project_id | 雪花 ID | 来源项目 ID |
| to_project_id | 雪花 ID | 目标项目 ID |
| title | varchar(255) | 约束标题 |
| content | text | 详细内容（Markdown） |
| category | varchar(16) | 分类（notice / dependency / api_change / other） |
| status | varchar(16) | 状态（pending / consumed） |
| priority | varchar(16) | 优先级（high / medium / low） |
| expire_days | int | 超时天数 |
| consumed_at | timestamp | 消费时间 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### Project 表（参考值）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | 雪花 ID | 主键 |
| name | varchar(128) | 项目名称（唯一） |
| alias_name | json | 别名列表，JSON 字符串数组 |
| description | text | 项目描述 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### Info 表扩展记录（参考值）

复用已有的 Info 表结构（key-value 模式），新增一行：

| 字段 | 类型 | 说明 |
|------|------|------|
| key | varchar(64) | 配置键，值为 "pin_expire_days" |
| value | varchar(64) | 配置值，值为 "30" |

## 使用场景

### 场景一：跨项目依赖通知

```
Agent A 正在处理 Project-A 的 API 变更：
  -> 发现 Project-B 调用了即将废弃的接口
  -> pin_push(
       from_project_id="Project-A",
       to_project_id="Project-B",
       title="/api/v1/users 接口即将废弃",
       content="Project-A 的 /api/v1/users 将在 v2.0 中移除...",
       category="接口变更",
       priority="high"
     )
  <- Pin 创建成功，状态为 pending
```

### 场景二：启动时消费待处理约束

```
Agent B 启动 Project-B 的新会话：
  -> pin_get(to_project_id="Project-B")
  <- 返回最老的一条 pending Pin（来自 Project-A 的接口变更通知）
     该 Pin 自动被标记为 consumed，consumed_at 写入当前时间
  -> Agent B 根据约束内容调整开发计划
  -> 再次调用 pin_get
  <- 如果没有更多 pending Pin，返回空（或提示"无待处理约束"）
```

### 场景三：超时约束处理

```
Agent B 在 Project-B 工作一段时间后：
  -> pin_list(to_project_id="Project-B")
  <- 返回所有 Pin（pending + consumed），按 created_at 排序
     其中一条 45 天前的 Pin 标记为"已过期"
  -> Agent B 检查过期 Pin 的内容
  -> 发现该约束仍然 relevant（接口变更尚未完成）
  -> 可以选择：
     a) 忽略（约束已处理或不再适用）
     b) 联系 Project-A 的 Agent 确认状态
     c) 如需要回顾细节，调用 pin_get(id="xxx", force=true)
```

## Pin 与 Memory 的区别

| 维度 | Pin | Memory |
|------|-----|--------|
| **核心目的** | 跨项目约束传递 | 项目内决策归档 |
| **数据流向** | 项目 A -> 项目 B（定向） | 项目内创建，项目内查询 |
| **消费模式** | FIFO 队列，顺序消费 | 条件检索，按需查询 |
| **生命周期** | pending -> consumed（不可逆） | active -> deprecated -> archived |
| **时效性** | 强时效性，关注"当前需要处理什么" | 弱时效性，关注"长期需要遵循什么" |
| **内容类型** | 约束、提醒、变更通知 | 架构决策、编码规范、技术选型 |
| **状态数量** | 2 种（pending / consumed） | 3 种（active / deprecated / archived） |
| **消费者** | 目标项目的 Agent（被动接收） | 任何查询的 Agent（主动检索） |
| **更新策略** | 创建后极少修改，消费即结束 | 可多次更新，长期维护 |

简单记忆：

- **Pin** = "别人需要你注意什么"（外来的、待处理的、一次性的）
- **Memory** = "你自己决定要记住什么"（内在的、长期的、可迭代的）

## 已知限制 / 未来考虑

### 当前限制

1. **消费主动性依赖**：Pin 不保证目标项目的 Agent 一定会消费。如果 Agent 从不调用 pin_get，pending 的 Pin 将一直堆积。系统不会主动推送或提醒。

2. **无消费确认机制**：MVP 阶段不支持 ACK/NACK 模式。推送方无法知道 Pin 是否被接收、是否被理解、是否被执行。消费方消费后不会向推送方发送反馈。

3. **alias_name 匹配边界**：alias_name 的模糊包含匹配在处理大小写、空格、连字符、下划线等特殊字符时可能存在边界情况。过于相似的别名可能导致意外的歧义匹配。

4. **单项目消费队列**：同一项目的所有 pending Pins 共享一个 FIFO 队列。高优先级 Pin 无法插队，如果队列前部有一条低优先级但长期未处理的 Pin，高优先级 Pin 必须等待。

5. **无批量消费**：pin_get 每次只返回一条 Pin。如果某项目积累了大量 Pins，Agent 需要多次调用才能清空队列。

### 未来考虑

1. **Q&A-Pin 联动**：Agent 在消费 Pin 后，可能希望将 Pin 内容转发给人类用户确认。未来可考虑支持将 Pin 内容一键创建为 Q&A Session 的消息，打通"约束传递 -> 人工确认"的链路。

2. **消费回调 Webhook**：支持配置 Webhook URL，当某项目的 Pin 被消费时，向该 URL 发送通知。这可以让推送方及时感知约束已被接收。

3. **Pin 依赖链可视化**：当多个项目之间存在复杂的 Pin 传递关系时（A -> B -> C），未来可考虑提供依赖链的可视化展示，帮助理解跨项目约束的传播路径。

4. **优先级队列**：未来可考虑引入优先级队列，让 high 优先级的 Pin 能够优先于 medium/low 被消费，而不是严格 FIFO。

5. **批量 pin_get**：支持一次消费多条 Pin，返回一个 Pin 列表，减少 Agent 的调用次数。
