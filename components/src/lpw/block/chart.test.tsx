/** @vitest-environment jsdom */
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import * as echartsLazy from './echarts-lazy'
import { ChartBlock, CHART_COLORS, buildOption } from './chart'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('ChartBlock & buildOption', () => {
  it('当 ECharts 真实清空容器 innerHTML 时，setReady 不应触发 React removeChild 异常', async () => {
    const initMock = vi.fn().mockImplementation((dom: HTMLElement) => {
      // 严格复现 ZRender/ECharts 对 root.innerHTML = '' 的真实行为
      dom.innerHTML = '<div><canvas></canvas></div>'
      return {
        setOption: vi.fn(),
        resize: vi.fn(),
        dispose: vi.fn(),
      }
    })

    vi.spyOn(echartsLazy, 'loadEcharts').mockResolvedValue({
      echarts: {
        init: initMock,
      } as unknown as Awaited<
        ReturnType<typeof echartsLazy.loadEcharts>
      >['echarts'],
    })

    expect(() => {
      render(
        <ChartBlock
          blockId="ch-repro"
          props={{
            chartType: 'line',
            categories: ['A', 'B'],
            series: [{ name: 'S1', data: [1, 2] }],
          }}
          depth={1}
        />
      )
    }).not.toThrow()

    // 等待异步加载和 setReady 状态提交
    await waitFor(() => {
      expect(initMock).toHaveBeenCalledTimes(1)
    })
  })

  it('loadEcharts 加载失败时应捕获异常并展示错误兜底状态', async () => {
    vi.spyOn(echartsLazy, 'loadEcharts').mockRejectedValue(new Error('Network error'))

    render(
      <ChartBlock
        blockId="ch-fail"
        props={{
          chartType: 'line',
          categories: ['A', 'B'],
          series: [{ name: 'S1', data: [1, 2] }],
        }}
        depth={1}
      />
    )

    await waitFor(() => {
      expect(screen.getByTestId('chart-error')).toBeTruthy()
      expect(screen.getByText('图表组件加载失败')).toBeTruthy()
      expect(screen.getByText('Network error')).toBeTruthy()
    })
  })

  it('buildOption 应正确构建 line、donut 与 radar 的基础结构与固定颜色序列', () => {
    const lineOpt = buildOption({
      chartType: 'line',
      categories: ['A', 'B'],
      series: [{ name: 'S1', data: [10, 20] }],
    })
    expect(lineOpt.color).toEqual(CHART_COLORS)
    const lineSeries = lineOpt.series as Array<{ type: string }>
    expect(lineSeries[0].type).toBe('line')

    const donutOpt = buildOption({
      chartType: 'donut',
      categories: ['C1', 'C2'],
      series: [{ name: 'S1', data: [30, 70] }],
    })
    const donutSeries = donutOpt.series as Array<{
      type: string
      radius: string[]
    }>
    expect(donutSeries[0].type).toBe('pie')
    expect(donutSeries[0].radius).toEqual(['45%', '70%'])

    const radarOpt = buildOption({
      chartType: 'radar',
      categories: ['R1', 'R2'],
      series: [{ name: 'S1', data: [3, 4] }],
    })
    expect(radarOpt.radar).toBeDefined()
    const radarSeries = radarOpt.series as Array<{ type: string }>
    expect(radarSeries[0].type).toBe('radar')
  })

  it('ChartBlock 应正确异步挂载 ECharts 并执行 dispose 清理', async () => {
    const setOptionMock = vi.fn()
    const resizeMock = vi.fn()
    const disposeMock = vi.fn()

    const mockEchartsInstance = {
      setOption: setOptionMock,
      resize: resizeMock,
      dispose: disposeMock,
    }

    const initMock = vi.fn().mockReturnValue(mockEchartsInstance)

    vi.spyOn(echartsLazy, 'loadEcharts').mockResolvedValue({
      echarts: {
        init: initMock,
      } as unknown as Awaited<
        ReturnType<typeof echartsLazy.loadEcharts>
      >['echarts'],
    })

    const { unmount } = render(
      <ChartBlock
        blockId="ch-1"
        props={{
          chartType: 'line',
          categories: ['A', 'B'],
          series: [{ name: 'S1', data: [1, 2] }],
        }}
        depth={1}
      />,
    )

    await waitFor(() => {
      expect(initMock).toHaveBeenCalledTimes(1)
      expect(setOptionMock).toHaveBeenCalledTimes(1)
    })

    unmount()
    expect(disposeMock).toHaveBeenCalledTimes(1)
  })

  it('数据为空时应渲染空态占位，不调用 init', async () => {
    const initMock = vi.fn()
    vi.spyOn(echartsLazy, 'loadEcharts').mockResolvedValue({
      echarts: { init: initMock } as unknown as Awaited<
        ReturnType<typeof echartsLazy.loadEcharts>
      >['echarts'],
    })

    render(
      <ChartBlock
        blockId="ch-empty"
        props={{
          chartType: 'line',
          series: [],
        }}
        depth={1}
      />,
    )

    expect(screen.getByTestId('chart-empty')).toBeTruthy()
    expect(initMock).not.toHaveBeenCalled()
  })
})
