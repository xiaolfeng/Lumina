import type { LpwIconName } from './icon-map'
import type { LpwRenderLocation } from './render-location'

export interface LpwMeta {
  title: string
  description?: string
  author?: string
  version?: string
  tags?: string[]
}

export type LpwNodeKind = 'layout' | 'container' | 'block'

export interface LpwNodeBase {
  id: string
  type: string
}

// ── Layout ──────────────────────────────────────────────
export type LpwLayoutPattern =
  | 'split'
  | 'alternating'
  | 'grid'
  | 'bento'
  | 'newspaper'
  | 'editorial-wrap'
  | 'flow'

export type LpwLayoutStrategy =
  | { type: 'equal' }
  | { type: 'ratio'; tracks: number[] }
  | {
      type: 'fixed-fluid'
      fixed: 'start' | 'end'
      size: 'sm' | 'md' | 'lg' | '25%' | '33%' | '40%' | '50%'
    }
  | { type: 'spans'; columns: 2 | 3 | 4 }
  | { type: 'columns'; count: 2 | 3 }
  | {
      type: 'media-wrap'
      mediaPosition: 'top-start' | 'top-end'
      mediaWidth: 'quarter' | 'third' | 'two-fifths' | 'half'
      mediaShape: 'square' | 'portrait' | 'landscape'
    }

export type LpwLayoutRole =
  | 'primary'
  | 'secondary'
  | 'media'
  | 'body'
  | 'lead'
  | 'aside'
  | 'full'

export interface LpwLayoutPlacement {
  nodeId: string
  role?: LpwLayoutRole
  colSpan?: 1 | 2 | 3 | 4
  rowSpan?: 1 | 2 | 3
  orderOnMobile?: number
}

export interface LpwLayoutProps {
  pattern: LpwLayoutPattern
  direction?: 'horizontal' | 'vertical'
  strategy?: LpwLayoutStrategy
  gap?: 'sm' | 'md' | 'lg'
  align?: 'start' | 'center' | 'stretch'
  alternateFrom?: 'media' | 'body'
  placements?: LpwLayoutPlacement[]
}

export interface LpwLayout extends LpwNodeBase {
  kind: 'layout'
  type: 'layout'
  props: LpwLayoutProps
  children: LpwLayoutChild[]
}

// ── Container ───────────────────────────────────────────
export interface LpwSectionContainerProps {
  variant?: 'article' | 'feature' | 'evidence'
  title: string
  icon?: LpwIconName
  collapsible?: boolean
  defaultOpen?: boolean
}

export interface LpwPanelContainerProps {
  variant?: 'summary' | 'dashboard' | 'aside'
  title?: string
  icon?: LpwIconName
}

export interface LpwDetailsContainerProps {
  variant?: 'supplement' | 'raw-data'
  summary: string
  defaultOpen?: boolean
}

export interface LpwTabsContainerProps {
  variant?: 'comparison' | 'reference' | 'gallery'
  title?: string
  items: Array<{ key: string; label: string }>
  defaultKey?: string
}

export type LpwContainerProps =
  | LpwSectionContainerProps
  | LpwPanelContainerProps
  | LpwDetailsContainerProps
  | LpwTabsContainerProps

export interface LpwContainer extends LpwNodeBase {
  kind: 'container'
  type: 'section' | 'panel' | 'details' | 'tabs'
  props: LpwContainerProps
  children: LpwBlock[]
}

// ── Block ───────────────────────────────────────────────
export interface LpwBlock<TProps = Record<string, unknown>> extends LpwNodeBase {
  kind: 'block'
  props: TProps
  annotation?: LpwAnnotation
}

export type LpwContentNode = LpwLayout | LpwContainer | LpwBlock
export type LpwLayoutChild = LpwContainer | LpwBlock

export interface LpwDocument {
  version: '1.1'
  meta?: LpwMeta
  content: LpwContentNode[]
}

// ── Annotation ──────────────────────────────────────────
export type LpwAnnotationKind =
  | 'note'
  | 'suggestion'
  | 'todo'
  | 'issue'
  | 'approved'
  | 'question'

export interface LpwAnnotationTarget {
  field: string
  pattern: string
  flags?: 'i' | 'u' | 'iu'
}

export interface LpwAnnotation {
  kind: LpwAnnotationKind
  label?: string
  message: string
  author?: string
  targets?: LpwAnnotationTarget[]
}

// ── 策略组与容器契约类型 ─────────────────────────────────────
export type LpwBlockGroup =
  | 'text'
  | 'media'
  | 'data'
  | 'decision'
  | 'process'
  | 'technical'
  | 'notice'

export interface LpwContainerVariantContract {
  AllowedGroups?: LpwBlockGroup[]
  AllowedTypes?: string[]
  DeniedTypes?: string[]
  MinItems: number
  MaxItems: number
  FirstOf?: string[]
  NoRepeatTypes?: string[]
}

// ── 26 种 Block Props ────────────────────────────────────
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

export interface LpwHeadingProps {
  level?: 1 | 2 | 3
  content: string
  icon?: LpwIconName
}

export interface LpwListItem {
  content: string
  checked?: boolean
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
  highlightLines?: number[]
}

export interface LpwImageProps {
  src: string
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
  width?: string
}

export type LpwChartType =
  | 'line'
  | 'bar'
  | 'area'
  | 'pie'
  | 'donut'
  | 'scatter'
  | 'radar'

export interface LpwChartSeries {
  name: string
  data: Array<number | [number, number]>
}

export interface LpwChartProps {
  chartType: LpwChartType
  title?: string
  categories?: string[]
  series: LpwChartSeries[]
  xLabel?: string
  yLabel?: string
  stacked?: boolean
  legend?: boolean
  height?: number
}

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
  value: number
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

export interface LpwTakeawayProps {
  title?: string
  content: string
}

export interface LpwGlanceItem {
  label: string
  text: string
}
export interface LpwGlanceProps {
  items: LpwGlanceItem[]
}

export interface LpwOpenItem {
  title: string
  detail?: string
  owner?: string
  due?: string
}
export interface LpwOpenItemsProps {
  title?: string
  items: LpwOpenItem[]
}

export interface LpwScorecardProps {
  title?: string
  criteria: Array<{ name: string; weight: number }>
  plans: Array<{ name: string; scores: number[]; recommended?: boolean }>
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

// ── 插槽 Props 契约 ──────────────────────────────────────
export interface LpwLayoutSlotProps<TProps = LpwLayoutProps> {
  nodeId: string
  props: TProps
  location: LpwRenderLocation
  children: LpwLayoutChild[]
}

export interface LpwContainerSlotProps<TProps = LpwContainerProps> {
  nodeId?: string
  blockId?: string
  props: TProps
  location?: LpwRenderLocation
  children?: LpwBlock[]
  childrenBlocks?: LpwBlock[]
  depth?: number
}

export interface LpwBlockSlotProps<TProps = Record<string, unknown>> {
  nodeId?: string
  props: TProps
  location?: LpwRenderLocation
  annotation?: LpwAnnotation
  // 兼容别名
  blockId?: string
  depth?: number
  childrenBlocks?: LpwBlock[]
}

// 兼容旧接口别名
export type LpwSectionProps = LpwSectionContainerProps
export type LpwTabsProps = LpwTabsContainerProps
export type LpwDetailsProps = LpwDetailsContainerProps
export type LpwColumnsProps = Record<string, unknown>
