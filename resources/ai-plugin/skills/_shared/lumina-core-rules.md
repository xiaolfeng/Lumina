# Lumina MCP 全局核心原则与红线约束 (Lumina Core Rules)

在使用 Lumina MCP 工具链时，所有 AI Agent 必须遵循以下全局核心原则与红线禁令。

## 技能怎么拆

按任务读对应技能的 `SKILL.md`，细节进该技能的 `references/` 与 `examples/`：

| 任务 | 技能 |
|---|---|
| 向用户提问并等待裁决 | `lumina-qa` |
| 展示原生 HTML/CSS/JS 或浏览器端 React/Vue 原型 | `lumina-preview` |
| 把核对过的预览晋升为持久页面 | `lumina-pages` |
| 跨项目约束推送 / 消费 | `lumina-pin` |
| 读已生成的仓库 Wiki | `lumina-repowiki` |

项目解析是公共步骤，见 [`project-resolver.md`](./project-resolver.md)。Preview 挂到 Q&A 的字段契约见 [`preview-qa-contract.md`](./preview-qa-contract.md)。


---

## 🧭 全局核心原则

1. **服从当前任务主线**：
   - MCP 工具是辅助任务推进的手段。禁止因为工具存在而主动创建无意义的项目、问答会话、前端预览或跨项目消息。
2. **不以猜测代替决策**：
   - 遇到重大技术选型、不可逆操作（如数据清理、破坏性重构）、多分支方案抉择时，使用 Q&A 模块向用户提问，由用户明确裁决。
3. **严格错误闭环**：
   - 当工具返回 `isError: true` 或输出明确的错误信息（如“参数缺失”、“会话不存在”）时，必须**立即根据错误信息修正参数或向用户如实报错**，严禁忽略错误并假装调用成功。
4. **单向状态机流转**：
   - Q&A 回答、Preview 文件核对、Pin 队列消费等均具有确定的生命周期状态机，必须严格按顺序推进，不得跳步。

---

## ⛔ 核心红线清单 (Hard Rules Checklist)

| 类别 | 规则等级 | 约束内容 |
|---|---|---|
| **工具调用** | **NEVER** | 严禁自创或猜测不存在的 MCP 工具名称、参数名称或题型枚举。 |
| **主动打开链接** | **MUST** | 凡是会话创建（Q&A `[URL]`）或原型预览（Preview `preview_url`）返回了浏览器链接，**必须立即通过 Bash 执行系统打开命令 (`open` / `xdg-open` / `start`) 主动弹出浏览器**，严禁让用户人工复制/输入链接。 |
| **Q&A 阻塞** | **NEVER** | 严禁使用 `bash sleep`、休眠命令或死循环脚本等待用户回答。`qa_get_answer` 本身内置安全阻塞等待机制（约 25s），无需外部休眠。 |
| **Q&A 轮询** | **NEVER** | 严禁使用 `qa_reget_answer` 轮询等待用户回答。`qa_reget_answer` 是非阻塞的，仅用于“重新获取已回答问题内容或多媒体附件”。等待必须使用 `qa_get_answer`。 |
| **Supplement 时序** | **MUST** | 当 `qa_push_question` 携带 `supplement: true` 时，必须先调用 `qa_push_supplement` 注入详情，然后再调用 `qa_get_answer`，否则前端将因等待详情而一直加载阻塞用户。 |
| **Preview 交付** | **MUST** | Preview 文件上传后必须调用 `preview_file_list` 核对，确保识别出 `entry_file`（HTML 入口）后再向用户提供 URL 或挂载到 Q&A。 |
| **Preview 联动** | **NEVER** | 挂载 Preview 到 Q&A 时，严禁传递 URL 或 Hash；必须原样传递 `qa_supplement.content`（纯 JSON 字符串 `{"session_id":"...","file_id":"..."}`），严禁添加 Markdown 代码围栏（\`\`\`json）。 |
| **Pin 消费** | **MUST** | 约束消费只能通过 `pin_consume` 单向流转（pending → consumed），`pin_update` 仅允许修改优先级与分类。 |
| **RepoWiki 定位** | **MUST** | RepoWiki MCP 工具（`repoWiki_list` / `repoWiki_query`）为纯只读知识库，生成与更新由 Git Webhook 自动触发，MCP 端不提供写入接口。 |
| **Pages 晋升** | **MUST** | 无 HTML 入口不得 `pages_promote`。密码与访问策略只在控制台 `/console/pages` 配置，MCP 禁止传密码。继续改已发布页面必须先 `pages_fork`。 |
