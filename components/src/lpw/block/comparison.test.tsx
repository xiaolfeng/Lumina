/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ComparisonBlock } from './comparison'

afterEach(() => {
  cleanup()
})

describe('ComparisonBlock', () => {
  it('应正确渲染对比表格、推荐 Badge 与 verdict 三态样式', () => {
    const { container } = render(
      <ComparisonBlock
        blockId="cp-1"
        props={{
          title: '方案选型',
          plans: [{ name: '方案 A', recommended: true }, { name: '方案 B' }],
          rows: [
            {
              dimension: '成本',
              values: [
                { text: '极低', verdict: 'good' },
                { text: '偏高', verdict: 'bad' },
              ],
            },
            {
              dimension: '上手度',
              values: [{ text: '一般', verdict: 'warn' }, { text: '容易' }],
            },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('方案选型')).toBeTruthy()
    expect(screen.getByText('方案 A')).toBeTruthy()
    expect(screen.getByText('推荐')).toBeTruthy()
    expect(screen.getByText('极低').className).toContain('text-kicker')
    expect(screen.getByText('偏高').className).toContain('text-destructive')
    expect(screen.getByText('一般').className).toContain('text-palm')
    expect(screen.getByText('容易').className).toContain('text-sea-ink')

    // Q-35: 维度列应为 th scope="row"，方案列头应为 th scope="col"
    const rowTh = screen.getByText('成本').closest('th')
    expect(rowTh?.getAttribute('scope')).toBe('row')
    const colTh = screen.getByText('方案 A').closest('th')
    expect(colTh?.getAttribute('scope')).toBe('col')

    // Q-35: 辅助阅读隐藏标签
    expect(screen.getByText('(优势)')).toBeTruthy()
    expect(screen.getByText('(劣势)')).toBeTruthy()
    expect(screen.getByText('(注意)')).toBeTruthy()

    // Q-12: 表格保底宽度与首列最小宽度
    const table = rowTh?.closest('table')
    expect(table?.className).toContain('min-w-[520px]')
    const dimHeaderTh = screen.getByText('维度').closest('th')
    expect(dimHeaderTh?.className).toContain('min-w-[100px]')
    expect(dimHeaderTh?.className).toContain('sm:w-1/4')

    const root = container.querySelector('[data-testid="comparison-block"]')
    expect(root?.className).toContain('min-w-0')
    expect(root?.className).toContain('max-w-full')
    expect(root?.className).toContain('overflow-hidden')
  })

  it('rows 为空时渲染暂无对比维度', () => {
    render(
      <ComparisonBlock
        blockId="cp-empty"
        props={{
          plans: [{ name: 'A' }, { name: 'B' }],
          rows: [],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('暂无对比维度')).toBeTruthy()
  })

  it('Q-01: 方案列表为空时渲染空态提示行而不是崩溃', () => {
    render(
      <ComparisonBlock
        blockId="cp-no-plans"
        props={{
          title: '空方案选型',
          plans: [],
          rows: [
            {
              dimension: '成本',
              values: [{ text: '无' }],
            },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('空方案选型')).toBeTruthy()
    expect(screen.getByText('暂无方案对比数据')).toBeTruthy()
  })

  it('Q-01: 方案数据列多于 plans 时应截断数据不解引用越界', () => {
    const { container } = render(
      <ComparisonBlock
        blockId="cp-overflow-values"
        props={{
          plans: [{ name: '单方案' }],
          rows: [
            {
              dimension: '成本',
              values: [
                { text: '方案1值', verdict: 'good' },
                { text: '多余值', verdict: 'bad' },
              ],
            },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('方案1值')).toBeTruthy()
    expect(screen.queryByText('多余值')).toBeNull()
    const tds = container.querySelectorAll('tbody td')
    expect(tds.length).toBe(1)
  })

  it('Q-18: 包含移动端横滑感知遮罩', () => {
    const { container } = render(
      <ComparisonBlock
        blockId="cp-scroll"
        props={{
          plans: [{ name: 'A' }, { name: 'B' }],
          rows: [{ dimension: 'D', values: [{ text: '1' }, { text: '2' }] }],
        }}
        depth={1}
      />,
    )

    const block = container.querySelector('[data-testid="comparison-block"]')
    expect(block).toBeTruthy()
    const mask = block?.querySelector('.sm\\:hidden')
    expect(mask).toBeTruthy()
  })
})
