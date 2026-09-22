# LPW 1.1 美学组合指南

LPW 的美感来自阅读顺序、内容层级和留白。节点数量与布局复杂度都服务于理解。

## 先定叙事，再选组件

先用一句话写出文档的阅读路径。例如：

- “先给结论，再给三项证据，最后列待办。”
- “先看四个指标，再看趋势，最后展开原始数据。”
- “先建立背景，再并列两个方案，最后给出推荐。”

这句话决定根节点顺序。文档根 `content` 本身就是纵向阅读流。只有一组内容需要统一 horizontal/vertical flow 时才添加 `flow` layout；一旦需要 split、grid、bento、alternating、newspaper 或 editorial-wrap，就把该 layout 与前后 container 作为根部兄弟节点，因为 layout 不能嵌套 layout。

## 视觉层级

1. `meta.title` 是唯一主标题。
2. `section.title` 负责章节层级。section 已有标题时，内部不再添加同义 `heading`。
3. block 自带的 `title` 只描述局部图表或数据集合。
4. `markdown` 负责正文，避免用 `#` / `##` 重复制造标题。
5. 一屏只安排一个最强强调元素。`takeaway`、推荐列、高危 `callout` 任选其一成为焦点。

## 布局决策

| 内容关系 | 推荐结构 | 使用条件 |
| --- | --- | --- |
| 线性说明、评审正文 | 根 `content` 顺序排列；可选单个 `flow` vertical | 默认选择，最稳定；flow 内不得再放 layout |
| 两组真实对照、左证据右图 | `split` | 恰好两个同级区域；桌面并排、窄视口自动堆叠 |
| 同权卡片、指标集合 | `grid` 或 `panel/dashboard` | 子项同权；选一个承载层即可 |
| 明确主次的焦点拼版 | `bento` | 2–6 项；`colSpan` 不超过 columns；每项有清晰权重 |
| 成组图文案例 | `alternating` | 偶数子项，每对由媒体与正文组成 |
| 长篇编辑稿 | `newspaper` | placements 覆盖全部 child，且只有一个 `body` markdown |
| 单张主图与连续正文 | `editorial-wrap` | 恰好一个 image 和一个 markdown |
| 紧凑标签或短卡片横排 | `flow` horizontal | 子项可在窄宽度自然换行，避免表格/代码 |

### 选择克制

- 每份文档只设一种主布局语言。可以有一个局部 split 或 grid，避免同时叠 bento、tabs、newspaper。
- `panel/dashboard` 已负责网格组织时，内部优先放单组 metrics、chart、progress 或 comparison，避免再套根级 grid。
- `tabs` 只用于用户确实需要切换的互斥内容；顺序阅读更合适时直接纵向排列。
- `newspaper` 只适合长文。短评审稿使用 flow 更清晰。

## 高频配方

### 评审稿：结论 → 证据 → 风险 → 行动

```text
meta
├─ panel/summary
│  ├─ takeaway
│  └─ glance
├─ split(ratio 3:2)
│  ├─ section/evidence → code|diff|table + callout
│  └─ section/article → mermaid|image + markdown
└─ section/article → open-items
```

规则：summary 最多 2 个块；split 两侧都使用有标题的 container；风险只用一个 warning/error callout。

### 运行简报：指标 → 趋势 → 明细

```text
meta
├─ panel/dashboard
│  ├─ metrics
│  ├─ chart
│  └─ progress
└─ details/raw-data → table|code
```

规则：metrics 放 2–4 个关键值；chart 只表达一个问题；原始数据折叠收纳。

### 决策稿：背景 → 对比 → 裁决

```text
meta
├─ section/article → markdown + list
├─ panel/dashboard → comparison
├─ tabs/comparison → comparison + scorecard（仅在已有权重与评分，且确实需要切换两种依据时替换上一行）
└─ panel/summary → takeaway
```

规则：comparison 与 scorecard 的方案名称保持一致；仅一个方案标记 `recommended:true`；takeaway 用一句可执行结论收尾。用户没有提供权重或分数时，使用 panel/dashboard + comparison，省略 scorecard 和 tabs，并把量化补充项放进 open-items。

### 叙事长文：引子 → 正文 → 补充

```text
meta
├─ section/feature → image|gallery + markdown
├─ section/article → markdown + quote + timeline
└─ details/supplement → code|tree|table
```

规则：feature 首块优先 image/gallery；article 可承载 process 组；补充材料不抢正文注意力。

## 组件选用

| 目标 | 首选 | 常见误用 |
| --- | --- | --- |
| 一句话结论 | takeaway | 连续放多个 takeaway |
| 2–4 项关键速览 | glance | 用 cards 模拟指标 |
| 数值与变化 | metrics | 把解释长文塞进 desc |
| 连续趋势 | chart | 同时堆太多 series |
| 方案差异 | comparison | 用普通 table 失去 verdict 语义 |
| 加权选择 | scorecard | 权重不闭合或方案 scores 缺位 |
| 执行流程 | steps | 用 heading + list 人工模拟序号 |
| 历史事件 | timeline | 用 steps 表示时间 |
| 尚待决策 | open-items | 当作普通任务清单滥用 |
| 补充证据 | details/raw-data | 在正文直接展开超长代码 |

## MCP 构建顺序

1. 调用 `preview_lpw_schema` 查当前不确定的 pattern、variant 或 block。
2. `preview_lpw_init` 写入简短、具体的 meta。
3. 按阅读顺序添加顶层 layout/container，先形成骨架。
4. 每次 `preview_lpw_node_add` 只添加一个节点；节点对象不带 `children`。
5. 用语义 ID，例如 `decision-summary`、`latency-chart`、`rollback-items`。
6. 骨架完成后调用一次 `preview_lpw_outline`；之后每新增约 6–8 个节点再核对一次。
7. 修正层级、缺位和重复强调，直到 `completeness_warnings` 为空。
8. 用 `preview_file_list` 获取入口并核对 `ready_for_review`。

## 收工前美学检查

- 5 秒内能看懂标题、结论与下一步。
- 从上到下只有一条清晰阅读路径。
- 标题没有重复，块标题描述的是内容。
- 同一区域没有重复边框、重复粗线或无意义装饰标签。
- 结构化块使用真实数据；缺数据时选择更简单的 block。
- 桌面宽度下信息有层级，窄视口下顺序仍符合阅读逻辑。
- 交互容器具有必要性：details 用于折叠，tabs 用于互斥切换。
- 文案简洁具体，没有“示例标题”“待补充”等占位语。
