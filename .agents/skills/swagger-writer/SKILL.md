---
name: swagger-writer
description: 专门为 Go handler 编写和补全 godoc-swagger 注释，严格对齐本项目 swag 风格与字段顺序。
argument-hint: [ handler-or-endpoint ]
allowed-tools: Read, Write, Edit, AskUserQuestion
---

# Swagger Writer Skill (swagger-writer)

用于给 `internal/handler/*.go` 中的接口方法编写或修正 Godoc + Swag 注释，目标是直接可用于 `swag init` 生成文档。

参考基准：`internal/handler/user.go` 的 `UserCurrent` 注释风格。

---

## 触发场景

当用户出现以下意图时自动启用：

- “给 handler 写 swagger 注释”
- “补齐 godoc”
- “修复 swag 注解”
- “给某个接口补文档”

---

## 核心目标

1. 只改注释，不改业务逻辑。
2. 注释顺序、字段语义、路径方法全部准确。
3. 与现有项目风格一致（中文描述、`xBase.BaseResponse` 响应模型、`@Tags` 命名习惯）。

---

## 标准工作流

1. 读取目标 handler 文件与对应函数。
2. 若路由/请求模型不明确，继续读取 router、dto、entity、logic 相关文件。
3. 先确定接口事实信息：
    - 路径（`@Router`）
    - 方法（GET/POST/PUT/DELETE...）
    - 入参位置（path/query/body/header）
    - 成功响应数据类型（`BaseResponse{data=...}`）
    - 失败状态码范围
4. 按固定顺序生成注释块并写回函数前。
5. 自检：注释内容与函数行为一致，不捏造字段。

---

## 注释顺序规范（必须）

按以下顺序输出：

```go
// FuncName 中文一句话说明
//
// @Summary     [玩家/管理/超管] 接口名
// @Description 说明接口做什么，输入输出是什么
// @Tags        模块接口
// @Accept      json
// @Produce     json
// @Param       ...
// @Success     200   {object}  xBase.BaseResponse{data=entity.XXX}  "成功"
// @Failure     400   {object}  xBase.BaseResponse                    "请求参数错误"
// @Failure     401   {object}  xBase.BaseResponse                    "未授权"
// @Failure     403   {object}  xBase.BaseResponse                    "无权限"
// @Failure     404   {object}  xBase.BaseResponse                    "资源不存在"
// @Router      /api/v1/xxx [GET]
```

说明：

- 无请求参数时可以省略 `@Param`。
- `@Failure` 只保留真实可能出现的状态码，不强行凑齐。
- `@Summary` 推荐使用 `[玩家/管理/超管] 动作` 结构，例如 `[玩家] 用户信息`。

---

## @Param 写法规则

### Path 参数

```go
// @Param id path int true "用户ID"
```

### Query 参数

```go
// @Param page query int false "页码"
// @Param size query int false "每页数量"
```

### Body 参数

```go
// @Param request body dto.CreateUserRequest true "创建用户请求"
```

### Header 参数（按需）

```go
// @Param Authorization header string true "Bearer Access Token"
```

---

## 响应模型规范

本项目优先使用：

```go
// @Success 200 {object} xBase.BaseResponse{data=entity.User} "成功"
```

若接口返回列表，按真实结构填写 data：

- `data=[]entity.User`
- `data=dto.UserListResponse`

禁止写与真实返回不一致的结构。

---

## 文案风格

- 注释文案使用中文，简洁、可读。
- `@Description` 说明“依据什么入参，返回什么结果”。
- `@Tags` 统一使用“中文模块 + 接口”，例如：`用户接口`、`认证接口`。
- 保持与文件内其他注释风格一致，不混用中英标点格式。

---

## 质量检查清单（写完必须自检）

- [ ] 注释块位于函数定义正上方。
- [ ] `@Router` 的路径与方法和实际注册一致。
- [ ] `@Param` 名称、位置、必填状态与绑定结构一致。
- [ ] `@Success` 的 `data=` 类型与函数真实输出一致。
- [ ] 未改动函数逻辑、返回流程、错误处理。

---

## AskUserQuestion 使用时机

在以下信息无法从代码推断时，使用 `AskUserQuestion`：

1. 同一函数被多个路由复用，无法确定主路由。
2. 返回数据模型存在多个候选（entity/dto 均可能）。
3. 业务要求的失败码文案有团队约定但代码中未体现。

优先先读代码再问，禁止在可推断场景下直接提问。

---

## 示例（贴合本项目风格）

```go
// UserCurrent 获取用户的信息
//
// @Summary     [玩家] 用户信息
// @Description 根据 AT 获取用户信息，获取到本程序的用户信息
// @Tags        用户接口
// @Accept      json
// @Produce     json
// @Success     200   {object}  xBase.BaseResponse{data=entity.User}  "登录成功"
// @Failure     400   {object}  xBase.BaseResponse                    "请求体格式不正确"
// @Failure     401   {object}  xBase.BaseResponse                    "用户名或密码错误"
// @Failure     403   {object}  xBase.BaseResponse                    "用户已禁用或账户已锁定"
// @Failure     404   {object}  xBase.BaseResponse                    "用户不存在"
// @Router      /api/v1/user/info [GET]
func (h *UserHandler) UserCurrent(ctx *gin.Context) {}
```

---

一句话准则：先读代码定事实，再写注释补表达；注释必须准确，不允许“想当然”。
