import type React from 'react'
import { DiagnosticCard } from '../diagnostics'
import { renderLayoutChildren } from '../renderer'
import type { LpwLayout, LpwLayoutProps, LpwLayoutSlotProps } from '../types'
import { renderLayoutPreset } from './presets'
import { validateLayoutPattern } from './validation'

export const LayoutNode: React.FC<LpwLayoutSlotProps<LpwLayoutProps>> = ({
  nodeId,
  props,
  location,
  children,
}) => {
  const layoutNode: LpwLayout = {
    id: nodeId,
    kind: 'layout',
    type: 'layout',
    props,
    children,
  }

  const diagnostics = validateLayoutPattern(layoutNode, location)
  if (diagnostics.length > 0) {
    return (
      <div className="space-y-4 my-6">
        {diagnostics.map((d, idx) => (
          <DiagnosticCard key={idx} diagnostic={d} />
        ))}
      </div>
    )
  }

  const renderedChildren = renderLayoutChildren(children, location)

  return (
    <section id={nodeId} data-testid="layout-node" className="my-4">
      {renderLayoutPreset({
        props,
        rawChildren: children,
        renderedChildren: Array.isArray(renderedChildren)
          ? renderedChildren
          : [renderedChildren],
      })}
    </section>
  )
}
