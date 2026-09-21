import type React from 'react'
import type { LpwBlockSlotProps } from '../types'

export const DividerBlock: React.FC<
  LpwBlockSlotProps<Record<string, unknown>>
> = () => (
  <div
    data-testid="divider-block"
    aria-hidden="true"
    className="my-10 h-px w-full bg-gradient-to-r from-transparent via-line to-transparent"
  />
)
