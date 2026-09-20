import type React from 'react'
import type { LpwBlockSlotProps, LpwHeadingProps } from '../types'

export const HeadingBlock: React.FC<LpwBlockSlotProps<LpwHeadingProps>> = ({
  blockId,
  props,
}) => {
  const level = props.level ?? 2
  const commonClass = 'font-semibold text-sea-ink mt-6 mb-3 scroll-mt-16'

  if (level === 1) {
    return (
      <h1 id={blockId} className={`text-3xl ${commonClass}`}>
        {props.content}
      </h1>
    )
  }

  if (level === 3) {
    return (
      <h3 id={blockId} className={`text-xl ${commonClass}`}>
        {props.content}
      </h3>
    )
  }

  return (
    <h2 id={blockId} className={`text-2xl ${commonClass}`}>
      {props.content}
    </h2>
  )
}
