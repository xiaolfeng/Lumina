/** @vitest-environment jsdom */
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { SessionProgressBar } from './session-progress-bar'

describe('SessionProgressBar', () => {
  it('shows answered, remaining, and total with a progressbar', () => {
    render(
      <SessionProgressBar progress={{ total: 8, answered: 3, remaining: 4 }} />,
    )

    expect(screen.getByLabelText('问题进度：已答 3，未答 4，共 8')).toBeTruthy()
    expect(screen.getByText('已答')).toBeTruthy()
    expect(screen.getByText('未答')).toBeTruthy()
    expect(screen.getByText('总计')).toBeTruthy()
    expect(screen.getByText('3/8')).toBeTruthy()

    const bar = screen.getByRole('progressbar')
    expect(bar.getAttribute('aria-valuenow')).toBe('3')
    expect(bar.getAttribute('aria-valuemax')).toBe('8')
  })
})
