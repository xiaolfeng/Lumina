import type React from 'react'
import { renderChildren } from '../render-children'
import type { LpwBlockSlotProps, LpwDetailsProps } from '../types'

export const DetailsContainer: React.FC<LpwBlockSlotProps<LpwDetailsProps>> = ({
  props,
  depth,
  childrenBlocks,
}) => {
  return (
    <details
      data-testid="details-container"
      open={props.defaultOpen ?? false}
      className="my-6 border border-line bg-surface/40 p-5 text-xs shadow-2xs font-sans group"
    >
      <summary className="cursor-pointer select-none font-serif font-semibold text-sm text-sea-ink hover:text-lagoon transition-colors flex items-center justify-between">
        <span>{props.summary}</span>
        <span className="font-mono text-[10px] text-sea-ink-soft/60 uppercase tracking-widest">
          ADDENDUM
        </span>
      </summary>
      <div className="mt-4 pt-3 border-t border-line/50 space-y-4">
        {renderChildren(childrenBlocks, depth)}
      </div>
    </details>
  )
}
