import type React from 'react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { loadEcharts } from '../echarts-lazy'
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
  const { chartType, title, categories = [], series = [] } = props
  const legendShow = props.legend ?? true

  const option: Record<string, unknown> = {
    color: CHART_COLORS,
    tooltip: {
      trigger: chartType === 'pie' || chartType === 'donut' ? 'item' : 'axis',
    },
    legend: {
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
      left: 40,
      right: 30,
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
      if (chartType === 'bar' && props.stacked) {
        sItem.stack = 'total'
      }
      return sItem
    })
  } else if (chartType === 'pie' || chartType === 'donut') {
    const firstSeries = series[0] as (typeof series)[0] | undefined
    const pieData = categories.map((cat, idx) => {
      const rawVal = firstSeries?.data[idx]
      const val = typeof rawVal === 'number' ? rawVal : 0
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
      left: 40,
      right: 30,
      top: title ? 36 : 20,
      bottom: legendShow ? 32 : 16,
    }
    option.tooltip = { trigger: 'item' }
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
  const [ready, setReady] = useState(false)
  const height = props.height ?? 280

  const hasData = useMemo(() => {
    if (props.series.length === 0) return false
    return props.series.some((s) => s.data.length > 0)
  }, [props.series])

  useEffect(() => {
    if (!hasData) return

    let chart: {
      setOption: (opt: unknown) => void
      resize: () => void
      dispose: () => void
    } | null = null
    let ro: ResizeObserver | null = null
    let disposed = false

    void loadEcharts().then(({ echarts }) => {
      if (disposed || !ref.current) return
      chart = echarts.init(ref.current, undefined, { renderer: 'canvas' })
      chart.setOption(buildOption(props))

      if (typeof ResizeObserver !== 'undefined') {
        ro = new ResizeObserver(() => chart?.resize())
        ro.observe(ref.current)
      }
      setReady(true)
    })

    return () => {
      disposed = true
      ro?.disconnect()
      chart?.dispose()
    }
  }, [props, hasData])

  if (!hasData) {
    return (
      <div
        data-testid="chart-empty"
        style={{ height }}
        className="my-4 flex items-center justify-center border border-dashed border-line bg-surface p-4 text-xs text-sea-ink-soft/60"
      >
        暂无图表数据
      </div>
    )
  }

  return (
    <div
      data-testid="chart-block"
      className="my-4 border border-line bg-surface p-4"
    >
      <div
        ref={ref}
        data-testid="chart-canvas-container"
        style={{ height }}
        className="relative w-full"
      >
        {!ready && (
          <div
            data-testid="chart-loading-skeleton"
            style={{ height }}
            className="flex items-center justify-center text-xs text-sea-ink-soft/40"
          >
            图表加载中…
          </div>
        )}
      </div>
    </div>
  )
}
