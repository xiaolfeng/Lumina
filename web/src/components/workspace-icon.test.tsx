/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { WorkspaceIcon } from './workspace-icon'

afterEach(() => {
  cleanup()
})

describe('WorkspaceIcon', () => {
  it('does not render short lucide names as text', () => {
    render(<WorkspaceIcon name="home" />)
    expect(screen.queryByText('home')).toBeNull()
    expect(screen.getByRole('img', { name: '空间图标' })).toBeTruthy()
  })

  it('renders composite emoji as the visual icon', () => {
    render(<WorkspaceIcon name="👨‍💻" label="开发空间图标" />)
    expect(screen.getByRole('img', { name: '开发空间图标' }).textContent).toBe(
      '👨‍💻',
    )
  })

  it('does not leak unknown icon identifiers as page text', () => {
    render(<WorkspaceIcon name="not-an-icon" />)
    expect(screen.queryByText('not-an-icon')).toBeNull()
    expect(screen.getByRole('img', { name: '空间图标' })).toBeTruthy()
  })
})
