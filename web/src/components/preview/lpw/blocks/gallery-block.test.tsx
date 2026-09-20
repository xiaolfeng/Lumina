/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { GalleryBlock } from './gallery-block'

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
})
