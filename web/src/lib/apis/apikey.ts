import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type { PageResponse } from '../models/response/page'
import type {
  ApikeyItem,
  ApikeyCreateResponse,
  ApikeyDetailResponse,
  ApikeyResetResponse,
} from '../models/response/apikey'
import type {
  ApikeyCreateRequest,
  ApikeyUpdateRequest,
  ApikeyListParams,
} from '../models/request/apikey'

export function getApikeyList(
  params?: ApikeyListParams,
): Promise<BaseResponse<PageResponse<ApikeyItem>>> {
  return apiClient.get('/api/v1/apikey', { params })
}

export function getApikeyDetail(
  id: string,
): Promise<BaseResponse<ApikeyDetailResponse>> {
  return apiClient.get(`/api/v1/apikey/${id}`)
}

export function createApikey(
  data: ApikeyCreateRequest,
): Promise<BaseResponse<ApikeyCreateResponse>> {
  return apiClient.post('/api/v1/apikey', data)
}

export function updateApikey(
  id: string,
  data: ApikeyUpdateRequest,
): Promise<BaseResponse<ApikeyItem>> {
  return apiClient.put(`/api/v1/apikey/${id}`, data)
}

export function deleteApikey(id: string): Promise<BaseResponse> {
  return apiClient.delete(`/api/v1/apikey/${id}`)
}

export function resetApikey(
  id: string,
): Promise<BaseResponse<ApikeyResetResponse>> {
  return apiClient.post(`/api/v1/apikey/${id}/reset`)
}
