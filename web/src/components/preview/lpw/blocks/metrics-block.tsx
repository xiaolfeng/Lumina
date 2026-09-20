import type React from 'react'
import type { LpwBlockSlotProps, LpwMetricsProps } from '../types'

export const MetricsBlock: React.FC<LpwBlockSlotProps<LpwMetricsProps>> = ({
  props,
}) => {
  return (
    <div className="my-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {props.items.map((item, idx) => {
        return (
          <div
            key={idx}
            data-testid={`metric-item-${idx}`}
            className="border border-line bg-surface p-4"
          >
            <div className="text-xs text-sea-ink-soft">{item.label}</div>
            <div className="mt-1.5 flex items-baseline gap-1">
              <span className="text-2xl font-semibold text-sea-ink">
                {item.value}
              </span>
              {item.unit && (
                <span className="text-xs font-normal text-sea-ink-soft">
                  {item.unit}
                </span>
              )}
            </div>
            {(item.trend || item.change) && (
              <div
                className={`mt-1 flex items-center gap-1 text-xs ${
                  item.trend === 'up'
                    ? 'text-kicker'
                    : item.trend === 'down'
                      ? 'text-palm'
                      : 'text-sea-ink-soft'
                }`}
              >
                {item.trend === 'up' && <span>↑</span>}
                {item.trend === 'down' && <span>↓</span>}
                {item.change && <span>{item.change}</span>}
              </div>
            )}
            {item.desc && (
              <p className="mt-1 text-[11px] text-sea-ink-soft/70">
                {item.desc}
              </p>
            )}
          </div>
        )
      })}
    </div>
  )
}
