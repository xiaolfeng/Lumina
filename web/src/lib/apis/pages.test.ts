import { afterEach, describe, expect, it, vi } from 'vitest'

import { apiClient, publicApiClient } from './client'
import {
  archivePage,
  checkPageAuth,
  forkPage,
  getPage,
  getPageMeta,
  getPages,
  getPageVersions,
  getPageVersionsByProject,
  promotePreviewSession,
  switchActiveVersion,
  unlockPage,
  updatePageAccessPolicy,
} from './pages'

vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
  },
  publicApiClient: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

afterEach(() => vi.clearAllMocks())

describe('Pages Public APIs (publicApiClient)', () => {
  it('checkPageAuth dispatches via publicApiClient', async () => {
    vi.mocked(publicApiClient.get).mockResolvedValueOnce({
      code: 200,
      data: { authenticated: false, password_required: true },
    })
    await checkPageAuth('my-project', 'landing')
    expect(publicApiClient.get).toHaveBeenCalledWith(
      '/api/v1/pages/by-project/my-project/landing/auth-check',
    )
  })

  it('unlockPage dispatches via publicApiClient with password payload', async () => {
    vi.mocked(publicApiClient.post).mockResolvedValueOnce({
      code: 200,
      message: '解锁成功',
    })
    await unlockPage('my-project', 'landing', 'secret123')
    expect(publicApiClient.post).toHaveBeenCalledWith(
      '/api/v1/pages/by-project/my-project/landing/unlock',
      { password: 'secret123' },
    )
  })

  it('getPageMeta dispatches via publicApiClient without version', async () => {
    vi.mocked(publicApiClient.get).mockResolvedValueOnce({
      code: 200,
      data: { page: {}, version: {}, files: [] },
    })
    await getPageMeta('my-project', 'landing')
    expect(publicApiClient.get).toHaveBeenCalledWith(
      '/api/v1/pages/by-project/my-project/landing/meta',
      { params: undefined },
    )
  })

  it('getPageMeta dispatches via publicApiClient with version query param', async () => {
    vi.mocked(publicApiClient.get).mockResolvedValueOnce({
      code: 200,
      data: { page: {}, version: {}, files: [] },
    })
    await getPageMeta('my-project', 'landing', 'v1.0.0')
    expect(publicApiClient.get).toHaveBeenCalledWith(
      '/api/v1/pages/by-project/my-project/landing/meta',
      { params: { v: 'v1.0.0' } },
    )
  })

  it('getPageVersionsByProject dispatches via publicApiClient', async () => {
    vi.mocked(publicApiClient.get).mockResolvedValueOnce({
      code: 200,
      data: { items: [] },
    })
    await getPageVersionsByProject('my-project', 'landing')
    expect(publicApiClient.get).toHaveBeenCalledWith(
      '/api/v1/pages/by-project/my-project/landing/versions',
    )
  })
})

describe('Pages Admin APIs (apiClient)', () => {
  it('getPages uses admin apiClient', async () => {
    vi.mocked(apiClient.get).mockResolvedValueOnce({
      code: 200,
      data: { items: [], total: 0 },
    })
    await getPages({ workspace_id: 'ws_1' })
    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/pages', {
      params: { workspace_id: 'ws_1' },
    })
  })

  it('getPage uses admin apiClient', async () => {
    vi.mocked(apiClient.get).mockResolvedValueOnce({ code: 200, data: {} })
    await getPage('page_1')
    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/pages/page_1')
  })

  it('getPageVersions uses admin apiClient', async () => {
    vi.mocked(apiClient.get).mockResolvedValueOnce({
      code: 200,
      data: { items: [] },
    })
    await getPageVersions('page_1')
    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/pages/page_1/versions')
  })

  it('switchActiveVersion uses admin apiClient', async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({ code: 200, data: {} })
    await switchActiveVersion('page_1', 'ver_2')
    expect(apiClient.post).toHaveBeenCalledWith(
      '/api/v1/pages/page_1/versions/ver_2/switch-active',
    )
  })

  it('forkPage uses admin apiClient', async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({ code: 200, data: {} })
    await forkPage('page_1', 'ver_1')
    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/pages/page_1/fork', {
      version_id: 'ver_1',
    })
  })

  it('archivePage uses admin apiClient', async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({ code: 200, data: {} })
    await archivePage('page_1')
    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/pages/page_1/archive')
  })

  it('updatePageAccessPolicy uses admin apiClient', async () => {
    vi.mocked(apiClient.put).mockResolvedValueOnce({ code: 200, data: {} })
    await updatePageAccessPolicy('page_1', {
      access_mode: 'password',
      password: 'pwd',
    })
    expect(apiClient.put).toHaveBeenCalledWith(
      '/api/v1/pages/page_1/access-policy',
      {
        access_mode: 'password',
        password: 'pwd',
      },
    )
  })

  it('promotePreviewSession uses admin apiClient', async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({ code: 200, data: {} })
    await promotePreviewSession('session_1', {
      slug: 'test',
      title: 'Title',
    })
    expect(apiClient.post).toHaveBeenCalledWith(
      '/api/v1/preview/sessions/session_1/promote',
      { slug: 'test', title: 'Title' },
    )
  })
})
