import React from 'react'
import type { LpwBlock } from './types'

interface State {
  hasError: boolean
  error: Error | null
}

export class BlockErrorBoundary extends React.Component<
  { block: LpwBlock; children: React.ReactNode },
  State
> {
  state: State = { hasError: false, error: null }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  render() {
    if (this.state.hasError) {
      return (
        <div
          data-testid={`error-${this.props.block.id}`}
          className="my-4 border border-destructive/40 bg-destructive/5 p-4 text-xs shadow-2xs font-sans"
        >
          <div className="flex items-center gap-2 font-mono font-semibold text-destructive">
            <span>块渲染失败</span>
            <span className="text-[10px] text-destructive/70">// ERRATUM</span>
            <span className="text-sea-ink-soft/70">
              [{this.props.block.type}#{this.props.block.id}]
            </span>
          </div>
          <p className="mt-1.5 font-mono text-sea-ink-soft leading-relaxed">
            {this.state.error?.message}
          </p>
        </div>
      )
    }
    return this.props.children
  }
}
