/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { QuoteBlock } from './quote'

afterEach(() => {
  cleanup()
})

describe('QuoteBlock', () => {
  it('应正确渲染引述正文且无署名时不渲染 footer', () => {
    const { container } = render(
      <QuoteBlock
        blockId="q-1"
        props={{ content: '单纯的一句名言。' }}
        depth={1}
      />,
    )

    expect(screen.getByText('单纯的一句名言。')).toBeTruthy()
    expect(container.querySelector('footer')).toBeNull()
  })

  it('同时存在 author 与 source 时正确拼接 footer', () => {
    const { container } = render(
      <QuoteBlock
        blockId="q-2"
        props={{
          content: '这是一段设计引述。',
          author: '泉此方',
          source: '幸运星',
        }}
        depth={1}
      />,
    )

    const footer = container.querySelector('footer')
    expect(footer).toBeTruthy()
    expect(footer?.textContent).toContain('— 泉此方')
    expect(footer?.textContent).toContain('·')
    expect(footer?.textContent).toContain('幸运星')
  })

  it('仅有 author 时正确渲染', () => {
    const { container } = render(
      <QuoteBlock
        blockId="q-3"
        props={{ content: '独自发言。', author: '筱锋' }}
        depth={1}
      />,
    )

    expect(container.querySelector('footer')?.textContent).toContain('— 筱锋')
  })

  it('Q-08 内边距为 p-4 sm:p-6 且文本容器包含换行类', () => {
    const { container } = render(
      <QuoteBlock
        blockId="q-paddings"
        props={{ content: '引言文本内容' }}
        depth={1}
      />,
    )
    const root = container.querySelector('blockquote')
    expect(root?.className).toContain('p-4')
    expect(root?.className).toContain('sm:p-6')
    expect(root?.className).not.toContain('p-5')

    const contentDiv = screen.getByText('引言文本内容').closest('div')
    expect(contentDiv?.className).toContain('break-words')
    expect(contentDiv?.className).toContain('[overflow-wrap:anywhere]')
  })

  it('Q-11 & Q-12 source 使用 cite 语义标签且 footer 包含 flex-wrap 与 items-baseline', () => {
    const { container } = render(
      <QuoteBlock
        blockId="q-cite"
        props={{
          content: '引言',
          source: '文集来源',
        }}
        depth={1}
      />,
    )
    const footer = container.querySelector('footer')
    expect(footer?.className).toContain('flex-wrap')
    expect(footer?.className).toContain('items-baseline')
    expect(footer?.className).toContain('gap-1.5')
    const cite = footer?.querySelector('cite')
    expect(cite).toBeTruthy()
    expect(cite?.textContent).toBe('文集来源')
    expect(footer?.textContent).not.toContain('—')
  })

  it('Q-14 根 DOM 节点显式补齐 id', () => {
    const { container } = render(
      <QuoteBlock
        blockId="q-id-node"
        props={{ content: '带 id 引述' }}
        depth={1}
      />,
    )
    expect(container.querySelector('blockquote')?.getAttribute('id')).toBe('q-id-node')
  })

  it('Q-31 在引语首部渲染 Quote 图标', () => {
    const { container } = render(
      <QuoteBlock
        blockId="q-icon"
        props={{ content: '带图标引述' }}
        depth={1}
      />,
    )
    const icon = container.querySelector('svg')
    expect(icon).toBeTruthy()
    expect(icon?.getAttribute('aria-hidden')).toBe('true')
  })
})
