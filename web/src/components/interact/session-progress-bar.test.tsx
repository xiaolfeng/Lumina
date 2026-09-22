/** @vitest-environment jsdom */
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest'

import { SessionProgressBar } from './session-progress-bar'

beforeAll(() => {
  class ResizeObserverStub {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
  vi.stubGlobal('ResizeObserver', ResizeObserverStub)
})

afterEach(() => {
  cleanup()
})

describe('SessionProgressBar', () => {
  it('renders a compact answered/total readout with a multi-segment progressbar', () => {
    render(
      <SessionProgressBar
        progress={{
          total: 8,
          answered: 3,
          cancelled: 1,
          skipped: 2,
          remaining: 2,
        }}
      />,
    )

    expect(
      screen.getByLabelText('问题进度：已答 3，跳过 2，取消 1，未答 2，共 8'),
    ).toBeTruthy()
    expect(screen.getByText('已答')).toBeTruthy()
    expect(screen.getByText('3')).toBeTruthy()
    expect(screen.getByText('8')).toBeTruthy()
    // 紧凑读数不再直出取消/未答明细
    expect(screen.queryByText('取消')).toBeNull()
    expect(screen.queryByText('未答')).toBeNull()

    const bar = screen.getByRole('progressbar')
    expect(bar.getAttribute('aria-valuenow')).toBe('3')
    expect(bar.getAttribute('aria-valuemax')).toBe('8')
  })

  it('paints dedicated segments for answered, skipped, and cancelled', () => {
    const { container } = render(
      <SessionProgressBar
        progress={{
          total: 8,
          answered: 3,
          cancelled: 1,
          skipped: 2,
          remaining: 2,
        }}
      />,
    )

    expect(container.querySelector('.bg-lagoon')).not.toBeNull()
    expect(container.querySelector('.bg-amber-400')).not.toBeNull()
    expect(container.querySelector('.bg-rose-500')).not.toBeNull()
  })

  it('reveals the full breakdown card when the readout is focused', async () => {
    render(
      <SessionProgressBar
        progress={{
          total: 8,
          answered: 3,
          cancelled: 1,
          skipped: 2,
          remaining: 2,
        }}
      />,
    )

    fireEvent.focus(screen.getByRole('button', { name: /问题进度/ }))

    // Radix Tooltip 会同步渲染一份 visually-hidden 的 role="tooltip" 镜像，故用 getAllByText
    await waitFor(() => {
      expect(screen.getAllByText('已回答').length).toBeGreaterThanOrEqual(1)
    })
    expect(screen.getAllByText('未回答').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('已跳过').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('已取消').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('总计').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('2').length).toBeGreaterThanOrEqual(2)
  })
})
