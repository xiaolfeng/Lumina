import type React from 'react'
import type { LpwBlockSlotProps } from '../types'

export const DividerBlock: React.FC<
  LpwBlockSlotProps<Record<string, unknown>>
> = () => {
  return <hr className="my-6 border-line" />
}
