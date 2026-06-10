import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type {
  SessionListResponse,
  SessionDetailResponse,
  QuestionDetailResponse,
  QaConfigResponse,
} from '../models/response/qa-admin'
import type {
  SessionListParams,
  UpdateQaConfigRequest,
} from '../models/request/qa-admin'

export function getSessionList(
  params?: SessionListParams,
): Promise<BaseResponse<SessionListResponse>> {
  return apiClient.get('/api/v1/qa/sessions', { params })
}

export function getSessionDetail(
  id: string,
): Promise<BaseResponse<SessionDetailResponse>> {
  return apiClient.get(`/api/v1/qa/sessions/${id}`)
}

export function deleteSession(id: string): Promise<BaseResponse> {
  return apiClient.delete(`/api/v1/qa/sessions/${id}`)
}

export function getQuestionDetail(
  sessionId: string,
  questionId: string,
): Promise<BaseResponse<QuestionDetailResponse>> {
  return apiClient.get(
    `/api/v1/qa/sessions/${sessionId}/questions/${questionId}`,
  )
}

export function getQaConfig(): Promise<BaseResponse<QaConfigResponse>> {
  return apiClient.get('/api/v1/qa/config')
}

export function updateQaConfig(
  data: UpdateQaConfigRequest,
): Promise<BaseResponse<QaConfigResponse>> {
  return apiClient.put('/api/v1/qa/config', data)
}
