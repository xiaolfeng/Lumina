import { Badge } from '../../ui/badge'
import type React from 'react'
import type { LpwBlockSlotProps, LpwTimelineProps } from '../types'

export const TimelineBlock: React.FC<LpwBlockSlotProps<LpwTimelineProps>> = ({
  props,
}) => {
  return (
    <ul className="space-y-7 border-l border-line pl-6 text-xs font-sans">
      {props.items.map((item, idx) => (
        <li
          key={idx}
          data-testid={`timeline-item-${idx}`}
          className="relative pl-2 group"
        >
          {/* 左侧编年志轴节点古典印记 */}
          <div className="absolute -left-6 top-1.5 h-3 w-3 -translate-x-1/2 rounded-full border-2 border-lagoon bg-surface" />

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
            <div className="mt-2 border-l border-line/70 pl-3 text-xs leading-relaxed text-sea-ink-soft">
              {item.content}
            </div>
          )}
        </li>
      ))}
    </ul>
  )
}
