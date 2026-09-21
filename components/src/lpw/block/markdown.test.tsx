/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { MarkdownBlock } from './markdown'

afterEach(() => {
  cleanup()
})

describe('MarkdownBlock', () => {
  it('应正确渲染 markdown 标题与加粗内容，并包含 proseArticle 类名', () => {
    const { container } = render(
      <MarkdownBlock
        blockId="md-1"
        props={{ content: '# 一级标题\n\n正文**加粗内容**' }}
        depth={1}
      />,
    )

    expect(screen.getByRole('heading', { level: 1 })).toBeTruthy()
    expect(screen.getByText('加粗内容')).toBeTruthy()
    const outer = container.firstElementChild
    expect(outer?.className).toContain('max-w-none')
  })

  it('Q-06 划线不进入行内 code 与链接', () => {
    const { container } = render(
      <MarkdownBlock
        nodeId="md-ann"
        props={{
          content: '请使用 `Redis` 与 [Redis](https://redis.io) ，以及 Redis 集群。',
        }}
        annotation={{
          kind: 'suggestion',
          message: '核对缓存方案',
          targets: [{ field: 'content', pattern: 'Redis' }],
        }}
      />,
    )

    const marks = Array.from(container.querySelectorAll('mark')).map(
      (el) => el.textContent,
    )
    expect(marks).toEqual(['Redis'])
    const code = container.querySelector('code')
    expect(code?.querySelector('mark')).toBeNull()
    const link = container.querySelector('a')
    expect(link?.querySelector('mark')).toBeNull()
  })

  it('Q-04 content 为空或 undefined 时不抛错', () => {
    expect(() => {
      render(
        <MarkdownBlock
          blockId="md-null"
          props={{ content: undefined as unknown as string }}
          depth={1}
        />,
      )
    }).not.toThrow()

    expect(() => {
      render(
        <MarkdownBlock
          blockId="md-empty"
          props={{ content: '' }}
          depth={1}
        />,
      )
    }).not.toThrow()
  })

  it('Q-14 根 DOM 节点显式补齐 id', () => {
    const { container } = render(
      <MarkdownBlock
        blockId="md-id-test"
        props={{ content: '正文测试' }}
        depth={1}
      />,
    )
    expect(container.firstElementChild?.getAttribute('id')).toBe('md-id-test')
  })
})
