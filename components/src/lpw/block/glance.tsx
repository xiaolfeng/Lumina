import type React from 'react'
import type { LpwBlockSlotProps, LpwGlanceProps } from '../types'

const ROMAN_NUMERALS = ['I', 'II', 'III', 'IV', 'V', 'VI', 'VII', 'VIII']

export const GlanceBlock: React.FC<LpwBlockSlotProps<LpwGlanceProps>> = ({
  props,
}) => {
  const len = props.items.length
  if (len === 0) {
    return (
      <div
        data-testid="glance-block"
        className="my-6 border border-line bg-surface/30 p-6 text-center text-xs font-serif italic text-sea-ink-soft/70"
      >
        暂无速览要点
      </div>
    )
  }

  let gridColClass = 'grid-cols-1'
  if (len === 4) {
    gridColClass = 'sm:grid-cols-2 lg:grid-cols-4'
  } else if (len > 1) {
    gridColClass = 'sm:grid-cols-2 md:grid-cols-3'
  }

  return (
    <div
      data-testid="glance-block"
      className={`my-6 grid gap-4 ${gridColClass}`}
    >
      {props.items.map((item, idx) => {
        const orderStr = String(idx + 1).padStart(2, '0')
        const roman = ROMAN_NUMERALS[idx] || String(idx + 1)
        return (
          <div
            key={idx}
            data-testid={`glance-item-${idx}`}
            className="relative border border-line bg-surface p-5 text-xs transition-all duration-150 hover:border-sea-ink/40 shadow-[0_1px_3px_rgba(0,0,0,0.02)] flex flex-col justify-between"
          >
            <div>
              {/* 顶部古典罗马数字与序号 */}
              <div className="flex items-center justify-between border-b border-line/60 pb-2 mb-3">
                <span className="font-serif text-base italic font-semibold text-lagoon">
                  {roman}.
                </span>
                <span className="font-mono text-[11px] text-sea-ink-soft/60">
                  {orderStr}
                </span>
              </div>
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
