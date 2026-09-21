/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ScorecardBlock } from './scorecard'

afterEach(() => {
  cleanup()
})

describe('ScorecardBlock', () => {
  it('应准确计算加权总分并高亮推荐列', () => {
    // 权重 30% 与 70%
    // 方案 A: 3*30 + 5*70 = 90 + 350 = 440 -> 4.4
    // 方案 B: 4*30 + 4*70 = 120 + 280 = 400 -> 4.0
    render(
      <ScorecardBlock
        blockId="sc-1"
        props={{
          title: '方案评分卡',
          criteria: [
            { name: '包体开销', weight: 30 },
            { name: '渲染能力', weight: 70 },
          ],
          plans: [
            { name: 'ECharts 按需', scores: [3, 5], recommended: true },
            { name: '全量引入', scores: [4, 4] },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('方案评分卡')).toBeTruthy()
    expect(screen.getByText('包体开销')).toBeTruthy()
    expect(screen.getByText('(30%)')).toBeTruthy()
    expect(screen.getByText('推荐')).toBeTruthy()

    // 验证总分
    expect(screen.getByTestId('scorecard-total-0').textContent).toBe('4.4')
    expect(screen.getByTestId('scorecard-total-1').textContent).toBe('4.0')
  })

  it('权重 30/70 与得分 [4, 5] 时总分应计算为 4.7', () => {
    // 4*30 + 5*70 = 120 + 350 = 470 -> 4.7
    render(
      <ScorecardBlock
        blockId="sc-2"
        props={{
          criteria: [
            { name: '易用性', weight: 30 },
            { name: '扩展性', weight: 70 },
          ],
          plans: [{ name: '单方案', scores: [4, 5] }],
        }}
        depth={1}
      />,
    )

    expect(screen.getByTestId('scorecard-total-0').textContent).toBe('4.7')
  })

  it('并列最高分时若无 recommended 字段则不渲染推荐 Badge', () => {
    render(
      <ScorecardBlock
        blockId="sc-tie"
        props={{
          criteria: [{ name: '维度', weight: 100 }],
          plans: [
            { name: '方案 1', scores: [5] },
            { name: '方案 2', scores: [5] },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.queryByText('推荐')).toBeNull()
  })

  it('Q-19: 动态累加 validWeight，仅累加有效分数项；未打分项不按 0 算失真', () => {
    // 准则 1 权重 20，准则 2 权重 30
    // plan 1: 准则 1 得分 4，准则 2 未打分 (undefined) -> validWeight = 20, sum = 4 * 20 = 80 -> 80 / 20 = 4.0
    // plan 2: 全未打分 -> validWeight = 0 -> '—'
    render(
      <ScorecardBlock
        blockId="sc-partial"
        props={{
          criteria: [
            { name: '安全性', weight: 20 },
            { name: '性能', weight: 30 },
          ],
          plans: [
            { name: '部分打分', scores: [4] },
            { name: '全未打分', scores: [] },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByTestId('scorecard-total-0').textContent).toBe('4.0')
    expect(screen.getByTestId('scorecard-total-1').textContent).toBe('—')
  })

  it('Q-18: 包含移动端横滑感知提示遮罩', () => {
    const { container } = render(
      <ScorecardBlock
        blockId="sc-scroll"
        props={{
          criteria: [{ name: 'A', weight: 100 }],
          plans: [{ name: 'P', scores: [5] }],
        }}
        depth={1}
      />,
    )

    const block = container.querySelector('[data-testid="scorecard-block"]')
    expect(block).toBeTruthy()
    const mask = block?.querySelector('.sm\\:hidden')
    expect(mask).toBeTruthy()
  })
})
