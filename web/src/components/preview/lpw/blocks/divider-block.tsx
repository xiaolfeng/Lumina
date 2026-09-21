import type React from 'react'
import type { LpwBlockSlotProps } from '../types'

export const DividerBlock: React.FC<
  LpwBlockSlotProps<Record<string, unknown>>
> = () => {
  return (
    <div className="relative my-10 flex items-center justify-center">
      <hr className="w-full border-t border-line" />
      <div className="absolute bg-surface-strong px-3 text-sea-ink-soft/40 font-mono text-[10px] uppercase tracking-widest">
        §
      </div>
    </div>
  )
}
