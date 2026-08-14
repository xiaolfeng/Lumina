import { describe, expect, it, vi } from 'vitest'
import { render } from '@testing-library/react'

// 捕获 Markdown 收到的 className，用于黑底回归守卫
const markdownProps = vi.hoisted(() => ({ className: undefined as string | undefined }))

// Mock 共享 Markdown，验证是否传入 proseArticle 排版类
vi.mock('@lumina/components/markdown', () => ({
  Markdown: ({
    className,
    children,
  }: {
    className?: string
    children: string
  }) => {
    markdownProps.className = className
    return <div data-testid="mock-markdown">{children}</div>
  },
  proseArticle: 'prose prose-slate [&_pre]:bg-white [&_pre]:border-line [&_pre]:text-gray-800',
}))

import { MarkdownRenderer } from './markdown-renderer'

describe('MarkdownRenderer', () => {
  it('将 proseArticle 排版类传给 Markdown（黑底回归守卫）', () => {
    render(<MarkdownRenderer content="# Hello\n\n```js\nx\n```" />)

    // Markdown 组件必须收到 proseArticle —— 若改回裸 <Markdown> 或丢 className，
    // 代码块会退回 typography 默认深色 --tw-prose-pre-bg，重现黑底 bug
    expect(markdownProps.className).toBeDefined()
    expect(markdownProps.className).toContain('[&_pre]:bg-white')
    expect(markdownProps.className).toContain('prose')
  })
})
