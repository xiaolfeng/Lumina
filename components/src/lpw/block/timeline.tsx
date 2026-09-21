import { Badge } from '../../ui/badge'
import type React from 'react'
import type { LpwBlockSlotProps, LpwTimelineProps } from '../types'

export const TimelineBlock: React.FC<LpwBlockSlotProps<LpwTimelineProps>> = ({
  props,
}) => {
  return (
    <ul className="my-8 border-l border-line pl-6 space-y-8 text-xs font-sans">
      {props.items.map((item, idx) => (
        <li
          key={idx}
          data-testid={`timeline-item-${idx}`}
          className="relative pl-2 group"
        >
          {/* 左侧编年志轴节点古典印记 */}
          <div className="absolute -left-[29px] top-1 h-3 w-3 rounded-full border-2 border-sea-ink bg-surface shadow-xs transition-transform group-hover:scale-125" />

          <div className="flex flex-wrap items-center gap-2.5">
            <span className="font-mono text-xs font-medium text-sea-ink-soft/80 tracking-wider">
              {item.time}
            </span>
            <span className="font-serif text-base font-semibold text-sea-ink tracking-tight">
              {item.title}
            </span>
            {item.tag && (
              <Badge
                variant="outline"
                className="text-[10px] font-mono border-line text-sea-ink-soft px-1.5 py-0"
              >
                {item.tag}
              </Badge>
            )}
          </div>

          {item.content && (
            <div className="mt-2 text-xs text-sea-ink-soft leading-relaxed bg-surface/40 border border-line/50 p-3.5 shadow-[0_1px_2px_rgba(0,0,0,0.01)]">
              {item.content}
            </div>
          )}
        </li>
      ))}
    </ul>
  )
}
