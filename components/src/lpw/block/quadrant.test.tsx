/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { QuadrantBlock } from './quadrant'

afterEach(() => {
  cleanup()
})

describe('QuadrantBlock', () => {
  it('应准确落位四个象限，空象限显示占位破折号，并显示轴名', () => {
    render(
      <QuadrantBlock
        blockId="qd-1"
        props={{
          title: '战略象限',
          xLabel: '代价',
          yLabel: '收益',
          items: [
            // 左上: low, high
            { label: '自研分发', x: 'low', y: 'high', note: '高 ROI' },
            // 右下: high, low
            { label: '全盘重构', x: 'high', y: 'low' },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('战略象限')).toBeTruthy()
    expect(screen.getByText('代价')).toBeTruthy()
    expect(screen.getByText('收益')).toBeTruthy()

    // 左上包含自研分发
    const leftTop = screen.getByTestId('quadrant-left-top')
    expect(leftTop.textContent).toContain('自研分发')
    expect(leftTop.textContent).toContain('(高 ROI)')

    // 右下包含全盘重构
    const rightBottom = screen.getByTestId('quadrant-right-bottom')
    expect(rightBottom.textContent).toContain('全盘重构')

    // 右上与左下为空象限，应显示「—」
    const rightTop = screen.getByTestId('quadrant-right-top')
    expect(rightTop.textContent).toContain('—')

    const leftBottom = screen.getByTestId('quadrant-left-bottom')
    expect(leftBottom.textContent).toContain('—')
  })

  it('Q-11 & Q-29: 包含动势箭头与移动端响应式网格布局类名', () => {
    const { container } = render(
      <QuadrantBlock
        blockId="qd-2"
        props={{
          xLabel: '影响面',
          yLabel: '可行性',
          items: [],
        }}
        depth={1}
      />,
    )

    const block = container.querySelector('[data-testid="quadrant-block"]')
    expect(block?.className).toContain('overflow-x-auto')
    expect(block?.className).toContain('min-w-0')
    expect(block?.className).toContain('p-3')
    expect(block?.className).toContain('sm:p-6')

    const grid = block?.querySelector('.min-w-\\[260px\\]')
    expect(grid).toBeTruthy()

    const cell = block?.querySelector('[data-testid="quadrant-left-top"]')
    expect(cell?.className).toContain('p-2.5')
    expect(cell?.className).toContain('sm:p-4')
    expect(cell?.className).toContain('min-h-24')
    expect(cell?.className).toContain('sm:min-h-28')

    // 动势箭头 svg
    const svgs = block?.querySelectorAll('svg')
    expect(svgs?.length).toBeGreaterThanOrEqual(2)
  })
})
