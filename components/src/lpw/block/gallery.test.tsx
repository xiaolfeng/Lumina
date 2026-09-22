/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { GalleryBlock } from './gallery'

afterEach(() => {
  cleanup()
})

describe('GalleryBlock', () => {
  it('应正确渲染多图画廊，单图错误仅该格变为占位卡', () => {
    render(
      <GalleryBlock
        blockId="gy-1"
        props={{
          images: [
            {
              src: '/img1.png',
              alt: '图一正常',
              caption: '正常图片',
            },
            {
              src: '/img2.png',
              alt: '图二异常',
            },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByRole('img', { name: '图一正常' })).toBeTruthy()
    expect(screen.getByText('正常图片')).toBeTruthy()

    const img2 = screen.getByRole('img', { name: '图二异常' })
    fireEvent.error(img2)

    // 图二变占位卡
    expect(screen.getByTestId('gallery-fallback-1')).toBeTruthy()
    expect(screen.getByText(/img2\.png/)).toBeTruthy()

    // 图一仍然正常存在
    expect(screen.getByRole('img', { name: '图一正常' })).toBeTruthy()
  })

  it('Q-10: 移动端网格与图片样式响应式布局', () => {
    render(
      <GalleryBlock
        blockId="gy-responsive"
        props={{
          images: [
            {
              src: '/img1.png',
              alt: '图一',
            },
          ],
        }}
        depth={1}
      />,
    )

    const grid = screen.getByTestId('gallery-block')
    expect(grid.className).toContain('repeat(auto-fit')
    expect(grid.className).toContain('min(100%,14rem)')
    expect(grid.className).toContain('gap-5')

    const img = screen.getByRole('img', { name: '图一' })
    expect(img.className).toContain('w-full')
    expect(img.className).toContain('max-h-72')
    expect(img.className).toContain('sm:h-40')
    expect(img.className).toContain('object-contain')
    expect(img.className).toContain('sm:object-cover')
    expect(img.className).toContain('bg-surface-muted/20')
  })

  it('Q-07: Fallback 提示文字包含 break-all', () => {
    render(
      <GalleryBlock
        blockId="gy-fallback-break"
        props={{
          images: [
            {
              src: '/very/long/nested/path/to/gallery/image/broken.png',
              alt: '长路径图',
            },
          ],
        }}
        depth={1}
      />,
    )

    const img = screen.getByRole('img', { name: '长路径图' })
    fireEvent.error(img)

    const fallbackText = screen.getByText(/broken\.png/)
    expect(fallbackText.className).toContain('break-all')
  })
})
