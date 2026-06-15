---
name: mcp-qa-test
description: 作为 Lumina MCP 测试工程师，对 Q&A 模块进行全链路集成测试与缺陷回归。当用户说"测试 MCP"、"测试 Q&A"、"模拟问答测试"、"复测缺陷"、"回归测试"、"验证修复"或要求模拟 Agent 与用户交互问答时自动激活。覆盖创建会话→推送问题→等待回答→核对返回格式→归档报告全流程。
argument-hint: [ scenario | regression <defect-id> | full ]
allowed-tools: Read, Write, Edit, Bash, AskUserQuestion, mcp__lumina__project_list, mcp__lumina__project_get, mcp__lumina__qa_session_create, mcp__lumina__qa_session_list, mcp__lumina__qa_session_get, mcp__lumina__qa_session_archive, mcp__lumina__qa_push_question, mcp__lumina__qa_push_supplement, mcp__lumina__qa_what_question, mcp__lumina__qa_get_answer, mcp__lumina__qa_reget_answer
---

# MCP Q&A 全链路测试技能 (mcp-qa-test)

作为 Lumina MCP 测试工程师，扮演 **AI Agent 角色**，通过 MCP 工具推送问题，用户在浏览器（Interact 页面）实时回答，验证 Q&A 模块端到端链路与返回格式正确性。

## 触发场景

| 用户意图 | 典型话术 | 执行模式 |
|---------|---------|---------|
| 全链路测试 | "测试一下 MCP"、"模拟问答"、"跑一遍 Q&A" | `full`（4 场景） |
| 单题型测试 | "测试 select/image/options" | `scenario` |
| 缺陷回归 | "复测 D-03"、"验证修复"、"回归一下" | `regression` |

---

## 五阶段标准流程

```text
① 环境准备  →  ② 推送问题  →  ③ 等待回答  →  ④ 核对格式  →  ⑤ 归档报告
```

### 阶段 ① 环境准备

1. `project_list` 获取 `project_id`（通常为 1 个项目）
2. `qa_session_list`(status=active, temporary) 检查是否复用 —— 推荐每次新建，避免历史污染
3. `qa_session_create`(project_id, temporary, title="MCP 测试 · <主题>", agent_name="zcode-tester")
4. **核对 D-01**：返回须包含三段标记
   - `[RESPONSE] 会话创建成功（ID: ...）`
   - `[URL] http://...`
   - `[TIP]` 含 macOS `open` / Linux `xdg-open` / Windows `start` 三平台命令 + 无头环境降级
5. `Bash: open "<URL>"`（macOS）主动为用户打开浏览器

> 能调用 `mcp__lumina__*` 即代表后端环境就绪，无需额外健康检查。

### 阶段 ② 推送问题

按测试场景推送对应题型，**遵循 supplement 协议**：

1. `qa_push_question` 推送问题
2. **核对 D-02**：若传 `supplement: true`，返回**必须**包含 `[REQUIRE_SUPPLEMENT]`（提示先推 supplement 再 get_answer，否则前端阻塞）
3. 若 `supplement: true`，**立即**调用 `qa_push_supplement` 为问题/各选项推送详情
   - 选项级：`option_id` 参数（每个选项单独推送）
   - 问题级：不传 `option_id`（总览/对比类内容）
4. `content_type` 选择：技术说明 → `markdown`；交互预览 → `html`

> ⚠️ 收到 `[REQUIRE_SUPPLEMENT]` 后**不要直接** `qa_get_answer`，否则前端持续加载阻塞用户。

### 阶段 ③ 等待回答

`qa_get_answer` 阻塞循环，按返回标记分支处理：

| 返回标记 | 含义 | 动作 |
|---------|------|------|
| `[ANSWERED]` / `[ANSWER]` | 用户已回答 | 进入阶段 ④ |
| `[NEED_SUPPLEMENT]` + `[USER_NOTE]` | 用户请求补充 | `qa_push_supplement` 推送对应内容 → 再次 `qa_get_answer` |
| `[PENDING]` + `[RETRY] n/max` | 暂未回答 | 重新调用 `qa_get_answer`（无需等待间隔） |
| `[STOPPED]` | 等待过久 | 暂停轮询，告知用户"回答后说继续" → 用户说继续后再次 `qa_get_answer` |
| `[SKIPPED]` | 用户跳过 | 记录，继续下一场景 |

> 单次阻塞约 25 秒，`[RETRY]` 计数到 max（默认 36）转 `[STOPPED]`。无需手动 sleep。

### 阶段 ④ 核对返回格式

对照 **`references/return-format-checklist.md`** 逐题型核对标记。核心验收点：

| 验收点 | 预期 | 涉及缺陷 |
|--------|------|---------|
| `qa_session_create` | `[RESPONSE]`+`[URL]`+`[TIP]` 三平台命令 | D-01 |
| `qa_push_question`(supplement:true) | 含 `[REQUIRE_SUPPLEMENT]` | D-02 |
| `image`/`file` 的 `[DOWNLOAD_PATH]` | **目录路径**（末尾 `/`，无 UUID 文件名） | D-03 |
| `image`/`file` 的 `[TIP]` | 明确 `PATH + FILE_NAME` 拼接 + curl/Invoke 示例 | D-03 |
| `options` 的用户 feedback | 应独立标记（如 `[FEEDBACK]`），不复用 `[SUPPLEMENT]` | D-04 |

**`image`/`file` 额外磁盘验证**：`Bash: ls -lh <DOWNLOAD_PATH>` + `file <path>` 确认文件存在、类型正确、大小合理。

> 测试 OTP 下载链路时，后端可能未运行（HTTP 000），属环境问题，不计入返回格式缺陷。

### 阶段 ⑤ 归档报告

1. `qa_session_archive`(session_id) 归档会话
2. 输出测试报告到 `docs/`：
   - **全链路测试** → `docs/mcp-qa-test-report-YYYY-MM-DD.md`
   - **缺陷回归** → 追加到 `docs/mcp-qa-defects-YYYY-MM-DD.md` 的「复测记录」章节
3. 报告须含：场景结果表 + 验收点核对 + 发现的问题（标注严重程度）

---

## 四大标准测试场景

| 场景 | 题型 | 覆盖能力 | 核心验收 |
|------|------|---------|---------|
| 头脑风暴 | `select` + `supplement` | 选项级 supplement 联动、`[NEED_SUPPLEMENT]` 闭环 | D-02、supplement 多通道 |
| 方案对比 | `options`(pros/cons) | 差异化卡片、用户 feedback 语义 | D-04 |
| 图片上传 | `image` | base64→磁盘→OTP 令牌链路 | D-03 |
| 文件上传 | `file` | `accept` 过滤、任意类型、OTP | D-03 |

不确定题型参数时，先 `qa_what_question`(question_type) 查询用法。

---

## 推送问题模板速查

### select + supplement（头脑风暴）

```text
qa_push_question:
  question_type: select
  content: "## <问题标题>\n\n<场景描述>"
  supplement: true
  options: [{label, description}, ...]  # 4 个左右
```
推送后**必须**为每个 option 推 supplement（技术对比表 + 适用/不适用场景）。

### options（方案对比）

```text
qa_push_question:
  question_type: options
  content: "<选择提示>"
  supplement: true
  options: [{label, description, pros:[...], cons:[...]}]  # pros/cons 必填
```

### image / file（媒体上传）

```text
qa_push_question:
  question_type: image  # 或 file
  content: "请上传..."
  config: {maxImages: 3, multiple: true}  # image
         {accept: [".json",".yaml"], maxFiles: 3}  # file
```
无需 supplement。返回后验证 `[DOWNLOAD_PATH]` 目录化 + 磁盘文件。

---

## supplement 内容写作规范

- **选项级**：单方案深度（包体积表、代码示例、适用场景）
- **问题级**：横向总对比（决策矩阵表、决策树、综合推荐）
- Markdown 支持：表格、代码块、Mermaid、KaTeX
- 收到 `[NEED_SUPPLEMENT]` + `[USER_NOTE]` 时，按 note 针对性补充（如"需要总对比"→ 推问题级总表）

---

## 缺陷回归模式

当 argument 为 `regression <defect-id>`（如 `regression D-03`）：

1. 只跑该缺陷涉及的最小题型场景
2. 严格对照缺陷报告的「期望」段落核对返回
3. 报告写入 `docs/mcp-qa-defects-YYYY-MM-DD.md`：
   - 修复 → 标记 ✅ + 实际返回证据
   - 未修复 → 标记 ❌ + 复现返回 + 用户反馈原话
4. 发现新问题 → 分配新编号（D-05、D-06...）

---

## 注意事项

1. **角色定位**：你是 AI Agent 端，用户是浏览器端。推送后用 `qa_get_answer` 等待，让用户在浏览器操作。
2. **supplement 时序**：`supplement: true` → 见 `[REQUIRE_SUPPLEMENT]` → 推 supplement → 再 `get_answer`，不可跳过。
3. **不臆测返回**：实际返回与期望不符时，原样记录返回内容作为缺陷证据，不要"修正"。
4. **前端 bug 区分**：返回格式正确但用户反馈前端展示异常，属前端缺陷（如 D-04/D-05），单独编号。
5. **环境问题豁免**：OTP 下载 HTTP 000、WebSocket 断连等环境问题不计入返回格式缺陷。
6. **完整核对表**：14 题型的完整返回标记核对，读取 `references/return-format-checklist.md`。
