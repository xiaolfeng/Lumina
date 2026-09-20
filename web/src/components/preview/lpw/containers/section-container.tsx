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
    <section id={blockId} data-testid="section-container" className="my-6">
      <div className="border-b border-line pb-2">
        {collapsible ? (
          <button
            type="button"
            aria-expanded={open}
            onClick={() => setOpen((o) => !o)}
            className="flex w-full cursor-pointer select-none items-center justify-between text-left"
          >
            <h2 className="text-xl font-semibold text-sea-ink">
              {props.title}
            </h2>
            <span className="font-mono text-xs text-sea-ink-soft/70">
              {open ? '▲' : '▼'}
            </span>
          </button>
        ) : (
          <h2 className="text-xl font-semibold text-sea-ink">{props.title}</h2>
        )}
      </div>

      {(!collapsible || open) && (
        <div data-testid="section-content" className="mt-4 space-y-4">
          {renderChildren(childrenBlocks, depth)}
        </div>
      )}
    </section>
  )
}
