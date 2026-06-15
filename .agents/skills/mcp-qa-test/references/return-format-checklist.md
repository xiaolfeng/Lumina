# 14 题型返回格式核对表

Q&A 模块 `qa_get_answer` / `qa_reget_answer` 返回的格式标记核对清单。按题型分组，每个标记说明来源（用户输入 vs Agent 推送）和核对要点。

> 标记来源约定：`[ANSWER]` 系列来自用户作答；`[SUPPLEMENT]` 来自 Agent 通过 `qa_push_supplement` 推送的内容；`[DESCRIPTION]`/`[OPTION_DESCRIPTION]` 来自推送问题时传入的字段。

---

## 选择类

### select（单选）

| 标记 | 来源 | 说明 |
|------|------|------|
| `[ANSWER] 用户选择：<label>` | 用户 | 选中选项的 label |
| `[DESCRIPTION]` | push 的 description | 问题级描述，为空则不输出 |
| `[OPTION_DESCRIPTION]` | option.description | 选中选项的描述，为空则不输出 |
| `[SUPPLEMENT]` | Agent push_supplement | 选中选项的 supplement，为空则不输出 |

**核对要点**：label 正确反查；description 区分问题级 vs 选项级；supplement 仅输出选中项。

### multi-select（多选）

| 标记 | 来源 | 说明 |
|------|------|------|
| `[ANSWER] 用户选择 N 项` | 用户 | 选择总数 |
| `[DESCRIPTION]` | push 的 description | 问题级描述 |
| `---` | 系统分隔符 | 多选项之间分隔 |
| `[OPTION] <label>` | 用户 | 每个选中项 |
| `[OPTION_DESCRIPTION]` | option.description | 该项描述 |
| `[SUPPLEMENT]` | Agent push_supplement | 该项 supplement |
| `[OPTION] __other__` + `[SUPPLEMENT]` | 用户"其他"输入 | 自定义文本 |

**核对要点**：多选项用 `---` 分隔不杂糅；`__other__` 段单独处理。

### options（差异化对比，pros/cons 必填）

| 标记 | 来源 | 说明 |
|------|------|------|
| `[ANSWER] <label>` | 用户 | 选中方案 |
| `[DESCRIPTION] <选中项 description>` | option.description | |
| `[SUPPLEMENT] <feedback>` | ⚠️ 用户填写 | **D-04 缺陷点**：当前复用此标记装用户"选择理由" |

**核对要点（D-04）**：用户 feedback 应使用独立标记（如 `[FEEDBACK]`），不应复用 `[SUPPLEMENT]`。当前实现中两者混用，属待修复缺陷。

---

## 输入类

### text（文本输入）

| 标记 | 来源 | 说明 |
|------|------|------|
| `[ANSWER] <文本>` | 用户 | 用户输入内容 |

**核对要点**：原样返回，multiline 文本保留换行。

### boolean（布尔确认）

| 标记 | 来源 | 说明 |
|------|------|------|
| `[ANSWER] 是` 或 `[ANSWER] 否` | 用户 | 二选一 |

### code（代码输入）

| 标记 | 来源 | 说明 |
|------|------|------|
| `[ANSWER] <代码>` | 用户 | 用户输入代码 |
| `[LANGUAGE] <语言>` | config.language | 与推送时 config.language 一致 |

**核对要点**：language 原样输出（如 `regex`/`go`/`python`）。

### image（图片上传）⭐ D-03 核心验证点

| 标记 | 来源 | 说明 |
|------|------|------|
| `[ANSWER] 用户已上传内容` | 用户 | 固定文案 |
| `---` | 系统分隔符 | 每张图之间 |
| `[FILE_NAME] <文件名>` | 用户 | 原始文件名（含扩展名） |
| `[DOWNLOAD_PATH] <目录>/` | 系统存储 | **目录路径**，末尾带 `/`，不含 UUID 文件名 |
| `[DOWNLOAD_URL]` + OTP 链接列表 | 系统生成 | `http://<domain>/api/v1/qa/download/cs_<token>`，每行一个 |
| `[IMPORTANT]` | 系统提示 | 一次性令牌使用即失效提示 |
| `[TIP]` | 系统提示 | `PATH + FILE_NAME` 拼接规则 + curl/Invoke 示例 |
| `[GIT_TIP]` | 系统提示 | `.lumina/cache/` 加入 .gitignore |

**核对要点（D-03）**：
1. `[DOWNLOAD_PATH]` 必须是目录（末尾 `/`），不是完整文件路径（不含 UUID）
2. `[TIP]` 必须说明最终路径 = `DOWNLOAD_PATH + FILE_NAME`
3. 磁盘验证：`ls -lh <PATH>` 确认 UUID 文件存在，`file <path>` 确认类型正确

**缺陷对照（D-03 修复前 vs 修复后）**：
- ❌ 修复前：`[DOWNLOAD_PATH] .lumina/cache/<sid>/<uuid>`（无扩展名，curl 保存不可用）
- ✅ 修复后：`[DOWNLOAD_PATH] .lumina/cache/<sid>/`（目录）+ TIP 拼接规则

### file（文件上传）⭐ 同 image，D-03 同样适用

格式与 image 完全一致，区别仅在 `accept` 过滤。核对要点相同。

---

## 展示类（供审阅决策）

### diff（差异对比）

| 标记 | 来源 | 说明 |
|------|------|------|
| approve 批准 | | |
| `[ANSWER] 用户已批准该修改` | 用户 | |
| `[FINAL] <config.after 代码>` | 系统 | 批准时为修改后代码 |
| reject 拒绝 | | |
| `[ANSWER] 用户已拒绝该修改` | 用户 | |
| `[FEEDBACK] <原因>` | 用户 | 拒绝原因 |
| edit 编辑后提交 | | |
| `[ANSWER] 用户修改后提交` | 用户 | |
| `[FINAL] <用户修改后代码>` | 用户 | 用户在 diff 基础上改的 |

**核对要点**：仅 approve/edit 有 `[FINAL]`；reject 无。approve 的 FINAL = config.after，edit 的 FINAL = 用户改后。

### plan（方案展示）

| 标记 | 来源 | 说明 |
|------|------|------|
| approve 批准 | | |
| `[ANSWER] 用户已批准该计划` | 用户 | |
| `[PLAN_DETAIL]` + 编号列表 | 系统 | 完整计划各 section |
| reject 拒绝 | | |
| `[ANSWER] 用户已拒绝该计划` | 用户 | |
| `[FEEDBACK] <原因>` | 用户 | |
| revise 需修订 | | |
| `[ANSWER] 用户要求修改该计划` | 用户 | |
| `[REVISIONS]` + `[<sectionId>] 意见` | 用户 | 每 section 修订意见 |
| `[FEEDBACK] <整体反馈>` | 用户 | 可选 |

### review（内容审阅）

| 标记 | 来源 | 说明 |
|------|------|------|
| `[ANSWER] 用户批准了该修改` | 用户 | approve |
| `[ANSWER] 用户要求修改` | 用户 | revise |
| `[FEEDBACK] <修改意见>` | 用户 | revise 时 |

**核对要点**：review 只有 approve/revise（无 reject）；approve 返回简洁不重复标记。

---

## 评分类

### slider（滑块评分）

| 标记 | 来源 | 说明 |
|------|------|------|
| `[ANSWER] <数值>` | 用户 | 用户拖动选择的数值 |

### rank（排序，需 options）

| 标记 | 来源 | 说明 |
|------|------|------|
| `[ANSWER] 1. <label> → 2. <label> → ...` | 用户 | 按优先级从高到低 |

**核对要点**：返回 label（反查后可读名称），非选项 ID。

### rate（星级评分，需 options，每项独立打分）

| 标记 | 来源 | 说明 |
|------|------|------|
| `[ANSWER] <label1>: <分>, <label2>: <分>` | 用户 | 每个选项独立评分 |

**核对要点**：每个 option 独立打分；无 options 会导致空评分错误。

---

## 补充：通用流程标记（非题型，来自 qa_get_answer）

| 标记 | 含义 | 处理 |
|------|------|------|
| `[NEED_SUPPLEMENT]` + `[TARGET]` + `[USER_NOTE]` | 用户请求补充内容 | 按 note `qa_push_supplement` → 再次 `get_answer` |
| `[PENDING]` + `[RETRY] n/max` | 暂未回答 | 重新调用 `qa_get_answer` |
| `[STOPPED]` | 等待过久 | 暂停轮询，告知用户"回答后说继续" |
| `[SKIPPED]` | 用户跳过 | 记录，继续下个场景 |
| `[RESPONSE]` | 工具操作确认 | session_create/push_question 等的成功回执 |

---

## qa_push_question 返回标记

| 标记 | 含义 | 触发条件 |
|------|------|---------|
| `[RESPONSE] 问题已推送（ID: ...）` | 推送成功 | 总是输出 |
| `[OPTIONS]` + `label → id` 列表 | 选项 ID 映射 | 有 options 时 |
| `[REQUIRE_SUPPLEMENT]` | 提醒先推 supplement | **`supplement: true` 时必须出现（D-02）** |
| `[TIP] 使用 qa_get_answer 等待用户回答` | 操作提示 | 总是输出 |

## qa_session_create 返回标记

| 标记 | 含义 |
|------|------|
| `[RESPONSE] 会话创建成功（ID: ...）` | 创建成功 |
| `[URL] http://...` | Interact 页面地址 |
| `[TIP]` 三平台 open 命令 | **D-01：必须出现** |
