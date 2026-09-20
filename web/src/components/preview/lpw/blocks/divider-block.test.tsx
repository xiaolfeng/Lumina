/** @vitest-environment jsdom */
import { cleanup, render } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { DividerBlock } from './divider-block'

afterEach(() => {
  cleanup()
})

describe('DividerBlock', () => {
  it('应渲染 hr 分割线元素', () => {
    const { container } = render(
      <DividerBlock blockId="dv-1" props={{}} depth={1} />,
    )

    const hr = container.querySelector('hr')
    expect(hr).toBeTruthy()
    expect(hr?.className).toContain('border-line')
  })
})
