import type React from 'react'
import { LpwBlockRenderer } from './lpw-block-renderer'
import type { LpwBlock } from './types'

export function renderChildren(
  blocks: LpwBlock[] | undefined,
  depth: number,
): React.ReactNode {
  return blocks?.map((child) => (
    <LpwBlockRenderer key={child.id} block={child} depth={depth + 1} />
  ))
}
