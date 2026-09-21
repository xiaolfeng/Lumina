/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ImageBlock } from './image'

afterEach(() => {
  cleanup()
})

describe('ImageBlock', () => {
  it('应正确渲染 img、alt、caption 与 width 样式', () => {
    render(
      <ImageBlock
        blockId="img-1"
        props={{
          src: '/preview.png',
          alt: '预览架构图',
          caption: '图 1. 架构流向',
          width: '300px',
        }}
        depth={1}
      />,
    )

    const img = screen.getByRole('img', { name: '预览架构图' })
    expect(img).toBeTruthy()
    expect(img.style.width).toBe('300px')
    expect(screen.getByText('图 1. 架构流向')).toBeTruthy()
  })

  it('图片加载失败触发 onError 时应切换为占位卡', () => {
    render(
      <ImageBlock
        blockId="img-err"
        props={{
          src: '/not-exist.png',
          alt: '丢失的图片',
        }}
        depth={1}
      />,
    )

    const img = screen.getByRole('img', { name: '丢失的图片' })
    fireEvent.error(img)

    expect(screen.getByTestId('image-fallback')).toBeTruthy()
    expect(screen.getByText('丢失的图片')).toBeTruthy()
    expect(screen.getByText(/not-exist\.png/)).toBeTruthy()
  })
})
