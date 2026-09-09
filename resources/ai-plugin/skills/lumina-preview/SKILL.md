---
name: lumina-preview
description: Lumina Preview 把 HTML/CSS/JS 原型推到沙盒页给用户看。需要展示组件、线框、交互稿或把预览挂到 Q&A 详情时使用。Preview 只是评审媒介，不能代替改真实仓库。
license: MIT
compatibility: Requires Lumina MCP (Streamable HTTP) and network access to the Lumina instance.
metadata:
  author: lumina
  version: "0.1.1"
argument-hint: [ session-id | filename ]
allowed-tools: Read, Write, Edit, Bash, AskUserQuestion, mcp__lumina__project_get, mcp__lumina__project_list, mcp__lumina__project_create, mcp__lumina__preview_session_create, mcp__lumina__preview_session_list, mcp__lumina__preview_file_upload, mcp__lumina__preview_file_list, mcp__lumina__preview_file_get, mcp__lumina__qa_push_supplement, mcp__lumina__qa_get_answer, mcp__plugin_lumina_lumina__project_get, mcp__plugin_lumina_lumina__project_list, mcp__plugin_lumina_lumina__project_create, mcp__plugin_lumina_lumina__preview_session_create, mcp__plugin_lumina_lumina__preview_session_list, mcp__plugin_lumina_lumina__preview_file_upload, mcp__plugin_lumina_lumina__preview_file_list, mcp__plugin_lumina_lumina__preview_file_get, mcp__plugin_lumina_lumina__qa_push_supplement, mcp__plugin_lumina_lumina__qa_get_answer
---

# Lumina 前端原型实时预览与可视化评审指南 (lumina-preview)

用于指导 AI Agent 构建轻量级前端原型预览会话，通过单层文件上传与沙盒隔离，向用户实时展示可视化的 HTML/CSS/JS 页面，并支持独立浏览器评审与 Q&A 题目挂载。

项目解析见 [`../_shared/project-resolver.md`](../_shared/project-resolver.md)。挂到 Q&A 时只读 [`../_shared/preview-qa-contract.md`](../_shared/preview-qa-contract.md)，不要另造字段。

## 按需加载

| 需要 | 读取 |
|---|---|
| 扁平文件名、相对引用、256 KiB、MIME | [`references/file-layout.md`](./references/file-layout.md) |
| 独立浏览器评审走查 | [`examples/standalone-review.md`](./examples/standalone-review.md) |
| 挂到 Q&A 详情走查 | [`examples/qa-mount.md`](./examples/qa-mount.md) |
| 入口 HTML 模板 | [`assets/index.html`](./assets/index.html) |

---

## 🎯 核心定位与设计哲学

- **媒介定位**：Preview 是快速对齐视觉和交互的**沟通与评审媒介**，不能替代对本地仓库真实源文件的实现、单元测试与交付。
- **运行环境**：前端采用 `iframe sandbox="allow-scripts"` 隔离环境渲染，同层文件通过相对路径互相引用。
- **文件支持**：HTML、CSS、JavaScript/MJS、JSON、SVG 与纯文本。

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
- **同层相对引用**：HTML 引入 CSS/JS 时一律使用同层文件名（如 `href="style.css"`，不要写成 `/style.css` 或 `assets/style.css`）。
- **文件大小上限**：单文件 UTF-8 编码字节数不得超过 **256 KiB**。

### 阶段 ④ 最终清单与入口核对 (`preview_file_list`)
所有文件上传完毕后，**必须调用 `preview_file_list(session_id)` 进行最终核对**。
- 工具会自动扫描是否存在有效的 HTML 入口文件（`entry_file`）。
- 确认返回的 `workflow.state` 为 `"ready_for_review"`。

### 阶段 ⑤ 双分支交付

#### 分支 A：独立视觉评审
若用户需要直接在浏览器中查看原型效果：
1. 从 `preview_file_list` 返回中获取绝对 `preview_url`（形如 `http://<domain>/preview?session=<hash>&file=index.html`）。
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

## 🛠️ 文件读取与覆写迭代

- **读取源码**：调用 `preview_file_get(session_id, filename)` 读取现有预览文件的完整内容。
- **覆写更新**：修改代码后，使用相同的 `session_id` 和 `filename` 再次调用 `preview_file_upload` 即可原位覆写。
- **重新核对**：覆写完成后，重新调用 `preview_file_list` 确认最新状态。

---

## ⛔ 核心红线 (Hard Rules)

1. **MUST**: 独立评审交付时，必须立即通过 Bash 执行 `open "<preview_url>"`（macOS）/ `xdg-open` / `start` 主动打开浏览器，严禁让用户手动复制输入。
2. **MUST**: 文件名严格扁平单层，禁止任何路径分隔符。
3. **MUST**: 必须上传至少一个 `.html` 文件作为渲染入口，且上传完必须用 `preview_file_list` 核对。
4. **NEVER**: 严禁向用户交付空的 `preview_url`。
5. **NEVER**: 挂载到 Q&A 时，严禁传递 URL 或 Hash，严禁给 JSON 添加 Markdown 代码围栏。
