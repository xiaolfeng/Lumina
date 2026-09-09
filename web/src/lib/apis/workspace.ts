import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type {
  WorkspaceItem,
  WorkspaceListResponse,
  DeleteWorkspaceResponse,
} from '../models/response/workspace'
import type {
  CreateWorkspaceRequest,
  UpdateWorkspaceRequest,
  WorkspaceListParams,
} from '../models/request/workspace'

export function getWorkspaceList(
  params?: WorkspaceListParams,
): Promise<BaseResponse<WorkspaceListResponse>> {
  return apiClient.get('/api/v1/workspace', { params })
}

export function getWorkspace(id: string): Promise<BaseResponse<WorkspaceItem>> {
  return apiClient.get(`/api/v1/workspace/${id}`)
}

export function createWorkspace(
  data: CreateWorkspaceRequest,
): Promise<BaseResponse<WorkspaceItem>> {
  return apiClient.post('/api/v1/workspace', data)
}

export function updateWorkspace(
  id: string,
  data: UpdateWorkspaceRequest,
): Promise<BaseResponse<WorkspaceItem>> {
  return apiClient.put(`/api/v1/workspace/${id}`, data)
}

export function deleteWorkspace(
  id: string,
): Promise<BaseResponse<DeleteWorkspaceResponse>> {
  return apiClient.delete(`/api/v1/workspace/${id}`)
}
