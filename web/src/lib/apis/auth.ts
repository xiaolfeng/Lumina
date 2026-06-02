import { apiClient } from './client'
import type {
  InitializeRequest,
  LoginRequest,
  RefreshRequest,
} from '../models/request/auth'
import type { TokenResponse, StatusResponse } from '../models/response/auth'
import type { BaseResponse } from '../models/response/common'

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
