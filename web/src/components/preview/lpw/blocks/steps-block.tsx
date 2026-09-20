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
    <ol className="my-4 flex flex-wrap gap-6 text-xs">
      {props.items.map((item, idx) => {
        const isCurrent = current !== undefined && current === idx
        const status = item.status ?? (isCurrent ? 'process' : 'wait')
        const color = STATUS_COLORS[status]

        return (
          <li
            key={idx}
            data-testid={`step-item-${idx}`}
            className="flex items-start gap-3"
          >
            <div
              className={`flex h-6 w-6 shrink-0 items-center justify-center border font-mono text-[11px] ${
                color.dot
              } ${isCurrent ? 'ring-2 ring-lagoon/40 font-bold' : ''}`}
            >
              {status === 'finish' ? '✓' : status === 'error' ? '!' : idx + 1}
            </div>
            <div className="flex flex-col">
              <span
                className={`font-semibold text-sea-ink ${
                  isCurrent ? 'text-lagoon-deep' : ''
                }`}
              >
                {item.title}
              </span>
              {item.desc && (
                <span className="mt-0.5 text-[11px] text-sea-ink-soft/70 leading-relaxed max-w-xs">
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
