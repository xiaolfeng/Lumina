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

// 认证相关的错误码（401xx），需要自动清除凭据并跳转登录页
const AUTH_ERROR_CODES = new Set([40101, 40102, 40103, 40104, 40105])

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
        // 认证相关错误：清除凭据并跳转登录页
        if (AUTH_ERROR_CODES.has(baseData.code)) {
          Cookies.remove('access_token', { path: '/' })
          Cookies.remove('refresh_token', { path: '/' })
          Cookies.remove('expires_at', { path: '/' })
          window.location.href = '/auth/login'
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
    // HTTP 层面的 401 错误（中间件直接拦截的情况）
    if (error.response?.status === 401) {
      Cookies.remove('access_token', { path: '/' })
      Cookies.remove('refresh_token', { path: '/' })
      Cookies.remove('expires_at', { path: '/' })
      window.location.href = '/auth/login'
    }
    return Promise.reject(error)
  },
)
