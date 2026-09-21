/** @vitest-environment jsdom */
import { cleanup, render } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { DividerBlock } from './divider'

afterEach(() => {
  cleanup()
})

describe('DividerBlock', () => {
  it('renders gradient hairline without hr', () => {
    const { container } = render(
      <DividerBlock blockId="dv-1" props={{}} depth={1} />,
    )

    const hr = container.querySelector('hr')
    expect(hr).toBeNull()

    const el = container.querySelector('[data-testid="divider-block"]')
    expect(el).toBeTruthy()
    expect(el?.className).toContain('bg-gradient-to-r')
    expect(el?.className).toContain('h-px')
    expect(el?.className).toContain('my-6')
    expect(el?.className).toContain('sm:my-10')
    expect(el?.getAttribute('role')).toBe('separator')
  })
})
