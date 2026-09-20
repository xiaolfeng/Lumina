import { Badge } from '@lumina/components/ui/badge'
import type React from 'react'
import type { LpwBlockSlotProps, LpwTimelineProps } from '../types'

export const TimelineBlock: React.FC<LpwBlockSlotProps<LpwTimelineProps>> = ({
  props,
}) => {
  return (
    <ul className="my-4 border-l border-line pl-4 space-y-6 text-xs">
      {props.items.map((item, idx) => (
        <li
          key={idx}
          data-testid={`timeline-item-${idx}`}
          className="relative pl-2"
        >
          {/* 左侧轴节点原点 */}
          <div className="absolute -left-[21px] top-1 h-2.5 w-2.5 border border-lagoon bg-surface" />

          <div className="flex items-center gap-2">
            <span className="font-mono text-xs text-sea-ink-soft">
              {item.time}
            </span>
            <span className="font-semibold text-sm text-sea-ink">
              {item.title}
            </span>
            {item.tag && (
              <Badge variant="outline" className="text-[10px]">
                {item.tag}
              </Badge>
            )}
          </div>

          {item.content && (
            <p className="mt-1 text-xs text-sea-ink-soft leading-relaxed">
              {item.content}
            </p>
          )}
        </li>
      ))}
    </ul>
  )
}
