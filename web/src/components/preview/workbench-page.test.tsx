/** @vitest-environment jsdom */
import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import * as previewApi from '#/lib/apis/preview'
import { PreviewWorkbenchPage } from './workbench-page'

const mockNavigate = vi.fn()

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mockNavigate,
}))

vi.mock('#/hooks/usePreviewHeader', () => ({
  usePreviewHeader: () => ({
    setTitle: vi.fn(),
    setControls: vi.fn(),
  }),
}))

vi.mock('#/hooks/usePreviewWebSocket', () => ({
  usePreviewWebSocket: () => ({
    status: 'connected',
    disconnect: vi.fn(),
    reconnect: vi.fn(),
  }),
}))

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
  mockNavigate.mockReset()
})

describe('PreviewWorkbenchPage (S-01 & Q-01)', () => {
  const sessionData = {
    session: {
      id: 'sess-1',
      hash: 'hash-123',
      title: '测试预览',
      status: 'active' as const,
      source_page_id: null,
      source_page_slug: null,
      created_at: '2026-09-20T00:00:00Z',
      updated_at: '2026-09-20T00:00:00Z',
      expires_at: '2026-09-30T00:00:00Z',
    },
    files: [
      {
        id: 'f-1',
        session_id: 'sess-1',
        filename: 'index.html',
        mime_type: 'text/html; charset=utf-8',
        size: 100,
        created_at: '2026-09-20T00:00:00Z',
        updated_at: '2026-09-20T00:00:00Z',
      },
      {
        id: 'f-2',
        session_id: 'sess-1',
        filename: 'about.html',
        mime_type: 'text/html; charset=utf-8',
        size: 100,
        created_at: '2026-09-20T00:00:00Z',
        updated_at: '2026-09-20T00:00:00Z',
      },
    ],
  }

  beforeEach(() => {
    vi.spyOn(previewApi, 'getPreviewSessionDetail').mockResolvedValue({
      code: 200,
      message: 'OK',
      data: sessionData as any,
    })
  })

  it('S-01: 来自未授权外部窗口（伪造 event.source）的 lumina:navigate 事件应被严格忽略', async () => {
    render(
      <PreviewWorkbenchPage
        sessionHash="hash-123"
        requestedFile="index.html"
      />,
    )

    await waitFor(() => {
      expect(screen.getByTitle('index.html')).toBeDefined()
    })

    // 模拟恶意外部弹窗/窗口发送 postMessage
    const fakeExternalWindow = {} as Window
    act(() => {
      const messageEvent = new MessageEvent('message', {
        data: { type: 'lumina:navigate', href: 'about.html' },
        source: fakeExternalWindow,
        origin: 'http://evil.com',
      })
      window.dispatchEvent(messageEvent)
    })

    // 外部伪造来源必须被拦截，不得触发路由跳转
    expect(mockNavigate).not.toHaveBeenCalledWith(
      expect.objectContaining({
        params: expect.objectContaining({ _splat: 'about.html' }),
      }),
    )
  })

  it('S-01 & Q-01: 来自合法 iframe contentWindow 的 lumina:navigate 应受信任并联动', async () => {
    const { container } = render(
      <PreviewWorkbenchPage
        sessionHash="hash-123"
        requestedFile="index.html"
      />,
    )

    await waitFor(() => {
      expect(screen.getByTitle('index.html')).toBeDefined()
    })

    const iframe = container.querySelector('iframe')
    expect(iframe).toBeTruthy()

    // 模拟来自 iframe.contentWindow 的合法内部导航事件
    act(() => {
      const messageEvent = new MessageEvent('message', {
        data: { type: 'lumina:navigate', href: 'about.html' },
        source: iframe?.contentWindow,
        origin: 'null',
      })
      window.dispatchEvent(messageEvent)
    })

    // 应允许导航并同步路由状态
    expect(mockNavigate).toHaveBeenCalledWith({
      to: '/preview/$sessionHash/$',
      params: { sessionHash: 'hash-123', _splat: 'about.html' },
      replace: true,
    })
  })
})
