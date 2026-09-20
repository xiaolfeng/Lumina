export interface LpwMeta {
  title: string
  description?: string
  author?: string
  version?: string
  tags?: string[]
}

export interface LpwBlock<TProps = Record<string, unknown>> {
  id: string
  type: string
  props: TProps
  children?: LpwBlock[] // 仅容器类型允许出现
}

export interface LpwDocument {
  version: '1.0'
  meta?: LpwMeta
  blocks: LpwBlock[]
}

// ── 内容块 props ──────────────────────────────────────────────
export interface LpwMarkdownProps {
  content: string
}

export interface LpwCalloutProps {
  level?: 'info' | 'success' | 'warning' | 'error'
  title?: string
  content: string
}

export interface LpwMetricItem {
  label: string
  value: string | number
  unit?: string
  trend?: 'up' | 'down'
  change?: string
  desc?: string
}
export interface LpwMetricsProps {
  items: LpwMetricItem[]
}

export interface LpwStepItem {
  title: string
  desc?: string
  status?: 'wait' | 'process' | 'finish' | 'error'
}
export interface LpwStepsProps {
  current?: number
  items: LpwStepItem[]
}

export interface LpwTimelineItem {
  time: string
  title: string
  content?: string
  tag?: string
}
export interface LpwTimelineProps {
  items: LpwTimelineItem[]
}

export interface LpwDiffProps {
  filename?: string
  language?: string
  oldCode: string
  newCode: string
  splitView?: boolean
}

export interface LpwTableColumn {
  key: string
  title: string
  width?: string
  align?: 'left' | 'center' | 'right'
}
export type LpwCellValue = string | number | boolean | null
export interface LpwTableProps {
  columns: LpwTableColumn[]
  data: Array<Record<string, LpwCellValue>>
  sortable?: boolean
}

// ── 容器块 props ──────────────────────────────────────────────
export interface LpwSectionProps {
  title: string
  collapsible?: boolean
  defaultOpen?: boolean
}
export interface LpwTabItem {
  key: string
  label: string
}
export interface LpwTabsProps {
  items: LpwTabItem[]
  defaultKey?: string
}
export interface LpwColumnsProps {
  ratio?: '1:1' | '1:2' | '2:1' | '1:1:1'
}

// ── 扩充内容块 props ──────────────────────────────────────────
export interface LpwHeadingProps {
  level?: 1 | 2 | 3
  content: string
}

export interface LpwListItem {
  content: string
  checked?: boolean // 仅 style='check' 生效
}
export interface LpwListProps {
  style?: 'ordered' | 'unordered' | 'check'
  items: LpwListItem[]
}

export interface LpwQuoteProps {
  content: string
  author?: string
  source?: string
}

export interface LpwCodeProps {
  language?: string
  filename?: string
  content: string
  showLineNumbers?: boolean
  highlightLines?: number[] // 1 起始行号
}

export interface LpwImageProps {
  src: string // 同会话文件名（相对解析）或 https?:// 外链
  alt: string
  caption?: string
  width?: string
}

export interface LpwCardItem {
  title: string
  description?: string
  href?: string
}
export interface LpwCardsProps {
  items: LpwCardItem[]
}

export interface LpwMermaidProps {
  content: string
  caption?: string
}

export type LpwChartType =
  'line' | 'bar' | 'area' | 'pie' | 'donut' | 'scatter' | 'radar'
export interface LpwChartSeries {
  name: string
  data: Array<number | [number, number]> // scatter 用点对，其余用 number
}
export interface LpwChartProps {
  chartType: LpwChartType
  title?: string
  categories?: string[] // scatter 不允许出现
  series: LpwChartSeries[]
  xLabel?: string
  yLabel?: string
  stacked?: boolean // 仅 bar / area
  legend?: boolean
  height?: number // 160–640，默认 280
}

// ── 扩充容器块 props ──────────────────────────────────────────
export interface LpwDetailsProps {
  summary: string
  defaultOpen?: boolean
}

// ── 第二批内容块 props ────────────────────────────────────────
export interface LpwComparisonCell {
  text: string
  verdict?: 'good' | 'warn' | 'bad'
}
export interface LpwComparisonProps {
  title?: string
  plans: Array<{ name: string; recommended?: boolean }>
  rows: Array<{ dimension: string; values: LpwComparisonCell[] }>
}

export interface LpwProgressItem {
  label: string
  value: number // 0–100
  status?: 'wait' | 'process' | 'finish' | 'error'
}
export interface LpwProgressProps {
  title?: string
  items: LpwProgressItem[]
}

export interface LpwTreeNode {
  label: string
  note?: string
  children?: LpwTreeNode[]
}
export interface LpwTreeProps {
  title?: string
  nodes: LpwTreeNode[]
}

// ── 第三批内容块 props（复刻 /report 模块） ───────────────────
export interface LpwTakeawayProps {
  title?: string // 默认「核心判断」
  content: string
}

export interface LpwGlanceItem {
  label: string
  text: string
}
export interface LpwGlanceProps {
  items: LpwGlanceItem[] // 建议 3 格：做成 / 卡住 / 下一步
}

export interface LpwOpenItem {
  title: string
  detail?: string
  owner?: string
  due?: string
}
export interface LpwOpenItemsProps {
  title?: string // 默认「未决事项」
  items: LpwOpenItem[]
}

export interface LpwScorecardProps {
  title?: string
  criteria: Array<{ name: string; weight: number }> // weight 1–100，Σ=100（logic 校验）
  plans: Array<{ name: string; scores: number[]; recommended?: boolean }> // 每项 1–5
}

export interface LpwQuadrantItem {
  label: string
  x: 'low' | 'high'
  y: 'low' | 'high'
  note?: string
}
export interface LpwQuadrantProps {
  title?: string
  xLabel?: string
  yLabel?: string
  xLow?: string
  xHigh?: string
  yLow?: string
  yHigh?: string
  items: LpwQuadrantItem[]
}

export interface LpwPersonnelItem {
  name: string
  role: string
  duties?: string
}
export interface LpwPersonnelProps {
  title?: string
  items: LpwPersonnelItem[]
}

export interface LpwGalleryImage {
  src: string
  alt: string
  caption?: string
}
export interface LpwGalleryProps {
  images: LpwGalleryImage[]
}

// ── 引擎组件统一契约与解析类型 ─────────────────────────────────
export interface LpwBlockSlotProps<TProps = Record<string, unknown>> {
  blockId: string // 块 id：锚点 / 错误定位
  props: TProps // 该块 props（组件内应用默认值）
  depth: number // 当前深度（顶层 = 1）
  childrenBlocks?: LpwBlock[] // 仅容器类型非空
}

export interface LpwParseError {
  message: string
  path?: string
}

export type LpwParseResult =
  | { document: LpwDocument; error?: undefined }
  | { document?: undefined; error: LpwParseError }
