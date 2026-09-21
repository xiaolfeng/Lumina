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
})
