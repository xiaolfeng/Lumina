import axios from 'axios'
import Cookies from 'js-cookie'
import type {
  InitializeRequest,
  LoginRequest,
  RefreshRequest,
} from '../models/request/auth'
import type { TokenResponse, StatusResponse } from '../models/response/auth'
import type { BaseResponse } from '../models/response/common'

const apiClient = axios.create({
  baseURL: '',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use((config) => {
  const token = Cookies.get('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use((response) => {
  const data = response.data
  if (data && typeof data === 'object' && 'code' in data) {
    const baseData = data as BaseResponse
    if (baseData.code !== 200) {
      throw new Error(baseData.message)
    }
  }
  return response.data
})

export function initialize(data: InitializeRequest): Promise<BaseResponse> {
  return apiClient.post('/api/v1/auth/initialize', data)
}

export function login(
  data: LoginRequest,
): Promise<BaseResponse<TokenResponse>> {
  return apiClient.post('/api/v1/auth/login', data)
}

export function logout(data: RefreshRequest): Promise<BaseResponse> {
  return apiClient.post('/api/v1/auth/logout', data)
}

export function refresh(
  data: RefreshRequest,
): Promise<BaseResponse<TokenResponse>> {
  return apiClient.post('/api/v1/auth/refresh', data)
}

export function getStatus(): Promise<BaseResponse<StatusResponse>> {
  return apiClient.get('/api/v1/auth/status')
}
