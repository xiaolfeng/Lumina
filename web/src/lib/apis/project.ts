import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type { ProjectItem, ProjectListResponse } from '../models/response/project'
import type {
  CreateProjectRequest,
  UpdateProjectRequest,
  ProjectListParams,
} from '../models/request/project'

export function getProjectList(
  params?: ProjectListParams,
): Promise<BaseResponse<ProjectListResponse>> {
  return apiClient.get('/api/v1/project', { params })
}

export function createProject(
  data: CreateProjectRequest,
): Promise<BaseResponse<ProjectItem>> {
  return apiClient.post('/api/v1/project', data)
}

export function updateProject(
  id: string,
  data: UpdateProjectRequest,
): Promise<BaseResponse<ProjectItem>> {
  return apiClient.put(`/api/v1/project/${id}`, data)
}

export function deleteProject(id: string): Promise<BaseResponse> {
  return apiClient.delete(`/api/v1/project/${id}`)
}
