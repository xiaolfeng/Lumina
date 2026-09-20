import { Badge } from '@lumina/components/ui/badge'
import type React from 'react'
import type { LpwBlockSlotProps, LpwPersonnelProps } from '../types'

export const PersonnelBlock: React.FC<LpwBlockSlotProps<LpwPersonnelProps>> = ({
  props,
}) => {
  const { title, items } = props

  return (
    <div
      data-testid="personnel-block"
      className="my-4 border border-line bg-surface p-4 text-xs"
    >
      {title && (
        <div className="mb-3 font-semibold text-sm text-sea-ink">{title}</div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {items.map((person, idx) => (
          <div
            key={idx}
            data-testid={`personnel-item-${idx}`}
            className="border border-line/60 bg-surface-muted/20 p-3"
          >
            <div className="flex items-center justify-between gap-2">
              <span className="font-semibold text-sm text-sea-ink">
                {person.name}
              </span>
              <Badge variant="outline" className="text-[10px]">
                {person.role}
              </Badge>
            </div>
            {person.duties && (
              <p className="mt-1.5 text-xs text-sea-ink-soft leading-relaxed">
                {person.duties}
              </p>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
