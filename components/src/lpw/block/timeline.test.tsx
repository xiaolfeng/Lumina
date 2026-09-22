/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { TimelineBlock } from './timeline'

afterEach(() => {
  cleanup()
})

describe('TimelineBlock', () => {
  it('应正确渲染时间轴 item，含 tag Badge 与可选 content', () => {
    render(
      <TimelineBlock
        blockId="tl-1"
        props={{
          items: [
            {
              time: '2026-09-01',
              title: '项目立项',
              content: '首期方案论证完毕。',
              tag: '里程碑',
            },
            {
              time: '2026-09-15',
              title: '架构定稿',
            },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('2026-09-01')).toBeTruthy()
    expect(screen.getByText('项目立项')).toBeTruthy()
    expect(screen.getByText('首期方案论证完毕。')).toBeTruthy()
    expect(screen.getByText('里程碑')).toBeTruthy()

    expect(screen.getByText('2026-09-15')).toBeTruthy()
    expect(screen.getByText('架构定稿')).toBeTruthy()
  })

  it('Q-15: 圆点定位采用 -left-6 top-1.5 -translate-x-1/2 精确对齐', () => {
    const { container } = render(
      <TimelineBlock
        blockId="tl-2"
        props={{
          items: [{ time: '2026-09-22', title: '测试基准对齐' }],
        }}
        depth={1}
      />,
    )

    const dot = container.querySelector('[data-testid="timeline-item-0"] > div.absolute')
    expect(dot?.className).toContain('absolute -left-6 top-1.5 h-3 w-3 -translate-x-1/2 rounded-full border-2 border-lagoon bg-surface')
  })
})
