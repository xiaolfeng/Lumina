import type React from 'react'
import { ICON_COMPONENTS } from '../icon-map'
import type { LpwContainerSlotProps, LpwPanelContainerProps } from '../types'
import { renderContainerBlocks } from '../renderer'
import { containerVariants } from '../contract'

export const PanelContainer: React.FC<
  LpwContainerSlotProps<LpwPanelContainerProps>
> = ({ nodeId, props, location, children }) => {
  const { variant = 'summary', title, icon } = props
  const Icon = icon ? ICON_COMPONENTS[icon] : undefined
  const contract = containerVariants.panel[variant]

  const variantClass =
    variant === 'summary'
      ? 'border-l-4 border-lagoon bg-surface/40'
      : variant === 'aside'
        ? 'border-line/70 bg-surface/20 text-sm'
        : 'border-line bg-surface/50'

  return (
    <div
      id={nodeId}
      data-testid="panel-container"
      data-variant={variant}
      className={`my-6 rounded-none border border-line p-5 shadow-2xs font-sans ${variantClass}`}
    >
      {title && (
        <div className="mb-4 flex items-center gap-2 font-serif font-semibold text-sea-ink">
          {Icon && <Icon className="h-4 w-4 text-lagoon" aria-hidden="true" />}
          <span>{title}</span>
        </div>
      )}
      <div
        className={
          variant === 'dashboard' ? 'grid gap-4 sm:grid-cols-2' : 'space-y-4'
        }
      >
        {renderContainerBlocks(
          children || [],
          location,
          contract,
          `panel/${variant}`,
        )}
      </div>
    </div>
  )
}
