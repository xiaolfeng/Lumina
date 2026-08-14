import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type {
  CreatePreviewSessionRequest,
  PreviewSessionDetailResponse,
  PreviewSessionListResponse,
  PreviewSessionItem,
  PreviewFileDetailResponse,
} from '../models/response/preview'

export interface PreviewSessionListParams {
  project_id?: string
  page?: number
  size?: number
}

export function getPreviewSessionByHash(
  hash: string,
): Promise<BaseResponse<PreviewSessionDetailResponse>> {
  return apiClient.get(`/api/v1/preview/sessions/${hash}`)
}

export function getPreviewFileByID(
  fileId: string,
): Promise<BaseResponse<PreviewFileDetailResponse>> {
  return apiClient.get(`/api/v1/preview/files/${fileId}`)
}

export function getPreviewSessions(
  params?: PreviewSessionListParams,
): Promise<BaseResponse<PreviewSessionListResponse>> {
  return apiClient.get('/api/v1/preview/sessions', { params })
}

export function createPreviewSession(
  data: CreatePreviewSessionRequest,
): Promise<BaseResponse<PreviewSessionItem>> {
  return apiClient.post('/api/v1/preview/sessions', data)
}

export function deletePreviewSession(id: string): Promise<BaseResponse> {
  return apiClient.delete(`/api/v1/preview/sessions/${id}`)
}

export function deletePreviewFile(id: string): Promise<BaseResponse> {
  return apiClient.delete(`/api/v1/preview/files/${id}`)
}
