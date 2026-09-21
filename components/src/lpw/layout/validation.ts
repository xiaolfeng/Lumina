import { layoutPatterns } from '../contract'
import type { LpwDiagnostic } from '../diagnostics'
import type { LpwRenderLocation } from '../render-location'
import type { LpwLayout } from '../types'

export function validateLayoutPattern(
  layout: LpwLayout,
  location: LpwRenderLocation,
): LpwDiagnostic[] {
  const diagnostics: LpwDiagnostic[] = []
  const pattern = (layout.props as { pattern?: string }).pattern
  const placements = layout.props.placements || []
  const children = layout.children

  // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
  if (!pattern || !layoutPatterns[pattern]) {
    diagnostics.push({
      code: 'INVALID_LAYOUT_SLOT',
      location,
      nodeId: layout.id,
      nodeType: 'layout',
      componentName: 'Layout',
      reason: `未知的 layout pattern "${String(pattern)}"`,
      suggestion:
        '请指定合法的 pattern (split, alternating, grid, bento, newspaper, editorial-wrap, flow)',
    })
    return diagnostics
  }

  const spec = layoutPatterns[pattern]
  if (children.length < spec.MinChildren || children.length > spec.MaxChildren) {
    diagnostics.push({
      code: 'INVALID_LAYOUT_SLOT',
      location,
      nodeId: layout.id,
      nodeType: 'layout',
      componentName: 'Layout',
      reason: `layout(${pattern}) 子节点数量 (${children.length}) 超出限制 [${spec.MinChildren}, ${spec.MaxChildren}]`,
      suggestion: `请调整子节点数量使其位于 [${spec.MinChildren}, ${spec.MaxChildren}] 之间`,
    })
  }

  if (spec.EvenOnly && children.length % 2 !== 0) {
    diagnostics.push({
      code: 'INVALID_LAYOUT_SLOT',
      location,
      nodeId: layout.id,
      nodeType: 'layout',
      componentName: 'Layout',
      reason: `layout(${pattern}) 子节点数量 (${children.length}) 必须为偶数`,
      suggestion: '请提供偶数个子节点',
    })
  }

  if (pattern === 'editorial-wrap') {
    let hasImg = false
    let hasMd = false
    for (const c of children) {
      if (c.kind !== 'block') {
        diagnostics.push({
          code: 'INVALID_LAYOUT_SLOT',
          location,
          nodeId: layout.id,
          nodeType: 'layout',
          componentName: 'Layout',
          reason: `layout(editorial-wrap) 子节点必须为 block，不能为 ${c.kind}`,
          suggestion: '只能包含一个 image 块和一个 markdown 块',
        })
      } else if (c.type === 'image' && !hasImg) {
        hasImg = true
      } else if (c.type === 'markdown' && !hasMd) {
        hasMd = true
      } else {
        diagnostics.push({
          code: 'INVALID_LAYOUT_SLOT',
          location,
          nodeId: layout.id,
          nodeType: 'layout',
          componentName: 'Layout',
          reason: 'layout(editorial-wrap) 必须恰好由一个 image 块和一个 markdown 块组成',
        })
      }
    }
  }

  // placements 检查
  const childrenIDSet = new Set(children.map((c) => c.id))
  const seenNodeIds = new Set<string>()
  const seenOrders = new Set<number>()
  let hasBodyMarkdown = false

  for (let i = 0; i < placements.length; i++) {
    const p = placements[i]
    if (!childrenIDSet.has(p.nodeId)) {
      diagnostics.push({
        code: 'INVALID_LAYOUT_SLOT',
        location,
        nodeId: layout.id,
        nodeType: 'layout',
        componentName: 'Layout',
        reason: `placements[${i}].nodeId "${p.nodeId}" 不是该 layout 的直接子节点`,
      })
    }
    if (seenNodeIds.has(p.nodeId)) {
      diagnostics.push({
        code: 'INVALID_LAYOUT_SLOT',
        location,
        nodeId: layout.id,
        nodeType: 'layout',
        componentName: 'Layout',
        reason: `placements[${i}].nodeId "${p.nodeId}" 重复出现`,
      })
    }
    seenNodeIds.add(p.nodeId)

    if (typeof p.orderOnMobile === 'number') {
      if (seenOrders.has(p.orderOnMobile)) {
        diagnostics.push({
          code: 'INVALID_LAYOUT_SLOT',
          location,
          nodeId: layout.id,
          nodeType: 'layout',
          componentName: 'Layout',
          reason: `orderOnMobile ${p.orderOnMobile} 重复`,
        })
      }
      seenOrders.add(p.orderOnMobile)
    }

    if (pattern === 'newspaper') {
      const targetChild = children.find((c) => c.id === p.nodeId)
      if (p.role === 'body') {
        if (
          !targetChild ||
          targetChild.kind !== 'block' ||
          targetChild.type !== 'markdown'
        ) {
          diagnostics.push({
            code: 'INVALID_LAYOUT_SLOT',
            location,
            nodeId: layout.id,
            nodeType: 'layout',
            componentName: 'Layout',
            reason: 'layout(newspaper) 中 role=body 的节点必须是 markdown 块',
          })
        }
        hasBodyMarkdown = true
      } else if (p.role !== 'full') {
        if (
          targetChild &&
          ['chart', 'table', 'code', 'mermaid'].includes(targetChild.type)
        ) {
          diagnostics.push({
            code: 'INVALID_LAYOUT_SLOT',
            location,
            nodeId: layout.id,
            nodeType: 'layout',
            componentName: 'Layout',
            reason: `layout(newspaper) 中 ${targetChild.type} 类型只能放在 role=full 位置`,
          })
        }
      }
    }
  }

  if (pattern === 'newspaper' && !hasBodyMarkdown) {
    diagnostics.push({
      code: 'INVALID_LAYOUT_SLOT',
      location,
      nodeId: layout.id,
      nodeType: 'layout',
      componentName: 'Layout',
      reason: 'layout(newspaper) 必须有且仅有一个 role=body 的 markdown 块',
    })
  }

  return diagnostics
}
