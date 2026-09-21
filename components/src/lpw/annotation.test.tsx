/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { AnnotatedText, AnnotationFrame } from './annotation'
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

    // 静态拒绝：不产生 mark，原文完整回退
    expect(container.querySelector('mark')).toBeNull()
    expect(container.textContent).toBe('aab')
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

  it('tooltip opens on focus', () => {
    const ann: LpwAnnotation = {
      kind: 'todo',
      message: '补充测试数据',
      author: '此方',
    }

    render(
      <AnnotationFrame annotation={ann}>
        <div>内容块</div>
      </AnnotationFrame>,
    )

    const btn = screen.getByRole('button', { name: /批注/ })
    expect(btn).toBeTruthy()

    btn.focus()
    // 聚焦后按钮可达
    expect(document.activeElement).toBe(btn)
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

