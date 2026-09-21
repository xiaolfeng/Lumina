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
})
