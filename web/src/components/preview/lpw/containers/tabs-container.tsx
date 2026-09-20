import React, { useState } from 'react'
import { renderChildren } from '../render-children'
import type { LpwBlockSlotProps, LpwTabsProps } from '../types'

export const TabsContainer: React.FC<LpwBlockSlotProps<LpwTabsProps>> = ({
  props,
  depth,
  childrenBlocks = [],
}) => {
  const { items, defaultKey } = props

  // 校验 defaultKey 是否合法，否则防御性回退到第一项
  const initialKey =
    defaultKey && items.some((it) => it.key === defaultKey)
      ? defaultKey
      : (items[0]?.key ?? '')

  const [activeKey, setActiveKey] = useState<string>(initialKey)

  // Q-06：preview_sync 增量更新可能移除/改名当前激活 key（组件位置不变、state 保留），
  // 此处按 items 实时归一：陈旧 key 自动回落到第一项，避免卡死在「缺槽」占位。
  const resolvedKey = items.some((it) => it.key === activeKey)
    ? activeKey
    : (items[0]?.key ?? '')
  const activeIndex = items.findIndex((it) => it.key === resolvedKey)
  const activeChild = activeIndex >= 0 ? childrenBlocks[activeIndex] : undefined

  return (
    <div
      data-testid="tabs-container"
      className="my-4 border border-line bg-surface"
    >
      {/* TabList 标头 */}
      <div
        role="tablist"
        className="flex overflow-x-auto border-b border-line bg-surface-muted/30"
      >
        {items.map((item) => {
          const isSelected = item.key === resolvedKey
          return (
            <button
              key={item.key}
              role="tab"
              type="button"
              aria-selected={isSelected}
              onClick={() => setActiveKey(item.key)}
              className={`cursor-pointer px-4 py-2 text-xs font-semibold select-none border-b-2 transition-colors ${
                isSelected
                  ? 'border-lagoon text-lagoon-deep bg-surface'
                  : 'border-transparent text-sea-ink-soft hover:text-sea-ink hover:bg-surface-muted/40'
              }`}
            >
              {item.label}
            </button>
          )
        })}
      </div>

      {/* 面板内容 */}
      <div role="tabpanel" className="p-4">
        {activeChild ? (
          renderChildren([activeChild], depth)
        ) : (
          <div
            data-testid="tabs-empty-slot"
            className="p-6 text-center text-xs text-sea-ink-soft/60"
          >
            该页签缺少内容块
          </div>
        )}
      </div>
    </div>
  )
}
