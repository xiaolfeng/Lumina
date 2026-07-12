import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type {
	CreateSshKeyRequest,
	UpdateSshKeyRequest,
	SshKeyListParams,
} from '../models/request/ssh'
import type {
	SshKeyItem,
	CreateSshKeyResponse,
	SshKeyListResponse,
} from '../models/response/ssh'

export function createSshKey(
	data: CreateSshKeyRequest,
): Promise<BaseResponse<CreateSshKeyResponse>> {
	return apiClient.post('/api/v1/ssh', data)
}

export function listSshKeys(
	params?: SshKeyListParams,
): Promise<BaseResponse<SshKeyListResponse>> {
	return apiClient.get('/api/v1/ssh', { params })
}

export function getSshKey(id: string): Promise<BaseResponse<SshKeyItem>> {
	return apiClient.get(`/api/v1/ssh/${id}`)
}

export function updateSshKey(
	id: string,
	data: UpdateSshKeyRequest,
): Promise<BaseResponse<SshKeyItem>> {
	return apiClient.put(`/api/v1/ssh/${id}`, data)
}

export function deleteSshKey(id: string): Promise<BaseResponse> {
	return apiClient.delete(`/api/v1/ssh/${id}`)
}

export function getSshPublicKey(id: string): Promise<Blob> {
	return apiClient.get(`/api/v1/ssh/${id}/public-key`, {
		responseType: 'blob',
	})
}
