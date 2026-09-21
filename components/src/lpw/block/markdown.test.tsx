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
})
