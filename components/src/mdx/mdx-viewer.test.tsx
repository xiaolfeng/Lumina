/** @vitest-environment jsdom */
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { PreviewMdxInlineViewer, PreviewMdxViewer } from './mdx-viewer'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('PreviewMdxViewer & PreviewMdxInlineViewer', () => {
  it('展示加载态并在成功后呈现 Frontmatter 头部与正文', async () => {
    const mockMdx = `---
title: 微明技术文档
description: 这是 MDX 独立预览测试
icon: BookOpen
tags: [test, preview]
---

# 正文标题

这是测试正文内容。`

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(mockMdx, { status: 200 }),
    )

    render(<PreviewMdxViewer src="/preview/test/doc.mdx" filename="doc.mdx" />)

    expect(screen.getByText('加载 MDX 文档中…')).toBeTruthy()

    await waitFor(() => {
      expect(screen.getByText('微明技术文档')).toBeTruthy()
      expect(screen.getByText('这是 MDX 独立预览测试')).toBeTruthy()
      expect(screen.getByText('test')).toBeTruthy()
      expect(screen.getByText('preview')).toBeTruthy()
      expect(screen.getByText('这是测试正文内容。')).toBeTruthy()
    })
  })

  it('网络请求失败时展示错误提示与重试按钮', async () => {
    vi.spyOn(globalThis, 'fetch').mockRejectedValueOnce(new Error('Network error'))

    render(<PreviewMdxViewer src="/preview/test/doc.mdx" filename="doc.mdx" />)

    await waitFor(() => {
      expect(screen.getByText(/加载失败: Network error/)).toBeTruthy()
      expect(screen.getByText('重试')).toBeTruthy()
    })
  })

  it('PreviewMdxInlineViewer 应正常内联渲染', async () => {
    const mockMdx = `---
title: 内联 MDX 卡片
---

内联内容段落`

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(mockMdx, { status: 200 }),
    )

    render(<PreviewMdxInlineViewer src="/preview/test/doc.mdx" filename="doc.mdx" />)

    await waitFor(() => {
      expect(screen.getByText('内联 MDX 卡片')).toBeTruthy()
      expect(screen.getByText('内联内容段落')).toBeTruthy()
    })
  })
})
