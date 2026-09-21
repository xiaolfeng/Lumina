import type React from 'react'
import type { LpwBlockSlotProps, LpwTreeNode, LpwTreeProps } from '../types'

const TreeNodeItem: React.FC<{ node: LpwTreeNode }> = ({ node }) => {
  const hasChildren = node.children && node.children.length > 0

  return (
    <div className="relative my-1">
      <div className="flex items-center gap-2 py-1">
        <span className="font-mono text-xs text-sea-ink font-semibold">
          {node.label}
        </span>
        {node.note && (
          <span className="font-serif italic text-[11.5px] text-sea-ink-soft/80">
            ({node.note})
          </span>
        )}
      </div>

      {hasChildren && (
        <div className="ml-3.5 border-l border-line/80 pl-3.5 space-y-1 my-0.5">
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
      className="my-6 border border-line bg-surface/30 p-6 text-xs shadow-2xs font-sans"
    >
      {title && (
        <div className="mb-4 pb-2 border-b border-line/60 font-serif font-semibold text-base text-sea-ink flex items-center justify-between">
          <span>{title}</span>
          <span className="font-mono text-[10px] text-sea-ink-soft uppercase tracking-widest">
            TAXONOMY TREE // 架构目录
          </span>
        </div>
      )}

      {nodes.length === 0 ? (
        <div className="p-6 text-center text-xs font-serif italic text-sea-ink-soft/70">
          暂无目录树节点
        </div>
      ) : (
        <div className="space-y-1 bg-surface/40 border border-line/40 p-4">
          {nodes.map((node, idx) => (
            <TreeNodeItem key={idx} node={node} />
          ))}
        </div>
      )}
    </div>
  )
}
