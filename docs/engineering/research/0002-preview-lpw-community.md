# LPW 文档格式与社区实现调研

> 调研日期：2026-09-14 · 仓库基线：`789812a`。
> 方法：查阅官方文档与部分公开源码；未安装候选依赖，未运行性能、安全或浏览器集成测试。

## 背景

Lumina 拟增加 `.lpw`（Lumina Preview 文件），用 JSON 描述文档块，由 `type` 映射到前端专用组件库。核心问题是：社区已有哪些可复用的机制，哪些文件、校验与呈现约束仍需 LPW 自己定义？

本次需求已确认以文档块为主，交互限于阅读所需的本地行为。“映射数据”指文件参数传给组件，不指运行时数据源绑定；专用组件库尚未设计。

## 发现

### 1. 社区已有类型到组件的映射机制

json-render 的 Catalog 定义组件及参数约束，Registry 将类型名映射到实际组件；Puck 的 Config 同样把 JSON 中的 `type`、`props` 映射到 `render`。Editor.js 与 Adaptive Cards 则提供按类型组织内容块的文档或卡片模型。类型化 JSON 驱动已有组件并非 LPW 独有机制。[R1][R2][R4][R5][R6]

| 候选 | 一手材料确认的机制 | 与 LPW 的交集 | 直接采用仍需处理的差异 |
| --- | --- | --- | --- |
| json-render | 组件目录、参数约束、注册表；现成 React 渲染器采用 `root + elements`，子节点通过 ID 引用；支持高层语义块 | 最接近 `type + props → React 组件`；组件可以是整张表格或时间线，不必暴露基础按钮 | 专用组件、文档排版和兼容约束仍由 Lumina 提供；自定义文件结构需转换或自建渲染器，不能直接传给现成 Renderer [R1–R3] |
| Editor.js | 保存结果包含有序 `blocks`；各块有 `id`、`type`、`data`，类型由配置的工具名称决定 | 接近文档按阅读顺序组织完整内容块的形态 | 保存格式属于编辑器输出；该材料没有提供覆盖 LPW 专用组件的统一 JSON Schema，也没有替 LPW 定义版本兼容 [R4] |
| Adaptive Cards | JSON 序列化卡片模型，包含版本、`body`、类型化元素、容器、输入和动作；提交方式由宿主决定 | 可参考显式版本、类型分发和有限布局 | 面向卡片；LPW 长文排版及专用内容块仍需适配，输入和提交不是本次需求 [R5] |
| Puck | Config 定义组件、字段及渲染函数；编辑器产生 JSON，Render 根据同一 Config 展示内容 | 专用组件注册和参数传递机制接近需求 | 可视化编辑不是当前目标；编辑字段定义不能直接等同于标准 JSON Schema 契约 [R6] |
| A2UI | JSON Schema 组件目录；组件结构与数据模型分离；以消息创建、更新和删除界面区域，支持操作回传 | 可参考受控组件目录、校验错误定位和组件能力协商 | 完整协议包含消息顺序、状态更新和双向交互，范围大于一份文档文件；并非必须使用 SSE，也支持 WebSocket 等传输 [R7] |
| Markdoc | Markdown 标签可组合；定义属性、允许的子节点、验证函数，并映射 React 组件 | 保留 Markdown 写作体验的可行备选 | 标签语法不是纯 JSON，其验证配置也不是标准 JSON Schema；与本次已确认的文件形态不同 [R8] |
| amis | 通过 JSON 配置生成后台页面的低代码框架 | 证明较复杂页面也可由声明式配置描述 | 后台页面目标大于 LPW 文档阅读；本轮仅作方向筛选，未评估运行时依赖与安全配置 [R9] |

JSON Forms 的 UI Schema 主要描述表单的控件、布局和规则。名称包含 Schema 不代表它直接满足文档块渲染需求，因此本轮不将其列为重点候选。[R10]

### 2. json-render 可复用，但不是现成的 LPW 文档系统

json-render 区分三部分：定义结构的 Schema、定义可用组件的 Catalog、提供组件实现的 Registry。其组件参数以 Zod 描述；JSON Schema 契约是否完整覆盖所需校验，需要针对选定版本验证，不能仅凭 TypeScript 类型安全推断。[R1][R2][R3]

现成 React 格式使用平铺元素表与子节点 ID。官方也给出了高层页面块示例，因此“使用 json-render”与“必须让 AI 拼基础 UI”不是同一个决定。反过来，如果 LPW 采用嵌套文档块数组，需要转换成现成格式，或按官方自定义指南另写渲染器。[R3]

源码核对还有两点影响文档完整性：

- `Renderer` 从 `spec.elements[spec.root]` 开始渲染；缺失根节点会返回空内容。未知类型未配置 fallback、缺失子节点也可能只警告并跳过。
- `ElementErrorBoundary` 捕获组件渲染异常后返回 `null`；现成渲染器还包含动态属性、事件、状态监听和重复渲染处理。采用它仍需限制输入能力，并保证错误能被读者看到，而不是让内容静默消失。

这些是源码行为，不是实际浏览器复现；没有据此断言该库存在安全漏洞或不能使用。[R11]

### 3. JSON Schema 约束数据，组件实现负责呈现和行为

JSON Schema 可以描述对象字段、类型、数组以及组合条件。LPW 可用互斥的 `type` 值关联各组件参数 Schema，但最终字段、组合方式和所选方言尚未确定。标准文档提醒，`oneOf` 需要验证各个分支，递归组合也可能增加校验成本；组件数量增加后需实测，而不是直接承诺生成或校验速度。[R12]

Schema 合法不等于显示内容完整、组件一定存在或内容绝对安全。LPW 还需要检查版本与注册表是否匹配、容器是否合法、资源使用是否超限，并由可信组件处理文本、链接和错误状态。JSON Schema 2020-12 本身也区分 `format` 注解与断言，不能把一个 URL 格式标记当作完整的 URL 安全策略。[R11][R12][R13]

### 4. 当前 Preview 接入存在配套工作

以下为仓库 `789812a` 的静态代码事实，均不代表 LPW 已实现：

| 位置 | 当前行为 | 对后续实施的影响 |
| --- | --- | --- |
| [`web/src/lib/preview-file.ts`](../../../web/src/lib/preview-file.ts) | 文件按扩展名分为 html、markdown、svg、code；`.lpw` 落入 code | 需要新增 LPW 识别与专用渲染分支 |
| [`web/src/components/preview/file-viewer.tsx`](../../../web/src/components/preview/file-viewer.tsx) | HTML/SVG 使用 iframe；其他内容读取后按 Markdown 或代码显示 | JSON 原文可读取不等于 LPW 可渲染 |
| [`internal/logic/preview_logic.go`](../../../internal/logic/preview_logic.go) | 上传校验文件名和大小；未知扩展名 MIME 回退纯文本；成功写入后通知 `OnPreviewChanged` | LPW 校验需接入公共上传逻辑，不能只改前端 |
| [`internal/constant/preview.go`](../../../internal/constant/preview.go) | 单文件上限为 `256 * 1024` 字节 | 本轮没有要求修改现有限额 |
| [`internal/mcp/preview_tools.go`](../../../internal/mcp/preview_tools.go) | `findPreviewEntry` 只选 HTML；工具工作流要求 HTML 入口 | LPW 成为独立入口时，入口判断、状态文案及工具说明需要一并更新 |
| [`web/src/components/interact/primitives/preview-frame.tsx`](../../../web/src/components/interact/primitives/preview-frame.tsx) | Q&A 的 Preview 引用解析文件后直接交给 iframe | 原始 LPW JSON 不能照搬此路径；挂载需要按文件类型调用对应渲染器 |
| [`components/src/markdown/fenced-components.tsx`](../../../components/src/markdown/fenced-components.tsx) | 已有 Callout、Card、Steps、Step；由 remark fenced 插件映射 | LPW 可复用底层原语，但现有 Markdown 插件不是 LPW 的文件 Schema 或专用组件库 |

Q&A 的 Preview 引用当前包含 `session_id + file_id`。支持 LPW 显示不需要把文档中的本地交互转成 Q&A 回答，也不需要建立 Preview 对 Q&A 的业务调用。[上述 MCP 与预览引用源码]

## 结论

**确定：** 类型化 JSON 到已注册组件的机制有社区实现；文档块与通用 UI 组件是组件目录的不同颗粒度，不是由 JSON 本身决定。json-render/Puck 的映射机制与 Editor.js/Adaptive Cards 的内容模型提供了相互独立的参照。

**倾向：** LPW 重点定义文档文件契约与专用组件目录。json-render 可作为复用候选，但当前需求不要求它的状态表达式、动作或流式消息能力；是否引入，留待具体结构与兼容验证。自建有限渲染分发与使用社区渲染器都是仍可比较的实现方式。

**局限与开放问题：**

- 已确认纯 JSON、文档块、专用组件库和本地阅读交互；完整组件清单、参数、容器嵌套、媒体资源和文档排版尚未设计。
- Schema 方言、唯一维护源及导出方式、前后端验证器尚未选定；不能宣称两端已经具有一致校验行为。
- 未测试候选库在当前 React 19、TanStack Start、主题和构建配置下的兼容性；未测包体、性能、无障碍或 AI 生成成功率。
- 未确认采用社区库的具体发布版本。json-render 源码基线的 core 包声明为 `0.20.0`，不据此认定 npm 最新版或已发布状态。其他未标版本页面只作为本次访问到的机制说明。
- Markdoc 的 `/docs/schema`、Adaptive Cards 的旧 Schema Explorer 及本轮尝试的 Schema 原始文件地址未取得内容，未用于证明能力；改用成功读取的官方指南。未以社区宣传中的“始终可靠”“完全安全”代替实测。

## 参考

下列外部资料访问日期均为 **2026-09-14**。未注明发布日期或包版本的页面保留为滚动文档，不推断其发布时间。仓库相对链接随工作区变化，事实基线为上文所列提交。

- **R1** · json-render Introduction / Catalog，滚动文档、未注明发布日期：https://json-render.dev/docs · https://json-render.dev/docs/catalog
- **R2** · json-render Registry，滚动文档、未注明发布日期：https://json-render.dev/docs/registry
- **R3** · json-render Specs / Schemas / Custom Schema & Renderer，滚动文档、未注明发布日期：https://json-render.dev/docs/specs · https://json-render.dev/docs/schemas · https://json-render.dev/docs/custom-schema
- **R4** · Editor.js Saving data，页面示例输出版本 `2.8.1`，不是当前最新版本证明：https://editorjs.io/saving-data/
- **R5** · Microsoft Learn，Getting Started - Adaptive Cards，页面元数据更新于 2025-07-03；示例 Schema 版本 `1.0`：https://learn.microsoft.com/en-us/adaptive-cards/authoring-cards/getting-started
- **R6** · Puck Component Configuration，滚动文档、未注明发布日期及版本：https://puckeditor.com/docs/integrating-puck/component-configuration
- **R7** · A2UI 官网及 `v0.9.1` 规范；访问时官网标记该版本为 Current、`v1.0` 为 Candidate。规范头部 Last Updated 为 2025-12-03，正文已有更晚示例，不把头部日期视为全文最新修订证明：https://a2ui.org/ · https://a2ui.org/specification/v0.9.1-a2ui/
- **R8** · Markdoc Tags，滚动文档、未注明发布日期及版本：https://markdoc.dev/docs/tags
- **R9** · baidu/amis README，master 分支快照，未锁定发布版本：https://github.com/baidu/amis
- **R10** · JSON Forms UI Schema，滚动文档、未注明发布日期及版本：https://jsonforms.io/docs/uischema/
- **R11** · json-render 源码，main 基线 `6c7164342a37fab055907f1666658e1333e2cb86`，core 包声明 `0.20.0`：[React renderer](https://github.com/vercel-labs/json-render/blob/6c7164342a37fab055907f1666658e1333e2cb86/packages/react/src/renderer.tsx) · [React schema](https://github.com/vercel-labs/json-render/blob/6c7164342a37fab055907f1666658e1333e2cb86/packages/react/src/schema.ts) · [core package.json](https://github.com/vercel-labs/json-render/blob/6c7164342a37fab055907f1666658e1333e2cb86/packages/core/package.json)
- **R12** · JSON Schema Boolean JSON Schema combination，滚动文档、未注明发布日期：https://json-schema.org/understanding-json-schema/reference/combining
- **R13** · JSON Schema Draft 2020-12，规范页面列出发布时间 2022-06-16：https://json-schema.org/draft/2020-12
