/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest'
import { useState } from 'react'
import { WorkspaceIconPicker } from './workspace-icon-picker'

beforeAll(() => {
  class ResizeObserverStub {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
  vi.stubGlobal('ResizeObserver', ResizeObserverStub)
  Element.prototype.scrollIntoView = vi.fn()
})

afterEach(() => {
  cleanup()
})

function Harness({ initial = '' }: { initial?: string }) {
  const [value, setValue] = useState(initial)
  return <WorkspaceIconPicker value={value} onChange={setValue} />
}

describe('WorkspaceIconPicker', () => {
  it('previews a selected lucide icon without showing the identifier as the trigger text', () => {
    render(<Harness initial="home" />)
    expect(screen.getByRole('combobox', { name: '空间图标：家' })).toBeTruthy()
    expect(screen.queryByText('home')).toBeNull()
  })

  it('can search a short lucide name and keep a live preview', () => {
    render(<Harness />)
    fireEvent.click(screen.getByRole('combobox'))
    fireEvent.change(screen.getByPlaceholderText('搜索图标名称或输入表情'), {
      target: { value: 'map' },
    })
    expect(screen.getByRole('option', { name: /地图/ })).toBeTruthy()
  })
})
