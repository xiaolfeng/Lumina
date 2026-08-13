import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type {
  PreviewSessionDetailResponse,
  PreviewFileDetailResponse,
} from '../models/response/preview'

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
