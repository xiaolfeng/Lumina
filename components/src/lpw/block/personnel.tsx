import { Briefcase } from 'lucide-react'
import { Badge } from '../../ui/badge'
import type React from 'react'
import type { LpwBlockSlotProps, LpwPersonnelProps } from '../types'

export const PersonnelBlock: React.FC<LpwBlockSlotProps<LpwPersonnelProps>> = ({
  props,
}) => {
  const { title, items } = props

  return (
    <div
      data-testid="personnel-block"
      className="border border-line bg-surface/30 p-5 text-xs font-sans sm:p-6"
    >
      {title && (
        <div className="mb-4 pb-2 border-b border-line/60 font-serif font-semibold text-base text-sea-ink flex items-center justify-between">
          <span>{title}</span>

        </div>
      )}

      {items.length === 0 ? (
        <div className="p-6 text-center text-xs font-serif italic text-sea-ink-soft/70">
          暂无干系人信息
        </div>
      ) : (
        <div className="grid grid-cols-[repeat(auto-fit,minmax(min(100%,13rem),1fr))] gap-4">
          {items.map((person, idx) => (
            <div
              key={idx}
              data-testid={`personnel-item-${idx}`}
              className="border border-line/60 bg-surface/60 p-4 transition-all duration-150 hover:border-sea-ink hover:bg-surface/90 shadow-2xs"
            >
              <div className="flex items-center justify-between gap-2 border-b border-line/40 pb-2 mb-2 min-w-0">
                <span className="font-serif font-semibold text-sm text-sea-ink truncate min-w-0">
                  {person.name}
                </span>
                <Badge
                  variant="outline"
                  className="font-mono text-[10px] uppercase border-line text-sea-ink-soft px-1.5 py-0 shrink-0 max-w-[50%] truncate flex items-center gap-1"
                >
                  <Briefcase className="lucide-briefcase h-3 w-3 shrink-0" />
                  <span className="truncate">{person.role}</span>
                </Badge>
              </div>
              {person.duties && (
                <p className="mt-1 text-xs text-sea-ink-soft leading-relaxed font-sans">
                  {person.duties}
                </p>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
