import type React from 'react'
import { useState } from 'react'
import { containerVariants } from '../contract'
import { renderContainerBlocks } from '../renderer'
import type { LpwContainerSlotProps, LpwSectionContainerProps } from '../types'

export const SectionContainer: React.FC<
  LpwContainerSlotProps<LpwSectionContainerProps>
> = ({ nodeId, blockId, props, location, children, childrenBlocks }) => {
  const actualId = nodeId || blockId || ''
  const actualChildren = children || childrenBlocks || []
  const variant = props.variant
  const contract = variant ? containerVariants.section[variant] : undefined
  const collapsible = Boolean(props.collapsible)
  const [open, setOpen] = useState<boolean>(props.defaultOpen ?? true)

  const isEvidence = variant === 'evidence'

  return (
    <section
      id={actualId}
      data-testid="section-container"
      data-variant={variant || 'article'}
      className={`my-10 font-sans ${isEvidence ? 'bg-surface/30 p-6 border border-line/40' : ''}`}
    >
      <div className="mb-6 border-l-[3px] border-palm/70 pl-4">
        {collapsible ? (
          <button
            type="button"
            aria-expanded={open}
            onClick={() => setOpen((o) => !o)}
            className="group flex w-full cursor-pointer select-none items-center justify-between text-left"
          >
            <h2 className="font-serif text-xl sm:text-2xl font-semibold tracking-tight text-sea-ink transition-colors group-hover:text-lagoon">
              {props.title}
            </h2>
            <span className="border border-line/60 bg-surface/50 px-2 py-0.5 font-mono text-xs text-sea-ink-soft/70">
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
          {renderContainerBlocks(
            actualChildren,
            location,
            contract,
            `section/${variant || 'article'}`,
          )}
        </div>
      )}
    </section>
  )
}
