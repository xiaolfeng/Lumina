import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type {
	RepoWikiConfigListParams,
	CreateRepoWikiConfigRequest,
	UpdateRepoWikiConfigRequest,
} from '../models/request/repowiki'
import type {
	RepoWikiConfigItem,
	RepoWikiConfigListResponse,
	RepoWikiVersionListResponse,
} from '../models/response/repowiki'

export function getRepoWikiConfigList(
	params?: RepoWikiConfigListParams,
): Promise<BaseResponse<RepoWikiConfigListResponse>> {
	return apiClient.get('/api/v1/repowiki/configs', { params })
}

export function getRepoWikiConfig(id: string): Promise<BaseResponse<RepoWikiConfigItem>> {
	return apiClient.get(`/api/v1/repowiki/configs/${id}`)
}

export function createRepoWikiConfig(
	data: CreateRepoWikiConfigRequest,
): Promise<BaseResponse<RepoWikiConfigItem>> {
	return apiClient.post('/api/v1/repowiki/configs', data)
}

export function updateRepoWikiConfig(
	id: string,
	data: UpdateRepoWikiConfigRequest,
): Promise<BaseResponse<RepoWikiConfigItem>> {
	return apiClient.put(`/api/v1/repowiki/configs/${id}`, data)
}

export function deleteRepoWikiConfig(id: string): Promise<BaseResponse> {
	return apiClient.delete(`/api/v1/repowiki/configs/${id}`)
}

export function analyzeRepoWiki(
	id: string,
	data?: Record<string, unknown>,
): Promise<BaseResponse> {
	return apiClient.post(`/api/v1/repowiki/configs/${id}/analyze`, data ?? {})
}

export function updateRepoWiki(id: string): Promise<BaseResponse> {
	return apiClient.put(`/api/v1/repowiki/configs/${id}/update`)
}

export function getRepoWikiVersionList(
	configId: string,
	params?: RepoWikiConfigListParams,
): Promise<BaseResponse<RepoWikiVersionListResponse>> {
	return apiClient.get(`/api/v1/repowiki/configs/${configId}/versions`, { params })
}

export function getRepoWikiVersionStatus(
	versionId: string,
): Promise<BaseResponse<{ status: string; message?: string }>> {
	return apiClient.get(`/api/v1/repowiki/versions/${versionId}/status`)
}
