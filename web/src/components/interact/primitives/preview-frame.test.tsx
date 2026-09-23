/** @vitest-environment jsdom */
import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import * as previewApi from '#/lib/apis/preview'
import { PreviewFrame, PreviewSupplement } from './preview-frame'

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
      expect(iframe.getAttribute('sandbox')).toBe('allow-scripts allow-popups')
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

describe('PreviewFrame', () => {
  it('初次加载完成后触发 onLoad 应处于 loaded 状态', async () => {
    const { container } = render(
      <PreviewFrame src="/preview/test/index.html" />,
    )
    const iframe = screen.getByTitle('前端预览')
    expect(iframe).toBeDefined()

    // 触发 onLoad
    iframe.dispatchEvent(new Event('load'))

    await waitFor(() => {
      // 初次加载完成之后，iframe 容器应存在
      expect(container.querySelector('iframe')).toBeTruthy()
    })
  })

  it('LUMINA-22: 当 src 更新时，不应弹出全屏硬核 LoadingSlot 覆盖层，且应启用 pointer-events-none 防止误触', async () => {
    const { rerender, container } = render(
      <PreviewFrame src="/preview/test/index.html?v=1" />,
    )
    const iframe = screen.getByTitle('前端预览')

    // 完成初次加载
    iframe.dispatchEvent(new Event('load'))

    await waitFor(() => {
      expect(screen.queryByText('正在加载中')).toBeNull()
    })

    // 更新 src（例如 WebSocket preview_sync 触发）
    rerender(<PreviewFrame src="/preview/test/index.html?v=2" />)

    // 严禁重新弹出硬核全屏 LoadingSlot ("正在加载中")
    expect(screen.queryByText('正在加载中')).toBeNull()

    // 在过渡/刷新期间，iframe 视口应被施加 pointer-events-none (S-02)
    const motionWrapper = container.querySelector(
      '[data-testid="preview-frame-viewport"]',
    )
    expect(motionWrapper).toBeTruthy()
    expect(motionWrapper?.className).toContain('pointer-events-none')

    // 细进度条应显示
    expect(screen.getByTestId('preview-sync-bar')).toBeDefined()

    // 新 iframe 加载完成
    iframe.dispatchEvent(new Event('load'))

    await waitFor(() => {
      expect(motionWrapper?.className).not.toContain('pointer-events-none')
      expect(screen.queryByTestId('preview-sync-bar')).toBeNull()
    })
  })

  it('Q-04: 当 iframe onLoad 未触发超时（5s）时，应自动恢复交互状态避免界面卡死', async () => {
    vi.useFakeTimers()
    const { rerender, container } = render(
      <PreviewFrame src="/preview/test/index.html?v=1" />,
    )
    const iframe = screen.getByTitle('前端预览')
    iframe.dispatchEvent(new Event('load'))

    // 更新 src
    rerender(<PreviewFrame src="/preview/test/index.html?v=2" />)
    const motionWrapper = container.querySelector(
      '[data-testid="preview-frame-viewport"]',
    )
    expect(motionWrapper?.className).toContain('pointer-events-none')

    // 快进 5000ms
    act(() => {
      vi.advanceTimersByTime(5000)
    })

    // 应自动解除 pointer-events-none
    expect(motionWrapper?.className).not.toContain('pointer-events-none')
    vi.useRealTimers()
  })

  it('Q-02 (subagent发现): 首次加载超时后，下一次 src 更新仍应被视为增量更新，不应重新弹出全屏 LoadingSlot', async () => {
    vi.useFakeTimers()
    const { rerender } = render(
      <PreviewFrame src="/preview/test/index.html?v=1" />,
    )

    // 首次不派发 load 事件，直接快进 5000ms 触发超时恢复
    act(() => {
      vi.advanceTimersByTime(5000)
    })

    // 超时后全屏 LoadingSlot 应已解除
    expect(screen.queryByText('正在加载中')).toBeNull()

    // 触发下一次增量更新
    rerender(<PreviewFrame src="/preview/test/index.html?v=2" />)

    // 不得再次退回全屏 LoadingSlot ("正在加载中")
    expect(screen.queryByText('正在加载中')).toBeNull()
    vi.useRealTimers()
  })
})
