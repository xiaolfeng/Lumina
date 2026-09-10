/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { useState } from 'react'
import type { WorkspaceItem } from '#/lib/models/response/workspace'
import { EditWorkspaceDialog } from './edit-dialog'

const { mutate, isPending } = vi.hoisted(() => ({
  mutate: vi.fn(),
  isPending: { value: false },
}))

vi.mock('#/hooks/useWorkspace', () => ({
  useUpdateWorkspace: () => ({
    mutate,
    get isPending() {
      return isPending.value
    },
  }),
}))

afterEach(() => {
  cleanup()
  mutate.mockReset()
  isPending.value = false
})

const item: WorkspaceItem = {
  id: 'ws-work',
  name: '工作',
  slug: 'space-keep-me',
  description: '办公',
  icon: 'briefcase',
  is_default: false,
  created_at: '2026-01-02T00:00:00Z',
  updated_at: '2026-01-02T00:00:00Z',
}

function Harness() {
  const [open, setOpen] = useState(true)
  return (
    <>
      <button type="button" onClick={() => setOpen(true)}>
        打开
      </button>
      <EditWorkspaceDialog open={open} onOpenChange={setOpen} item={item} />
    </>
  )
}

describe('EditWorkspaceDialog', () => {
  it('hides slug and keeps the original slug on save', () => {
    render(<Harness />)

    expect(screen.queryByText('标识')).toBeNull()
    expect(screen.queryByDisplayValue('space-keep-me')).toBeNull()
    expect(screen.queryByText('space-keep-me')).toBeNull()

    fireEvent.change(screen.getByLabelText('名称 *'), {
      target: { value: '工作室' },
    })
    fireEvent.click(screen.getByRole('button', { name: '保存' }))

    expect(mutate).toHaveBeenCalledTimes(1)
    expect(mutate.mock.calls[0]?.[0]).toMatchObject({
      id: 'ws-work',
      data: {
        name: '工作室',
        slug: 'space-keep-me',
      },
    })
  })

  it('restores the original values after cancel and reopen', () => {
    render(<Harness />)
    fireEvent.change(screen.getByLabelText('名称 *'), {
      target: { value: '改过的名字' },
    })
    fireEvent.click(screen.getByRole('button', { name: '取消' }))
    fireEvent.click(screen.getByRole('button', { name: '打开' }))
    expect(screen.getByLabelText('名称 *')).toHaveProperty('value', '工作')
  })

  it('disables controls while saving', () => {
    isPending.value = true
    render(<Harness />)
    expect(screen.getByLabelText('名称 *')).toHaveProperty('disabled', true)
    expect(screen.getByRole('button', { name: '取消' })).toHaveProperty(
      'disabled',
      true,
    )
    expect(screen.getByRole('button', { name: '保存中...' })).toHaveProperty(
      'disabled',
      true,
    )
  })
})
