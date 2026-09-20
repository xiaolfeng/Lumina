import type React from 'react'
import type { LpwBlockSlotProps, LpwGlanceProps } from '../types'

export const GlanceBlock: React.FC<LpwBlockSlotProps<LpwGlanceProps>> = ({
  props,
}) => {
  const len = props.items.length
  let gridColClass = 'grid-cols-1'
  if (len === 4) {
    gridColClass = 'sm:grid-cols-2 lg:grid-cols-4'
  } else if (len > 1) {
    gridColClass = 'sm:grid-cols-2 md:grid-cols-3'
  }

  return (
    <div
      data-testid="glance-block"
      className={`my-4 grid gap-4 ${gridColClass}`}
    >
      {props.items.map((item, idx) => {
        const orderStr = String(idx + 1).padStart(2, '0')
        return (
          <div
            key={idx}
            data-testid={`glance-item-${idx}`}
            className="relative border border-line bg-surface p-4 text-xs"
          >
            <div className="flex items-center justify-between">
              <span className="font-semibold text-sm text-sea-ink">
                {item.label}
              </span>
              <span className="font-mono text-xs text-sea-ink-soft/60">
                {orderStr}
              </span>
            </div>
            <p className="mt-2 text-xs text-sea-ink-soft leading-relaxed">
              {item.text}
            </p>
          </div>
        )
      })}
    </div>
  )
}
