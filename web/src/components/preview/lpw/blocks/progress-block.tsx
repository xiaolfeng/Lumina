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
      className="my-4 border border-line bg-surface p-4 text-xs"
    >
      {title && (
        <div className="mb-3 font-semibold text-sm text-sea-ink">{title}</div>
      )}

      <div className="space-y-3">
        {items.map((item, idx) => {
          const status = resolveStatus(item)
          const barColor = STATUS_BAR_COLORS[status]
          const clampedVal = Math.max(0, Math.min(100, item.value))

          return (
            <div
              key={idx}
              data-testid={`progress-item-${idx}`}
              className="space-y-1"
            >
              <div className="flex items-center justify-between text-xs text-sea-ink">
                <span>{item.label}</span>
                <span className="font-mono text-sea-ink-soft">
                  {clampedVal}%
                </span>
              </div>
              {/* 进度条轨道 */}
              <div className="h-2 w-full bg-surface-muted border border-line/40 overflow-hidden">
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
