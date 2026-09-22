/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { lpwRegistry } from '../registry'
import { SectionContainer } from './section'

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
        childrenBlocks={[{ id: 'c1', kind: 'block', type: 'dummy-text', props: {} }]}
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
        childrenBlocks={[{ id: 'c2', kind: 'block', type: 'dummy-text', props: {} }]}
      />,
    )

    const btn = screen.getByRole('button', { name: /第二章 详情/ })
    expect(btn.getAttribute('aria-expanded')).toBe('false')
    expect(btn.getAttribute('aria-controls')).toBe('sec-2-content')
    expect(screen.queryByText('可折叠子块')).toBeNull()

    fireEvent.click(btn)
    expect(btn.getAttribute('aria-expanded')).toBe('true')
    expect(screen.getByText('可折叠子块')).toBeTruthy()

    const region = screen.getByRole('region')
    expect(region.id).toBe('sec-2-content')
    expect(region.getAttribute('aria-labelledby')).toBe('sec-2-title')
  })

  it('渲染契约受控图标', () => {
    const { container } = render(
      <SectionContainer
        blockId="sec-icon"
        props={{
          title: '带图标章节',
          icon: 'bookmark',
        }}
        depth={1}
      />,
    )

    const heading = screen.getByRole('heading', { level: 2, name: '带图标章节' })
    expect(heading).toBeTruthy()
    const svg = container.querySelector('h2 svg')
    expect(svg).toBeTruthy()
    expect(svg?.classList.contains('text-palm')).toBe(true)
    expect(svg?.classList.contains('shrink-0')).toBe(true)
  })

  it('title wrapper uses a quiet divider without a competing accent bar', () => {
    const { container } = render(
      <SectionContainer
        blockId="sec-border"
        props={{ title: '无底横线章节' }}
        depth={1}
      />,
    )

    const titleWrapper = container.querySelector('.border-b')
    expect(titleWrapper).toBeTruthy()
    expect(titleWrapper?.className).toContain('border-line/70')
    expect(titleWrapper?.className).not.toContain('border-sea-ink')
  })
})
