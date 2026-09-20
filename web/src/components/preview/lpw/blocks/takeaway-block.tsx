import type React from 'react'
import type { LpwBlockSlotProps, LpwTakeawayProps } from '../types'

export const TakeawayBlock: React.FC<LpwBlockSlotProps<LpwTakeawayProps>> = ({
  props,
}) => {
  const title = props.title ?? '核心判断'

  return (
    <div data-testid="takeaway-block" className="my-4 bg-sea-ink p-5 text-foam">
      <div className="font-mono text-xs tracking-widest uppercase text-foam/70">
        {title}
      </div>
      <div className="mt-2 text-lg font-medium leading-relaxed text-foam">
        {props.content}
      </div>
    </div>
  )
}
