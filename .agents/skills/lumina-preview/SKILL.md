---
name: lumina-preview
description: Lumina Preview 把原生 HTML/CSS/JS 或浏览器端 React/Vue 原型推到沙盒页给用户看。需要展示组件、线框、交互稿或把预览挂到 Q&A 详情时使用。Preview 只是评审媒介，不能代替改真实仓库。
license: MIT
compatibility: Requires Lumina MCP (Streamable HTTP) and network access to the Lumina instance.
metadata:
  author: lumina
  version: "0.1.2"
argument-hint: [ session-id | filename ]
allowed-tools: Read, Write, Edit, Bash, AskUserQuestion, mcp__lumina__project_get, mcp__lumina__project_list, mcp__lumina__project_create, mcp__lumina__preview_session_create, mcp__lumina__preview_session_list, mcp__lumina__preview_file_upload, mcp__lumina__preview_file_edit, mcp__lumina__preview_file_delete, mcp__lumina__preview_file_list, mcp__lumina__preview_file_get, mcp__lumina__preview_lpw_init, mcp__lumina__preview_lpw_node_add, mcp__lumina__preview_lpw_node_edit, mcp__lumina__preview_lpw_node_remove, mcp__lumina__preview_lpw_node_sort, mcp__lumina__preview_lpw_meta_set, mcp__lumina__preview_lpw_outline, mcp__lumina__qa_push_supplement, mcp__lumina__qa_get_answer, mcp__plugin_lumina_lumina__project_get, mcp__plugin_lumina_lumina__project_list, mcp__plugin_lumina_lumina__project_create, mcp__plugin_lumina_lumina__preview_session_create, mcp__plugin_lumina_lumina__preview_session_list, mcp__plugin_lumina_lumina__preview_file_upload, mcp__plugin_lumina_lumina__preview_file_edit, mcp__plugin_lumina_lumina__preview_file_delete, mcp__plugin_lumina_lumina__preview_file_list, mcp__plugin_lumina_lumina__preview_file_get, mcp__plugin_lumina_lumina__preview_lpw_init, mcp__plugin_lumina_lumina__preview_lpw_node_add, mcp__plugin_lumina_lumina__preview_lpw_node_edit, mcp__plugin_lumina_lumina__preview_lpw_node_remove, mcp__plugin_lumina_lumina__preview_lpw_node_sort, mcp__plugin_lumina_lumina__preview_lpw_meta_set, mcp__plugin_lumina_lumina__preview_lpw_outline, mcp__plugin_lumina_lumina__qa_push_supplement, mcp__plugin_lumina_lumina__qa_get_answer
---

# Lumina 前端原型实时预览与可视化评审指南 (lumina-preview)

用于指导 AI Agent 构建轻量级前端原型预览会话，通过单层文件上传与沙盒隔离，向用户实时展示原生 HTML/CSS/JavaScript 或浏览器端 React/Vue 页面，并支持独立浏览器评审与 Q&A 题目挂载。

项目解析见 [`../_shared/project-resolver.md`](../_shared/project-resolver.md)。挂到 Q&A 时只读 [`../_shared/preview-qa-contract.md`](../_shared/preview-qa-contract.md)，不要另造字段。

## 按需加载

| 需要 | 读取 |
|---|---|
| 扁平文件名、相对引用、256 KiB、MIME | [`references/file-layout.md`](./references/file-layout.md) |
| React/Vue、CDN、ESM、自动刷新边界 | [`references/framework-runtime.md`](./references/framework-runtime.md) |
| LPW 节点契约 / variant 允许表 / 批注规则 | [`references/lpw-nodes.md`](./references/lpw-nodes.md) |
| LPW 端到端构建示例 | [`examples/lpw-node-workflow.md`](./examples/lpw-node-workflow.md) |
| React 无构建浏览器示例 | [`examples/react-browser-runtime.md`](./examples/react-browser-runtime.md) |
| Vue 无构建浏览器示例 | [`examples/vue-browser-runtime.md`](./examples/vue-browser-runtime.md) |
| 独立浏览器评审走查 | [`examples/standalone-review.md`](./examples/standalone-review.md) |
| 挂到 Q&A 详情走查 | [`examples/qa-mount.md`](./examples/qa-mount.md) |
| 入口 HTML 模板 | [`assets/index.html`](./assets/index.html) |

---

## 🎯 核心定位与设计哲学

- **媒介定位**：Preview 是快速对齐视觉和交互的**沟通与评审媒介**，不能替代对本地仓库真实源文件的实现、单元测试与交付。
- **运行环境**：前端采用 `iframe sandbox="allow-scripts"` 隔离环境渲染，同层 HTML 可通过相对路径加载经典 CSS/JS；完整加载边界见框架运行约束。
- **文件支持**：HTML、CSS、JavaScript/MJS、JSON、SVG 与纯文本。
- **LPW 文档支持**：结构化交互文档（`.lpw`）支持 1.1 节点语义增量生成与渐进生长，直渲 31 种微明视觉组件与容器。
- **框架支持**：React/Vue 可使用固定版本 CDN 的浏览器构建，无需在 Lumina 安装 npm 依赖；这不等于提供 Node.js、打包器、SSR、Vue SFC 或本地多文件 ESM 运行时。
- **实时含义**：文件变更经 WebSocket 通知工作台重新加载 iframe；这是自动刷新，不是保留组件状态的 HMR。

---

## 🧱 LPW 精度模式（1.1 节点工作流）

`.lpw` 是 LPW 1.1 结构化文档：根字段 `content`，节点分三类同级 kind——layout（纯结构）/ container（美化整体）/ block（原子内容）。层级硬规则见 references/lpw-nodes.md。

1. **初始化** `preview_lpw_init`：写 meta（title 必填），content 留空。
2. **先结构后内容** `preview_lpw_node_add`：先加 layout / container，再把 block 挂进去。单次调用只提交**一个**节点；children 通过多次 add 逐个挂载。parent_id 缺省 = 根 content；layout 下只能挂 container/block；container 下只能挂契约允许的 block。
3. **中途核对** `preview_lpw_outline`：看 kind / type / pattern_or_variant / json_path，取最新 revision；`completeness_warnings` 非空说明文档还没满足完成态契约（如 layout 子节点不足、newspaper 缺 role=body），收工前必须补齐到该列表为空。
4. **局部修正** `preview_lpw_node_edit`（props 浅合并；annotation 仅 block 可设，传 null 清除）/ `preview_lpw_node_sort` / `preview_lpw_node_remove`。
5. **终核** `preview_file_list` → 交付 preview_url 或挂 Q&A。

**LPW 红线**：
- NEVER 调用 preview_lpw_block_add / _edit / _remove / _sort（已删除，会直接报错）。
- NEVER 写 `blocks` 根字段或 version "1.0"。
- NEVER 给 layout 写正文、给 block 加 children。
- NEVER 往 container 塞不匹配 variant 的 block——先查 references/lpw-nodes.md 契约表。
- NEVER 用 preview_file_upload 覆写分节点构建中的 LPW（JSON 损坏修复除外）。

---

## 🔄 Preview 标准五阶段生命周期

```text
① 解析 project_id  ──▶  ② 会话复用/创建  ──▶  ③ 逐文件上传 (upload)
                                                       │
                                                       ▼
⑤ 交付 (URL评审 或 挂载Q&A)  ◀──  ④ 清单与入口核对 (list) ◀─┘
```

### 阶段 ① 项目上下文解析
依照 `../_shared/project-resolver.md` 标准，通过 `project_get` / `project_list` 确定目标项目的 `project_id`。

### 阶段 ② 会话复用与创建
1. 调用 `preview_session_list(project_id)` 检查当前任务是否已有可复用的 Preview 会话。
2. 若无合适会话，调用 `preview_session_create` 创建空会话：
   ```json
   {
     "project_id": "<project_id>",
     "title": "购物车组件交互原型"
   }
   ```
   > ⚠️ **注意**：刚创建的会话是空的，**绝对不要**在此刻向用户交付或打开返回的空 `preview_url`。

### 阶段 ③ 逐文件上传 (`preview_file_upload`)
将原型的各个文件逐一上传到该会话：
```json
{
  "session_id": "<session_id>",
  "filename": "index.html",
  "content": "<!DOCTYPE html>\n<html>\n<head>\n  <link rel=\"stylesheet\" href=\"style.css\">\n</head>\n<body>\n  <div id=\"app\"></div>\n  <script src=\"app.js\"></script>\n</body>\n</html>"
}
```
**文件编写核心约束**：
- **扁平单层文件名**：文件名只能包含合法名称（如 `index.html`、`style.css`、`app.js`、`data.json`），**严禁出现 `/`、`\` 或 `..`**。
- **同层相对引用**：HTML 可用同层相对路径引入经典 CSS/JS（如 `href="style.css"`、`src="app.js"`）；不要写 `/style.css` 或 `assets/app.js`。沙盒是 opaque origin，本地多文件 ESM 不属于可靠支持范围。
- **框架依赖**：React/Vue 优先使用固定版本 CDN 的浏览器构建；需要 JSX/TSX、Vue SFC、npm 包解析或构建插件时，先在真实项目构建，再上传静态产物。具体读 [`references/framework-runtime.md`](./references/framework-runtime.md)。
- **文件大小上限**：单文件 UTF-8 编码字节数不得超过 **256 KiB**。

### 阶段 ④ 最终清单与入口核对 (`preview_file_list`)
所有文件上传完毕后，**必须调用 `preview_file_list(session_id)` 进行最终核对**。
- 工具会自动扫描是否存在有效的 HTML 入口文件（`entry_file`）。
- 确认返回的 `workflow.state` 为 `"ready_for_review"`。

### 阶段 ⑤ 双分支交付

#### 分支 A：独立视觉评审
若用户需要直接在浏览器中查看原型效果：
	1. 从 `preview_file_list` 返回中获取绝对 `preview_url`（形如 `http://<domain>/preview/<hash>/index.html`）。Preview 必须登录；打开后若跳到登录页属预期。需要对外持久分享时改走 `lumina-pages`。
2. **[CRITICAL] 主动打开浏览器**：Agent **必须立即通过 Bash 执行系统打开命令为用户弹出预览页面**，严禁要求用户手动复制或输入链接：
   - **macOS**: `open "<preview_url>"`
   - **Linux**: `xdg-open "<preview_url>"`
   - **Windows**: `start "<preview_url>"`
   > 仅在无 GUI/无头环境下，才降级为在对话中输出可点击的 Markdown URL。

#### 分支 B：作为 Q&A 问题的详情面板挂载 (跨模块协同)
若该原型是配合 Q&A 问题或选项呈现给用户的富交互面板：
1. 从 `preview_file_list` 返回中提取 `qa_supplement.content` 字段（这是一个纯 JSON 字符串，形如 `{"session_id":"123","file_id":"456"}`）。
2. 调用 `qa_push_supplement`，严格遵守 `../_shared/preview-qa-contract.md`：
   ```json
   {
     "session_id": "<QA会话ID>",
     "question_id": "<问题ID>",
     "option_id": "<选项ID（可选）>",
     "content_type": "preview",
     "content": "{\"session_id\":\"123\",\"file_id\":\"456\"}"
   }
   ```
3. 调用 `qa_get_answer` 阻塞等待用户在结合了预览面板后作出的回答。

---

## 🛠️ 文件读取、行级编辑与删除迭代

- **读取源码**：调用 `preview_file_get(session_id, filename)` 读取现有预览文件的完整内容；定位修改点或大文件时传 `start_line`/`end_line` 行区间（1 起始闭区间），返回内容按「行号| 文本」格式，行号可直接用于编辑定位。
- **行级编辑（局部修改首选）**：调用 `preview_file_edit` 只传输变更片段，无需整文件重传：
  ```json
  {
    "session_id": "<session_id>",
    "filename": "app.js",
    "operation": "replace",
    "start_line": 12,
    "end_line": 14,
    "content": "function mount() {\n  render();\n}"
  }
  ```
  - `insert`：把 content 各行插入到 `start_line` 之前（省略 `start_line` 时追加到文件末尾）；
  - `replace`：替换 `[start_line, end_line]` 闭区间（空 content 等价删除该区间）；
  - `delete`：删除该闭区间（不带 content）。
  - 编辑后核对返回的 `edited_region`（变更主体 ±3 行、带行号），确认落点正确；不确定行号时先读取再编辑，禁止盲猜。
- **整文件重写**：结构调整或大面积改写时，使用相同的 `session_id` 和 `filename` 调用 `preview_file_upload` 原位覆写。
- **删除文件**：废弃文件调用 `preview_file_delete(session_id, filename)`；删除的是 HTML 入口时必须重新补齐入口再交付。
- **重新核对**：变更完成后，重新调用 `preview_file_list` 确认最新状态。

---

## ⛔ 核心红线 (Hard Rules)

1. **MUST**: 独立评审交付时，必须立即通过 Bash 执行 `open "<preview_url>"`（macOS）/ `xdg-open` / `start` 主动打开浏览器，严禁让用户手动复制输入。
2. **MUST**: 文件名严格扁平单层，禁止任何路径分隔符。
3. **MUST**: 必须上传至少一个 `.html` 文件作为渲染入口，且上传完必须用 `preview_file_list` 核对。
4. **MUST**: 使用外部框架时固定 CDN 版本，并向用户说明运行依赖网络；需要 npm 构建链时先在真实项目构建。
5. **NEVER**: 不要把自动刷新描述成 React Fast Refresh、Vue HMR 或状态保持热更新。
6. **NEVER**: 严禁向用户交付空的 `preview_url`。
7. **NEVER**: 挂载到 Q&A 时，严禁传递 URL 或 Hash，严禁给 JSON 添加 Markdown 代码围栏。
8. **NEVER**: 禁止用 `preview_file_upload` 整体覆写分节点构建中的 LPW（除非修复严重损坏的 JSON；写回内容必须符合 1.1 结构契约，允许合规的渐进中间态）。
9. **NEVER**: NEVER 在 LPW 文档里使用旧版 blocks 结构或 block 系工具。
10. **NEVER**: 禁止对 `.lpw` 文件调用 `preview_file_edit` 进行日常局部编辑（除非紧急修复损坏的 JSON；LPW 局部修改必须使用 `preview_lpw_node_*` 节点语义工具族）。
