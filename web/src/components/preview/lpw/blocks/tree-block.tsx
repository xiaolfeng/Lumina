import type React from 'react'
import type { LpwBlockSlotProps, LpwTreeNode, LpwTreeProps } from '../types'

const TreeNodeItem: React.FC<{ node: LpwTreeNode }> = ({ node }) => {
  const hasChildren = node.children && node.children.length > 0

  return (
    <div className="relative my-1">
      <div className="flex items-center gap-1.5 py-0.5">
        <span className="font-mono text-xs text-sea-ink font-medium">
          {node.label}
        </span>
        {node.note && (
          <span className="text-[11px] text-sea-ink-soft/70">
            ({node.note})
          </span>
        )}
      </div>

      {hasChildren && (
        <div className="ml-3 border-l border-line/60 pl-3 space-y-0.5">
          {node.children?.map((child, idx) => (
            <TreeNodeItem key={idx} node={child} />
          ))}
        </div>
      )}
    </div>
  )
}

export const TreeBlock: React.FC<LpwBlockSlotProps<LpwTreeProps>> = ({
  props,
}) => {
  const { title, nodes } = props

  return (
    <div
      data-testid="tree-block"
      className="my-4 border border-line bg-surface p-4 text-xs"
    >
      {title && (
        <div className="mb-3 font-semibold text-sm text-sea-ink">{title}</div>
      )}

      {nodes.length === 0 ? (
        <div className="p-4 text-center text-xs text-sea-ink-soft/60">
          暂无目录树节点
        </div>
      ) : (
        <div className="space-y-1">
          {nodes.map((node, idx) => (
            <TreeNodeItem key={idx} node={node} />
          ))}
        </div>
      )}
    </div>
  )
}
