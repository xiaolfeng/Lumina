import type React from 'react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { loadEcharts } from './echarts-lazy'
import type { LpwBlockSlotProps, LpwChartProps } from '../types'

/**
 * 微明主题固定图表色系（退化十六进制常量备用）：
 * 1. lagoon: #c9883a
 * 2. kicker: #7a4e1a
 * 3. palm: #b87050
 * 4. sea-ink: #2b2018
 * 5. sea-ink-soft (替代 sand-deep): #8a7c6e
 */
export const CHART_COLORS = [
  '#c9883a',
  '#7a4e1a',
  '#b87050',
  '#2b2018',
  '#8a7c6e',
]

export function buildOption(props: LpwChartProps): Record<string, unknown> {
  const { chartType, title, categories = [] } = props
  const series = (props.series as typeof props.series | undefined) ?? []
  const legendShow = props.legend ?? true

  const option: Record<string, unknown> = {
    color: CHART_COLORS,
    tooltip: {
      trigger: chartType === 'pie' || chartType === 'donut' ? 'item' : 'axis',
      confine: true,
    },
    legend: {
      type: 'scroll',
      show: legendShow,
      bottom: 4,
      textStyle: { fontSize: 11, color: '#8a7c6e' },
    },
  }

  if (title) {
    option.title = {
      text: title,
      textStyle: { fontSize: 13, fontWeight: '600', color: '#2b2018' },
      left: 0,
      top: 0,
    }
  }

  if (chartType === 'line' || chartType === 'bar' || chartType === 'area') {
    option.grid = {
      containLabel: true,
      left: 16,
      right: 16,
      top: title ? 36 : 20,
      bottom: legendShow ? 32 : 16,
    }
    option.xAxis = {
      type: 'category',
      data: categories,
      name: props.xLabel,
      axisLabel: { fontSize: 11, color: '#8a7c6e' },
      axisLine: { lineStyle: { color: '#e8e2d8' } },
    }
    option.yAxis = {
      type: 'value',
      name: props.yLabel,
      axisLabel: { fontSize: 11, color: '#8a7c6e' },
      splitLine: { lineStyle: { color: '#f0ebe1' } },
    }
    option.series = series.map((s) => {
      const sItem: Record<string, unknown> = {
        name: s.name,
        type: chartType === 'bar' ? 'bar' : 'line',
        data: s.data,
      }
      if (chartType === 'area') {
        sItem.areaStyle = { opacity: 0.25 }
      }
      if ((chartType === 'bar' || chartType === 'area') && props.stacked) {
        sItem.stack = 'total'
      }
      return sItem
    })
  } else if (chartType === 'pie' || chartType === 'donut') {
    const firstSeries = series[0] as (typeof series)[0] | undefined
    const pieData = (firstSeries?.data ?? []).map((rawVal, idx) => {
      const cat = categories[idx] ?? (categories.length > 0 ? `Item ${idx + 1}` : String(idx))
      let val = 0
      if (typeof rawVal === 'number') {
        val = rawVal
      } else if (Array.isArray(rawVal)) {
        val = rawVal[1]
      }
      return { name: cat, value: val }
    })
    const sItem: Record<string, unknown> = {
      name: firstSeries?.name ?? title ?? '',
      type: 'pie',
      data: pieData,
      radius: chartType === 'donut' ? ['45%', '70%'] : '65%',
      center: ['50%', '50%'],
      label: { fontSize: 11, color: '#2b2018' },
    }
    option.series = [sItem]
  } else if (chartType === 'scatter') {
    option.grid = {
      containLabel: true,
      left: 16,
      right: 16,
      top: title ? 36 : 20,
      bottom: legendShow ? 32 : 16,
    }
    option.tooltip = { trigger: 'item', confine: true }
    option.xAxis = {
      type: 'value',
      name: props.xLabel,
      axisLabel: { fontSize: 11, color: '#8a7c6e' },
      splitLine: { lineStyle: { color: '#f0ebe1' } },
    }
    option.yAxis = {
      type: 'value',
      name: props.yLabel,
      axisLabel: { fontSize: 11, color: '#8a7c6e' },
      splitLine: { lineStyle: { color: '#f0ebe1' } },
    }
    option.series = series.map((s) => ({
      name: s.name,
      type: 'scatter',
      data: s.data,
    }))
  } else {
    // chartType === 'radar'
    let maxVal = 5
    for (const s of series) {
      for (const d of s.data) {
        if (typeof d === 'number' && d > maxVal) {
          maxVal = Math.ceil(d)
        }
      }
    }
    option.radar = {
      indicator: categories.map((c) => ({ name: c, max: maxVal })),
      axisName: { color: '#8a7c6e', fontSize: 11 },
      splitLine: { lineStyle: { color: '#f0ebe1' } },
    }
    option.series = [
      {
        type: 'radar',
        data: series.map((s) => ({
          value: s.data,
          name: s.name,
        })),
      },
    ]
  }

  return option
}

export const ChartBlock: React.FC<LpwBlockSlotProps<LpwChartProps>> = ({
  props,
}) => {
  const ref = useRef<HTMLDivElement>(null)
  const chartRef = useRef<{
    setOption: (opt: unknown, notMerge?: boolean) => void
    resize: () => void
    dispose: () => void
  } | null>(null)
  const [ready, setReady] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const height = props.height ?? 280

  const series = (props.series as typeof props.series | undefined) ?? []

  const hasData = useMemo(() => {
    if (series.length === 0) return false
    return series.some((s) => s.data.length > 0)
  }, [series])

  // 初始化 ECharts 实例
  useEffect(() => {
    if (!hasData) return

    let ro: ResizeObserver | null = null
    let disposed = false

    void loadEcharts()
      .then(({ echarts }) => {
        if (disposed || !ref.current) return
        const chart = echarts.init(ref.current, undefined, { renderer: 'canvas' })
        chartRef.current = chart
        chart.setOption(buildOption(props), true)

        if (typeof ResizeObserver !== 'undefined') {
          ro = new ResizeObserver(() => {
            chart.resize()
          })
          ro.observe(ref.current)
        }
        setReady(true)
      })
      .catch((err: unknown) => {
        if (disposed) return
        const msg = err instanceof Error ? err.message : '图表库加载失败'
        setLoadError(msg)
      })

    return () => {
      disposed = true
      ro?.disconnect()
      chartRef.current?.dispose()
      chartRef.current = null
    }
  }, [hasData]) // 仅当从无数据切换到有数据或挂载时初始化一次

  // props 发生变化时执行增量 setOption
  useEffect(() => {
    chartRef.current?.setOption(buildOption(props), true)
  }, [props, hasData])

  if (!hasData) {
    return (
      <div
        data-testid="chart-empty"
        style={{ height }}
        className="my-8 flex items-center justify-center border border-dashed border-line bg-surface/40 p-6 text-xs font-serif italic text-sea-ink-soft/70 shadow-2xs"
      >
        暂无图表数据
      </div>
    )
  }

  if (loadError) {
    return (
      <div
        data-testid="chart-error"
        style={{ height }}
        className="my-8 flex flex-col items-center justify-center border border-dashed border-destructive/40 bg-destructive/5 p-6 text-xs text-destructive shadow-2xs font-sans"
      >
        <span className="font-semibold">图表组件加载失败</span>
        <span className="mt-1 text-2xs text-sea-ink-soft/70">{loadError}</span>
      </div>
    )
  }

  return (
    <div
      data-testid="chart-block"
      role="img"
      aria-label={props.title || '数据图表'}
      className="my-8 border-t-2 border-b-2 border-sea-ink bg-surface/30 p-6 shadow-2xs font-sans"
    >
      <div style={{ height }} className="relative w-full">
        {!ready && (
          <div
            data-testid="chart-loading-skeleton"
            style={{ height }}
            className="absolute inset-0 flex items-center justify-center bg-surface-muted/40 font-serif italic text-xs text-sea-ink-soft/60"
          >
            图表渲染中...
          </div>
        )}
        <div ref={ref} style={{ height }} className="w-full" />
      </div>
    </div>
  )
}
