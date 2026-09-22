---
name: lumina-preview-lpw
description: Lumina Preview LPW 把技术评审、架构说明、运行简报、方案对比、决策稿和结构化长文组合成具有明确阅读路径与出版级美感的 `.lpw` 文档。只要用户要用 Lumina 展示结构化内容、要求更好的排版/美学、提到 LPW、preview_lpw、技术简报、评审材料、决策矩阵、指标看板或需要通过 MCP 增量生成文档，就应使用本 skill；普通 HTML/CSS/JS 交互原型继续使用 lumina-preview。
license: MIT
compatibility: Requires Lumina MCP (Streamable HTTP) and network access to the Lumina instance.
metadata:
  author: lumina
  version: "0.1.0"
argument-hint: [ session-id | document-brief ]
allowed-tools: Read, Write, Edit, Bash, AskUserQuestion, mcp__lumina__project_get, mcp__lumina__project_list, mcp__lumina__project_create, mcp__lumina__preview_session_create, mcp__lumina__preview_session_list, mcp__lumina__preview_file_upload, mcp__lumina__preview_file_get, mcp__lumina__preview_file_delete, mcp__lumina__preview_file_list, mcp__lumina__preview_lpw_init, mcp__lumina__preview_lpw_node_add, mcp__lumina__preview_lpw_node_edit, mcp__lumina__preview_lpw_node_remove, mcp__lumina__preview_lpw_node_sort, mcp__lumina__preview_lpw_meta_set, mcp__lumina__preview_lpw_outline, mcp__lumina__preview_lpw_schema, mcp__lumina__qa_push_supplement, mcp__lumina__qa_get_answer, mcp__plugin_lumina_lumina__project_get, mcp__plugin_lumina_lumina__project_list, mcp__plugin_lumina_lumina__project_create, mcp__plugin_lumina_lumina__preview_session_create, mcp__plugin_lumina_lumina__preview_session_list, mcp__plugin_lumina_lumina__preview_file_upload, mcp__plugin_lumina_lumina__preview_file_get, mcp__plugin_lumina_lumina__preview_file_delete, mcp__plugin_lumina_lumina__preview_file_list, mcp__plugin_lumina_lumina__preview_lpw_init, mcp__plugin_lumina_lumina__preview_lpw_node_add, mcp__plugin_lumina_lumina__preview_lpw_node_edit, mcp__plugin_lumina_lumina__preview_lpw_node_remove, mcp__plugin_lumina_lumina__preview_lpw_node_sort, mcp__plugin_lumina_lumina__preview_lpw_meta_set, mcp__plugin_lumina_lumina__preview_lpw_outline, mcp__plugin_lumina_lumina__preview_lpw_schema, mcp__plugin_lumina_lumina__qa_push_supplement, mcp__plugin_lumina_lumina__qa_get_answer
---

# Lumina LPW 结构化文档设计与 MCP 构建指南

LPW（Lumina Paper Workshop）适合需要清晰叙事、结构化证据和稳定排版的评审文档。它使用受控节点表达内容关系，由渲染器统一处理桌面、窄侧栏和移动端布局。

普通网页交互、表单、动画或自由品牌页面使用 `lumina-preview`。LPW 负责文档型内容：技术评审、架构说明、运行简报、方案对比、决策记录和带证据的长文。

## 按需加载

| 需要 | 读取 |
| --- | --- |
| 视觉层级、布局决策、高频配方与收工检查 | [`references/composition.md`](./references/composition.md) |
| 三类节点、pattern、variant、block 与批注契约 | [`references/nodes.md`](./references/nodes.md) |
| 评审稿完整 MCP 调用顺序 | [`examples/review-document.md`](./examples/review-document.md) |
| 项目解析 | [`../_shared/project-resolver.md`](../_shared/project-resolver.md) |
| 挂载到 Q&A | [`../_shared/preview-qa-contract.md`](../_shared/preview-qa-contract.md) |

开始构建 LPW 前，必须先读 composition；遇到字段或容器归属不确定时再读 nodes，并调用 `preview_lpw_schema` 获取运行时真值。

## 工作原则

1. **先写阅读路径**：用一句话确定“用户先看到什么、据此理解什么、最后做什么”。
2. **选择一种主布局**：根 `content` 已提供纵向顺序。只有需要统一分组时才添加 flow；需要 split/grid/bento 等局部布局时，将它与前后 container 放成根部兄弟节点。layout 绝不嵌套 layout。
3. **先骨架后内容**：先建立 layout/container，再逐个加入 block。单次 add 只提交一个节点。
4. **内容决定组件**：有真实数值才用 metrics/chart，有真实比较维度才用 comparison，已有权重和评分才用 scorecard。缺少量化依据时省略对应数据组件，把待补信息放入 open-items；不要用占位分数满足 schema。
5. **强调有预算**：一屏只保留一个最强焦点。优先把结论、风险或推荐方案中的一个设为焦点。
6. **持续核对**：完成骨架后、每新增约 6–8 个节点后、交付前调用 outline。

## 标准流程

### 1. 解析项目与会话

按照 `../_shared/project-resolver.md` 解析 `project_id`。调用 `preview_session_list` 复用当前任务的 LPW 会话；没有合适会话时调用 `preview_session_create`。

会话标题使用文档目的，例如“缓存改造技术评审”，避免“LPW 测试”“新预览”等无信息名称。

### 2. 形成设计简报

在调用工具前，内部整理五项内容：

- **读者**：谁会看，具备什么背景。
- **任务**：读完需要理解、评审或决定什么。
- **阅读路径**：一句话说明顺序。
- **证据清单**：已有结论、指标、代码、图表、风险和待办。
- **视觉焦点**：整份文档最应被看见的一项内容。

信息不足且会改变文档结构时，使用 Q&A 让用户裁决。其余情况选择保守、清晰的根节点顺序；只有整组内容无需第二种 layout 时才使用 flow。

### 3. 查询契约

调用 `preview_lpw_schema` 查询即将使用的不确定 pattern、container 或 block。运行时 schema 是字段真值；references 提供选择方法。

不要发明 props。常见无效字段包括 `subtitle`、`tag`、`defaultExpanded` 和 tab item 的 `icon`。

### 4. 初始化文档

调用 `preview_lpw_init`：

- `meta.title` 使用具体主题，是唯一主标题。
- `description` 用一句话说明文档目的。
- tags 保留 1–3 个可检索主题。
- content 从空数组开始。

`.lpw` 本身可以成为 Preview 入口，无需额外上传空 HTML。

### 5. 准备伴随资源

当 image/gallery/editorial-wrap 需要本地图片或 SVG 时，使用 `preview_file_upload` 上传同一会话的扁平文件名资源，再在 block 的 `src` 中写该文件名。单文件保持在 256 KiB 内；更新前可用 `preview_file_get` 核对，废弃资源用 `preview_file_delete` 清理。

`preview_file_upload` 只用于伴随资源，不用于创建或覆写 `.lpw` 文档。LPW 正文始终由节点工具维护。已有 HTTPS 图片可直接引用，并在交付说明中注明网络依赖。

### 6. 增量构建

按 `references/composition.md` 选择配方并构建：

1. 添加顶层 container 或必要的 layout。根节点按阅读顺序排列；layout 的 `parent_id` 必须省略，使其始终位于根 `content`。
2. 给 section/panel 写能说明内容的标题。
3. 逐个添加 block；节点对象不带 `children`。
4. 使用语义化、稳定的 ID，例如 `decision-summary`、`latency-chart`、`rollback-items`。
5. 同一概念只出现一次标题。section 已有 title 时，不再添加同义 heading。
6. markdown 负责正文，避免再次使用 `#` / `##` 制造重复层级。

### 7. 结构与美学巡检

调用 `preview_lpw_outline`，按顺序检查：

- `completeness_warnings` 为空。
- 阅读顺序符合设计简报。
- container variant 接受其所有 block。
- tabs 的 items 数与 children 数一致。
- newspaper placements 覆盖每个 child，且唯一 body 是 markdown。
- bento 的 colSpan 不超过 columns，移动端 order 唯一。
- 没有重复标题、连续 takeaway、重复推荐或无数据图表。

需要调整时使用 node edit/sort/remove。props 编辑为浅合并；删除字段或大幅换型前先读取 outline，避免遗留旧字段。

### 8. 交付

调用 `preview_file_list`，确认：

- LPW 文件存在且为入口。
- `workflow.state` 为 `ready_for_review`。
- `preview_url` 非空。

独立评审时主动打开 `preview_url`。挂到 Q&A 时，读取 `../_shared/preview-qa-contract.md`，原样使用 `qa_supplement.content`；传 `session_id + file_id`，不传 URL 或 hash。

## 硬约束

- 使用 LPW 1.1 根字段 `content`，节点 kind 为 layout/container/block。
- layout 只含 container/block，且 layout 的 parent 只能是文档根；container 只含契约允许的 block；block 不含 children。
- 每次 `preview_lpw_node_add` 只添加一个节点，node 对象不携带 children。
- `preview_file_upload` 只上传伴随资源，禁止创建或覆写 `.lpw` 文档。
- LPW 日常编辑只使用 `preview_lpw_node_*` 工具族。
- 构建完成前必须让 `completeness_warnings` 清空。
- 每份文档必须通过 composition 的“收工前美学检查”。
- Preview 是评审媒介；真实项目实现、测试和交付仍在真实仓库完成。
