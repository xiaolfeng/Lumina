import type React from 'react'
import type { LpwBlockSlotProps, LpwOpenItemsProps } from '../types'

export const OpenItemsBlock: React.FC<LpwBlockSlotProps<LpwOpenItemsProps>> = ({
  props,
}) => {
  const title = props.title ?? '未决事项'

  return (
    <div
      data-testid="open-items-block"
      className="my-4 border border-dashed border-line bg-surface p-4 text-xs"
    >
      <div className="mb-3 font-semibold text-sm text-sea-ink">{title}</div>

      <div className="divide-y divide-line/60">
        {props.items.map((item, idx) => (
          <div
            key={idx}
            data-testid={`open-item-${idx}`}
            className="flex flex-col gap-1 py-2.5 first:pt-0 last:pb-0 sm:flex-row sm:items-start sm:justify-between"
          >
            <div className="flex-1">
              <div className="font-semibold text-sea-ink">{item.title}</div>
              {item.detail && (
                <p className="mt-0.5 text-xs text-sea-ink-soft leading-relaxed">
                  {item.detail}
                </p>
              )}
            </div>

            {(item.owner || item.due) && (
              <div className="flex shrink-0 items-center gap-2 font-mono text-[11px] text-sea-ink-soft/70">
                {item.owner && <span>负责人: {item.owner}</span>}
                {item.due && <span>截止: {item.due}</span>}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
