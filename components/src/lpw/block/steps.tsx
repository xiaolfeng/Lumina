import { AlertCircle, Check } from 'lucide-react'
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
    <ol className="grid grid-cols-[repeat(auto-fit,minmax(min(100%,13rem),1fr))] gap-5 text-xs font-sans">
      {props.items.map((item, idx) => {
        const isCurrent = current !== undefined && current === idx
        const derivedStatus =
          current !== undefined
            ? idx < current
              ? 'finish'
              : idx === current
                ? 'process'
                : 'wait'
            : 'wait'
        const status = item.status ?? derivedStatus
        const color = STATUS_COLORS[status]

        return (
          <li
            key={idx}
            data-testid={`step-item-${idx}`}
            className="group relative flex w-full items-start gap-3"
          >
            {/* 移动端左侧流程连接视觉元素 */}
            {idx < props.items.length - 1 && (
              <span
                aria-hidden="true"
                className="absolute left-3 top-7 -bottom-4 w-px -translate-x-1/2 bg-line sm:hidden"
              />
            )}

            {/* 序号章 */}
            <div
              className={`flex h-6 w-6 shrink-0 items-center justify-center border font-mono text-[11px] shadow-2xs ${
                color.dot
              } ${isCurrent ? 'ring-2 ring-lagoon/40 font-bold' : ''}`}
            >
              {status === 'finish' ? (
                <Check className="h-3.5 w-3.5" />
              ) : status === 'error' ? (
                <AlertCircle className="h-3.5 w-3.5" />
              ) : (
                idx + 1
              )}
            </div>

            {/* 标题与描述 */}
            <div className="flex flex-col flex-1 min-w-0">
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
