// @vitest-environment jsdom

import { describe, it, expect, vi, beforeEach } from 'vitest'
import axios from 'axios'
import type * as AuthModule from './auth'

vi.mock('axios')

describe('auth API', () => {
  const mockGet = vi.fn()
  const mockPost = vi.fn()
  let capturedRequestHandler: ((config: any) => any) | undefined
  let capturedResponseHandler: ((response: any) => any) | undefined

  let authModule: typeof AuthModule

  beforeEach(async () => {
    vi.clearAllMocks()
    vi.resetModules()

    mockGet.mockReset()
    mockPost.mockReset()
    capturedRequestHandler = undefined
    capturedResponseHandler = undefined

    const mockClient = {
      get: async (...args: any[]) => {
        const response = await mockGet(...args)
        if (capturedResponseHandler) {
          return capturedResponseHandler(response)
        }
        return response
      },
      post: async (...args: any[]) => {
        const response = await mockPost(...args)
        if (capturedResponseHandler) {
          return capturedResponseHandler(response)
        }
        return response
      },
      interceptors: {
        request: {
          use: vi.fn((handler) => {
            capturedRequestHandler = handler
          }),
        },
        response: {
          use: vi.fn((handler) => {
            capturedResponseHandler = handler
          }),
        },
      },
    }

    vi.mocked(axios.create).mockReturnValue(mockClient as any)

    authModule = await import('./auth')
  })

  describe('getStatus', () => {
    it('returns status response when is_initial is true', async () => {
      mockGet.mockResolvedValueOnce({
        data: { code: 200, message: 'ok', data: { is_initial: true } },
      })

      const result = await authModule.getStatus()

      expect(mockGet).toHaveBeenCalledWith('/api/v1/auth/status')
      expect(result.data.is_initial).toBe(true)
    })

    it('returns status response when is_initial is false', async () => {
      mockGet.mockResolvedValueOnce({
        data: { code: 200, message: 'ok', data: { is_initial: false } },
      })

      const result = await authModule.getStatus()

      expect(mockGet).toHaveBeenCalledWith('/api/v1/auth/status')
      expect(result.data.is_initial).toBe(false)
    })
  })

  describe('login', () => {
    it('returns token response on success', async () => {
      const tokenData = {
        access_token: 'access_123',
        refresh_token: 'refresh_456',
        expires_in: 3600,
      }
      mockPost.mockResolvedValueOnce({
        data: { code: 200, message: 'ok', data: tokenData },
      })

      const result = await authModule.login({
        account: 'admin',
        password: 'secret',
      })

      expect(mockPost).toHaveBeenCalledWith('/api/v1/auth/login', {
        account: 'admin',
        password: 'secret',
      })
      expect(result.data.access_token).toBe('access_123')
      expect(result.data.refresh_token).toBe('refresh_456')
      expect(result.data.expires_in).toBe(3600)
    })

    it('throws error when response code is not 200', async () => {
      mockPost.mockResolvedValueOnce({
        data: { code: 401, message: 'Unauthorized', data: null },
      })

      await expect(
        authModule.login({ account: 'admin', password: 'wrong' }),
      ).rejects.toThrow('Unauthorized')
    })
  })

  describe('initialize', () => {
    it('returns success response', async () => {
      mockPost.mockResolvedValueOnce({
        data: { code: 200, message: 'ok', data: null },
      })

      const result = await authModule.initialize({
        username: 'admin',
        email: 'admin@example.com',
        password: 'secret',
      })

      expect(mockPost).toHaveBeenCalledWith('/api/v1/auth/initialize', {
        username: 'admin',
        email: 'admin@example.com',
        password: 'secret',
      })
      expect(result.code).toBe(200)
    })

    it('throws error on conflict (e.g., already initialized)', async () => {
      mockPost.mockResolvedValueOnce({
        data: { code: 409, message: 'Already initialized', data: null },
      })

      await expect(
        authModule.initialize({
          username: 'admin',
          email: 'admin@example.com',
          password: 'secret',
        }),
      ).rejects.toThrow('Already initialized')
    })
  })

  describe('logout', () => {
    it('returns success response', async () => {
      mockPost.mockResolvedValueOnce({
        data: { code: 200, message: 'ok', data: null },
      })

      const result = await authModule.logout({
        refresh_token: 'refresh_456',
      })

      expect(mockPost).toHaveBeenCalledWith('/api/v1/auth/logout', {
        refresh_token: 'refresh_456',
      })
      expect(result.code).toBe(200)
    })
  })

  describe('refresh', () => {
    it('returns new token response', async () => {
      const tokenData = {
        access_token: 'new_access_789',
        refresh_token: 'new_refresh_012',
        expires_in: 7200,
      }
      mockPost.mockResolvedValueOnce({
        data: { code: 200, message: 'ok', data: tokenData },
      })

      const result = await authModule.refresh({
        refresh_token: 'refresh_456',
      })

      expect(mockPost).toHaveBeenCalledWith('/api/v1/auth/refresh', {
        refresh_token: 'refresh_456',
      })
      expect(result.data.access_token).toBe('new_access_789')
      expect(result.data.expires_in).toBe(7200)
    })
  })

  describe('response interceptor', () => {
    it('throws error when code !== 200', () => {
      expect(capturedResponseHandler).toBeDefined()

      const response = {
        data: { code: 500, message: 'Internal Server Error' },
      }

      expect(() => capturedResponseHandler!(response)).toThrow(
        'Internal Server Error',
      )
    })

    it('returns data when code === 200', () => {
      expect(capturedResponseHandler).toBeDefined()

      const response = {
        data: { code: 200, message: 'ok', data: { foo: 'bar' } },
      }

      const result = capturedResponseHandler!(response)
      expect(result).toEqual({ code: 200, message: 'ok', data: { foo: 'bar' } })
    })

    it('passes through non-object data without code check', () => {
      expect(capturedResponseHandler).toBeDefined()

      const response = { data: 'plain text' }

      const result = capturedResponseHandler!(response)
      expect(result).toBe('plain text')
    })
  })

  describe('request interceptor', () => {
    it('injects Bearer header when access_token cookie exists', () => {
      expect(capturedRequestHandler).toBeDefined()

      Object.defineProperty(document, 'cookie', {
        value: 'access_token=my_token_123; other=value',
        writable: true,
      })

      const config = { headers: {} }
      const result = capturedRequestHandler!(config)

      expect(result.headers.Authorization).toBe('Bearer my_token_123')
    })

    it('does not inject header when access_token cookie is absent', () => {
      expect(capturedRequestHandler).toBeDefined()

      Object.defineProperty(document, 'cookie', {
        value: 'other=value',
        writable: true,
      })

      const config = { headers: {} }
      const result = capturedRequestHandler!(config)

      expect(result.headers.Authorization).toBeUndefined()
    })
  })
})
