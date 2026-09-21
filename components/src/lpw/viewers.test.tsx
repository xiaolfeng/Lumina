/** @vitest-environment jsdom */
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { PreviewLpwInlineViewer, PreviewLpwViewer } from './viewers'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('PreviewLpwViewer & PreviewLpwInlineViewer', () => {
  const sampleDoc = JSON.stringify({
    version: '1.1',
    meta: { title: '远程文档' },
    content: [],
  })

  it('PreviewLpwViewer 加载成功时应渲染文档', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(sampleDoc, { status: 200 }),
    )

    render(<PreviewLpwViewer src="/test.lpw" filename="test.lpw" />)

    expect(screen.getByText('加载中…')).toBeTruthy()

    await waitFor(() => {
      expect(screen.getByText('远程文档')).toBeTruthy()
    })
  })

  it('PreviewLpwViewer 请求失败时展示错误与重试按钮', async () => {
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response('', { status: 404, statusText: 'Not Found' }),
      )
      .mockResolvedValueOnce(new Response(sampleDoc, { status: 200 }))

    render(<PreviewLpwViewer src="/test.lpw" filename="test.lpw" />)

    await waitFor(() => {
      expect(screen.getByText(/读取 LPW 文件失败/)).toBeTruthy()
    })

    const retryBtn = screen.getByRole('button', { name: '重试' })
    fireEvent.click(retryBtn)

    await waitFor(() => {
      expect(screen.getByText('远程文档')).toBeTruthy()
    })

    expect(fetchMock).toHaveBeenCalledTimes(2)
  })

  it('PreviewLpwInlineViewer 加载成功时应渲染文档', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(sampleDoc, { status: 200 }),
    )

    render(<PreviewLpwInlineViewer src="/test.lpw" filename="test.lpw" />)

    await waitFor(() => {
      expect(screen.getByText('远程文档')).toBeTruthy()
    })
  })
})
