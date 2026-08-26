---
name: lumina-qa
description: Lumina Q&A 向用户发起结构化富交互问答并阻塞等待裁决。需要方案确认、技术选型、代码评审、计划审批、收集文件/图片或打分时使用；不要用猜测代替用户决策。
license: MIT
compatibility: Requires Lumina MCP (Streamable HTTP) and network access to the Lumina instance.
metadata:
  author: lumina
  version: "0.1.0"
argument-hint: [ question-type | session-id ]
allowed-tools: Read, Write, Edit, Bash, AskUserQuestion, mcp__lumina__project_get, mcp__lumina__project_list, mcp__lumina__project_create, mcp__lumina__qa_session_create, mcp__lumina__qa_session_list, mcp__lumina__qa_session_get, mcp__lumina__qa_session_archive, mcp__lumina__qa_what_question, mcp__lumina__qa_push_question, mcp__lumina__qa_push_supplement, mcp__lumina__qa_get_answer, mcp__lumina__qa_reget_answer, mcp__lumina__qa_cancel_question
---

# Lumina Q&A 富交互问答与决策协同指南 (lumina-qa)

用于指导 AI Agent 在面临多方案决策、关键参数确认或需要多模态输入时，通过 Lumina Q&A 模块向用户发起结构化、富交互的问答，并在浏览器端获得高质量用户裁决。

全局红线与项目解析见 [`../_shared/lumina-core-rules.md`](../_shared/lumina-core-rules.md) 与 [`../_shared/project-resolver.md`](../_shared/project-resolver.md)。Preview 挂载契约见 [`../_shared/preview-qa-contract.md`](../_shared/preview-qa-contract.md)。

## 按需加载

| 需要 | 读取 |
|---|---|
| 14 种题型选型 | [`references/overview.md`](./references/overview.md) |
| 某题型字段与返回标记 | `references/<question_type>.md`，或先调 `qa_what_question` |
| 完整会话走查 | [`examples/select-with-supplement.md`](./examples/select-with-supplement.md) |
| 等待状态机样例 | [`examples/answer-markers.md`](./examples/answer-markers.md) |

---

## 🎯 核心定位与适用场景

- **何时使用**：
  - 遇到多个实现方案需用户拍板（如架构选型、库依赖对比）。
  - 需要用户输入敏感信息、配置参数、环境标识。
  - 需要代码差异比对（Diff 评审）、分段实施计划审批（Plan）。
  - 需要用户上传本地文件、配置文件或截图（File/Image）。
  - 任务结束需满意度或多维度打分反馈（Rate/Slider/Rank）。
- **何时不要使用**：
  - 琐碎且明确的代码单行修改（直接执行即可，无需打扰用户）。
  - 纯日志输出（直接输出给用户，无需建 Q&A 会话）。

---

## 🔄 Q&A 标准六阶段生命周期

```text
① 项目上下文解析  ──▶  ② 会话复用/创建  ──▶  ③ 题型查询与构造
                                                    │
                                                    ▼
⑥ 会话归档/清理   ◀──  ⑤ 阻塞等待与状态机  ◀──  ④ 问题与详情推送
```

### 阶段 ① 项目上下文解析
参照 `../_shared/project-resolver.md` 标准四步法，通过 `project_get` / `project_list` 确定目标项目的 `project_id`。

### 阶段 ② 会话复用与创建
1. 调用 `qa_session_list(status="active", session_type="temporary")` 检查是否有可复用的活跃会话。
2. 若存在活跃会话，直接复用其 `session_id`；若无则调用 `qa_session_create` 创建新会话：
   ```json
   {
     "project_id": "<project_id>",
     "session_type": "temporary",
     "title": "方案选型与决策问答",
     "agent_name": "lumina-agent"
   }
   ```
3. **[CRITICAL] 主动打开浏览器**：`qa_session_create` 返回的内容包含 `[URL] http://...` 交互地址。Agent **必须立即通过 Bash 执行系统打开命令为用户弹出浏览器**，严禁让用户人工手动输入/复制链接：
   - **macOS**: `open "<URL>"`
   - **Linux**: `xdg-open "<URL>"`
   - **Windows**: `start "<URL>"`
   > 仅在明确处于无头服务器/容器等无 GUI 环境下，才降级为在对话中输出可点击 Markdown 链接。

### 阶段 ③ 题型查询与参数构造
- **[RULE] 严禁猜测题型与字段**：在推送前，必须查阅对应题型的权威定义。
- 快捷查阅：调用 `qa_what_question(question_type="<类型>")` 获取官方 JSON 示例。
- 本地参考：查阅 `references/` 目录下各题型的精细拆解文档（详见文末导航）。

### 阶段 ④ 问题推送与详情注入 (Supplement)
1. **推送问题**：调用 `qa_push_question`。
2. **Supplement 注入**：若设置了 `supplement: true` 或选项需要深度技术说明，**必须立即**调用 `qa_push_supplement`：
   - **Markdown 详情**：`content_type="markdown"`，支持表格、代码块、Mermaid、KaTeX。
   - **Preview 挂载**：`content_type="preview"`，content 必须严格遵守 `../_shared/preview-qa-contract.md` 规范。
   - **作用域**：传 `option_id` 关联具体选项；不传 `option_id` 关联问题全局。

### 阶段 ⑤ 阻塞等待与状态机流转 (qa_get_answer)
调用 `qa_get_answer(session_id)` 阻塞等待用户提交（单次约 25s），根据返回文本标记执行相应分支：

| 返回标记 | 含义与场景 | Agent 响应动作 |
|---|---|---|
| `[ANSWERED]` / `[ANSWER]` | 用户已成功提交回答 | 解析答案文本与结构，进入下一步业务处理。 |
| `[NEED_SUPPLEMENT]` + `[USER_NOTE]` | 用户在界面点击请求补充更多信息 | 根据 `[USER_NOTE]` 补充内容，调用 `qa_push_supplement` 后**再次调用 `qa_get_answer`**。 |
| `[PENDING]` | 暂无回答，单次阻塞超时 | **立即重新调用 `qa_get_answer`** 继续阻塞等待（严禁 sleep！）。 |
| `[STOPPED]` | 等待过久，进入休眠保护 | 停止调用任何工具，向用户输出提示：“如果您已在浏览器回答，请回复「继续」，我将获取您的答案。”收到回复后再次调用 `qa_get_answer`。 |
| `[SKIPPED]` | 用户主动跳过该题 | 记录跳过状态，继续后续逻辑。 |

### 阶段 ⑥ 任务收尾与归档
- 所有问答完成并得出结论后，调用 `qa_session_archive(session_id)` 将会话置为只读。
- 若中途决策变更需撤销未回答问题，调用 `qa_cancel_question(session_id, question_id)`。

---

## 📚 14 种题型精细拆解导航 (References)

按交互族系划分，查阅 `references/` 下独立文档：

| 族系 | 题型名称 | 核心适用场景 | 详情文档 |
|---|---|---|---|
| **选型决策** | `overview` | 14 种题型全景概览与快速选型树 | [`references/overview.md`](./references/overview.md) |
| **选择类** | `select` | 单选方案、环境选择（支持选项级 Supplement） | [`references/select.md`](./references/select.md) |
| | `multi-select` | 多选功能模块、权限分配 | [`references/multi-select.md`](./references/multi-select.md) |
| | `options` | 差异化对比专用（强制要求 pros/cons 利弊分析） | [`references/options.md`](./references/options.md) |
| **输入类** | `text` | 自由文本输入、场景描述（单行/多行） | [`references/text.md`](./references/text.md) |
| | `boolean` | 高危操作确认、二选一开关 | [`references/boolean.md`](./references/boolean.md) |
| | `code` | 代码片段、正则表达式、配置编辑器 | [`references/code.md`](./references/code.md) |
| | `image` | 架构图、UI 截图上传（一次性 OTP 令牌下载） | [`references/image.md`](./references/image.md) |
| | `file` | 配置文件、数据包上传（一次性 OTP 令牌下载） | [`references/file.md`](./references/file.md) |
| **展示类** | `diff` | 代码修改前后差异审阅（返回修改后代码） | [`references/diff.md`](./references/diff.md) |
| | `plan` | 分段实施计划审批（支持分段修订意见） | [`references/plan.md`](./references/plan.md) |
| | `review` | 逐段设计与文档评审（approve/revise） | [`references/review.md`](./references/review.md) |
| **评分类** | `slider` | 数值区间滑动选择（满意度、百分比） | [`references/slider.md`](./references/slider.md) |
| | `rank` | 需求或任务优先级拖拽排序 | [`references/rank.md`](./references/rank.md) |
| | `rate` | 多维度独立星级打分 | [`references/rate.md`](./references/rate.md) |

---

## ⛔ 核心红线与避坑指南 (Hard Rules)

1. **MUST**: 会话创建返回 `[URL]` 后，必须立即通过 Bash 执行 `open "<URL>"`（macOS）/ `xdg-open` / `start` 主动打开浏览器，严禁让用户手动复制输入。
2. **MUST**: 推送前先核对题型字段（查阅 `references/` 或调用 `qa_what_question`），严禁自造字段。
3. **MUST**: 设为 `supplement: true` 时，推送问题后必须立即调用 `qa_push_supplement`，防止前端持续加载。
4. **NEVER**: 严禁使用 `bash sleep` 或外部死循环进行休眠等待。
5. **NEVER**: 严禁使用 `qa_reget_answer` 进行轮询等待（`qa_reget_answer` 仅用于已回答内容的重取或下载令牌失效时刷新）。
6. **MUST**: 收到 `[PENDING]` 时立即重试 `qa_get_answer`；收到 `[STOPPED]` 时停下提示用户。
