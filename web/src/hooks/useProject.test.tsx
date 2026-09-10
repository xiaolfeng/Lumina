/** @vitest-environment jsdom */
import { act, cleanup, renderHook } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { ReactNode } from 'react'
import { useUpdateProject } from './useProject'
import { updateProject } from '#/lib/apis/project'

vi.mock('#/lib/apis/project', () => ({ updateProject: vi.fn() }))
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))
afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

function setup() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const invalidate = vi.spyOn(client, 'invalidateQueries')
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  )
  return {
    ...renderHook(() => useUpdateProject(), { wrapper }),
    invalidate,
    client,
  }
}

describe('useUpdateProject', () => {
  it('invalidates both workspace lists, project details and dependent lists after migration', async () => {
    vi.mocked(updateProject).mockResolvedValue({ data: {} } as never)
    const { result, invalidate, client } = setup()
    const keys = [
      ['project', 'list', { workspace_id: 'source' }],
      ['project', 'list', { workspace_id: 'target' }],
      ['project', 'detail', 'project'],
      ['qa', 'sessions', { workspace_id: 'source' }],
      ['preview', 'sessions', { workspace_id: 'target' }],
      ['pin', 'list', { workspace_id: 'source' }],
    ]
    keys.forEach((key) => client.setQueryData(key, { fixture: true }))
    await act(async () => {
      await result.current.mutateAsync({
        id: 'project',
        data: { name: 'Test', workspace_id: 'target' },
      })
    })
    expect(invalidate).toHaveBeenCalledTimes(4)
    keys.forEach((key) =>
      expect(client.getQueryState(key)?.isInvalidated).toBe(true),
    )
  })

  it('leaves dependent lists alone when only editing project fields', async () => {
    vi.mocked(updateProject).mockResolvedValue({ data: {} } as never)
    const { result, invalidate } = setup()
    await act(async () => {
      await result.current.mutateAsync({
        id: 'project',
        data: { name: 'Renamed' },
      })
    })
    expect(invalidate.mock.calls).toEqual([[{ queryKey: ['project'] }]])
  })

  it('does not invalidate or move anything when the update fails', async () => {
    vi.mocked(updateProject).mockRejectedValue(new Error('空间不存在'))
    const { result, invalidate } = setup()
    await act(async () => {
      await expect(
        result.current.mutateAsync({
          id: 'project',
          data: { name: 'Test', workspace_id: 'missing' },
        }),
      ).rejects.toThrow('空间不存在')
    })
    expect(invalidate).not.toHaveBeenCalled()
  })
})
