import React, { useState } from 'react'
import { renderChildren } from '../render-children'
import type { LpwBlockSlotProps, LpwSectionProps } from '../types'

export const SectionContainer: React.FC<LpwBlockSlotProps<LpwSectionProps>> = ({
  blockId,
  props,
  depth,
  childrenBlocks,
}) => {
  const collapsible = Boolean(props.collapsible)
  const [open, setOpen] = useState<boolean>(props.defaultOpen ?? true)

  return (
    <section
      id={blockId}
      data-testid="section-container"
      className="my-10 font-sans"
    >
      <div className="border-b border-sea-ink/80 pb-3 mb-6">
        {collapsible ? (
          <button
            type="button"
            aria-expanded={open}
            onClick={() => setOpen((o) => !o)}
            className="flex w-full cursor-pointer select-none items-center justify-between text-left group"
          >
            <h2 className="font-serif text-xl sm:text-2xl font-semibold tracking-tight text-sea-ink group-hover:text-lagoon transition-colors">
              {props.title}
            </h2>
            <span className="font-mono text-xs text-sea-ink-soft/70 px-2 py-0.5 border border-line/60 bg-surface/50">
              {open ? '收起 ▲' : '展开 ▼'}
            </span>
          </button>
        ) : (
          <h2 className="font-serif text-xl sm:text-2xl font-semibold tracking-tight text-sea-ink">
            {props.title}
          </h2>
        )}
      </div>

      {(!collapsible || open) && (
        <div data-testid="section-content" className="space-y-6">
          {renderChildren(childrenBlocks, depth)}
        </div>
      )}
    </section>
  )
}
