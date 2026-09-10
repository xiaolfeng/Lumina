/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { useState } from 'react'
import { CreateWorkspaceDialog } from './create-dialog'

const { mutate, isPending } = vi.hoisted(() => ({
  mutate: vi.fn(),
  isPending: { value: false },
}))

vi.mock('#/hooks/useWorkspace', () => ({
  useCreateWorkspace: () => ({
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

function Harness({ startOpen = true }: { startOpen?: boolean }) {
  const [open, setOpen] = useState(startOpen)
  return (
    <>
      <button type="button" onClick={() => setOpen(true)}>
        打开
      </button>
      <CreateWorkspaceDialog open={open} onOpenChange={setOpen} />
    </>
  )
}

describe('CreateWorkspaceDialog', () => {
  it('hides slug fields and submits a generated hidden slug', () => {
    render(<Harness />)

    expect(screen.queryByText('标识')).toBeNull()
    expect(screen.queryByLabelText('标识 *')).toBeNull()
    expect(screen.queryByPlaceholderText('例如：work')).toBeNull()

    fireEvent.change(screen.getByLabelText('名称 *'), {
      target: { value: '工作' },
    })
    fireEvent.click(screen.getByRole('button', { name: '创建' }))

    expect(mutate).toHaveBeenCalledTimes(1)
    const payload = mutate.mock.calls[0]?.[0] as {
      name: string
      slug: string
    }
    expect(payload.name).toBe('工作')
    expect(payload.slug).toMatch(
      /^space-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/,
    )
  })

  it('resets values after cancel and reopen', () => {
    render(<Harness />)
    fireEvent.change(screen.getByLabelText('名称 *'), {
      target: { value: '临时名称' },
    })
    fireEvent.click(screen.getByRole('button', { name: '取消' }))
    fireEvent.click(screen.getByRole('button', { name: '打开' }))
    expect(screen.getByLabelText('名称 *')).toHaveProperty('value', '')
  })

  it('disables controls while creating', () => {
    isPending.value = true
    render(<Harness />)
    expect(screen.getByLabelText('名称 *')).toHaveProperty('disabled', true)
    expect(screen.getByRole('button', { name: '取消' })).toHaveProperty(
      'disabled',
      true,
    )
    expect(screen.getByRole('button', { name: '创建中...' })).toHaveProperty(
      'disabled',
      true,
    )
  })
})
