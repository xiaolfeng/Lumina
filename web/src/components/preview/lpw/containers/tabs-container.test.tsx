/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { lpwRegistry } from '../lpw-registry'
import { TabsContainer } from './tabs-container'

afterEach(() => {
  cleanup()
})

describe('TabsContainer', () => {
  it('应按索引对应 children，初始渲染激活项，切换后渲染第二项', () => {
    lpwRegistry.register('tab-child-1', false, () => <div>第一个面板内容</div>)
    lpwRegistry.register('tab-child-2', false, () => <div>第二个面板内容</div>)

    render(
      <TabsContainer
        blockId="tabs-1"
        props={{
          items: [
            { key: 't1', label: '面板一' },
            { key: 't2', label: '面板二' },
          ],
        }}
        depth={1}
        childrenBlocks={[
          { id: 'c1', type: 'tab-child-1', props: {} },
          { id: 'c2', type: 'tab-child-2', props: {} },
        ]}
      />,
    )

    expect(screen.getByText('第一个面板内容')).toBeTruthy()
    expect(screen.queryByText('第二个面板内容')).toBeNull()

    // 切换到面板二
    const tab2Btn = screen.getByRole('tab', { name: '面板二' })
    fireEvent.click(tab2Btn)

    expect(screen.getByText('第二个面板内容')).toBeTruthy()
    expect(screen.queryByText('第一个面板内容')).toBeNull()
  })

  it('defaultKey 命中有效 key 时应初始激活对应面板，未命中时回退到第一项', () => {
    lpwRegistry.register('item-a', false, () => <div>内容 A</div>)
    lpwRegistry.register('item-b', false, () => <div>内容 B</div>)

    // 命中 b
    render(
      <TabsContainer
        blockId="tabs-2"
        props={{
          items: [
            { key: 'a', label: 'A' },
            { key: 'b', label: 'B' },
          ],
          defaultKey: 'b',
        }}
        depth={1}
        childrenBlocks={[
          { id: 'ca', type: 'item-a', props: {} },
          { id: 'cb', type: 'item-b', props: {} },
        ]}
      />,
    )
    expect(screen.getByText('内容 B')).toBeTruthy()

    cleanup()

    // defaultKey 乱写，回退到 a
    render(
      <TabsContainer
        blockId="tabs-3"
        props={{
          items: [
            { key: 'a', label: 'A' },
            { key: 'b', label: 'B' },
          ],
          defaultKey: 'non-exist',
        }}
        depth={1}
        childrenBlocks={[
          { id: 'ca', type: 'item-a', props: {} },
          { id: 'cb', type: 'item-b', props: {} },
        ]}
      />,
    )
    expect(screen.getByText('内容 A')).toBeTruthy()
  })

  it('缺槽位时应渲染该页签缺少内容块占位提示', () => {
    render(
      <TabsContainer
        blockId="tabs-empty"
        props={{
          items: [
            { key: 'first', label: '第一' },
            { key: 'second', label: '第二' },
          ],
          defaultKey: 'second',
        }}
        depth={1}
        childrenBlocks={[{ id: 'c1', type: 'item-a', props: {} }]} // 只有 1 个 child
      />,
    )

    expect(screen.getByTestId('tabs-empty-slot')).toBeTruthy()
    expect(screen.getByText('该页签缺少内容块')).toBeTruthy()
  })
})
