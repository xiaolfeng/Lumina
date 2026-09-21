import type React from 'react'
import type { LpwBlockSlotProps, LpwHeadingProps } from '../types'

export const HeadingBlock: React.FC<LpwBlockSlotProps<LpwHeadingProps>> = ({
  blockId,
  props,
}) => {
  const level = props.level ?? 2
  const commonClass =
    'font-serif font-semibold text-sea-ink tracking-tight scroll-mt-16'

  if (level === 1) {
    return (
      <h1
        id={blockId}
        className={`text-2xl sm:text-3xl lg:text-4xl mt-10 mb-4 pb-2 border-b border-sea-ink/80 ${commonClass}`}
      >
        {props.content}
      </h1>
    )
  }

  if (level === 3) {
    return (
      <h3
        id={blockId}
        className={`text-base sm:text-lg mt-6 mb-2 ${commonClass}`}
      >
        {props.content}
      </h3>
    )
  }

  return (
    <h2
      id={blockId}
      className={`text-xl sm:text-2xl mt-8 mb-3 pb-1 border-b border-line/80 ${commonClass}`}
    >
      {props.content}
    </h2>
  )
}
