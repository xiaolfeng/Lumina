/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

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
})
