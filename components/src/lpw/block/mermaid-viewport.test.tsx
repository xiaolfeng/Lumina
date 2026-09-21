/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { MermaidViewport } from './mermaid-viewport'

afterEach(() => {
  cleanup()
})

describe('MermaidViewport', () => {
  it('adds viewBox when missing', () => {
    const { container } = render(
      <MermaidViewport>
        <svg width="100" height="50">
          <circle cx="50" cy="25" r="20" />
        </svg>
      </MermaidViewport>,
    )

    const svg = container.querySelector('[data-testid="mermaid-svg-container"] svg')
    expect(svg?.getAttribute('viewBox')).toBe('0 0 100 50')
  })

  it('fit mode sets full width', () => {
    const { container } = render(
      <MermaidViewport>
        <svg width="200" height="100">
          <rect width="200" height="100" />
        </svg>
      </MermaidViewport>,
    )

    const svg = container.querySelector<SVGElement>('[data-testid="mermaid-svg-container"] svg')
    expect(svg?.style.width).toBe('100%')
    expect(svg?.style.height).toBe('auto')
  })

  it('actual mode restores intrinsic size', () => {
    const { container } = render(
      <MermaidViewport>
        <svg width="200" height="100">
          <rect width="200" height="100" />
        </svg>
      </MermaidViewport>,
    )

    const actualBtn = screen.getByRole('button', { name: '原始比例' })
    fireEvent.click(actualBtn)

    const svg = container.querySelector<SVGElement>('[data-testid="mermaid-svg-container"] svg')
    expect(svg?.style.width).toBe('')
    expect(svg?.style.height).toBe('')
  })

  it('toolbar buttons have aria-labels', () => {
    render(
      <MermaidViewport>
        <svg width="100" height="50" />
      </MermaidViewport>,
    )

    expect(screen.getByRole('button', { name: '适应宽度' })).toBeTruthy()
    expect(screen.getByRole('button', { name: '原始比例' })).toBeTruthy()
    expect(screen.getByRole('button', { name: '全屏查看' })).toBeTruthy()
  })

  it('Q-22: 过滤百分比宽度 height，优先提取 client 尺寸或 BBox，避免将 "100%" 误解析为 100', () => {
    const { container } = render(
      <MermaidViewport>
        <svg width="100%" height="100%">
          <rect width="300" height="150" />
        </svg>
      </MermaidViewport>,
    )

    const svg = container.querySelector<SVGElement>('[data-testid="mermaid-svg-container"] svg')
    // 由于 width/height 是 "100%"，在 jsdom 中 clientWidth/clientHeight 默认为 0，且无 getBBox 返回正数，
    // 不应设置类似 '0 0 100 100' 的畸变 viewBox
    expect(svg?.getAttribute('viewBox')).not.toBe('0 0 100 100')
  })

  it('Q-07: 全屏打开时，原视口应展示占位卡，保证 DOM 中只有一份 children 防止 SVG ID 冲突', () => {
    render(
      <MermaidViewport>
        <svg id="unique-mermaid-id" width="200" height="100">
          <circle cx="10" cy="10" r="5" />
        </svg>
      </MermaidViewport>,
    )

    expect(document.querySelectorAll('#unique-mermaid-id').length).toBe(1)

    const fullscreenBtn = screen.getByRole('button', { name: '全屏查看' })
    fireEvent.click(fullscreenBtn)

    // 原视口应出现全屏占位提示
    expect(screen.getByTestId('mermaid-fullscreen-placeholder')).toBeTruthy()
    // 全局中依然严格只有一份 unique-mermaid-id
    expect(document.querySelectorAll('#unique-mermaid-id').length).toBe(1)
  })
})
