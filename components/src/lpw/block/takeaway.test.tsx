/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { TakeawayBlock } from './takeaway'

afterEach(() => {
  cleanup()
})

describe('TakeawayBlock', () => {
  it('应默认以核心判断为标题并具有墨底类名', () => {
    const { container } = render(
      <TakeawayBlock
        blockId="tk-1"
        props={{ content: '本季度架构演进必须收敛。' }}
        depth={1}
      />,
    )

    expect(screen.getByText('核心判断')).toBeTruthy()
    expect(screen.getByText('本季度架构演进必须收敛。')).toBeTruthy()
    const box = container.querySelector('[data-testid="takeaway-block"]')
    expect(box?.className).toContain('bg-sea-ink')
    expect(box?.className).toContain('text-foam')
  })

  it('支持自定义标题', () => {
    render(
      <TakeawayBlock
        blockId="tk-2"
        props={{ title: '关键决策', content: '结论内容。' }}
        depth={1}
      />,
    )

    expect(screen.getByText('关键决策')).toBeTruthy()
  })

  it('移动端边距调优、印信截断与 Lucide 图标', () => {
    const { container } = render(
      <TakeawayBlock
        blockId="tk-3"
        props={{
          title: '超长超长超长超长超长超长超长的印信标题',
          content: '核心文本内容',
        }}
        depth={1}
      />,
    )

    const box = container.querySelector('[data-testid="takeaway-block"]')
    expect(box?.className).toContain('p-4')
    expect(box?.className).toContain('sm:p-7')

    const tag = container.querySelector('.bg-sea-ink.px-3')
    expect(tag?.className).toContain('max-w-[calc(100%-3rem)]')
    expect(tag?.className).toContain('truncate')

    const icon = container.querySelector('.lucide-quote, .lucide-sparkles')
    expect(icon).toBeTruthy()

    const contentWrapper = screen.getByText('核心文本内容').closest('div')
    expect(contentWrapper?.className).toContain('break-words')
  })
})
