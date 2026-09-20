import { CalloutBlock } from './blocks/callout-block'
import { CardsBlock } from './blocks/cards-block'
import { ChartBlock } from './blocks/chart-block'
import { CodeBlock } from './blocks/code-block'
import { ComparisonBlock } from './blocks/comparison-block'
import { DiffBlock } from './blocks/diff-block'
import { DividerBlock } from './blocks/divider-block'
import { GalleryBlock } from './blocks/gallery-block'
import { GlanceBlock } from './blocks/glance-block'
import { HeadingBlock } from './blocks/heading-block'
import { ImageBlock } from './blocks/image-block'
import { ListBlock } from './blocks/list-block'
import { MarkdownBlock } from './blocks/markdown-block'
import { MermaidBlock } from './blocks/mermaid-block'
import { MetricsBlock } from './blocks/metrics-block'
import { OpenItemsBlock } from './blocks/open-items-block'
import { PersonnelBlock } from './blocks/personnel-block'
import { ProgressBlock } from './blocks/progress-block'
import { QuadrantBlock } from './blocks/quadrant-block'
import { QuoteBlock } from './blocks/quote-block'
import { ScorecardBlock } from './blocks/scorecard-block'
import { StepsBlock } from './blocks/steps-block'
import { TableBlock } from './blocks/table-block'
import { TakeawayBlock } from './blocks/takeaway-block'
import { TimelineBlock } from './blocks/timeline-block'
import { TreeBlock } from './blocks/tree-block'
import { ColumnsContainer } from './containers/columns-container'
import { DetailsContainer } from './containers/details-container'
import { SectionContainer } from './containers/section-container'
import { TabsContainer } from './containers/tabs-container'
import { lpwRegistry } from './lpw-registry'

export * from './types'
export { parseLpwSource } from './lpw-parser'
export { lpwRegistry } from './lpw-registry'
export { LpwBlockRenderer } from './lpw-block-renderer'
export { LpwDocumentViewer } from './document-viewer'
export { PreviewLpwViewer, PreviewLpwInlineViewer } from './viewers'
export { renderChildren } from './render-children'

// 文本类内容块导出 (9 种)
export { MarkdownBlock } from './blocks/markdown-block'
export { CalloutBlock } from './blocks/callout-block'
export { HeadingBlock } from './blocks/heading-block'
export { ListBlock } from './blocks/list-block'
export { QuoteBlock } from './blocks/quote-block'
export { CodeBlock } from './blocks/code-block'
export { ImageBlock } from './blocks/image-block'
export { DividerBlock } from './blocks/divider-block'
export { CardsBlock } from './blocks/cards-block'

// 数据与图表类内容块导出 (10 种)
export { MetricsBlock } from './blocks/metrics-block'
export { StepsBlock } from './blocks/steps-block'
export { TimelineBlock } from './blocks/timeline-block'
export { DiffBlock } from './blocks/diff-block'
export { TableBlock } from './blocks/table-block'
export { MermaidBlock } from './blocks/mermaid-block'
export { ChartBlock } from './blocks/chart-block'
export { ComparisonBlock } from './blocks/comparison-block'
export { ProgressBlock } from './blocks/progress-block'
export { TreeBlock } from './blocks/tree-block'

// 复刻 /report 模块的内容块导出 (7 种)
export { TakeawayBlock } from './blocks/takeaway-block'
export { GlanceBlock } from './blocks/glance-block'
export { OpenItemsBlock } from './blocks/open-items-block'
export { ScorecardBlock } from './blocks/scorecard-block'
export { QuadrantBlock } from './blocks/quadrant-block'
export { PersonnelBlock } from './blocks/personnel-block'
export { GalleryBlock } from './blocks/gallery-block'

// 容器组织块导出 (4 种)
export { SectionContainer } from './containers/section-container'
export { TabsContainer } from './containers/tabs-container'
export { ColumnsContainer } from './containers/columns-container'
export { DetailsContainer } from './containers/details-container'

/**
 * 注册所有已实现的 LPW 组件 (26 种内容块 + 4 种容器块 = 30 种)
 */
export function registerAll(): void {
  // Plan 2: 9 种文本类
  lpwRegistry.register('markdown', false, MarkdownBlock)
  lpwRegistry.register('callout', false, CalloutBlock)
  lpwRegistry.register('heading', false, HeadingBlock)
  lpwRegistry.register('list', false, ListBlock)
  lpwRegistry.register('quote', false, QuoteBlock)
  lpwRegistry.register('code', false, CodeBlock)
  lpwRegistry.register('image', false, ImageBlock)
  lpwRegistry.register('divider', false, DividerBlock)
  lpwRegistry.register('cards', false, CardsBlock)

  // Plan 3: 10 种数据与图表类
  lpwRegistry.register('metrics', false, MetricsBlock)
  lpwRegistry.register('steps', false, StepsBlock)
  lpwRegistry.register('timeline', false, TimelineBlock)
  lpwRegistry.register('diff', false, DiffBlock)
  lpwRegistry.register('table', false, TableBlock)
  lpwRegistry.register('mermaid', false, MermaidBlock)
  lpwRegistry.register('chart', false, ChartBlock)
  lpwRegistry.register('comparison', false, ComparisonBlock)
  lpwRegistry.register('progress', false, ProgressBlock)
  lpwRegistry.register('tree', false, TreeBlock)

  // Plan 4: 7 种复刻 /report 内容块
  lpwRegistry.register('takeaway', false, TakeawayBlock)
  lpwRegistry.register('glance', false, GlanceBlock)
  lpwRegistry.register('open-items', false, OpenItemsBlock)
  lpwRegistry.register('scorecard', false, ScorecardBlock)
  lpwRegistry.register('quadrant', false, QuadrantBlock)
  lpwRegistry.register('personnel', false, PersonnelBlock)
  lpwRegistry.register('gallery', false, GalleryBlock)

  // Plan 5: 4 种容器组织块
  lpwRegistry.register('section', true, SectionContainer)
  lpwRegistry.register('tabs', true, TabsContainer)
  lpwRegistry.register('columns', true, ColumnsContainer)
  lpwRegistry.register('details', true, DetailsContainer)
}

// 默认执行全量注册
registerAll()
