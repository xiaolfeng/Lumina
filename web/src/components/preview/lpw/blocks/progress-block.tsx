import type React from 'react'
import type {
  LpwBlockSlotProps,
  LpwProgressItem,
  LpwProgressProps,
} from '../types'

const STATUS_BAR_COLORS: Record<
  NonNullable<LpwProgressItem['status']>,
  string
> = {
  wait: 'bg-line',
  process: 'bg-lagoon',
  finish: 'bg-kicker',
  error: 'bg-destructive',
}

function resolveStatus(
  item: LpwProgressItem,
): NonNullable<LpwProgressItem['status']> {
  if (item.status) return item.status
  if (item.value >= 100) return 'finish'
  if (item.value > 0) return 'process'
  return 'wait'
}

export const ProgressBlock: React.FC<LpwBlockSlotProps<LpwProgressProps>> = ({
  props,
}) => {
  const { title, items } = props

  return (
    <div
      data-testid="progress-block"
      className="my-6 border border-line bg-surface/30 p-6 text-xs shadow-2xs font-sans"
    >
      {title && (
        <div className="mb-4 pb-2 border-b border-line/60 font-serif font-semibold text-base text-sea-ink flex items-center justify-between">
          <span>{title}</span>
          <span className="font-mono text-[10px] text-sea-ink-soft uppercase tracking-widest">
            GAUGE // 进展规制
          </span>
        </div>
      )}

      <div className="space-y-4">
        {items.map((item, idx) => {
          const status = resolveStatus(item)
          const barColor = STATUS_BAR_COLORS[status]
          const clampedVal = Math.max(0, Math.min(100, item.value))

          return (
            <div
              key={idx}
              data-testid={`progress-item-${idx}`}
              className="space-y-1.5"
            >
              <div className="flex items-center justify-between text-xs text-sea-ink">
                <span className="font-medium text-sea-ink-soft">
                  {item.label}
                </span>
                <span className="font-mono text-xs font-semibold text-sea-ink">
                  {clampedVal}%
                </span>
              </div>
              {/* 进度条轨道 */}
              <div className="h-1.5 w-full bg-surface-muted border border-line/40 overflow-hidden">
                <div
                  data-testid={`progress-bar-${idx}`}
                  style={{ width: `${clampedVal}%` }}
                  className={`h-full transition-all duration-300 ${barColor}`}
                />
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
