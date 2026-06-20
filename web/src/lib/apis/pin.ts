import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type { PinItem, PinListResponse } from '../models/response/pin'
import type {
  CreatePinRequest,
  UpdatePinRequest,
  PinListParams,
} from '../models/request/pin'

export function getPinList(
  params?: PinListParams,
): Promise<BaseResponse<PinListResponse>> {
  return apiClient.get('/api/v1/pin', { params })
}

export function getPinDetail(id: string): Promise<BaseResponse<PinItem>> {
  return apiClient.get(`/api/v1/pin/${id}`)
}

export function createPin(
  data: CreatePinRequest,
): Promise<BaseResponse<PinItem>> {
  return apiClient.post('/api/v1/pin', data)
}

export function updatePin(
  id: string,
  data: UpdatePinRequest,
): Promise<BaseResponse<PinItem>> {
  return apiClient.put(`/api/v1/pin/${id}`, data)
}

export function deletePin(id: string): Promise<BaseResponse> {
  return apiClient.delete(`/api/v1/pin/${id}`)
}
