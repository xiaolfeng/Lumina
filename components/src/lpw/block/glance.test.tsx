/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { GlanceBlock } from './glance'

afterEach(() => {
  cleanup()
})

describe('GlanceBlock', () => {
  it('应正确渲染 3 项速览格并保持低噪声视觉层级', () => {
    const { container } = render(
      <GlanceBlock
        blockId="gl-3"
        props={{
          items: [
            { label: '做成', text: '核心引擎已落地' },
            { label: '卡住', text: '等待用户裁决' },
            { label: '下一步', text: '接入后端逻辑' },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('做成')).toBeTruthy()
    expect(container.querySelectorAll('[class~="bg-lagoon/70"]')).toHaveLength(3)
    expect(
      container.querySelector('[data-testid="glance-block"]')?.className,
    ).toContain('repeat(auto-fit')
  })

  it('多项使用 auto-fit 自适应列宽，单项保持通栏', () => {
    const { container, rerender } = render(
      <GlanceBlock
        blockId="gl-4"
        props={{
          items: [
            { label: 'A', text: '1' },
            { label: 'B', text: '2' },
            { label: 'C', text: '3' },
            { label: 'D', text: '4' },
          ],
        }}
        depth={1}
      />,
    )
    expect(
      container.querySelector('[data-testid="glance-block"]')?.className,
    ).toContain('min(100%,11rem)')

    rerender(
      <GlanceBlock
        blockId="gl-1"
        props={{
          items: [{ label: '独项', text: '通栏展示' }],
        }}
        depth={1}
      />,
    )
    expect(
      container.querySelector('[data-testid="glance-block"]')?.className,
    ).toContain('grid-cols-1')
  })

  it('items 为空时安全展示空态', () => {
    render(
      <GlanceBlock
        blockId="gl-empty"
        props={{
          items: [],
        }}
        depth={1}
      />,
    )
    expect(screen.getByText('暂无速览要点')).toBeTruthy()
  })
})
