/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { MetricsBlock } from './metrics-block'

afterEach(() => {
  cleanup()
})

describe('MetricsBlock', () => {
  it('应正确渲染指标看板的 label、value、unit 与 trend 状态色', () => {
    render(
      <MetricsBlock
        blockId="mt-1"
        props={{
          items: [
            {
              label: '响应延迟',
              value: 42,
              unit: 'ms',
              trend: 'down',
              change: '-12%',
              desc: 'P99 延迟指标',
            },
            {
              label: 'QPS',
              value: '10k',
              trend: 'up',
              change: '+8%',
            },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('响应延迟')).toBeTruthy()
    expect(screen.getByText('42')).toBeTruthy()
    expect(screen.getByText('ms')).toBeTruthy()
    expect(screen.getByText('-12%')).toBeTruthy()
    expect(screen.getByText('P99 延迟指标')).toBeTruthy()

    const item0 = screen.getByTestId('metric-item-0')
    expect(item0.innerHTML).toContain('text-palm')

    const item1 = screen.getByTestId('metric-item-1')
    expect(item1.innerHTML).toContain('text-kicker')
  })
})
