/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { lpwRegistry } from '../registry'
import { rootLocation } from '../render-location'
import { PanelContainer } from './panel'

afterEach(() => {
  cleanup()
})

describe('PanelContainer', () => {
  const loc = rootLocation()

  it('renders summary variant with title and icon', () => {
    lpwRegistry.register('block', 'panel-text', {
      displayName: 'Text',
      Component: () => <div data-testid="test-block">总结内容</div>,
      groups: ['text'],
      annotatableFields: [],
    })

    const { container } = render(
      <PanelContainer
        nodeId="panel-summary"
        location={loc}
        props={{
          variant: 'summary',
          title: '概要分析',
          icon: 'chart',
        }}
        children={[{ id: 'b1', kind: 'block', type: 'panel-text', props: {} }]}
      />,
    )

    const panel = screen.getByTestId('panel-container')
    expect(panel.getAttribute('data-variant')).toBe('summary')
    expect(panel.className).toContain('border')
    expect(panel.className).toContain('border-lagoon/30')
    expect(screen.getByText('概要分析')).toBeTruthy()

    const svg = container.querySelector('svg')
    expect(svg).toBeTruthy()
    expect(svg?.classList.contains('text-lagoon')).toBe(true)
    expect(screen.getByText('总结内容')).toBeTruthy()
  })

  it('renders aside variant with appropriate styling', () => {
    render(
      <PanelContainer
        nodeId="panel-aside"
        location={loc}
        props={{
          variant: 'aside',
          title: '边栏说明',
        }}
        children={[]}
      />,
    )

    const panel = screen.getByTestId('panel-container')
    expect(panel.getAttribute('data-variant')).toBe('aside')
    expect(panel.className).toContain('border')
    expect(panel.className).toContain('bg-surface/35')
    expect(screen.getByText('边栏说明')).toBeTruthy()
  })

  it('renders dashboard variant with an intrinsic responsive grid and normalized spacing', () => {
    lpwRegistry.register('block', 'dash-item', {
      displayName: 'DashItem',
      Component: () => <div data-testid="dashboard-block">指标项</div>,
      groups: ['data'],
      annotatableFields: [],
    })

    const { container } = render(
      <PanelContainer
        nodeId="panel-dashboard"
        location={loc}
        props={{
          variant: 'dashboard',
        }}
        children={[{ id: 'd1', kind: 'block', type: 'dash-item', props: {} }]}
      />,
    )

    const panel = screen.getByTestId('panel-container')
    expect(panel.getAttribute('data-variant')).toBe('dashboard')

    const innerWrapper = container.querySelector(
      "div[class*='grid-cols-[repeat(auto-fit']",
    )
    expect(innerWrapper).toBeTruthy()
    expect(innerWrapper?.className).toContain('gap-5')
  })
})
