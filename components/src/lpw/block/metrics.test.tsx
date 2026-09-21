/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { MetricsBlock } from './metrics'

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

  it('Q-08: props.items 缺失时应安全解构判空回退空数组，且大字号数值包含 min-w-0 break-all', () => {
    render(<MetricsBlock blockId="mt-empty" props={{} as any} depth={1} />)
    expect(screen.getByTestId('metrics-block')).toBeTruthy()

    render(
      <MetricsBlock
        blockId="mt-overflow"
        props={{
          items: [{ label: '超长数值', value: '123456789012345678901234567890' }],
        }}
        depth={1}
      />,
    )
    const valEl = screen.getByText('123456789012345678901234567890')
    expect(valEl.className).toContain('min-w-0')
    expect(valEl.className).toContain('break-all')
  })

  it('Q-11 & Q-49: 趋势箭头改用 Lucide 图标并补充 aria-hidden 与 sr-only 说明', () => {
    const { container } = render(
      <MetricsBlock
        blockId="mt-icons"
        props={{
          items: [
            { label: '增长项', value: 100, trend: 'up', change: '+10%' },
            { label: '下降项', value: 80, trend: 'down', change: '-5%' },
          ],
        }}
        depth={1}
      />,
    )

    // 不再使用原生 ↑ ↓ 字符
    expect(container.textContent).not.toContain('↑')
    expect(container.textContent).not.toContain('↓')

    // 包含无障碍 sr-only 读屏文案
    expect(screen.getByText('上升')).toBeTruthy()
    expect(screen.getByText('下降')).toBeTruthy()

    // 图标应有 aria-hidden="true"
    const svgs = container.querySelectorAll('svg')
    expect(svgs.length).toBeGreaterThanOrEqual(2)
    svgs.forEach((svg) => {
      expect(svg.getAttribute('aria-hidden')).toBe('true')
    })
  })
})
