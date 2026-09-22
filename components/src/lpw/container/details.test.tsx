/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { lpwRegistry } from '../registry'
import { DetailsContainer } from './details'

afterEach(() => {
  cleanup()
})

describe('DetailsContainer', () => {
  it('renders as a bordered disclosure with a clear state label', () => {
    render(
      <DetailsContainer
        blockId="det-1"
        props={{ summary: '更多详细信息' }}
        depth={1}
      />,
    )

    const container = screen.getByTestId('details-container')
    expect(container.className).toContain('border')
    expect(container.className).toContain('border-line')
    expect(container.className).toContain('bg-surface/35')
    const badge = screen.getByText('查看内容')
    expect(badge).toBeTruthy()
    expect(badge.className).toContain('hidden')
    expect(badge.className).toContain('sm:inline')
    const summarySpan = screen.getByText('更多详细信息')
    expect(summarySpan).toBeTruthy()
    expect(summarySpan.className).toContain('min-w-0')
    expect(summarySpan.className).toContain('break-words')
    expect(summarySpan.className).toContain('flex-1')
  })

  it('button toggles open state and aria-expanded', () => {
    lpwRegistry.register('det-child', false, () => <div>折叠内详细内容</div>)

    render(
      <DetailsContainer
        blockId="det-toggle"
        props={{ summary: '点击展开', defaultOpen: false }}
        depth={1}
        childrenBlocks={[{ id: 'c-1', kind: 'block', type: 'det-child', props: {} }]}
      />,
    )

    const btn = screen.getByRole('button')
    expect(btn.getAttribute('aria-expanded')).toBe('false')
    expect(screen.queryByText('折叠内详细内容')).toBeNull()

    fireEvent.click(btn)
    expect(btn.getAttribute('aria-expanded')).toBe('true')
    expect(screen.getByText('折叠内详细内容')).toBeTruthy()

    fireEvent.click(btn)
    expect(btn.getAttribute('aria-expanded')).toBe('false')
    expect(screen.queryByText('折叠内详细内容')).toBeNull()
  })

  it('chevron rotates when open', () => {
    const { container } = render(
      <DetailsContainer
        blockId="det-chevron"
        props={{ summary: '旋转测试', defaultOpen: true }}
        depth={1}
      />,
    )

    const svg = container.querySelector('svg')
    expect(svg?.classList.contains('rotate-90')).toBe(true)
  })

  it('content region is aria-controlled by summary', () => {
    render(
      <DetailsContainer
        blockId="det-aria"
        props={{ summary: '可访问性验证', defaultOpen: true }}
        depth={1}
      />,
    )

    const btn = screen.getByRole('button')
    expect(btn.getAttribute('aria-controls')).toBe('det-aria-content')

    const region = screen.getByRole('region')
    expect(region.id).toBe('det-aria-content')
    expect(region.getAttribute('aria-labelledby')).toBe('det-aria-summary')
  })
})
