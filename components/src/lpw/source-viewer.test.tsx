/** @vitest-environment jsdom */
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { lpwRegistry, registerAll } from './index'
import { LpwSourceViewer } from './source-viewer'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

beforeEach(() => {
  lpwRegistry.clear()
  registerAll()
})

describe('source-viewer', () => {
  it('computes assetBaseUrl from src', async () => {
    const docWithImage = JSON.stringify({
      version: '1.1',
      meta: { title: '相对图测试' },
      content: [
        {
          id: 'img-1',
          kind: 'block',
          type: 'image',
          props: { src: 'cover.png', alt: '封面' },
        },
      ],
    })

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(docWithImage, { status: 200 }),
    )

    const { container } = render(
      <LpwSourceViewer
        src="/preview/abc123/index.lpw"
        filename="index.lpw"
        variant="page"
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('相对图测试')).toBeTruthy()
    })

    const img = container.querySelector('img')
    expect(img?.getAttribute('src')).toBe('/preview/abc123/cover.png')
  })

  it('keeps absolute http src untouched', async () => {
    const docWithHttp = JSON.stringify({
      version: '1.1',
      meta: { title: '外链测试' },
      content: [
        {
          id: 'img-2',
          kind: 'block',
          type: 'image',
          props: { src: 'https://example.com/logo.png', alt: '外链' },
        },
      ],
    })

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(docWithHttp, { status: 200 }),
    )

    const { container } = render(
      <LpwSourceViewer
        src="/preview/abc123/index.lpw"
        filename="index.lpw"
        variant="page"
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('外链测试')).toBeTruthy()
    })

    const img = container.querySelector('img')
    expect(img?.getAttribute('src')).toBe('https://example.com/logo.png')
  })

  it('resolveAsset hook overrides base', async () => {
    const docWithImage = JSON.stringify({
      version: '1.1',
      meta: { title: '自定义解析' },
      content: [
        {
          id: 'img-3',
          kind: 'block',
          type: 'image',
          props: { src: 'pic.png', alt: '图' },
        },
      ],
    })

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(docWithImage, { status: 200 }),
    )

    const { container } = render(
      <LpwSourceViewer
        src="/preview/abc123/index.lpw"
        filename="index.lpw"
        variant="page"
        runtimeConfig={{
          resolveAsset: (src) => `https://custom-cdn.com/${src}`,
        }}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('自定义解析')).toBeTruthy()
    })

    const img = container.querySelector('img')
    expect(img?.getAttribute('src')).toBe('https://custom-cdn.com/pic.png')
  })

  it('fetch failure shows retry in inline variant', async () => {
    const docOk = JSON.stringify({
      version: '1.1',
      meta: { title: '重试成功' },
      content: [],
    })

    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response('', { status: 500, statusText: 'Server Error' }),
      )
      .mockResolvedValueOnce(new Response(docOk, { status: 200 }))

    render(
      <LpwSourceViewer
        src="/preview/abc123/index.lpw"
        filename="index.lpw"
        variant="inline"
      />,
    )

    await waitFor(() => {
      expect(screen.getByText(/读取 LPW 文件失败/)).toBeTruthy()
    })

    const retryBtn = screen.getByRole('button', { name: '重试' })
    fireEvent.click(retryBtn)

    await waitFor(() => {
      expect(screen.getByText('重试成功')).toBeTruthy()
    })
  })
})
