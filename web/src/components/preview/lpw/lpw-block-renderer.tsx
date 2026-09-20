import type React from 'react'
import { BlockErrorBoundary } from './block-error-boundary'
import { LpwFallbackBlock } from './fallback-block'
import { lpwRegistry } from './lpw-registry'
import type { LpwBlock } from './types'

const MAX_CONTAINER_DEPTH = 3

export interface LpwBlockRendererProps {
  block: LpwBlock
  depth?: number
}

export const LpwBlockRenderer: React.FC<LpwBlockRendererProps> = ({
  block,
  depth = 1,
}) => {
  if (depth > MAX_CONTAINER_DEPTH) {
    return (
      <LpwFallbackBlock
        block={block}
        reason={`容器嵌套深度超过最大限制（${MAX_CONTAINER_DEPTH} 层）`}
      />
    )
  }

  const entry = lpwRegistry.get(block.type)
  if (!entry) {
    return (
      <LpwFallbackBlock
        block={block}
        reason={`未注册的组件类型: ${block.type}`}
      />
    )
  }

  return (
    <BlockErrorBoundary block={block}>
      <entry.Component
        blockId={block.id}
        props={block.props}
        depth={depth}
        childrenBlocks={block.children}
      />
    </BlockErrorBoundary>
  )
}
