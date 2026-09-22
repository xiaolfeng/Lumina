import type React from 'react'
import type { LpwBlockSlotProps } from '../types'

export const DividerBlock: React.FC<
  LpwBlockSlotProps<Record<string, unknown>>
> = () => (
  <div
    data-testid="divider-block"
    role="separator"
    aria-hidden="true"
    className="my-1 h-px w-full bg-line/70"
  />
)

