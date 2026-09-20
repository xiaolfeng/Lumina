/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { LpwBlockRenderer } from './lpw-block-renderer'
import { LpwDocumentViewer } from './document-viewer'
import { lpwRegistry, registerAll } from './index'

vi.mock('./echarts-lazy', () => ({
  loadEcharts: vi.fn().mockResolvedValue({
    echarts: {
      init: vi.fn().mockReturnValue({
        setOption: vi.fn(),
        resize: vi.fn(),
        dispose: vi.fn(),
      }),
    },
  }),
}))

afterEach(() => {
  cleanup()
})

beforeEach(() => {
  lpwRegistry.clear()
  registerAll()
})

describe('LpwDocumentViewer', () => {
  it('应能正确渲染合法文档的 meta title、description 与 tags', () => {
    const source = JSON.stringify({
      version: '1.0',
      meta: {
        title: 'LPW 引擎自检',
        description: 'plan-1 验收样例',
        author: 'xiao_lfeng',
        version: '1.0.0',
        tags: ['demo', 'engine'],
      },
      blocks: [],
    })

    render(<LpwDocumentViewer source={source} />)

    expect(screen.getByText('LPW 引擎自检')).toBeTruthy()
    expect(screen.getByText('plan-1 验收样例')).toBeTruthy()
    expect(screen.getByText('xiao_lfeng')).toBeTruthy()
    expect(screen.getByText('v1.0.0')).toBeTruthy()
    expect(screen.getByText('demo')).toBeTruthy()
    expect(screen.getByText('engine')).toBeTruthy()
    expect(screen.getByTestId('document-empty')).toBeTruthy()
  })

  it('遇到未知组件类型时应渲染 Fallback 占位卡', () => {
    const source = JSON.stringify({
      version: '1.0',
      meta: { title: '测试文档' },
      blocks: [
        {
          id: 'block-unknown',
          type: 'not-a-type',
          props: { message: 'hello' },
        },
      ],
    })

    render(<LpwDocumentViewer source={source} />)

    expect(screen.getByText('测试文档')).toBeTruthy()
    expect(screen.getByTestId('fallback-block-unknown')).toBeTruthy()
    expect(screen.getByText(/未注册的组件类型: not-a-type/)).toBeTruthy()
    expect(screen.getByText(/\[not-a-type#block-unknown\]/)).toBeTruthy()
  })

  it('文档解析语法错误时应渲染文档级错误卡片', () => {
    render(<LpwDocumentViewer source="invalid-json" />)

    expect(screen.getByTestId('document-error')).toBeTruthy()
    expect(screen.getByText('LPW 文档解析失败')).toBeTruthy()
  })

  it('容器嵌套超过 3 层（块 depth 5）时应渲染超深占位卡', () => {
    render(
      <LpwBlockRenderer
        block={{
          id: 'deep-b',
          type: 'markdown',
          props: {},
        }}
        depth={5}
      />,
    )
    expect(screen.getByTestId('fallback-deep-b')).toBeTruthy()
    expect(screen.getByText(/容器嵌套深度超过最大限制/)).toBeTruthy()
  })

  it('组件渲染抛错时错误边界应捕获并展示诊断信息', () => {
    const CrashingComponent = () => {
      throw new Error('组件内部发生未知故障')
    }
    lpwRegistry.register('crasher', false, CrashingComponent)

    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    const source = JSON.stringify({
      version: '1.0',
      blocks: [{ id: 'b-crash', type: 'crasher', props: {} }],
    })

    render(<LpwDocumentViewer source={source} />)

    expect(screen.getByTestId('error-b-crash')).toBeTruthy()
    expect(screen.getByText('块渲染失败')).toBeTruthy()
    expect(screen.getByText(/\[crasher#b-crash\]/)).toBeTruthy()
    expect(screen.getByText('组件内部发生未知故障')).toBeTruthy()

    consoleSpy.mockRestore()
  })

  it('验收测试：Plan 2 文本块样例全部真实渲染且不再是占位卡', () => {
    const sample = JSON.stringify({
      version: '1.0',
      meta: { title: 'plan-2 文本块验收' },
      blocks: [
        {
          id: 'h-1',
          type: 'heading',
          props: { level: 2, content: '文本块验收' },
        },
        {
          id: 'md-1',
          type: 'markdown',
          props: { content: '## 小标题\n\n正文 **加粗** 与 `行内代码`。' },
        },
        {
          id: 'co-1',
          type: 'callout',
          props: {
            level: 'warning',
            title: '注意',
            content: '黄色警示应显示在左侧粗边。',
          },
        },
        {
          id: 'ls-1',
          type: 'list',
          props: {
            style: 'check',
            items: [
              { content: '已完成项', checked: true },
              { content: '未完成项' },
            ],
          },
        },
        {
          id: 'qt-1',
          type: 'quote',
          props: {
            content: '引述内容。',
            author: '设计者',
            source: 'design 0003',
          },
        },
        {
          id: 'cd-1',
          type: 'code',
          props: {
            language: 'go',
            filename: 'main.go',
            content: 'line1\nline2\nline3',
            showLineNumbers: true,
            highlightLines: [2],
          },
        },
        { id: 'dv-1', type: 'divider', props: {} },
        {
          id: 'ca-1',
          type: 'cards',
          props: {
            items: [
              {
                title: '卡片一',
                description: '说明',
                href: 'https://example.com',
              },
              { title: '卡片二' },
            ],
          },
        },
      ],
    })

    const { container } = render(<LpwDocumentViewer source={sample} />)

    // 不应有任何 Fallback 卡片
    expect(screen.queryByText('组件占位卡')).toBeNull()

    // 验证各块真实渲染
    expect(
      screen.getByRole('heading', { level: 2, name: '文本块验收' }),
    ).toBeTruthy()
    expect(screen.getByText('加粗')).toBeTruthy()
    expect(screen.getByText('黄色警示应显示在左侧粗边。')).toBeTruthy()
    expect(screen.getByText('已完成项')).toBeTruthy()
    expect(screen.getByText('引述内容。')).toBeTruthy()
    expect(screen.getByText('main.go')).toBeTruthy()
    expect(container.querySelector('hr.border-line')).toBeTruthy()
    expect(screen.getByRole('link', { name: /卡片一/ })).toBeTruthy()
    expect(screen.getByText('卡片二')).toBeTruthy()
  })

  it('验收测试：Plan 3 数据与图表块样例全部真实渲染且不再是占位卡', () => {
    const sample = JSON.stringify({
      version: '1.0',
      meta: { title: 'plan-3 数据图表块验收' },
      blocks: [
        {
          id: 'mt-1',
          type: 'metrics',
          props: {
            items: [
              {
                label: 'P99',
                value: '42ms',
                trend: 'down',
                change: '-8%',
              },
              { label: 'QPS', value: 12000, unit: 'req/s', trend: 'up' },
            ],
          },
        },
        {
          id: 'st-1',
          type: 'steps',
          props: {
            current: 1,
            items: [
              { title: '引擎' },
              { title: '内容块', status: 'process' },
              { title: '容器' },
            ],
          },
        },
        {
          id: 'tl-1',
          type: 'timeline',
          props: {
            items: [
              { time: '09-01', title: '立项', tag: '里程碑' },
              { time: '09-14', title: 'ADR 定稿' },
            ],
          },
        },
        {
          id: 'df-1',
          type: 'diff',
          props: { language: 'go', oldCode: 'a := 1', newCode: 'a := 2' },
        },
        {
          id: 'tb-1',
          type: 'table',
          props: {
            columns: [
              { key: 'name', title: '名称' },
              { key: 'size', title: '体积' },
            ],
            data: [
              { name: 'echarts', size: '300KB' },
              { name: '总控', size: null },
            ],
            sortable: true,
          },
        },
        {
          id: 'mm-1',
          type: 'mermaid',
          props: { content: 'flowchart LR\n  A-->B' },
        },
        {
          id: 'ch-1',
          type: 'chart',
          props: {
            chartType: 'line',
            categories: ['周一', '周二', '周三'],
            series: [{ name: '请求量', data: [120, 200, 150] }],
            title: '趋势',
          },
        },
        {
          id: 'cp-1',
          type: 'comparison',
          props: {
            plans: [{ name: '方案A', recommended: true }, { name: '方案B' }],
            rows: [
              {
                dimension: '成本',
                values: [
                  { text: '低', verdict: 'good' },
                  { text: '高', verdict: 'bad' },
                ],
              },
            ],
          },
        },
        {
          id: 'pg-1',
          type: 'progress',
          props: {
            items: [
              { label: 'PR1', value: 100 },
              { label: 'PR2', value: 60 },
              { label: 'PR3', value: 0 },
            ],
          },
        },
        {
          id: 'tr-1',
          type: 'tree',
          props: {
            nodes: [
              {
                label: 'internal/',
                note: '后端',
                children: [{ label: 'logic/' }],
              },
            ],
          },
        },
      ],
    })

    render(<LpwDocumentViewer source={sample} />)

    // 不应有任何 Fallback 占位卡
    expect(screen.queryByText('组件占位卡')).toBeNull()

    // 验证各真实渲染块
    expect(screen.getByText('P99')).toBeTruthy()
    expect(screen.getByText('42ms')).toBeTruthy()
    expect(screen.getByText('引擎')).toBeTruthy()
    expect(screen.getByText('09-01')).toBeTruthy()
    expect(screen.getByText('里程碑')).toBeTruthy()
    expect(screen.getByText('echarts')).toBeTruthy()
    expect(screen.getByTestId('mermaid-block')).toBeTruthy()
    expect(screen.getByTestId('chart-block')).toBeTruthy()
    expect(screen.getByText('方案A')).toBeTruthy()
    expect(screen.getByText('PR1')).toBeTruthy()
    expect(screen.getByText('internal/')).toBeTruthy()
    expect(screen.getByText('(后端)')).toBeTruthy()
  })

  it('验收测试：Plan 4 复刻报表块样例全部真实渲染且不再是占位卡', () => {
    const sample = JSON.stringify({
      version: '1.0',
      meta: { title: 'plan-4 复刻报表块验收' },
      blocks: [
        {
          id: 'tk-1',
          type: 'takeaway',
          props: {
            content: 'LPW 已具备承载内部研究特刊的表达力，可以进入实施。',
          },
        },
        {
          id: 'gl-1',
          type: 'glance',
          props: {
            items: [
              { label: '做成', text: '引擎与 26 种内容块' },
              { label: '卡住', text: '后端入口未接' },
              { label: '下一步', text: '容器块与 MCP 工具' },
            ],
          },
        },
        {
          id: 'oi-1',
          type: 'open-items',
          props: {
            items: [
              {
                title: 'Schema 方言复核',
                owner: '筱锋',
                due: '09-30 前',
              },
            ],
          },
        },
        {
          id: 'sc-1',
          type: 'scorecard',
          props: {
            criteria: [
              { name: '包体', weight: 30 },
              { name: '能力', weight: 70 },
            ],
            plans: [
              { name: 'ECharts 按需', scores: [3, 5], recommended: true },
              { name: '全量', scores: [1, 5] },
            ],
          },
        },
        {
          id: 'qd-1',
          type: 'quadrant',
          props: {
            xLabel: '代价',
            yLabel: '收益',
            items: [
              { label: '自建分发', x: 'low', y: 'high' },
              { label: '全量引入', x: 'high', y: 'low' },
            ],
          },
        },
        {
          id: 'pn-1',
          type: 'personnel',
          props: {
            items: [{ name: '张三', role: '前端', duties: '组件实现' }],
          },
        },
      ],
    })

    render(<LpwDocumentViewer source={sample} />)

    // 不应有任何 Fallback 占位卡
    expect(screen.queryByText('组件占位卡')).toBeNull()

    expect(
      screen.getByText('LPW 已具备承载内部研究特刊的表达力，可以进入实施。'),
    ).toBeTruthy()
    expect(screen.getByText('引擎与 26 种内容块')).toBeTruthy()
    expect(screen.getByText('Schema 方言复核')).toBeTruthy()
    expect(screen.getByText('ECharts 按需')).toBeTruthy()
    expect(screen.getByText('自建分发')).toBeTruthy()
    expect(screen.getByText('张三')).toBeTruthy()
  })

  it('验收测试：Plan 5 容器功能（折叠/页签切换/分栏/开合）正常工作且无占位卡', () => {
    // 还原 plan-5 原始样例结构：section > tabs > columns > 叶子（块 depth 4 = 3 层容器 + 叶子）
    // Q-01 回归：该结构后端校验放行，前端也必须无占位卡渲染
    const sample = JSON.stringify({
      version: '1.0',
      meta: { title: 'plan-5 容器验收' },
      blocks: [
        {
          id: 'sec-1',
          type: 'section',
          props: { title: '章节一', collapsible: true },
          children: [
            {
              id: 'md-1',
              type: 'markdown',
              props: { content: '章节内正文。' },
            },
            {
              id: 'tab-1',
              type: 'tabs',
              props: {
                items: [
                  { key: 'a', label: '页签A' },
                  { key: 'b', label: '页签B' },
                ],
                defaultKey: 'b',
              },
              children: [
                {
                  id: 'ls-1',
                  type: 'list',
                  props: { items: [{ content: '页签A 内容' }] },
                },
                {
                  id: 'col-1',
                  type: 'columns',
                  props: { ratio: '1:2' },
                  children: [
                    {
                      id: 'ca-1',
                      type: 'cards',
                      props: { items: [{ title: '左列' }] },
                    },
                    {
                      id: 'tk-1',
                      type: 'takeaway',
                      props: { content: '右列结论。' },
                    },
                  ],
                },
              ],
            },
          ],
        },
        {
          id: 'det-1',
          type: 'details',
          props: { summary: '附录' },
          children: [
            { id: 'cd-1', type: 'code', props: { content: 'SELECT 1;' } },
          ],
        },
      ],
    })

    render(<LpwDocumentViewer source={sample} />)

    // 不应有任何 Fallback 占位卡
    expect(screen.queryByText('组件占位卡')).toBeNull()

    // 章节一折叠
    expect(screen.getByRole('button', { name: /章节一/ })).toBeTruthy()
    expect(screen.getByText('章节内正文。')).toBeTruthy()

    // tabs 默认激活 B，其内 columns(1:2) 的两列叶子（depth 4）正常渲染
    expect(screen.getByText('左列')).toBeTruthy()
    expect(screen.getByText('右列结论。')).toBeTruthy()

    // details 附录存在
    expect(screen.getByText('附录')).toBeTruthy()
    expect(screen.getByText('SELECT 1;')).toBeTruthy()
  })

  it('深度超限测试：4 层容器嵌套时第 5 层叶子渲染超限占位卡，前 3 层容器正常', () => {
    // 构造 4 层容器: section(1) > tabs(2) > columns(3) > section(4) > leaf(5 超限!)
    const deepSample = JSON.stringify({
      version: '1.0',
      blocks: [
        {
          id: 's-1',
          type: 'section',
          props: { title: '第 1 层 Section' },
          children: [
            {
              id: 't-2',
              type: 'tabs',
              props: { items: [{ key: 'tab1', label: '第 2 层' }] },
              children: [
                {
                  id: 'col-3',
                  type: 'columns',
                  props: { ratio: '1:1' },
                  children: [
                    {
                      id: 's-4',
                      type: 'section',
                      props: { title: '第 4 层 Section' },
                      children: [
                        {
                          id: 'ca-5',
                          type: 'cards',
                          props: { items: [{ title: '第 5 层超限 Cards' }] },
                        },
                      ],
                    },
                  ],
                },
              ],
            },
          ],
        },
      ],
    })

    render(<LpwDocumentViewer source={deepSample} />)

    // 第 1、2、3 层容器正常渲染
    expect(screen.getByText('第 1 层 Section')).toBeTruthy()
    expect(screen.getByText('第 2 层')).toBeTruthy()
    expect(screen.getByText('第 4 层 Section')).toBeTruthy()

    // 第 5 层叶子渲染超限占位卡
    expect(screen.getByTestId('fallback-ca-5')).toBeTruthy()
    expect(screen.getByText(/容器嵌套深度超过最大限制/)).toBeTruthy()
  })
})
