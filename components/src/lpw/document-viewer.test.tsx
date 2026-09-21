/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { LpwDocumentViewer } from './document-viewer'
import { lpwRegistry, registerAll } from './index'
import { LpwNodeRenderer } from './renderer'

vi.mock('./block/echarts-lazy', () => ({
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
      version: '1.1',
      meta: {
        title: 'LPW 引擎自检',
        description: 'plan-1 验收样例',
        author: 'xiao_lfeng',
        version: '1.0.0',
        tags: ['demo', 'engine'],
      },
      content: [],
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

  it('Q-32: DocumentViewer 元信息区（作者、版本、标签）与空状态包含矢量图标与标题折行防护', () => {
    const source = JSON.stringify({
      version: '1.1',
      meta: {
        title: 'SupercalifragilisticexpialidociousLongWordTitleWithoutSpacesToTestWrapping',
        author: 'xiao_lfeng',
        version: '1.0.0',
        tags: ['demo'],
      },
      content: [],
    })

    const { container } = render(<LpwDocumentViewer source={source} />)

    // 标题 break-words [overflow-wrap:anywhere] 防长英文撑爆
    const h1 = container.querySelector('h1')
    expect(h1?.className).toContain('break-words')
    expect(h1?.className).toContain('[overflow-wrap:anywhere]')

    // 元信息区图标：User, GitBranch, Tag
    expect(container.querySelector('svg.lucide-user')).toBeTruthy()
    expect(container.querySelector('svg.lucide-git-branch')).toBeTruthy()
    expect(container.querySelector('svg.lucide-tag')).toBeTruthy()

    // 空文档状态包含 BookOpen 水墨图标
    const emptyState = screen.getByTestId('document-empty')
    expect(emptyState.querySelector('svg.lucide-book-open')).toBeTruthy()
  })

  it('遇到未知组件类型时应渲染 Fallback 诊断卡', () => {
    const source = JSON.stringify({
      version: '1.1',
      meta: { title: '测试文档' },
      content: [
        {
          id: 'block-unknown',
          kind: 'block',
          type: 'not-a-type',
          props: { message: 'hello' },
        },
      ],
    })

    render(<LpwDocumentViewer source={source} />)

    expect(screen.getByText('测试文档')).toBeTruthy()
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
    expect(screen.getByText(/未注册的 block 类型: not-a-type/)).toBeTruthy()
  })

  it('文档解析语法错误时应渲染文档级错误卡片', () => {
    render(<LpwDocumentViewer source="invalid-json" />)

    expect(screen.getByTestId('document-error')).toBeTruthy()
    expect(screen.getByText('LPW 文档解析失败')).toBeTruthy()
  })

  it('非法父子层级时应渲染诊断卡', () => {
    render(
      <LpwNodeRenderer
        node={{
          id: 'deep-b',
          kind: 'container',
          type: 'section',
          props: { variant: 'article', title: 't' },
          children: [],
        }}
        parentKind="container"
      />,
    )
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
    expect(screen.getByTestId('diagnostic-code').textContent).toBe(
      'INVALID_CONTAINER_CHILD',
    )
  })

  it('组件渲染抛错时错误边界应捕获并展示诊断信息', () => {
    const CrashingComponent = () => {
      throw new Error('组件内部发生未知故障')
    }
    lpwRegistry.register('block', 'crasher', {
      displayName: 'Crasher',
      Component: CrashingComponent as never,
      groups: ['text'],
      annotatableFields: [],
    })

    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    const source = JSON.stringify({
      version: '1.1',
      content: [{ id: 'b-crash', kind: 'block', type: 'crasher', props: {} }],
    })

    render(<LpwDocumentViewer source={source} />)

    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
    expect(screen.getByText('RENDER_EXCEPTION')).toBeTruthy()
    expect(screen.getByText(/crasher#b-crash/)).toBeTruthy()
    expect(screen.getByText('组件内部发生未知故障')).toBeTruthy()

    consoleSpy.mockRestore()
  })

  it('document shell renders at most one decorative bar', () => {
    const source = JSON.stringify({
      version: '1.1',
      meta: { title: '测试文档' },
      content: [],
    })
    const { container } = render(<LpwDocumentViewer source={source} />)
    const seaInkBorders = container.querySelectorAll(
      '.border-sea-ink, .border-sea-ink\\/80',
    )
    expect(seaInkBorders.length).toBe(0)
  })

  it('Q-07 文档壳不叠刊头与页脚装饰文案', () => {
    const source = JSON.stringify({
      version: '1.1',
      meta: { title: '测试文档', author: '筱锋' },
      content: [],
    })
    render(<LpwDocumentViewer source={source} />)
    expect(screen.getByText('测试文档')).toBeTruthy()
    expect(screen.getByText('筱锋')).toBeTruthy()
    expect(screen.queryByText(/LUMINA MONOGRAPH/)).toBeNull()
    expect(screen.queryByText(/KNOWLEDGE BLUEPRINT/)).toBeNull()
    expect(screen.queryByText(/ARCHITECTURE PRESS/)).toBeNull()
    expect(screen.queryByText(/IMPRIMATUR/)).toBeNull()
  })

  it('Q-04 批注栏在纸张外部，没有视口 fixed FAB 和右上角 label', async () => {
    const source = JSON.stringify({
      version: '1.1',
      meta: { title: '批注纸外' },
      content: [
        {
          id: 'md-1',
          kind: 'block',
          type: 'markdown',
          props: { content: '第一段 Redis 说明。' },
          annotation: {
            kind: 'issue',
            message: '复核缓存',
            targets: [{ field: 'content', pattern: 'Redis' }],
          },
        },
        {
          id: 'md-2',
          kind: 'block',
          type: 'markdown',
          props: { content: '第二段补充。' },
          annotation: { kind: 'note', message: '可再展开' },
        },
      ],
    })
    const { container } = render(<LpwDocumentViewer source={source} />)
    const gutter = await screen.findByTestId('annotation-gutter')
    expect(gutter).toBeTruthy()
    const paper = await screen.findByTestId('lpw-paper')
    expect(paper.querySelector('[data-testid="annotation-gutter"]')).toBeNull()
    expect(container.querySelector('.fixed')).toBeNull()
    expect(screen.queryByRole('button', { name: /批注：/ })).toBeNull()
    expect(document.getElementById('frame-md-1')).toBeTruthy()
  })

  it('验收测试：Plan 2 文本块样例全部真实渲染且不再是占位卡', () => {
    const sample = JSON.stringify({
      version: '1.1',
      meta: { title: 'plan-2 文本块验收' },
      content: [
        {
          id: 'h-1',
          kind: 'block',
          type: 'heading',
          props: { level: 2, content: '文本块验收' },
        },
        {
          id: 'md-1',
          kind: 'block',
          type: 'markdown',
          props: { content: '## 小标题\n\n正文 **加粗** 与 `行内代码`。' },
        },
        {
          id: 'co-1',
          kind: 'block',
          type: 'callout',
          props: {
            level: 'warning',
            title: '注意',
            content: '黄色警示应显示在左侧粗边。',
          },
        },
        {
          id: 'ls-1',
          kind: 'block',
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
          kind: 'block',
          type: 'quote',
          props: {
            content: '引述内容。',
            author: '设计者',
            source: 'design 0003',
          },
        },
        {
          id: 'cd-1',
          kind: 'block',
          type: 'code',
          props: {
            language: 'go',
            filename: 'main.go',
            content: 'line1\nline2\nline3',
            showLineNumbers: true,
            highlightLines: [2],
          },
        },
        { id: 'dv-1', kind: 'block', type: 'divider', props: {} },
        {
          id: 'ca-1',
          kind: 'block',
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
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()

    // 验证各块真实渲染
    expect(
      screen.getByRole('heading', { level: 2, name: '文本块验收' }),
    ).toBeTruthy()
    expect(screen.getByText('加粗')).toBeTruthy()
    expect(screen.getByText('黄色警示应显示在左侧粗边。')).toBeTruthy()
    expect(screen.getByText('已完成项')).toBeTruthy()
    expect(screen.getByText('引述内容。')).toBeTruthy()
    expect(screen.getByText('main.go')).toBeTruthy()
    expect(container.querySelector('[data-testid="divider-block"]')).toBeTruthy()
    expect(screen.getByRole('link', { name: /卡片一/ })).toBeTruthy()
    expect(screen.getByText('卡片二')).toBeTruthy()
  })

  it('验收测试：Plan 3 数据与图表块样例全部真实渲染且不再是占位卡', () => {
    const sample = JSON.stringify({
      version: '1.1',
      meta: { title: 'plan-3 数据图表块验收' },
      content: [
        {
          id: 'mt-1',
          kind: 'block',
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
          kind: 'block',
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
          kind: 'block',
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
          kind: 'block',
          type: 'diff',
          props: { language: 'go', oldCode: 'a := 1', newCode: 'a := 2' },
        },
        {
          id: 'tb-1',
          kind: 'block',
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
          kind: 'block',
          type: 'mermaid',
          props: { content: 'flowchart LR\n  A-->B' },
        },
        {
          id: 'ch-1',
          kind: 'block',
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
          kind: 'block',
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
          kind: 'block',
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
          kind: 'block',
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
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()

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
      version: '1.1',
      meta: { title: 'plan-4 复刻报表块验收' },
      content: [
        {
          id: 'tk-1',
          kind: 'block',
          type: 'takeaway',
          props: {
            content: 'LPW 已具备承载内部研究特刊的表达力，可以进入实施。',
          },
        },
        {
          id: 'gl-1',
          kind: 'block',
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
          kind: 'block',
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
          kind: 'block',
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
          kind: 'block',
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
          kind: 'block',
          type: 'personnel',
          props: {
            items: [{ name: '张三', role: '前端', duties: '组件实现' }],
          },
        },
      ],
    })

    render(<LpwDocumentViewer source={sample} />)

    // 不应有任何 Fallback 占位卡
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()

    expect(
      screen.getByText('LPW 已具备承载内部研究特刊的表达力，可以进入实施。'),
    ).toBeTruthy()
    expect(screen.getByText('引擎与 26 种内容块')).toBeTruthy()
    expect(screen.getByText('Schema 方言复核')).toBeTruthy()
    expect(screen.getByText('ECharts 按需')).toBeTruthy()
    expect(screen.getByText('自建分发')).toBeTruthy()
    expect(screen.getByText('张三')).toBeTruthy()
  })

  it('验收测试：容器与布局正常工作且无占位卡', () => {
    const sample = JSON.stringify({
      version: '1.1',
      meta: { title: '容器与布局验收' },
      content: [
        {
          id: 'lay-1',
          kind: 'layout',
          type: 'layout',
          props: { pattern: 'split' },
          children: [
            {
              id: 'sec-1',
              kind: 'container',
              type: 'section',
              props: { variant: 'article', title: '章节一', collapsible: true },
              children: [
                {
                  id: 'md-1',
                  kind: 'block',
                  type: 'markdown',
                  props: { content: '章节内正文。' },
                },
              ],
            },
            {
              id: 'det-1',
              kind: 'container',
              type: 'details',
              props: { variant: 'raw-data', summary: '附录' },
              children: [
                {
                  id: 'cd-1',
                  kind: 'block',
                  type: 'code',
                  props: { content: 'SELECT 1;' },
                },
              ],
            },
          ],
        },
      ],
    })

    render(<LpwDocumentViewer source={sample} />)

    // 不应有任何 Fallback 占位卡
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()

    expect(screen.getByText('章节一')).toBeTruthy()
    expect(screen.getByText('章节内正文。')).toBeTruthy()
    expect(screen.getByText('附录')).toBeTruthy()
  })
})
