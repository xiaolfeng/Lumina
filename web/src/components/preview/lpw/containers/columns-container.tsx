import type React from 'react'
import { renderChildren } from '../render-children'
import type { LpwBlockSlotProps, LpwColumnsProps } from '../types'

const RATIO_CLASSES: Record<NonNullable<LpwColumnsProps['ratio']>, string> = {
  '1:1': 'md:grid-cols-2',
  '1:2': 'md:grid-cols-[1fr_2fr]',
  '2:1': 'md:grid-cols-[2fr_1fr]',
  '1:1:1': 'md:grid-cols-3',
}

export const ColumnsContainer: React.FC<LpwBlockSlotProps<LpwColumnsProps>> = ({
  props,
  depth,
  childrenBlocks = [],
}) => {
  const ratio = props.ratio ?? '1:1'
  const colClass = RATIO_CLASSES[ratio]

  return (
    <div
      data-testid="columns-container"
      data-ratio={ratio}
      className={`my-4 grid grid-cols-1 gap-4 ${colClass}`}
    >
      {childrenBlocks.map((child) => (
        <div key={child.id} className="min-w-0 flex-1">
          {renderChildren([child], depth)}
        </div>
      ))}
    </div>
  )
}
