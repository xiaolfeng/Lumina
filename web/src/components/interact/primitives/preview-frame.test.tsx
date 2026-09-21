/** @vitest-environment jsdom */
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import * as previewApi from '#/lib/apis/preview'
import { PreviewSupplement } from './preview-frame'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('PreviewSupplement', () => {
  it('当引用文件为 .html 时，应渲染 PreviewFrame 内的 iframe', async () => {
    vi.spyOn(previewApi, 'getPreviewFileByID').mockResolvedValueOnce({
      code: 200,
      message: 'OK',
      data: {
        id: 'file-html',
        session_id: 'sess-1',
        session_hash: 'hash-abc',
        filename: 'index.html',
        mime_type: 'text/html; charset=utf-8',
        size: 100,
        created_at: '2026-09-20T00:00:00Z',
        updated_at: '2026-09-20T00:00:00Z',
      },
    })

    const content = JSON.stringify({
      session_id: 'sess-1',
      file_id: 'file-html',
    })
    render(<PreviewSupplement content={content} />)

    await waitFor(() => {
      const iframe = screen.getByTitle('前端预览')
      expect(iframe).toBeTruthy()
      expect(iframe.tagName.toLowerCase()).toBe('iframe')
    })
  })

  it('当引用文件为 .lpw 时，应分流渲染 PreviewLpwInlineViewer 而非 iframe', async () => {
    vi.spyOn(previewApi, 'getPreviewFileByID').mockResolvedValueOnce({
      code: 200,
      message: 'OK',
      data: {
        id: 'file-lpw',
        session_id: 'sess-1',
        session_hash: 'hash-abc',
        filename: 'report.lpw',
        mime_type: 'application/vnd.lumina.preview+json; charset=utf-8',
        size: 100,
        created_at: '2026-09-20T00:00:00Z',
        updated_at: '2026-09-20T00:00:00Z',
      },
    })

    // fetch 模拟返回原始 JSON
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ version: '1.1', content: [] }), {
        status: 200,
      }),
    )

    const content = JSON.stringify({
      session_id: 'sess-1',
      file_id: 'file-lpw',
    })
    const { container } = render(<PreviewSupplement content={content} />)

    await waitFor(() => {
      // 绝不能出现 iframe
      expect(container.querySelector('iframe')).toBeNull()
      // 出现空文档或 LPW 渲染容器
      expect(screen.getByTestId('document-empty')).toBeTruthy()
    })
  })
})
