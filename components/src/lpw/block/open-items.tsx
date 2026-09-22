import { Calendar, Circle, User } from 'lucide-react'
import type React from 'react'
import type { LpwBlockSlotProps, LpwOpenItemsProps } from '../types'

export const OpenItemsBlock: React.FC<LpwBlockSlotProps<LpwOpenItemsProps>> = ({
  props,
}) => {
  const title = props.title ?? '未决事项'

  return (
    <div
      data-testid="open-items-block"
      className="border border-line bg-surface/40 p-5 text-xs font-sans sm:p-6"
    >
      <div className="mb-4 pb-2 border-b border-dashed border-line/60 font-serif font-semibold text-base text-sea-ink flex flex-wrap items-center justify-between gap-2">
        <span>{title}</span>

      </div>

      {props.items.length === 0 ? (
        <div className="p-6 text-center text-xs font-serif italic text-sea-ink-soft/70">
          暂无待决事项
        </div>
      ) : (
        <div className="divide-y divide-line/60">
          {props.items.map((item, idx) => (
            <div
              key={idx}
              data-testid={`open-item-${idx}`}
              className="flex flex-col gap-1.5 py-3 first:pt-0 last:pb-0 sm:flex-row sm:items-start sm:justify-between hover:bg-surface/40 px-1 transition-colors"
            >
              <div className="flex-1">
                <div className="font-serif font-semibold text-sm text-sea-ink flex items-center gap-1.5">
                  <Circle className="lucide-circle h-3 w-3 shrink-0 text-sea-ink-soft/70" />
                  <span>{item.title}</span>
                </div>
                {item.detail && (
                  <p className="mt-1 text-xs text-sea-ink-soft leading-relaxed font-sans">
                    {item.detail}
                  </p>
                )}
              </div>

              {(item.owner || item.due) && (
                <div className="flex flex-wrap shrink-0 items-center gap-2 font-mono text-[11px] text-sea-ink-soft sm:pl-4 sm:pt-0.5">
                  {item.owner && (
                    <span className="bg-surface-muted/60 border border-line/40 px-2 py-0.5 flex items-center gap-1">
                      <User className="lucide-user h-3 w-3 shrink-0 text-sea-ink-soft/80" />
                      <span>负责人: {item.owner}</span>
                    </span>
                  )}
                  {item.due && (
                    <span className="bg-surface-muted/60 border border-line/40 px-2 py-0.5 flex items-center gap-1">
                      <Calendar className="lucide-calendar h-3 w-3 shrink-0 text-sea-ink-soft/80" />
                      <span>截止: {item.due}</span>
                    </span>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
