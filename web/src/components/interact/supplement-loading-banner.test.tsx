/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import {
  SupplementLoadingBanner,
  SupplementWaitOverlay,
} from './supplement-loading-banner'

afterEach(() => {
  cleanup()
})

describe('SupplementLoadingBanner', () => {
  it('shows the waiting hint', () => {
    render(<SupplementLoadingBanner onDismiss={() => {}} />)
    expect(screen.getByRole('status').textContent).toContain('正在加载补充内容')
  })
})

describe('SupplementWaitOverlay', () => {
  it('renders children without a mask when idle', () => {
    render(
      <SupplementWaitOverlay loading={false}>
        <button type="button">选项甲</button>
      </SupplementWaitOverlay>,
    )
    expect(screen.queryByRole('status')).toBeNull()
    expect(screen.getByRole('button', { name: '选项甲' })).toBeTruthy()
    expect(screen.queryByText('正在加载补充内容…')).toBeNull()
  })

  it('shows the banner and greyscales children while waiting', () => {
    const { container } = render(
      <SupplementWaitOverlay loading onDismiss={() => {}}>
        <button type="button">选项乙</button>
      </SupplementWaitOverlay>,
    )
    expect(screen.getByRole('status').textContent).toContain('正在加载补充内容')
    const busy = container.querySelector('[aria-busy="true"]')
    expect(busy).toBeTruthy()
    expect(busy?.hasAttribute('inert')).toBe(true)
    expect(container.querySelector('.grayscale')).toBeTruthy()
  })
})
