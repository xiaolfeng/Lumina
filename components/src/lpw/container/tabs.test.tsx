/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { lpwRegistry } from '../registry'
import { rootLocation } from '../render-location'
import { TabsContainer } from './tabs'

afterEach(() => {
  cleanup()
})

describe('TabsContainer', () => {
  const loc = rootLocation()

  it('应按索引对应 children，初始渲染激活项，切换后渲染第二项', () => {
    lpwRegistry.register('block', 'tab-child-1', {
      displayName: 'Tab1',
      Component: () => <div>第一个面板内容</div>,
      groups: ['text'],
      annotatableFields: [],
    })
    lpwRegistry.register('block', 'tab-child-2', {
      displayName: 'Tab2',
      Component: () => <div>第二个面板内容</div>,
      groups: ['text'],
      annotatableFields: [],
    })

    render(
      <TabsContainer
        nodeId="tabs-1"
        location={loc}
        props={{
          items: [
            { key: 't1', label: '面板一' },
            { key: 't2', label: '面板二' },
          ],
        }}
        children={[
          { id: 'c1', kind: 'block', type: 'tab-child-1', props: {} },
          { id: 'c2', kind: 'block', type: 'tab-child-2', props: {} },
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
    lpwRegistry.register('block', 'item-a', {
      displayName: 'ItemA',
      Component: () => <div>内容 A</div>,
      groups: ['text'],
      annotatableFields: [],
    })
    lpwRegistry.register('block', 'item-b', {
      displayName: 'ItemB',
      Component: () => <div>内容 B</div>,
      groups: ['text'],
      annotatableFields: [],
    })

    // 命中 b
    render(
      <TabsContainer
        nodeId="tabs-2"
        location={loc}
        props={{
          items: [
            { key: 'a', label: 'A' },
            { key: 'b', label: 'B' },
          ],
          defaultKey: 'b',
        }}
        children={[
          { id: 'ca', kind: 'block', type: 'item-a', props: {} },
          { id: 'cb', kind: 'block', type: 'item-b', props: {} },
        ]}
      />,
    )
    expect(screen.getByText('内容 B')).toBeTruthy()

    cleanup()

    // defaultKey 乱写，回退到 a
    render(
      <TabsContainer
        nodeId="tabs-3"
        location={loc}
        props={{
          items: [
            { key: 'a', label: 'A' },
            { key: 'b', label: 'B' },
          ],
          defaultKey: 'not-exist',
        }}
        children={[
          { id: 'ca', kind: 'block', type: 'item-a', props: {} },
          { id: 'cb', kind: 'block', type: 'item-b', props: {} },
        ]}
      />,
    )
    expect(screen.getByText('内容 A')).toBeTruthy()
  })

  it('缺槽位时应渲染该页签缺少内容块占位提示', () => {
    render(
      <TabsContainer
        nodeId="tabs-empty"
        location={loc}
        props={{
          items: [
            { key: 'a', label: 'A' },
            { key: 'b', label: 'B' },
          ],
        }}
        children={[]}
      />,
    )

    expect(screen.getByTestId('tabs-empty-slot')).toBeTruthy()
    expect(screen.getByText('该页签缺少内容块')).toBeTruthy()
  })

  it('Q-06 回归：items 更新移除激活 key 后应回落第一项，而不是卡死在缺槽占位', () => {
    lpwRegistry.register('block', 'q6-child-a', {
      displayName: 'Q6A',
      Component: () => <div>新版第一页内容</div>,
      groups: ['text'],
      annotatableFields: [],
    })
    lpwRegistry.register('block', 'q6-child-b', {
      displayName: 'Q6B',
      Component: () => <div>新版第二页内容</div>,
      groups: ['text'],
      annotatableFields: [],
    })

    const { rerender } = render(
      <TabsContainer
        nodeId="tabs-q6"
        location={loc}
        props={{
          items: [
            { key: 'a', label: 'A' },
            { key: 'b', label: 'B' },
          ],
          defaultKey: 'b',
        }}
        children={[
          { id: 'ca', kind: 'block', type: 'q6-child-a', props: {} },
          { id: 'cb', kind: 'block', type: 'q6-child-b', props: {} },
        ]}
      />,
    )
    // 初始激活 B
    expect(screen.getByText('新版第二页内容')).toBeTruthy()

    // 模拟 preview_sync 后的文档更新：key 'b' 被移除（组件同位置复用，state 保留）
    rerender(
      <TabsContainer
        nodeId="tabs-q6"
        location={loc}
        props={{
          items: [
            { key: 'a', label: 'A' },
            { key: 'c', label: 'C' },
          ],
        }}
        children={[
          { id: 'ca', kind: 'block', type: 'q6-child-a', props: {} },
          { id: 'cc', kind: 'block', type: 'q6-child-b', props: {} },
        ]}
      />,
    )

    // 应自动回落到 items[0]，即 'a' 对应的内容，而不是空白或死循环
    expect(screen.getByText('新版第一页内容')).toBeTruthy()
    expect(screen.queryByTestId('tabs-empty-slot')).toBeNull()
  })

  it('WAI-ARIA APG Tabs 属性、标题、横滑类名与键盘导航', () => {
    lpwRegistry.register('block', 'aria-child-1', {
      displayName: 'Aria1',
      Component: () => <div data-testid="t1-block">内容 1</div>,
      groups: ['text'],
      annotatableFields: [],
    })
    lpwRegistry.register('block', 'aria-child-2', {
      displayName: 'Aria2',
      Component: () => <div data-testid="t2-block">内容 2</div>,
      groups: ['text'],
      annotatableFields: [],
    })

    render(
      <TabsContainer
        nodeId="tabs-aria"
        location={loc}
        props={{
          title: '配置清单',
          items: [
            { key: 'opt1', label: '选项一' },
            { key: 'opt2', label: '选项二' },
          ],
        }}
        children={[
          { id: 'c1', kind: 'block', type: 'aria-child-1', props: {} },
          { id: 'c2', kind: 'block', type: 'aria-child-2', props: {} },
        ]}
      />,
    )

    // 标题渲染验证
    expect(screen.getByText('配置清单')).toBeTruthy()

    // TabList 与 Tab 属性验证
    const tablist = screen.getByRole('tablist')
    const tab1 = screen.getByRole('tab', { name: '选项一' })
    const tab2 = screen.getByRole('tab', { name: '选项二' })

    expect(tab1.id).toBe('tabs-aria-tab-opt1')
    expect(tab1.getAttribute('aria-controls')).toBe('tabs-aria-panel')
    expect(tab1.getAttribute('aria-selected')).toBe('true')
    expect(tab1.getAttribute('tabindex')).toBe('0')
    expect(tab1.className).toContain('shrink-0')
    expect(tab1.className).toContain('whitespace-nowrap')

    expect(tab2.id).toBe('tabs-aria-tab-opt2')
    expect(tab2.getAttribute('aria-controls')).toBe('tabs-aria-panel')
    expect(tab2.getAttribute('aria-selected')).toBe('false')
    expect(tab2.getAttribute('tabindex')).toBe('-1')

    // TabPanel 属性与统一 gap 验证
    const panel = screen.getByRole('tabpanel')
    expect(panel.id).toBe('tabs-aria-panel')
    expect(panel.getAttribute('aria-labelledby')).toBe('tabs-aria-tab-opt1')
    expect(panel.className).toContain('flex')
    expect(panel.className).toContain('gap-6')

    // 键盘导航：ArrowRight 切换到第二个 tab
    fireEvent.keyDown(tablist, { key: 'ArrowRight' })
    expect(screen.getByText('内容 2')).toBeTruthy()
    expect(tab2.getAttribute('aria-selected')).toBe('true')
    expect(tab2.getAttribute('tabindex')).toBe('0')
    expect(tab1.getAttribute('aria-selected')).toBe('false')
    expect(tab1.getAttribute('tabindex')).toBe('-1')
    expect(panel.getAttribute('aria-labelledby')).toBe('tabs-aria-tab-opt2')

    // 键盘导航：再次 ArrowRight 循环回第一个 tab
    fireEvent.keyDown(tablist, { key: 'ArrowRight' })
    expect(screen.getByText('内容 1')).toBeTruthy()
    expect(tab1.getAttribute('aria-selected')).toBe('true')

    // 键盘导航：ArrowLeft 循环切换到最后一个 tab
    fireEvent.keyDown(tablist, { key: 'ArrowLeft' })
    expect(screen.getByText('内容 2')).toBeTruthy()
    expect(tab2.getAttribute('aria-selected')).toBe('true')
  })
})
