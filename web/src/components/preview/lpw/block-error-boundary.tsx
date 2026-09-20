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
          className="my-2 border border-red-500/30 bg-red-500/5 p-3 text-xs"
        >
          <div className="flex items-center gap-1.5 font-semibold text-red-600">
            <span>块渲染失败</span>
            <span className="text-sea-ink-soft/60">
              [{this.props.block.type}#{this.props.block.id}]
            </span>
          </div>
          <p className="mt-1 font-mono text-sea-ink-soft">
            {this.state.error?.message}
          </p>
        </div>
      )
    }
    return this.props.children
  }
}
