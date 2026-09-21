import React from 'react'
import { DiagnosticCard, summarizeProps } from './diagnostics'
import type { LpwDiagnostic } from './diagnostics'
import { rootLocation } from './render-location'
import type { LpwRenderLocation } from './render-location'
import type { LpwBlock } from './types'

interface Props {
  diagnostic?: Omit<LpwDiagnostic, 'reason'>
  block?: LpwBlock
  location?: LpwRenderLocation
  children: React.ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class BlockErrorBoundary extends React.Component<Props, State> {
  state: State = { hasError: false, error: null }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo): void {
    console.error('LPW Block Error:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      const { block, location, diagnostic } = this.props
      const loc = diagnostic?.location ?? location ?? rootLocation()
      const nodeId = diagnostic?.nodeId ?? block?.id ?? 'unknown'
      const nodeType = diagnostic?.nodeType ?? block?.type ?? 'unknown'

      const fullDiagnostic: LpwDiagnostic = {
        code: 'RENDER_EXCEPTION',
        location: loc,
        nodeId,
        nodeType,
        componentName: diagnostic?.componentName,
        reason: this.state.error?.message ?? '未知渲染异常',
        suggestion:
          diagnostic?.suggestion ?? '请检查组件属性格式或图表/语法配置',
        propsSummary:
          diagnostic?.propsSummary ??
          (block?.props ? summarizeProps(block.props) : undefined),
        stack: this.state.error?.stack,
      }

      return (
        <div data-testid={`error-${nodeId}`}>
          <DiagnosticCard diagnostic={fullDiagnostic} />
        </div>
      )
    }

    return this.props.children
  }
}
