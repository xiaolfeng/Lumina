import type React from 'react'
import type { LpwBlockSlotProps, LpwGlanceProps } from '../types'

export const GlanceBlock: React.FC<LpwBlockSlotProps<LpwGlanceProps>> = ({
  props,
}) => {
  const len = props.items.length
  if (len === 0) {
    return (
      <div
        data-testid="glance-block"
        className="border border-line bg-surface/30 p-6 text-center text-sm text-sea-ink-soft"
      >
        暂无速览要点
      </div>
    )
  }

  let gridColClass = 'grid-cols-1'
  if (len === 4) {
    gridColClass = 'grid-cols-[repeat(auto-fit,minmax(min(100%,11rem),1fr))]'
  } else if (len > 1) {
    gridColClass = 'grid-cols-[repeat(auto-fit,minmax(min(100%,13rem),1fr))]'
  }

  return (
    <div
      data-testid="glance-block"
      className={`grid gap-3 ${gridColClass}`}
    >
      {props.items.map((item, idx) => {
        return (
          <div
            key={idx}
            data-testid={`glance-item-${idx}`}
            className="relative flex min-h-32 flex-col justify-between border border-line bg-surface/70 p-4 text-xs"
          >
            <div>
              {/* 顶部古典罗马数字与序号 */}
              <div className="mb-3 h-0.5 w-8 bg-lagoon/70" />
              <div className="font-serif text-sm font-semibold text-sea-ink tracking-tight">
                {item.label}
              </div>
            </div>
            <p className="mt-2.5 text-xs text-sea-ink-soft leading-relaxed font-sans">
              {item.text}
            </p>
          </div>
        )
      })}
    </div>
  )
}
