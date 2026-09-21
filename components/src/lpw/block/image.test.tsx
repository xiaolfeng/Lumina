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

  it('Q-25: 外部 props.src 变化后应重置 hasError 状态', () => {
    const { rerender } = render(
      <ImageBlock
        blockId="img-reset"
        props={{
          src: '/broken.png',
          alt: '重置测试',
        }}
        depth={1}
      />,
    )

    const img = screen.getByRole('img', { name: '重置测试' })
    fireEvent.error(img)

    expect(screen.getByTestId('image-fallback')).toBeTruthy()

    // 外部传入新 src
    rerender(
      <ImageBlock
        blockId="img-reset"
        props={{
          src: '/fixed.png',
          alt: '重置测试',
        }}
        depth={1}
      />,
    )

    // 应恢复渲染正常 img，不再是 fallback
    expect(screen.queryByTestId('image-fallback')).toBeNull()
    const newImg = screen.getByRole('img', { name: '重置测试' })
    expect(newImg).toBeTruthy()
    expect(newImg.getAttribute('src')).toBe('/fixed.png')
  })

  it('Q-07: Fallback 提示文字增加 break-all 防撑爆', () => {
    render(
      <ImageBlock
        blockId="img-break-all"
        props={{
          src: '/very/long/nested/path/that/might/overflow/the/container/without/break/all/image.png',
          alt: '溢出测试',
        }}
        depth={1}
      />,
    )

    const img = screen.getByRole('img', { name: '溢出测试' })
    fireEvent.error(img)

    const fallbackText = screen.getByText(/very\/long\/nested\/path/)
    expect(fallbackText.className).toContain('break-all')
  })

  it('Q-08: 外层 figure 与包含 img 的 div 添加 max-w-full min-w-0', () => {
    const { container } = render(
      <ImageBlock
        blockId="img-constraints"
        props={{
          src: '/valid.png',
          alt: '约束测试',
        }}
        depth={1}
      />,
    )

    const figure = container.querySelector('figure')
    expect(figure?.className).toContain('max-w-full')
    expect(figure?.className).toContain('min-w-0')

    const imgWrapper = figure?.querySelector('div')
    expect(imgWrapper?.className).toContain('max-w-full')
    expect(imgWrapper?.className).toContain('min-w-0')
  })
})
