/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { CardsBlock } from './cards-block'

afterEach(() => {
  cleanup()
})

describe('CardsBlock', () => {
  it('应为安全 href 渲染链接卡片，无 href 渲染普通卡片', () => {
    render(
      <CardsBlock
        blockId="ca-1"
        props={{
          items: [
            {
              title: '安全外链',
              description: '去往官网',
              href: 'https://example.com',
            },
            {
              title: '普通卡片',
              description: '不可点击',
            },
          ],
        }}
        depth={1}
      />,
    )

    const link = screen.getByRole('link', { name: /安全外链/ })
    expect(link).toBeTruthy()
    expect(link.getAttribute('href')).toBe('https://example.com')

    expect(screen.getByText('普通卡片')).toBeTruthy()
    expect(screen.queryByRole('link', { name: /普通卡片/ })).toBeNull()
  })

  it('危险协议 (javascript:/data:/vbscript:) 应被过滤并退化为 div', () => {
    const { container } = render(
      <CardsBlock
        blockId="ca-2"
        props={{
          items: [
            {
              title: 'XSS 攻击卡片',
              href: 'javascript:alert("hacked")',
            },
          ],
        }}
        depth={1}
      />,
    )

    expect(container.querySelector('a')).toBeNull()
    expect(screen.getByText('XSS 攻击卡片')).toBeTruthy()
  })
})
