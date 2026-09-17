import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type {
  ForkPageResponse,
  PageAuthCheckResponse,
  PageItem,
  PageListResponse,
  PagePublicMetaResponse,
  PageVersionListResponse,
  PromoteSessionResponse,
} from '../models/response/pages'
import type { PromotePreviewSessionRequest } from './preview'

export interface PageListParams {
  project_id?: string
  workspace_id?: string
  page?: number
  size?: number
}

export function getPages(
  params?: PageListParams,
): Promise<BaseResponse<PageListResponse>> {
  return apiClient.get('/api/v1/pages', { params })
}

export function getPage(id: string): Promise<BaseResponse<PageItem>> {
  return apiClient.get(`/api/v1/pages/${id}`)
}

export function getPageVersions(
  id: string,
): Promise<BaseResponse<PageVersionListResponse>> {
  return apiClient.get(`/api/v1/pages/${id}/versions`)
}

export function getPageVersionsByProject(
  projectName: string,
  slug: string,
): Promise<BaseResponse<PageVersionListResponse>> {
  return apiClient.get(`/api/v1/pages/by-project/${projectName}/${slug}/versions`)
}

export function switchActiveVersion(
  id: string,
  versionId: string,
): Promise<BaseResponse<unknown>> {
  return apiClient.post(`/api/v1/pages/${id}/versions/${versionId}/switch-active`)
}

export function forkPage(
  id: string,
  versionId?: string,
): Promise<BaseResponse<ForkPageResponse>> {
  return apiClient.post(`/api/v1/pages/${id}/fork`, {
    version_id: versionId || undefined,
  })
}

export function archivePage(id: string): Promise<BaseResponse<PageItem>> {
  return apiClient.post(`/api/v1/pages/${id}/archive`)
}

export function updatePageAccessPolicy(
  id: string,
  data: { access_mode: 'public' | 'password'; password?: string },
): Promise<BaseResponse<PageItem>> {
  return apiClient.put(`/api/v1/pages/${id}/access-policy`, data)
}

export function checkPageAuth(
  projectName: string,
  slug: string,
): Promise<BaseResponse<PageAuthCheckResponse>> {
  return apiClient.get(
    `/api/v1/pages/by-project/${projectName}/${slug}/auth-check`,
  )
}

export function unlockPage(
  projectName: string,
  slug: string,
  password: string,
): Promise<BaseResponse> {
  return apiClient.post(
    `/api/v1/pages/by-project/${projectName}/${slug}/unlock`,
    {
      password,
    },
  )
}

export function getPageMeta(
  projectName: string,
  slug: string,
  version?: string,
): Promise<BaseResponse<PagePublicMetaResponse>> {
  return apiClient.get(`/api/v1/pages/by-project/${projectName}/${slug}/meta`, {
    params: version ? { v: version } : undefined,
  })
}

export function promotePreviewSession(
  id: string,
  data: PromotePreviewSessionRequest,
): Promise<BaseResponse<PromoteSessionResponse>> {
  return apiClient.post(`/api/v1/preview/sessions/${id}/promote`, data)
}
