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
})
