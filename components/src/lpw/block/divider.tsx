import type React from 'react'
import type { LpwBlockSlotProps } from '../types'

export const DividerBlock: React.FC<
  LpwBlockSlotProps<Record<string, unknown>>
> = () => (
  <div
    data-testid="divider-block"
    role="separator"
    aria-hidden="true"
    className="my-6 sm:my-10 h-px w-full bg-gradient-to-r from-transparent via-line to-transparent"
  />
)

