import { afterEach, describe, expect, it, vi } from 'vitest'
import { apiClient } from './client'
import { getWorkspaceOptions } from './workspace'

vi.mock('./client', () => ({ apiClient: { get: vi.fn() } }))
afterEach(() => vi.clearAllMocks())

describe('getWorkspaceOptions', () => {
  it('includes spaces beyond the first page for switching and migration', async () => {
    const first = Array.from({ length: 100 }, (_, i) => ({ id: String(i) }))
    vi.mocked(apiClient.get)
      .mockResolvedValueOnce({ data: { items: first, total: 101 } })
      .mockResolvedValueOnce({ data: { items: [{ id: 'last' }], total: 101 } })
    const result = await getWorkspaceOptions()
    expect(result).toHaveLength(101)
    expect(result.at(-1)?.id).toBe('last')
    expect(apiClient.get).toHaveBeenNthCalledWith(2, '/api/v1/workspace', {
      params: { page: 2, size: 100 },
    })
  })

  it('returns an empty list without requesting additional pages', async () => {
    vi.mocked(apiClient.get).mockResolvedValueOnce({
      data: { items: [], total: 0 },
    })
    expect(await getWorkspaceOptions()).toEqual([])
    expect(apiClient.get).toHaveBeenCalledTimes(1)
  })

  it('propagates failures instead of returning incomplete migration targets', async () => {
    vi.mocked(apiClient.get)
      .mockResolvedValueOnce({ data: { items: [{ id: 'first' }], total: 2 } })
      .mockRejectedValueOnce(new Error('network error'))
    await expect(getWorkspaceOptions()).rejects.toThrow('network error')
  })
})
