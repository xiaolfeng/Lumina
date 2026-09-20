/** @vitest-environment jsdom */
import { SVGRenderer } from 'echarts/renderers'
import { afterEach, describe, expect, it } from 'vitest'

// 直接加载真实 echarts-module（不走 chart-block 的 mock），用 SVG 渲染器
// 在 jsdom 中验证按需注册完整性：Title 与 Radar 坐标系组件缺失时，
// title 文本与雷达轴标签不会出现在渲染产物中。
describe('echarts-module 按需注册完整性', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('注册 Title 与 Radar 组件：图表标题与雷达轴标签可渲染进 SVG', async () => {
    const { echarts } = await import('./echarts-module')
    echarts.use([SVGRenderer])

    const host = document.createElement('div')
    document.body.appendChild(host)
    const chart = echarts.init(host, null, {
      renderer: 'svg',
      width: 400,
      height: 300,
    })

    chart.setOption({
      title: { text: '雷达回归标题' },
      radar: {
        indicator: [
          { name: 'R1', max: 5 },
          { name: 'R2', max: 5 },
        ],
      },
      series: [{ type: 'radar', data: [{ value: [3, 4] }] }],
    })

    const svg = host.querySelector('svg')
    expect(svg).toBeTruthy()
    // TitleComponent 缺失时标题文本不会渲染（Q-02 复现点）
    expect(svg?.innerHTML).toContain('雷达回归标题')
    // 雷达坐标系与轴标签（echarts 6 的 RadarChart 自带坐标系安装，此处为回归护栏）
    expect(svg?.innerHTML).toContain('R1')
    expect(svg?.innerHTML).toContain('R2')

    chart.dispose()
  })
})
