import axios from 'axios'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { apiClient } from './client'
import { getPreviewSessionDetail, isPreviewSessionGoneError } from './preview'

vi.mock('./client', () => ({ apiClient: { get: vi.fn() } }))
afterEach(() => vi.clearAllMocks())

/** 构造带 HTTP 状态码的 axios 错误（响应体缺失时拦截器会原样透传） */
function axiosErrorWithStatus(status: number): unknown {
  return Object.assign(new axios.AxiosError('Request failed'), {
    response: { status },
  })
}

describe('getPreviewSessionDetail', () => {
  it('requests the session snapshot with the encoded hash', async () => {
    vi.mocked(apiClient.get).mockResolvedValueOnce({
      data: { session: null, files: [] },
    })
    await getPreviewSessionDetail('a b/#1')
    expect(apiClient.get).toHaveBeenCalledWith(
      '/api/v1/preview/sessions/a%20b%2F%231',
    )
  })
})

describe('isPreviewSessionGoneError', () => {
  it('treats axios 404/410 responses as session gone', () => {
    expect(isPreviewSessionGoneError(axiosErrorWithStatus(404))).toBe(true)
    expect(isPreviewSessionGoneError(axiosErrorWithStatus(410))).toBe(true)
  })

  it('keeps retrying on network errors and other statuses', () => {
    expect(isPreviewSessionGoneError(new axios.AxiosError('network'))).toBe(
      false,
    )
    expect(isPreviewSessionGoneError(axiosErrorWithStatus(500))).toBe(false)
    expect(isPreviewSessionGoneError(axiosErrorWithStatus(401))).toBe(false)
  })

  it('recognizes interceptor-rewritten business errors by known message', () => {
    expect(isPreviewSessionGoneError(new Error('预览会话不存在'))).toBe(true)
    expect(isPreviewSessionGoneError(new Error('查询预览会话失败'))).toBe(false)
  })

  it('ignores non-error values', () => {
    expect(isPreviewSessionGoneError(undefined)).toBe(false)
    expect(isPreviewSessionGoneError('预览会话不存在')).toBe(false)
  })
})
