import type React from 'react'
import { renderChildren } from '../render-children'
import type { LpwBlockSlotProps, LpwDetailsProps } from '../types'

export const DetailsContainer: React.FC<LpwBlockSlotProps<LpwDetailsProps>> = ({
  props,
  depth,
  childrenBlocks,
}) => {
  return (
    <details
      data-testid="details-container"
      open={props.defaultOpen ?? false}
      className="my-4 border border-line bg-surface p-4 text-xs"
    >
      <summary className="cursor-pointer select-none font-semibold text-sm text-sea-ink hover:text-lagoon-deep">
        {props.summary}
      </summary>
      <div className="mt-3 space-y-4">
        {renderChildren(childrenBlocks, depth)}
      </div>
    </details>
  )
}
