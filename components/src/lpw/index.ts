import { ensureRegistered } from './register-all'

export * from './types'
export * from './contract'
export {
  parseLpwSource,
  type LpwParseResult,
  type LpwParseError,
} from './parser'
export { lpwRegistry } from './registry'
export {
  LpwNodeRenderer,
  LpwBlockRenderer,
  renderLayoutChildren,
  renderContainerBlocks,
} from './renderer'
export { LpwDocumentViewer } from './document-viewer'
export {
  LpwSourceViewer,
  PreviewLpwViewer,
  PreviewLpwInlineViewer,
  useLpwSource,
  type LpwSourceViewerProps,
  type UseLpwSourceResult,
} from './source-viewer'
export {
  LpwRuntimeProvider,
  useLpwRuntime,
  resolveAssetSrcPure,
  type LpwRuntimeConfig,
} from './runtime-provider'
export {
  DiagnosticCard,
  summarizeProps,
  type LpwDiagnostic,
  type LpwDiagnosticCode,
  type DiagnosticCardProps,
} from './diagnostics'
export {
  rootLocation,
  childLocation,
  formatLocation,
  type LpwRenderLocation,
} from './render-location'
export {
  AnnotationFrame,
  AnnotatedText,
  type AnnotationFrameProps,
  type AnnotatedTextProps,
} from './annotation'
export {
  MermaidViewport,
  type MermaidViewportProps,
} from './block/mermaid-viewport'
export { LPW_ICON_NAMES, isLpwIconName, type LpwIconName } from './icon-map'
export {
  default as ReactDiffViewer,
  DiffMethod,
} from 'react-diff-viewer-continued'

// 导出 26 种 Block 组件
export * from './block'

// 导出 4 种 Container 组件
export * from './container'

// 导出 Layout 组件
export * from './layout'

// 导出与执行组件自动注册
export { registerAll, ensureRegistered } from './register-all'

// 模块加载时默认自注册全部 31 种内置节点
ensureRegistered()
