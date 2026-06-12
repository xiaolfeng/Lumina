import axios from 'axios'
import Cookies from 'js-cookie'
import type { BaseResponse } from '../models/response/common'

export const apiClient = axios.create({
  baseURL: '',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 认证相关的错误码，需要清除凭据并跳转登录页
const AUTH_ERROR_CODES = new Set([40101, 40102, 40103, 40104, 40105])
// 仅因 Access Token 过期导致的错误码，可尝试用 Refresh Token 刷新
const AT_EXPIRED_CODES = new Set([40101])

const REFRESH_URL = '/api/v1/auth/refresh'

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

function clearAuthAndRedirect() {
  Cookies.remove('access_token', { path: '/' })
  Cookies.remove('refresh_token', { path: '/' })
  Cookies.remove('expires_at', { path: '/' })
  window.location.href = '/auth/login'
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

      const expiresDays = tokenData.expires_in / 86400
      Cookies.set('access_token', tokenData.access_token, {
        expires: expiresDays,
        path: '/',
        sameSite: 'Lax',
      })
      Cookies.set('refresh_token', tokenData.refresh_token, {
        expires: 30,
        path: '/',
        sameSite: 'Lax',
      })
      Cookies.set(
        'expires_at',
        String(Date.now() + tokenData.expires_in * 1000),
        {
          expires: expiresDays,
          path: '/',
          sameSite: 'Lax',
        },
      )

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
    return Promise.reject(error)
  },
)
