import type React from 'react'
import type { LpwBlock } from './types'

export interface LpwFallbackBlockProps {
  block: LpwBlock
  reason: string
}

export const LpwFallbackBlock: React.FC<LpwFallbackBlockProps> = ({
  block,
  reason,
}) => {
  return (
    <div
      data-testid={`fallback-${block.id}`}
      className="my-2 border border-palm/40 bg-palm/5 p-3 text-xs"
    >
      <div className="flex items-center gap-1.5 font-semibold text-palm">
        <span>组件占位卡</span>
        <span className="text-sea-ink-soft/60">
          [{block.type}#{block.id}]
        </span>
      </div>
      <p className="mt-1 text-sea-ink-soft">{reason}</p>
      {Object.keys(block.props).length > 0 && (
        <details className="mt-2 text-sea-ink-soft/70">
          <summary className="cursor-pointer select-none text-[11px] hover:text-sea-ink">
            查看属性 (props)
          </summary>
          <pre className="mt-1 overflow-x-auto bg-surface-muted/50 p-2 font-mono text-[11px] text-sea-ink">
            {JSON.stringify(block.props, null, 2)}
          </pre>
        </details>
      )}
    </div>
  )
}
