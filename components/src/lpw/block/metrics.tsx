import { ArrowDownRight, ArrowUpRight } from 'lucide-react'
import type React from 'react'
import type { LpwBlockSlotProps, LpwMetricsProps } from '../types'

export const MetricsBlock: React.FC<LpwBlockSlotProps<LpwMetricsProps>> = ({
  props,
}) => {
  const items = (props.items as typeof props.items | undefined) ?? []
  const len = items.length
  let gridCols = 'sm:grid-cols-2 lg:grid-cols-3'
  if (len === 4) {
    gridCols = 'sm:grid-cols-2 lg:grid-cols-4'
  } else if (len === 2) {
    gridCols = 'sm:grid-cols-2'
  } else if (len === 1) {
    gridCols = 'grid-cols-1'
  }

  return (
    <div
      data-testid="metrics-block"
      className="my-8 border-t border-b border-sea-ink py-6 bg-surface/20"
    >
      <div className={`grid gap-6 sm:gap-0 ${gridCols}`}>
        {items.map((item, idx) => {
          const isLast = idx === len - 1
          return (
            <div
              key={idx}
              data-testid={`metric-item-${idx}`}
              className={`flex flex-col justify-between sm:px-5 ${
                !isLast ? 'sm:border-r sm:border-line/70' : ''
              }`}
            >
              {/* 学术图版编号与标签 */}
              <div>
                <div className="font-mono text-[10px] text-sea-ink-soft/60 mb-1 tracking-wider uppercase">
                  [FIG. 1.{idx + 1}]
                </div>
                <div className="text-xs font-semibold text-sea-ink-soft tracking-wide">
                  {item.label}
                </div>

                {/* 主数值 */}
                <div className="mt-2 mb-2 flex items-baseline gap-1.5 min-w-0">
                  <span className="font-serif text-3xl lg:text-4xl font-bold text-sea-ink tracking-tight leading-none min-w-0 break-all">
                    {item.value}
                  </span>
                  {item.unit && (
                    <span className="text-xs font-mono text-sea-ink-soft shrink-0">
                      {item.unit}
                    </span>
                  )}
                </div>
              </div>

              {/* 细比例规线与趋势注解 */}
              <div className="mt-2 pt-2 border-t border-line/40">
                {(item.trend || item.change) && (
                  <div
                    className={`flex items-center gap-1 text-xs font-mono font-medium ${
                      item.trend === 'up'
                        ? 'text-kicker'
                        : item.trend === 'down'
                          ? 'text-palm'
                          : 'text-sea-ink-soft'
                    }`}
                  >
                    {item.trend === 'up' && (
                      <>
                        <ArrowUpRight className="h-3.5 w-3.5" aria-hidden="true" />
                        <span className="sr-only">上升</span>
                      </>
                    )}
                    {item.trend === 'down' && (
                      <>
                        <ArrowDownRight className="h-3.5 w-3.5" aria-hidden="true" />
                        <span className="sr-only">下降</span>
                      </>
                    )}
                    {item.change && <span>{item.change}</span>}
                  </div>
                )}
                {item.desc && (
                  <p className="mt-1 text-[11.5px] text-sea-ink-soft/80 italic font-serif leading-relaxed">
                    {item.desc}
                  </p>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
