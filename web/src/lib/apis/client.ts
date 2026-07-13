import axios from 'axios'
import Cookies from 'js-cookie'
import JSONBig from 'json-bigint'
import { writeTokenCookies } from '../auth/cookie-utils'
import type { BaseResponse } from '../models/response/common'

const JSONBigString = JSONBig({ storeAsString: true })

function bigintTransformResponse(data: string): unknown {
  if (typeof data !== 'string') return data
  try {
    return JSONBigString.parse(data)
  } catch {
    return data
  }
}

function convertIdStringsToBigInt(data: unknown): unknown {
  if (typeof data === 'string' && /^\d{15,19}$/.test(data)) {
    try {
      return BigInt(data)
    } catch {
      return data
    }
  }
  if (Array.isArray(data)) return data.map(convertIdStringsToBigInt)
  if (data && typeof data === 'object') {
    const result: Record<string, unknown> = {}
    for (const key of Object.keys(data)) {
      result[key] = convertIdStringsToBigInt((data as Record<string, unknown>)[key])
    }
    return result
  }
  return data
}

function bigintTransformRequest(data: unknown, headers?: Record<string, string>): string {
  if (headers) {
    headers['Content-Type'] = 'application/json'
  }
  return JSONBigString.stringify(convertIdStringsToBigInt(data))
}

export const apiClient = axios.create({
  baseURL: '',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
  transformResponse: [bigintTransformResponse],
  transformRequest: [bigintTransformRequest],
})

// 认证相关的错误码，需要清除凭据并跳转登录页
const AUTH_ERROR_CODES = new Set([40101, 40102, 40103, 40104, 40105])
// 仅因 Access Token 过期导致的错误码，可尝试用 Refresh Token 刷新
const AT_EXPIRED_CODES = new Set([40101])

const REFRESH_URL = '/api/v1/auth/refresh'
const LOGIN_PATH = '/auth/login'
const AUTH_PATH_PREFIX = '/auth/'

// ── 安全重定向工具 ──

export function getSafeRedirect(
  redirect: unknown,
  fallback = '/console/dashboard',
  currentOrigin = typeof window !== 'undefined' ? window.location.origin : '',
): string {
  if (!redirect || typeof redirect !== 'string') return fallback

  const trimmed = redirect.trim()
  if (!trimmed) return fallback

  let target: string
  if (trimmed.startsWith('/') && !trimmed.startsWith('//')) {
    target = trimmed
  } else {
    try {
      const url = new URL(trimmed)
      if (!currentOrigin || url.origin !== currentOrigin) return fallback
      target = url.pathname + url.search + url.hash
    } catch {
      return fallback
    }
  }

  // 避免跳转回登录/鉴权页造成循环
  if (target === LOGIN_PATH || target.startsWith(AUTH_PATH_PREFIX)) {
    return fallback
  }

  return target
}

// ── 刷新状态机 ──
let isRefreshing = false
let refreshPromise: Promise<string> | null = null
let refreshSubscribers: Array<(token: string, error?: Error) => void> = []

function subscribeTokenRefresh(callback: (token: string, error?: Error) => void) {
  refreshSubscribers.push(callback)
}

function onTokenRefreshed(newToken: string, error?: Error) {
  refreshSubscribers.forEach((cb) => cb(newToken, error))
  refreshSubscribers = []
}

function clearAuthAndRedirect(currentPath?: string) {
  Cookies.remove('access_token', { path: '/' })
  Cookies.remove('refresh_token', { path: '/' })
  Cookies.remove('expires_at', { path: '/' })

  const raw =
    currentPath ??
    (typeof window !== 'undefined'
      ? window.location.pathname + window.location.search + window.location.hash
      : '')
  const redirect = getSafeRedirect(raw, '/console/dashboard')
  window.location.href = `${LOGIN_PATH}?redirect=${encodeURIComponent(redirect)}`
}

function refreshToken(): Promise<string> {
  const refreshTokenValue = Cookies.get('refresh_token')
  if (!refreshTokenValue) {
    return Promise.reject(new Error('No refresh token available'))
  }

  return apiClient
    .post('/api/v1/auth/refresh', { refresh_token: refreshTokenValue })
    .then((res: any) => {
      const tokenData = res.data
      if (!tokenData) {
        throw new Error('Refresh response missing token data')
      }

      writeTokenCookies(tokenData)

      return tokenData.access_token as string
    })
}

function doRefresh(): Promise<string> {
  if (!isRefreshing) {
    isRefreshing = true
    refreshPromise = refreshToken()
      .then((token) => {
        onTokenRefreshed(token)
        return token
      })
      .catch((err) => {
        onTokenRefreshed('', err)
        throw err
      })
      .finally(() => {
        isRefreshing = false
        refreshPromise = null
      })
  }
  return refreshPromise!
}

function handle401Error(originalRequest: any): Promise<any> {
  // 如果 refresh 请求本身 401，说明 RT 也失效，直接跳转避免死循环
  if (originalRequest.url === REFRESH_URL) {
    clearAuthAndRedirect()
    return Promise.reject(new Error('Refresh token expired'))
  }

  if (!Cookies.get('refresh_token')) {
    clearAuthAndRedirect()
    return Promise.reject(new Error('No refresh token available'))
  }

  if (originalRequest._retry) {
    clearAuthAndRedirect()
    return Promise.reject(new Error('Token refresh failed after retry'))
  }

  originalRequest._retry = true

  if (isRefreshing) {
    return new Promise((resolve, reject) => {
      subscribeTokenRefresh((token, err) => {
        if (err || !token) {
          clearAuthAndRedirect()
          reject(err || new Error('Token refresh failed'))
          return
        }
        originalRequest.headers.Authorization = `Bearer ${token}`
        resolve(apiClient(originalRequest))
      })
    })
  }

  return doRefresh()
    .then((token) => {
      originalRequest.headers.Authorization = `Bearer ${token}`
      return apiClient(originalRequest)
    })
    .catch((err) => {
      clearAuthAndRedirect()
      return Promise.reject(err)
    })
}

apiClient.interceptors.request.use((config) => {
  const token = Cookies.get('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => {
    const data = response.data
    if (data && typeof data === 'object' && 'code' in data) {
      const baseData = data as BaseResponse
      if (baseData.code !== 200) {
        // Access Token 过期：尝试用 Refresh Token 刷新并重试原请求
        if (AT_EXPIRED_CODES.has(baseData.code)) {
          return handle401Error(response.config)
        }
        // 其他认证错误：直接清除凭据并跳转
        if (AUTH_ERROR_CODES.has(baseData.code)) {
          clearAuthAndRedirect()
          return Promise.reject(
            new Error(baseData.error_message ?? baseData.message),
          )
        }
        return Promise.reject(
          new Error(baseData.error_message ?? baseData.message),
        )
      }
    }
    return response.data
  },
  (error) => {
    // HTTP 层面的 401 错误
    if (error.response?.status === 401) {
      return handle401Error(error.config)
    }
    const respData = error.response?.data
    if (respData && typeof respData === 'object' && 'error_message' in respData) {
      const msg = respData.error_message ?? respData.message
      return Promise.reject(new Error(msg ?? `Request failed with status code ${error.response?.status}`))
    }
    return Promise.reject(error)
  },
)
