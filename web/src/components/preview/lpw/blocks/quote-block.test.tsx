/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { QuoteBlock } from './quote-block'

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
    expect(footer?.textContent).toContain('— 泉此方 · 幸运星')
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
})
