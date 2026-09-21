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
      className="my-4 border border-palm/40 bg-palm/5 p-4 text-xs shadow-2xs font-sans"
    >
      <div className="flex items-center gap-2 font-mono font-semibold text-palm">
        <span>组件占位卡</span>
        <span className="text-[10px] text-palm/70">// LACUNA</span>
        <span className="text-sea-ink-soft/70">
          [{block.type}#{block.id}]
        </span>
      </div>
      <p className="mt-1.5 text-sea-ink-soft leading-relaxed">{reason}</p>
      {Object.keys(block.props).length > 0 && (
        <details className="mt-2 text-sea-ink-soft/80">
          <summary className="cursor-pointer select-none font-mono text-[11px] hover:text-sea-ink">
            查看属性 (props)
          </summary>
          <pre className="mt-1.5 overflow-x-auto bg-surface/60 border border-line/40 p-2.5 font-mono text-[11px] text-sea-ink">
            {JSON.stringify(block.props, null, 2)}
          </pre>
        </details>
      )}
    </div>
  )
}
