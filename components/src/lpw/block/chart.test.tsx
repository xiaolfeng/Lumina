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

  it('Q-02: props.series 缺失或为空时应安全回退渲染空态，不发生崩溃', () => {
    expect(() => render(<ChartBlock blockId="ch-no-series" props={{ chartType: 'line' } as any} depth={1} />)).not.toThrow()
    expect(screen.getByTestId('chart-empty')).toBeTruthy()
  })

  it('Q-04: 饼图无 categories 时回退使用 series 索引或名称，且 area 图表在 stacked=true 时支持堆叠', () => {
    const pieOptNoCat = buildOption({
      chartType: 'pie',
      series: [{ name: '占比', data: [10, 20] }],
    })
    const pieSeries = pieOptNoCat.series as Array<{
      data: Array<{ name: string; value: number }>
    }>
    expect(pieSeries[0].data).toEqual([
      { name: '0', value: 10 },
      { name: '1', value: 20 },
    ])

    const areaStackedOpt = buildOption({
      chartType: 'area',
      categories: ['Q1', 'Q2'],
      stacked: true,
      series: [
        { name: 'S1', data: [10, 20] },
        { name: 'S2', data: [30, 40] },
      ],
    })
    const areaSeries = areaStackedOpt.series as Array<{
      areaStyle?: { opacity: number }
      stack?: string
    }>
    expect(areaSeries[0].areaStyle).toBeDefined()
    expect(areaSeries[0].stack).toBe('total')
    expect(areaSeries[1].stack).toBe('total')
  })

  it('Q-20 & Q-21: buildOption 视口约束包含 containLabel, left/right 16, tooltip.confine 与 legend.type scroll', () => {
    const opt = buildOption({
      chartType: 'line',
      categories: ['A', 'B'],
      series: [{ name: 'S1', data: [1, 2] }],
    })
    const grid = opt.grid as { containLabel?: boolean; left?: number; right?: number }
    expect(grid.containLabel).toBe(true)
    expect(grid.left).toBe(16)
    expect(grid.right).toBe(16)

    const tooltip = opt.tooltip as { confine?: boolean }
    expect(tooltip.confine).toBe(true)

    const legend = opt.legend as { type?: string }
    expect(legend.type).toBe('scroll')
  })

  it('Q-11 & Q-20: 外层应有 role="img" 和 aria-label，且 props 更新时触发增量 setOption 而非多次 init', async () => {
    const setOptionMock = vi.fn()
    const resizeMock = vi.fn()
    const disposeMock = vi.fn()
    const mockChart = {
      setOption: setOptionMock,
      resize: resizeMock,
      dispose: disposeMock,
    }
    const initMock = vi.fn().mockReturnValue(mockChart)

    vi.spyOn(echartsLazy, 'loadEcharts').mockResolvedValue({
      echarts: { init: initMock } as unknown as Awaited<
        ReturnType<typeof echartsLazy.loadEcharts>
      >['echarts'],
    })

    const { rerender } = render(
      <ChartBlock
        blockId="ch-update"
        props={{
          title: '访问走势',
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

    const container = screen.getByTestId('chart-block')
    expect(container.getAttribute('role')).toBe('img')
    expect(container.getAttribute('aria-label')).toBe('访问走势')

    // 更新 props
    rerender(
      <ChartBlock
        blockId="ch-update"
        props={{
          title: '访问走势 (新)',
          chartType: 'line',
          categories: ['A', 'B'],
          series: [{ name: 'S1', data: [5, 10] }],
        }}
        depth={1}
      />,
    )

    await waitFor(() => {
      expect(initMock).toHaveBeenCalledTimes(1) // 实例单例初始化一次
      expect(setOptionMock).toHaveBeenCalledTimes(2) // 增量更新 setOption
    })
  })
})
