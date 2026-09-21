import type React from 'react'
import type { LpwBlockSlotProps, LpwStepItem, LpwStepsProps } from '../types'

const STATUS_COLORS: Record<
  NonNullable<LpwStepItem['status']>,
  { dot: string; text: string }
> = {
  wait: {
    dot: 'border-line bg-surface-muted text-sea-ink-soft/70',
    text: 'text-sea-ink-soft',
  },
  process: {
    dot: 'border-lagoon bg-lagoon/10 text-lagoon-deep font-semibold',
    text: 'text-lagoon-deep',
  },
  finish: {
    dot: 'border-kicker bg-kicker/10 text-kicker font-semibold',
    text: 'text-kicker',
  },
  error: {
    dot: 'border-destructive bg-destructive/10 text-destructive font-semibold',
    text: 'text-destructive',
  },
}

export const StepsBlock: React.FC<LpwBlockSlotProps<LpwStepsProps>> = ({
  props,
}) => {
  const current = props.current

  return (
    <ol className="my-6 flex flex-wrap items-start gap-6 sm:gap-8 text-xs font-sans">
      {props.items.map((item, idx) => {
        const isCurrent = current !== undefined && current === idx
        const status = item.status ?? (isCurrent ? 'process' : 'wait')
        const color = STATUS_COLORS[status]

        return (
          <li
            key={idx}
            data-testid={`step-item-${idx}`}
            className="flex items-start gap-3 relative group max-w-xs"
          >
            {/* 序号章 */}
            <div
              className={`flex h-6 w-6 shrink-0 items-center justify-center border font-mono text-[11px] shadow-2xs ${
                color.dot
              } ${isCurrent ? 'ring-2 ring-lagoon/40 font-bold' : ''}`}
            >
              {status === 'finish' ? '✓' : status === 'error' ? '!' : idx + 1}
            </div>

            {/* 标题与描述 */}
            <div className="flex flex-col">
              <span
                className={`font-serif text-sm font-semibold tracking-tight text-sea-ink ${
                  isCurrent ? 'text-lagoon-deep' : ''
                }`}
              >
                {item.title}
              </span>
              {item.desc && (
                <span className="mt-1 text-[11px] text-sea-ink-soft leading-relaxed">
                  {item.desc}
                </span>
              )}
            </div>
          </li>
        )
      })}
    </ol>
  )
}
