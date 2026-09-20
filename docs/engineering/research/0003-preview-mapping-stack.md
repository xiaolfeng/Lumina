> 调研日期：2026-09-14（承接 [0002](./0002-preview-lpw-community.md)）

## 背景

ADR 0008 冻结了 LPW 文档块与专用组件契约，但尚未确定前端实现采用何种映射技术栈。
核心问题：LPW 前端组件映射应引入社区库 json-render，还是自建极简映射分发器？

## 发现

### 1. 映射层在 LPW 渲染链路中的真实职责

LPW 渲染链路分为四段：JSON 解析与 Schema 校验、类型映射分发、组件属性绑定、错误边界防护。

```
.lpw 文件 (JSON) 
  → JSON Schema 校验 (合法块 AST) 
  → 映射分发器 (按 block.type 查表) 
  → 专用 React 组件 (<Component {...block.props} />) 
  → 渲染视图 (含局部 BlockErrorBoundary)
```

映射层的纯逻辑职责只有两项：
1. 维护 `type` 字符串到 React 组件的键值映射表；
2. 接收结构化块数据，将其 `props` 透传给对应组件，并在未知类型或渲染异常时挂载可见错误占位。

该职责不包含跨块状态绑定、动态表达式计算、远程组件加载或通用表单提交。[R1]

### 2. LPW 渲染链路全景与部件分布图

实现 LPW 文档映射与呈现涉及输入存储、前端路由分流、映射分发核心、专用组件库以及共享基础设施五个层次。

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ 1. 输入、存储与入口识别层 (Ingestion & MCP Entry)                            │
│    • 文件格式: .lpw (JSON, ≤ 256 KiB, 声明 schema 版本)                     │
│    • 后端逻辑: internal/logic/preview_logic.go (校验与落库)                 │
│    • MCP 工具: internal/mcp/preview_tools.go (findPreviewEntry 识别入口)     │
│    • 实时通知: internal/websocket/ (preview_sync 广播文件变更)              │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ 浏览器端获取文件内容
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 2. 前端文件识别与分流层 (File Kind & Viewer Branching)                      │
│    • 扩展名推导: web/src/lib/preview-file.ts (previewKindFromFilename: 'lpw')│
│    • 预览主入口: web/src/components/preview/file-viewer.tsx (分发至 LPW 视图) │
│    • Q&A 嵌入层: web/src/components/interact/primitives/preview-frame.tsx   │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ 载入 LPW 数据流
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 3. LPW 映射分发引擎 (LPW Mapping Engine - 自建分发核心)                      │
│    ┌───────────────────────────────────────────────────────────────────┐    │
│    │ 文档壳层 (LpwDocumentViewer): 解析 JSON、校验版本、渲染外层排版与目录    │    │
│    └─────────────────────────────────┬─────────────────────────────────┘    │
│                                      │ 遍历 blocks 数组                     │
│    ┌─────────────────────────────────▼─────────────────────────────────┐    │
│    │ 块级错误边界 (BlockErrorBoundary): 捕获单块异常，渲染可见错误卡片    │    │
│    └─────────────────────────────────┬─────────────────────────────────┘    │
│                                      │ 传递合法块 { id, type, props }       │
│    ┌─────────────────────────────────▼─────────────────────────────────┐    │
│    │ 映射分发器 (LpwBlockRenderer) ◄───► 组件注册表 (LpwRegistry)      │    │
│    │ • 按 type 查表分发                 • 维护 type -> Component 映射   │    │
│    │ • 递归渲染 container 容器子节点     • 未知类型分发至 FallbackBlock  │    │
│    └─────────────────────────────────┬─────────────────────────────────┘    │
└──────────────────────────────────────┼──────────────────────────────────────┘
                                       │ 实例化专用组件
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 4. LPW 专用组件库层 (Dedicated Component Library)                            │
│    • 基础内容块: Markdown, Callout, Metrics, Timeline, Steps, Diff, Table   │
│    • 容器组织块: Section, Tabs, Columns (有限深度递归)                      │
│    • 容错指示块: UnknownBlock, ErrorBlock (错误可见，不丢块)                │
│    • 本地交互态: Tab 切换、折叠展开、表格客户端排序 (纯本地无副作用)         │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ 消费底层样式与原语
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ 5. 共享设计系统与基础设施层 (@lumina/components)                            │
│    • UI 原语: components/src/ui/ (Radix UI, Lucide, Tailwind v4)            │
│    • 排版与高亮: components/src/markdown/ (react-markdown, codeblocks)      │
│    • 主题契约: components/src/styles/theme.css (静烛 v1 语义色盘与直角)     │
└─────────────────────────────────────────────────────────────────────────────┘
```

各层次包含的具体部件、源码路径与职责契约如下表：

| 层次 | 部件名称 | 物理代码路径 | 核心职责与契约 | 在映射方案中的定位 |
| --- | --- | --- | --- | --- |
| **存储与入口** | `.lpw` 文件契约 | 规范文档 / 文件系统 | UTF-8 JSON，声明版本，单文件 ≤ 256 KiB [R1] | 数据源输入 |
| | 后端上传校验 | `internal/logic/preview_logic.go` | 基础大小与格式校验、MIME 分配、广播 `preview_sync` [R1] | 输入网关校验 |
| | MCP 入口推导 | `internal/mcp/preview_tools.go` | 扩展 `findPreviewEntry`，识别 `.lpw` 为独立预览入口 [R1] | 工作流入口感知 |
| **类型分流** | 类型推导 | `web/src/lib/preview-file.ts` | 扩展 `PreviewKind` 增加 `'lpw'` 分支 [R5] | 前端文件识别 |
| | 渲染主容器 | `web/src/components/preview/file-viewer.tsx` | 拦截 `'lpw'` 类型，挂载 `PreviewLpwView` [R5] | 前端路由分流 |
| | Q&A 预览引用 | `web/src/components/interact/primitives/preview-frame.tsx` | 支持在 Q&A 详情中内嵌渲染 LPW 文档，不走 iframe [R1] | 跨场景渲染复用 |
| **映射引擎** | 文档壳层 | `web/src/components/preview/lpw/document-viewer.tsx` | 文档元数据（标题/版本）、长文滚动容器、全局布局 [R1] | 宿主容器 |
| | 解析校验器 | `web/src/components/preview/lpw/parser.ts` | JSON 解析与版本校验，产出文档级解析错误 [R1] | 前端防御解析 |
| | 组件注册表 | `web/src/components/preview/lpw/registry.ts` | `Map<string, ComponentType>` 字典，提供组件注册与检索 | **映射核心（自建）** |
| | 块级分发器 | `web/src/components/preview/lpw/block-renderer.tsx` | 按 `block.type` 查表并实例化，递归渲染容器 children | **分发核心（自建）** |
| | 块级错误边界 | `web/src/components/preview/lpw/block-error-boundary.tsx` | 独立捕获单块异常，渲染可见错误卡片，杜绝静默吞块 [R1] | **安全核心（防崩溃）** |
| | 兜底占位块 | `web/src/components/preview/lpw/fallback-block.tsx` | 未知类型或参数缺失时的可见警告卡片 [R1] | 容错呈现 |
| **专用组件** | 基础内容块 | `web/src/components/preview/lpw/blocks/*.tsx` | Callout, Metrics, Timeline, Steps, Diff, Table 等 [R1] | 业务表现层（自建） |
| | 容器组织块 | `web/src/components/preview/lpw/containers/*.tsx` | Section, Tabs, Columns (有限递归深度) [R1] | 结构组织（自建） |
| | 本地交互原语 | 专用组件内状态 hook | Tabs 激活、折叠展开、表格排序筛选（纯本地 React 状态） [R1] | 本地阅读交互 |
| **设计底座** | 共享 UI 原语 | `@lumina/components/ui/` | 按钮、标签、卡片、滚动区、分割线等底层组件 [R5] | 基础原语复用 |
| | Markdown 渲染 | `@lumina/components/markdown/` | 富文本安全解析与代码高亮 [R5] | 正文原语复用 |
| | 主题系统 | `@lumina/components/theme.css` | 静烛 v1 视觉规范（全平直角、`--sea-ink`、`--lagoon`） [R5] | 统一视觉基准 |

### 3. 自建极简分发器与 json-render 的机制对比

根据 ADR 0008 契约与 json-render 源码实测，两个实现路径在数据结构、错误处理、依赖体积和安全边界上存在实质差异。

| 对比维度 | 自建极简映射分发器 | 引入社区库 json-render (core 0.20.0) |
| --- | --- | --- |
| 数据结构契约 | 直接消费有序块数组 `blocks: LpwBlock[]`，支持有限容器递归嵌套 | 要求平铺结构 `{ root: string, elements: Record<string, Element> }`，子节点按 ID 引用 [R2][R3] |
| 结构适配成本 | 零转换。数据与 LPW 文档流一一对应 | 需在前端增加 AST 转平铺表的适配层，或另写自定义渲染器 [R3] |
| 渲染错误捕获 | 块级 ErrorBoundary 捕获后渲染可见错误卡片，标明块 ID、类型与错误原因 [R1] | 内置 `ElementErrorBoundary` 捕获异常后返回 `null`，导致崩溃块静默消失 [R4] |
| 未知类型处理 | 渲染带明显提示的占位块，告知读者缺失组件定义，不丢弃上下文 [R1] | 控制台输出 warning 并跳过渲染，页面表现为空白 [R4] |
| 运行时功能集 | 仅保留类型查表与 props 透传，无多余逻辑 | 内置动态状态表达式、动作监听（Actions）及流式增量协议 [R2][R4] |
| 代码与依赖负担 | 约 80–150 行原生 TypeScript 代码，0 新增外部依赖 | 新增 `json-render` 核心包及相关依赖；API 处于 `0.x` 迭代期 [R4] |
| 专用组件复用 | 直接导入并封装 `@lumina/components` 原语，共享 Tailwind v4 样式 [R5] | 仍需全量手写 LPW 专用组件，json-render 不提供文档排版组件 [R2] |

### 4. json-render 当前不宜作为主选的核心冲突

第一，错误处理契约冲突。ADR 0008 第 6 条明确规定「校验与渲染失败必须可见、可定位，不能静默丢失文档内容」。json-render 的 `renderer.tsx` 源码显示，`ElementErrorBoundary` 在 `componentDidCatch` 捕获组件渲染异常后直接返回 `null`。当第三方或 AI 生成的参数导致组件内部抛出异常时，该块在视图中完全消失，直接违反 LPW 可见错误原则。[R1][R4]

第二，数据抽象维度错位。LPW 面向长文档阅读，文件格式为顺序文档块列表；json-render 面向通用 UI 树构建，强制使用 `root + elements` 的平铺 ID 字典。若引入 json-render，必须在前端增加一次数组到字典的双向转换，徒增序列化开销且无架构收益。[R1][R3]

第三，能力溢出与攻击面。ADR 0008 第 5 条规定「文件不声明脚本、事件处理程序、表单提交、业务接口或 MCP 调用，也不引入通用数据源、状态表达式及绑定语言」。json-render 内置了状态求值和动作分发逻辑，引入后需额外做防御性裁剪，防止非预期表达式解析。[R1][R4]

### 5. 保留为候选的技术前提

保留 json-render 作为候选，仅适用于未来需求发生以下结构性变化时：
1. LPW 演进至跨块动态响应式文档（例如输入框动态联动图表渲染）；
2. 需要将 LPW 规范向 React 以外的宿主（如 Vue、React Native、原生客户端）做跨平台映射，且各端已具备成熟的 json-render 适配器；
3. 需要按块粒度实现细粒度增量流式渲染（Streaming UI），且自建分发层维护成本超过外部依赖。

在上述需求出现之前，引入 json-render 属于过度设计。[R1][R2]

### 6. 2026-09-20 合并后复核：React 直渲管线现状

同步 master（合并提交 `7ce1551`，含 `ec7f904`）后，本仓库前端已存在一条 React 直接渲染管线，与本调研的自建分发方案直接相关。以下为静态代码事实：

| 位置 | 现状 | 与 LPW 的关系 |
| --- | --- | --- |
| `web/src/components/preview/file-viewer.tsx` | `PreviewFileViewer` 统一分发：html/svg → `PreviewFrame`（iframe `sandbox="allow-scripts"`）；markdown → `PreviewMarkdownView`（`@lumina/components/markdown`，React 树内直渲）；其余 → CodeMirror 源码视图 | `.md` 的直渲分支就是 LPW 分发的现成落点：`'lpw'` kind 增加同层分支即可 |
| `web/src/components/preview/workbench-canvas.tsx` | Preview 工作台画布复用 `PreviewFileViewer`，源码检查态强制 `kind='code'` | LPW 在工作台自动获得渲染/源码双态 |
| `web/src/components/pages/showcase-shell.tsx` | Pages 展示态复用同一 `PreviewFileViewer`；`isRenderable` 仅匹配 `html\|htm\|md`，其余文件进「高级资源」折叠区 | `.lpw` 需加入 `isRenderable` 才能进入展示态页面直切区 |
| `web/src/components/interact/primitives/preview-frame.tsx` | Q&A `PreviewSupplement` 解析 `file_id` 后统一构造 `/preview/:hash/:file?lumina_frame=1` 交给 iframe，未按文件类型分流 | LPW 内嵌 Q&A 需要在此处按 kind 分流到 React 直渲 |
| `internal/mcp/preview_handlers.go` | 入口判定移入 snapshot 构建；`previewWriteWorkflow` 仍只有 HTML 入口语义（`awaiting_html_entry` / `reviewable_unverified`）；preview 工具增至 7 个（新增 `preview_file_edit` / `preview_file_delete`） | 原 `findPreviewEntry`（`preview_tools.go`）已不存在；LPW 入口判定需改在 snapshot 构建处扩展 |
| Pages 晋升（`internal/logic/pages_logic.go`） | 会话文件全量深拷贝为 `PageFile` 不可变快照；版本有 `EntryFilename` 入口字段 | 会话内 `.lpw` 会自然进入快照；版本入口选择需考虑 `.lpw` |

寻址已全面路径式化：Preview 为 `/preview/:session_hash/:filename`（登录态），Pages 为 `/pages/:project_name/:slug/:filename`（公开或密码门）。[R6]

## 结论

事实结论：
1. 映射层仅需类型查表与块级错误边界，自建极简分发器仅需约百行代码，天然契约对齐且零外部依赖。
2. json-render 的 `ElementErrorBoundary` 返回 `null` 与平铺 ID 树结构，直接违反 ADR 0008 的可见错误与文档顺序契约。
3. 专用组件实现、文档排版与样式完全依赖 `@lumina/components`，引入 json-render 并不能减少组件开发工作量。

倾向：
前端采用自建极简映射分发器（`LpwBlockRenderer` + `LpwRegistry`）作为最终技术栈；json-render 仅在跨块响应式交互或多端渲染诉求确立时再做评估——方案拍板由 RFC 承接。

开放问题：
1. 自建分发器在 React 19 严格模式下的并发渲染性能与长文档虚拟化滚动边界尚未实测。
2. 容器块（如 Section、Tabs）递归嵌套的最大深度限制，应在 Schema 阶段还是前端分发层进行硬约束。

## 参考

- **R1** · 本地 ADR 0008 · `docs/engineering/adr/0008-preview-lpw-document-contract.md`
- **R2** · json-render 官方文档 · https://json-render.dev/docs · https://json-render.dev/docs/catalog
- **R3** · json-render Custom Schema & Specs · https://json-render.dev/docs/custom-schema · https://json-render.dev/docs/specs
- **R4** · json-render 源码 (commit `6c7164342a37fab055907f1666658e1333e2cb86`) · https://github.com/vercel-labs/json-render/blob/6c7164342a37fab055907f1666658e1333e2cb86/packages/react/src/renderer.tsx
- **R5** · 本地组件库配置 · `components/package.json` 与 `components/src/`
- **R6** · 本地合并后代码 · 合并提交 `7ce1551`（含 origin/master `ec7f904`）· `web/src/components/preview/file-viewer.tsx` · `web/src/components/pages/showcase-shell.tsx` · `web/src/components/interact/primitives/preview-frame.tsx` · `internal/mcp/preview_handlers.go`
