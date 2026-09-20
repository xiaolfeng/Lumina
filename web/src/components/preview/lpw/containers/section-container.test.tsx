/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { lpwRegistry } from '../lpw-registry'
import { SectionContainer } from './section-container'

afterEach(() => {
  cleanup()
})

describe('SectionContainer', () => {
  it('无 collapsible 时常驻渲染 h2 标题与子内容块', () => {
    lpwRegistry.register('dummy-text', false, () => <div>子块内容</div>)

    render(
      <SectionContainer
        blockId="sec-1"
        props={{ title: '第一章 概述' }}
        depth={1}
        childrenBlocks={[{ id: 'c1', type: 'dummy-text', props: {} }]}
      />,
    )

    expect(
      screen.getByRole('heading', { level: 2, name: '第一章 概述' }),
    ).toBeTruthy()
    expect(screen.getByText('子块内容')).toBeTruthy()
  })

  it('collapsible 且 defaultOpen 为 false 时初始折叠，点击展开', () => {
    lpwRegistry.register('dummy-text', false, () => <div>可折叠子块</div>)

    render(
      <SectionContainer
        blockId="sec-2"
        props={{
          title: '第二章 详情',
          collapsible: true,
          defaultOpen: false,
        }}
        depth={1}
        childrenBlocks={[{ id: 'c2', type: 'dummy-text', props: {} }]}
      />,
    )

    const btn = screen.getByRole('button', { name: /第二章 详情/ })
    expect(btn.getAttribute('aria-expanded')).toBe('false')
    expect(screen.queryByText('可折叠子块')).toBeNull()

    // 点击展开
    fireEvent.click(btn)
    expect(btn.getAttribute('aria-expanded')).toBe('true')
    expect(screen.getByText('可折叠子块')).toBeTruthy()
  })
})
