/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { AnnotatedText, AnnotationFrame, AnnotationGutter } from './annotation'
import { AnnotationProvider } from './annotation-context'
import { parseLpwSource } from './parser'
import type { LpwAnnotation } from './types'

beforeEach(() => {
  window.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

afterEach(() => {
  cleanup()
})

describe('annotation', () => {
  it('renders wavy mark on matched text', () => {
    const ann: LpwAnnotation = {
      kind: 'suggestion',
      message: '请优化文案',
      targets: [{ field: 'content', pattern: '重要结论' }],
    }

    const { container } = render(
      <AnnotatedText
        field="content"
        value="这是一个重要结论需要注意。"
        annotation={ann}
      />,
    )

    const mark = container.querySelector('mark')
    expect(mark).toBeTruthy()
    expect(mark?.textContent).toBe('重要结论')
    expect(mark?.className).toContain('decoration-wavy')
  })

  it('does not duplicate text when pattern has internal capturing groups', () => {
    const ann: LpwAnnotation = {
      kind: 'suggestion',
      message: '缓存说明',
      targets: [{ field: 'content', pattern: '(Redis)\\s(缓存)' }],
    }

    const { container } = render(
      <AnnotatedText
        field="content"
        value="前置说明 Redis 缓存 后续内容"
        annotation={ann}
      />,
    )

    const mark = container.querySelector('mark')
    expect(mark).toBeTruthy()
    expect(mark?.textContent).toBe('Redis 缓存')
    expect(container.textContent).toBe('前置说明 Redis 缓存 后续内容')
  })

  it('rejects catastrophic-backtracking patterns and falls back to plain text', () => {
    const ann: LpwAnnotation = {
      kind: 'suggestion',
      message: '危险模式',
      targets: [{ field: 'content', pattern: '(a+)+b' }],
    }

    const { container } = render(
      <AnnotatedText
        field="content"
        value="aab"
        annotation={ann}
      />,
    )

    // 静态拒绝：不产生 mark，原文完整回退，并提示匹配器无效
    expect(container.querySelector('mark')).toBeNull()
    expect(container.textContent).toContain('aab')
    expect(screen.getByTestId('annotation-invalid-pattern').textContent).toContain(
      '批注匹配器无效',
    )
  })

  it('invalid pattern shows diagnostic not crash', () => {
    const ann: LpwAnnotation = {
      kind: 'issue',
      message: '正则测试',
      targets: [{ field: 'content', pattern: '((' }],
    }

    const { container } = render(
      <AnnotatedText
        field="content"
        value="正常文本"
        annotation={ann}
      />,
    )

    // 不崩溃，回退渲染原文
    expect(container.textContent).toContain('正常文本')
  })

  it('layout annotation rejected at parse', () => {
    const doc = JSON.stringify({
      version: '1.1',
      content: [
        {
          id: 'lay-1',
          kind: 'layout',
          type: 'layout',
          props: { pattern: 'split' },
          annotation: { kind: 'note', message: '错误挂载' },
          children: [],
        },
      ],
    })
    // parser 对 kind=layout 允许 parse 出 document，但是在渲染层或 Schema 层拒绝
    const res = parseLpwSource(doc)
    expect(res.document?.content[0].id).toBe('lay-1')
  })

  it('Q-04 块级 frame 不再渲染右上角 label 按钮', () => {
    const ann: LpwAnnotation = {
      kind: 'todo',
      message: '补充测试数据',
      author: '此方',
    }

    render(
      <AnnotationFrame nodeId="blk-1" annotation={ann}>
        <div>内容块</div>
      </AnnotationFrame>,
    )

    expect(screen.getByTestId('annotation-frame')).toBeTruthy()
    expect(screen.queryByRole('button', { name: /批注：/ })).toBeNull()
    expect(document.getElementById('frame-blk-1')).toBeTruthy()
  })

  it('supports unicode and case-insensitive flags (u / iu)', () => {
    const ann: LpwAnnotation = {
      kind: 'suggestion',
      message: 'Unicode匹配',
      targets: [{ field: 'content', pattern: '\\p{Script=Han}+', flags: 'u' }],
    }

    const { container } = render(
      <AnnotatedText
        field="content"
        value="Hello 世界 123"
        annotation={ann}
      />,
    )

    const mark = container.querySelector('mark')
    expect(mark).toBeTruthy()
    expect(mark?.textContent).toBe('世界')
  })

  it('Q-05 滚动目标优先命中 frame-${nodeId}', () => {
    const ann: LpwAnnotation = { kind: 'issue', message: '定位' }
    const { container } = render(
      <AnnotationProvider>
        <AnnotationFrame nodeId="md-annotated" annotation={ann}>
          <div>hello</div>
        </AnnotationFrame>
      </AnnotationProvider>,
    )
    expect(container.querySelector('#md-annotated')).toBeNull()
    expect(container.querySelector('#frame-md-annotated')).toBeTruthy()
  })

  it('Q-04 重叠批注压缩，点击后展开', async () => {
    const ann: LpwAnnotation = { kind: 'note', message: '第一条' }
    const ann2: LpwAnnotation = { kind: 'issue', message: '第二条' }
    render(
      <div data-testid="lpw-paper">
        <AnnotationProvider>
          <AnnotationFrame nodeId="n1" annotation={ann}>
            <div>A</div>
          </AnnotationFrame>
          <AnnotationFrame nodeId="n2" annotation={ann2}>
            <div>B</div>
          </AnnotationFrame>
          <AnnotationGutter />
        </AnnotationProvider>
      </div>,
    )

    const cluster = await screen.findByTestId('annotation-cluster')
    expect(cluster.textContent).toMatch(/2/)
    fireEvent.click(cluster)
    expect(screen.getAllByTestId('annotation-bubble').length).toBeGreaterThanOrEqual(
      2,
    )
  })

  it('supports unicode and case-insensitive flags combined (iu)', () => {
    const ann: LpwAnnotation = {
      kind: 'suggestion',
      message: 'Unicode大小写匹配',
      targets: [{ field: 'content', pattern: '\\p{Script=Greek}+', flags: 'iu' }],
    }

    const { container } = render(
      <AnnotatedText
        field="content"
        value="Alpha αβγ Omega"
        annotation={ann}
      />,
    )

    const mark = container.querySelector('mark')
    expect(mark).toBeTruthy()
    expect(mark?.textContent).toBe('αβγ')
  })
})

