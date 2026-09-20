/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { PagesBrandHeader } from './brand-header'

afterEach(() => {
  cleanup()
})

describe('PagesBrandHeader', () => {
  it('renders brand logo and Pages label', () => {
    render(<PagesBrandHeader projectName="lumina" slug="docs" />)
    expect(screen.getByText('Lumina')).toBeTruthy()
    expect(screen.getByText('Pages')).toBeTruthy()
  })

  it('renders title when provided', () => {
    render(
      <PagesBrandHeader
        projectName="lumina"
        slug="docs"
        title="微明设计系统与组件规范"
      />,
    )
    expect(
      screen.getByRole('heading', { name: '微明设计系统与组件规范' }),
    ).toBeTruthy()
  })

  it('falls back to projectName / slug when title is null or empty', () => {
    render(<PagesBrandHeader projectName="lumina" slug="design-system" />)
    expect(
      screen.getByRole('heading', { name: 'lumina / design-system' }),
    ).toBeTruthy()
  })
})
