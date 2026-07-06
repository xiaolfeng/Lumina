import { apiClient } from './client'
import type { BaseResponse } from '../models/response/common'
import type {
  CreateProviderRequest,
  UpdateProviderRequest,
  CreateModelRequest,
  UpdateModelRequest,
  UpdateAgentModelRequest,
} from '../models/request/llm'
import type {
  Provider,
  Model,
  AgentModelAssignment,
  ProviderListResponse,
  ModelListResponse,
} from '../models/response/llm'

export function listProviders(
  page = 1,
  size = 20,
): Promise<BaseResponse<ProviderListResponse>> {
  return apiClient.get('/api/v1/llm/provider', { params: { page, size } })
}

export function getProvider(id: string): Promise<BaseResponse<Provider>> {
  return apiClient.get(`/api/v1/llm/provider/${id}`)
}

export function createProvider(
  data: CreateProviderRequest,
): Promise<BaseResponse<Provider>> {
  return apiClient.post('/api/v1/llm/provider', data)
}

export function updateProvider(
  id: string,
  data: UpdateProviderRequest,
): Promise<BaseResponse> {
  return apiClient.put(`/api/v1/llm/provider/${id}`, data)
}

export function deleteProvider(id: string): Promise<BaseResponse> {
  return apiClient.delete(`/api/v1/llm/provider/${id}`)
}

export function listModels(
  page = 1,
  size = 20,
  providerId?: string,
): Promise<BaseResponse<ModelListResponse>> {
  return apiClient.get('/api/v1/llm/model', {
    params: { page, size, provider_id: providerId },
  })
}

export function getModel(id: string): Promise<BaseResponse<Model>> {
  return apiClient.get(`/api/v1/llm/model/${id}`)
}

export function createModel(
  data: CreateModelRequest,
): Promise<BaseResponse<Model>> {
  return apiClient.post('/api/v1/llm/model', data)
}

export function updateModel(
  id: string,
  data: UpdateModelRequest,
): Promise<BaseResponse> {
  return apiClient.put(`/api/v1/llm/model/${id}`, data)
}

export function deleteModel(id: string): Promise<BaseResponse> {
  return apiClient.delete(`/api/v1/llm/model/${id}`)
}

export function getAgentModel(
  role: string,
): Promise<BaseResponse<AgentModelAssignment>> {
  return apiClient.get(`/api/v1/llm/agent/${role}/model`)
}

export function updateAgentModel(
  role: string,
  data: UpdateAgentModelRequest,
): Promise<BaseResponse> {
  return apiClient.put(`/api/v1/llm/agent/${role}/model`, data)
}
