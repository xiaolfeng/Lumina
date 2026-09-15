import axios from 'axios'

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
  workspace_id?: string
  page?: number
  size?: number
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

export function getPreviewSessionDetail(
  hash: string,
): Promise<BaseResponse<PreviewSessionDetailResponse>> {
  return apiClient.get(`/api/v1/preview/sessions/${encodeURIComponent(hash)}`)
}

// 会话确认不存在的服务端消息（哈希查询与可用性校验均返回该文案，HTTP 404）
const SESSION_NOT_FOUND_MESSAGE = '预览会话不存在'

/**
 * 判断预览会话探活错误是否表示「会话确实不存在/已删除」
 *
 * apiClient 响应拦截器会把带 error_message 的错误体转写为普通 Error（HTTP
 * 状态码被剥离），因此除 axios 原生 404/410 外还需按已知消息识别。
 */
export function isPreviewSessionGoneError(error: unknown): boolean {
  if (axios.isAxiosError(error)) {
    return error.response?.status === 404 || error.response?.status === 410
  }
  return error instanceof Error && error.message === SESSION_NOT_FOUND_MESSAGE
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

export interface PromotePreviewSessionRequest {
  slug: string
  title: string
  description?: string
  version?: string
  changelog?: string
  set_as_active?: boolean
  confirm_conflict?: boolean
}
