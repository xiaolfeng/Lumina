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
})
