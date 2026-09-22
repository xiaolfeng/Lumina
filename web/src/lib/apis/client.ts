import axios from 'axios'
import Cookies from 'js-cookie'
import JSONBig from 'json-bigint'
import { writeTokenCookies } from '../auth/cookie-utils'
import {
  isSessionNeutralRequest,
  refreshTokenFromRequest,
  shouldDiscardSessionAfterRefreshFailure,
} from '../auth/session'
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
      result[key] = convertIdStringsToBigInt(
        (data as Record<string, unknown>)[key],
      )
    }
    return result
  }
  return data
}

function bigintTransformRequest(
  data: unknown,
  headers?: Record<string, string>,
): string {
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

function subscribeTokenRefresh(
  callback: (token: string, error?: Error) => void,
) {
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

function clearSessionAfterRefreshFailure(
  attemptedRefreshToken: string | undefined,
  currentPath?: string,
) {
  const current = Cookies.get('refresh_token')
  if (
    !shouldDiscardSessionAfterRefreshFailure(attemptedRefreshToken, current)
  ) {
    return
  }
  clearAuthAndRedirect(currentPath)
}

function withRefreshLock<T>(fn: () => Promise<T>): Promise<T> {
  const locks = typeof navigator !== 'undefined' ? navigator.locks : undefined
  if (!locks?.request) return fn()
  return locks.request('lumina-auth-refresh', fn)
}

function refreshToken(): Promise<string> {
  return withRefreshLock(async () => {
    const refreshTokenValue = Cookies.get('refresh_token')
    if (!refreshTokenValue) {
      throw new Error('No refresh token available')
    }

    const res = await apiClient.post('/api/v1/auth/refresh', {
      refresh_token: refreshTokenValue,
    })
    const tokenData = (
      res as { data?: Parameters<typeof writeTokenCookies>[0] }
    ).data
    if (!tokenData?.access_token) {
      throw new Error('Refresh response missing token data')
    }

    writeTokenCookies(tokenData)
    return tokenData.access_token
  })
}

/** 与 401 拦截器共用同一条刷新，避免页面定时器和请求失败各打一次。 */
export function refreshAccessToken(): Promise<void> {
  return doRefresh().then(() => undefined)
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

function readErrorMessage(cause: unknown, fallback: string): string {
  if (cause && typeof cause === 'object' && 'response' in cause) {
    const data = (
      cause as {
        response?: { data?: { error_message?: unknown; message?: unknown } }
      }
    ).response?.data
    const message = data?.error_message ?? data?.message
    if (typeof message === 'string' && message !== '') return message
  }
  if (cause instanceof Error && cause.message !== '') return cause.message
  return fallback
}

function handle401Error(originalRequest: any, cause?: unknown): Promise<any> {
  if (!originalRequest) {
    return Promise.reject(
      cause instanceof Error ? cause : new Error('Request failed'),
    )
  }
  if (isSessionNeutralRequest(originalRequest?.url)) {
    return Promise.reject(new Error(readErrorMessage(cause, 'Request failed')))
  }

  // 刷新请求自己失败：只丢掉这次提交的那把令牌，别把别的标签页刚换上的新令牌清掉
  if (originalRequest?.url === REFRESH_URL) {
    clearSessionAfterRefreshFailure(refreshTokenFromRequest(originalRequest))
    return Promise.reject(new Error('Refresh token expired'))
  }

  if (!Cookies.get('refresh_token')) {
    clearSessionAfterRefreshFailure(undefined)
    return Promise.reject(new Error('No refresh token available'))
  }

  if (originalRequest._retry) {
    clearSessionAfterRefreshFailure(Cookies.get('refresh_token'))
    return Promise.reject(new Error('Token refresh failed after retry'))
  }

  originalRequest._retry = true
  const attempted = Cookies.get('refresh_token')

  if (isRefreshing) {
    return new Promise((resolve, reject) => {
      subscribeTokenRefresh((token, err) => {
        if (err || !token) {
          clearSessionAfterRefreshFailure(attempted)
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
      clearSessionAfterRefreshFailure(attempted)
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
      return handle401Error(error.config, error)
    }
    const respData = error.response?.data
    if (
      respData &&
      typeof respData === 'object' &&
      'error_message' in respData
    ) {
      const msg = respData.error_message ?? respData.message
      return Promise.reject(
        new Error(
          msg ?? `Request failed with status code ${error.response?.status}`,
        ),
      )
    }
    return Promise.reject(error)
  },
)

// ── 面向访客公开端点的独立客户端 ──
//
// 供 Pages 展示态、密码门等无需控制台登录的公开接口消费：
// 1. withCredentials: true 发送/接收密码门 HttpOnly Cookie
// 2. 无 Bearer Token 自动注入，无 Token 刷新状态机
// 3. 401 错误仅作为普通业务错误上抛，绝不强制跳转 /auth/login
export const publicApiClient = axios.create({
  baseURL: '',
  timeout: 10000,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
  transformResponse: [bigintTransformResponse],
  transformRequest: [bigintTransformRequest],
})

publicApiClient.interceptors.response.use(
  (response) => {
    const data = response.data
    if (data && typeof data === 'object' && 'code' in data) {
      const baseData = data as BaseResponse
      if (baseData.code !== 200) {
        return Promise.reject(
          new Error(
            baseData.error_message ??
              baseData.message ??
              `Request failed with code ${baseData.code}`,
          ),
        )
      }
    }
    return response.data
  },
  (error) => {
    const respData = error.response?.data
    if (
      respData &&
      typeof respData === 'object' &&
      'error_message' in respData
    ) {
      const msg = respData.error_message ?? respData.message
      return Promise.reject(
        new Error(
          msg ?? `Request failed with status code ${error.response?.status}`,
        ),
      )
    }
    return Promise.reject(error)
  },
)
