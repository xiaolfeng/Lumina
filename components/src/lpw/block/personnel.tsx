import { Badge } from '../../ui/badge'
import type React from 'react'
import type { LpwBlockSlotProps, LpwPersonnelProps } from '../types'

export const PersonnelBlock: React.FC<LpwBlockSlotProps<LpwPersonnelProps>> = ({
  props,
}) => {
  const { title, items } = props

  return (
    <div
      data-testid="personnel-block"
      className="my-6 border border-line bg-surface/30 p-6 text-xs shadow-2xs font-sans"
    >
      {title && (
        <div className="mb-4 pb-2 border-b border-line/60 font-serif font-semibold text-base text-sea-ink flex items-center justify-between">
          <span>{title}</span>
          <span className="font-mono text-[10px] text-sea-ink-soft uppercase tracking-widest">
            COLOPHON // 干系人
          </span>
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {items.map((person, idx) => (
          <div
            key={idx}
            data-testid={`personnel-item-${idx}`}
            className="border border-line/60 bg-surface/60 p-4 transition-all duration-150 hover:border-sea-ink hover:bg-surface/90 shadow-2xs"
          >
            <div className="flex items-center justify-between gap-2 border-b border-line/40 pb-2 mb-2">
              <span className="font-serif font-semibold text-sm text-sea-ink">
                {person.name}
              </span>
              <Badge
                variant="outline"
                className="font-mono text-[10px] uppercase border-line text-sea-ink-soft px-1.5 py-0"
              >
                {person.role}
              </Badge>
            </div>
            {person.duties && (
              <p className="mt-1 text-xs text-sea-ink-soft leading-relaxed font-sans">
                {person.duties}
              </p>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
