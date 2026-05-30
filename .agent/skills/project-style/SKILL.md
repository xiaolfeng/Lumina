---
name: project-style
description: 统一项目多层架构写作风格，明确 handler、logic、repository（含 cache）职责边界与协作规则，避免跨层越界。
argument-hint: [ topic-or-module ]
allowed-tools: Read, Write, Edit, AskUserQuestion
---

# Project Writing Style (project-style)

用于规范本项目在多层架构下的代码写作风格与职责隔离。目标是：

1. 每层只做自己该做的事。
2. 层与层之间只通过稳定接口协作。
3. 禁止“顺手做掉”导致的跨层污染。

---

## 适用范围

- `handler`（接口/传输层）
- `logic`（业务编排层）
- `repository`（数据访问层）
- `repository/cache`（缓存读写与失效策略）

---

## 架构总览

```text
HTTP / RPC Request
   -> handler
   -> logic
   -> repository (+ cache)
   -> DB / Redis / External Service
```

### 依赖方向（只能向下）

- `handler` -> `logic`
- `logic` -> `repository`
- `repository` -> `cache/client/db`

禁止反向依赖与旁路依赖：

- `handler` 直接调用 `repository`（禁止）
- `handler` 直接操作 `cache`（禁止）
- `logic` 直接解析 HTTP 参数（禁止）
- `repository` 承载业务决策（禁止）

---

## 分层职责定义

## 1) handler 层（接口适配层）

**负责：**

- 接收请求（HTTP/RPC）并做基础绑定与格式校验。
- 将 transport DTO 转换为 logic 入参。
- 调用 logic 并将结果映射为响应 DTO。
- 统一处理错误到协议状态码/错误码。

**不负责：**

- 业务规则判断（如权限矩阵、状态流转、价格计算）。
- 持久化细节或缓存策略。
- 事务控制。

**写作要点：**

- 函数要短，重点是“转换 + 调用 + 映射”。
- 不在 handler 中出现 repository 类型。
- 错误文案面向接口消费者，不泄漏底层细节。

---

## 2) logic 层（业务编排层）

**负责：**

- 承载核心业务用例（use case）。
- 组合多个 repository 完成业务流程。
- 执行业务校验、幂等、事务边界、领域规则。
- 定义稳定的输入输出模型（面向上层/下层）。

**不负责：**

- HTTP 参数提取、响应序列化。
- SQL/Redis 语句细节。
- 框架耦合代码（Gin Context、GORM Session 细节）外泄到方法签名。

**写作要点：**

- 方法名用业务语义（如 `CreateUser`, `BanPlayer`, `RefreshProfile`）。
- 返回错误使用可判定类型（业务错误 vs 系统错误）。
- 业务流程可读性优先：按“校验 -> 执行 -> 收敛结果”组织。

---

## 3) repository 层（数据访问层）

**负责：**

- 面向实体/聚合提供稳定 CRUD 接口。
- 隔离数据库实现细节（GORM/SQL/索引策略）。
- 提供查询条件、分页、排序等数据访问能力。

**不负责：**

- 业务流程编排。
- 业务策略分支（例如“是否允许创建”）。
- transport DTO 适配。

**写作要点：**

- 输入尽量是明确 query/filter 结构，而非无序参数堆。
- repository 返回领域可用的数据模型，不返回 HTTP 语义。
- 错误包装要保留可观测信息（操作、主键、关键参数）。

---

## 4) repository/cache 层（缓存策略子层）

**负责：**

- cache-aside / read-through 等缓存模式落地。
- key 设计、TTL 管理、失效策略。
- 缓存命中与回源逻辑。

**不负责：**

- 业务决策（例如封禁判定、权限判定）。
- HTTP/RPC 协议转换。

**写作要点：**

- Key 命名统一：`<domain>:<entity>:<id|hash>`。
- 所有写操作明确失效点（写后删缓存、双删或事件失效）。
- 允许缓存失败降级，但必须记录观测日志。

---

## 越界禁止清单（Hard Rules）

1. **MUST**: handler 只能依赖 logic，不得直接依赖 repository/cache。
2. **MUST**: logic 只能通过 repository 接触数据源，不得拼接 SQL/Redis 命令。
3. **MUST**: repository 不承载业务分支，不做“是否允许”的业务判断。
4. **MUST**: cache 只做性能优化，不得成为业务正确性的唯一来源。
5. **NEVER**: 在下层返回上层协议对象（例如 repository 返回 HTTP 状态码）。
6. **NEVER**: 跨层复用“顺手函数”破坏边界（例如 handler 调 util 直连 DB）。

---

## 分层目录建议

```text
internal/
  handler/
    user_handler.go
  logic/
    user_logic.go
  repository/
    user_repository.go
    cache/
      user_cache_repository.go
  model/
    dto/
    entity/
```

说明：

- `handler`：协议适配与响应格式。
- `logic`：业务用例。
- `repository`：持久化与查询。
- `repository/cache`：缓存实现细节。

---

## 示例 1：正确分层（伪代码）

```go
// handler
func (h *UserHandler) Create(ctx *gin.Context) {
    req := bindCreateUserRequest(ctx)
    out, err := h.userLogic.CreateUser(ctx, req.ToLogicInput())
    renderCreateUserResponse(ctx, out, err)
}

// logic
func (l *UserLogic) CreateUser(ctx context.Context, in CreateUserInput) (CreateUserOutput, error) {
    if err := l.validator.ValidateCreateUser(in); err != nil {
        return CreateUserOutput{}, err
    }
    if exists, _ := l.userRepo.ExistsByUsername(ctx, in.Username); exists {
        return CreateUserOutput{}, ErrUsernameTaken
    }
    user, err := l.userRepo.Create(ctx, in.ToEntity())
    if err != nil {
        return CreateUserOutput{}, err
    }
    return ToCreateUserOutput(user), nil
}

// repository/cache
func (r *UserRepository) GetByID(ctx context.Context, id int64) (User, error) {
    if user, ok := r.cache.GetUser(ctx, id); ok {
        return user, nil
    }
    user, err := r.db.GetUserByID(ctx, id)
    if err != nil {
        return User{}, err
    }
    _ = r.cache.SetUser(ctx, id, user)
    return user, nil
}
```

---

## 示例 2：错误分层（反例）

```go
// 错误：handler 直接依赖 repository 并做业务判断
func (h *UserHandler) Ban(ctx *gin.Context) {
    user := h.userRepo.GetByID(ctx, id) // 越界：handler -> repository
    if user.Role != "ADMIN" {          // 越界：业务规则放在 handler
        ctx.JSON(403, ...)
        return
    }
    ...
}
```

---

## 评审检查清单（PR Checklist）

- [ ] handler 仅做协议适配，无业务规则分支。
- [ ] logic 完整表达业务用例，不含 transport/db 框架泄漏。
- [ ] repository 仅做数据访问，不承载业务决策。
- [ ] cache 策略与失效点明确，失败可降级。
- [ ] 依赖方向单向向下，无跨层旁路调用。

---

## 何时需要 AskUserQuestion

当以下信息不清晰时必须询问：

1. 该需求属于“新增用例”还是“扩展现有用例”？
2. 事务边界放在 logic 还是 repository（项目约定）？
3. 缓存策略采用 cache-aside 还是 read-through？
4. 错误语义是否有统一业务错误码体系？

---

## 一句话原则

> handler 负责“说人话（协议）”，logic 负责“做决策（业务）”，repository 负责“拿数据（存储）”，cache 负责“提性能（加速）”。
