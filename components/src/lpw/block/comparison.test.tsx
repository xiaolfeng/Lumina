/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ComparisonBlock } from './comparison'

afterEach(() => {
  cleanup()
})

describe('ComparisonBlock', () => {
  it('应正确渲染对比表格、推荐 Badge 与 verdict 三态样式', () => {
    render(
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
})
