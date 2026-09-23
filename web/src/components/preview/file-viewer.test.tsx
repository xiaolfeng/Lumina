/** @vitest-environment jsdom */
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { PreviewFileViewer } from './file-viewer'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('PreviewFileViewer', () => {
  it('当 kind 为 mdx 时正确渲染 PreviewMdxViewer', async () => {
    const mockMdx = `---
title: 测试 MDX 页面
description: MDX 预览集成测试
icon: ShieldCheck
---

# 一级标题
正文内容`

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(mockMdx, { status: 200 }),
    )

    render(
      <PreviewFileViewer
        kind="mdx"
        src="/preview/test/overview.mdx"
        filename="overview.mdx"
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('测试 MDX 页面')).toBeDefined()
      expect(screen.getByText('MDX 预览集成测试')).toBeDefined()
      expect(screen.getByText('正文内容')).toBeDefined()
    })
  })

  it('Q-03: 当同文件 src 增量更新时，PreviewSourceViewer 应保留现有内容，不应闪现“整理预览中…”占位', async () => {
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response('console.log("version 1")', { status: 200 }),
      )
      .mockImplementationOnce(() => new Promise(() => {})) // 模拟第二次请求 pending

    const { rerender } = render(
      <PreviewFileViewer
        kind="code"
        src="/preview/test/app.js?_lumina_sync=1"
        filename="app.js"
      />,
    )

    await waitFor(() => {
      expect(screen.getByText(/version 1/)).toBeDefined()
    })

    // 触发增量更新（相同 filename，不同 src）
    rerender(
      <PreviewFileViewer
        kind="code"
        src="/preview/test/app.js?_lumina_sync=2"
        filename="app.js"
      />,
    )

    // 不应退回到“整理预览中…”占位符
    expect(screen.queryByText('整理预览中…')).toBeNull()
  })
})
