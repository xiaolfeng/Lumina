/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { PreviewBrandHeader } from './brand-header'

afterEach(() => {
  cleanup()
})

describe('PreviewBrandHeader', () => {
  it('does not show the redundant 前端可视化预览 label', () => {
    render(<PreviewBrandHeader title="购物车原型" />)
    expect(screen.queryByText('前端可视化预览')).toBeNull()
    expect(screen.getByRole('heading', { name: '购物车原型' })).toBeTruthy()
  })

  it('hides the title slot when empty', () => {
    render(<PreviewBrandHeader title={null} />)
    expect(screen.queryByRole('heading')).toBeNull()
  })

  it('renders 4-in-1 button group (桌面, 平板, 手机, 源码) when controls are provided', () => {
    const setDevice = vi.fn()
    const setSourceMode = vi.fn()
    const onPromoteClick = vi.fn()

    render(
      <PreviewBrandHeader
        title="购物车原型"
        controls={{
          device: 'desktop',
          setDevice,
          sourceMode: false,
          setSourceMode,
          onPromoteClick,
        }}
      />,
    )

    // 验证 4 个按钮都存在
    expect(screen.getByRole('button', { name: '桌面' })).toBeTruthy()
    expect(screen.getByRole('button', { name: '平板' })).toBeTruthy()
    expect(screen.getByRole('button', { name: '手机' })).toBeTruthy()
    expect(screen.getByRole('button', { name: '源码' })).toBeTruthy()

    // 点击手机
    fireEvent.click(screen.getByRole('button', { name: '手机' }))
    expect(setSourceMode).toHaveBeenCalledWith(false)
    expect(setDevice).toHaveBeenCalledWith('mobile')

    // 点击源码
    fireEvent.click(screen.getByRole('button', { name: '源码' }))
    expect(setSourceMode).toHaveBeenCalledWith(true)
    expect(setDevice).toHaveBeenCalledWith('desktop')
  })

  it('renders promote button at the right side next to title', () => {
    const onPromoteClick = vi.fn()
    render(
      <PreviewBrandHeader
        title="设计系统"
        controls={{
          device: 'desktop',
          setDevice: vi.fn(),
          sourceMode: false,
          setSourceMode: vi.fn(),
          onPromoteClick,
          sourcePageSlug: 'design-system',
        }}
      />,
    )

    const titleEl = screen.getByRole('heading', { name: '设计系统' })
    const promoteBtn = screen.getByRole('button', { name: '晋升为 Pages' })
    expect(titleEl).toBeTruthy()
    expect(promoteBtn).toBeTruthy()

    fireEvent.click(promoteBtn)
    expect(onPromoteClick).toHaveBeenCalled()
  })
})
