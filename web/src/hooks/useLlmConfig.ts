import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as llmApi from '../lib/apis/llm'
import type {
  CreateProviderRequest,
  UpdateProviderRequest,
  CreateModelRequest,
  UpdateModelRequest,
  UpdateAgentModelRequest,
} from '../lib/models/request/llm'

export function useProviders(page = 1, size = 20) {
  return useQuery({
    queryKey: ['llm', 'providers', page, size],
    queryFn: () => llmApi.listProviders(page, size),
  })
}

export function useProvider(id: string) {
  return useQuery({
    queryKey: ['llm', 'providers', id],
    queryFn: () => llmApi.getProvider(id),
    enabled: !!id,
  })
}

export function useCreateProvider() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateProviderRequest) => llmApi.createProvider(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['llm', 'providers'] })
    },
  })
}

export function useUpdateProvider() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateProviderRequest }) =>
      llmApi.updateProvider(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['llm', 'providers'] })
    },
  })
}

export function useDeleteProvider() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => llmApi.deleteProvider(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['llm', 'providers'] })
    },
  })
}

export function useModels(page = 1, size = 20, providerId?: string) {
  return useQuery({
    queryKey: ['llm', 'models', page, size, providerId],
    queryFn: () => llmApi.listModels(page, size, providerId),
  })
}

export function useModel(id: string) {
  return useQuery({
    queryKey: ['llm', 'models', id],
    queryFn: () => llmApi.getModel(id),
    enabled: !!id,
  })
}

export function useCreateModel() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateModelRequest) => llmApi.createModel(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['llm', 'models'] })
    },
  })
}

export function useUpdateModel() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateModelRequest }) =>
      llmApi.updateModel(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['llm', 'models'] })
    },
  })
}

export function useDeleteModel() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => llmApi.deleteModel(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['llm', 'models'] })
    },
  })
}

export function useAgentModel(role: string) {
  return useQuery({
    queryKey: ['llm', 'agent', role],
    queryFn: () => llmApi.getAgentModel(role),
    enabled: !!role,
  })
}

export function useAgentModels(module: string) {
  return useQuery({
    queryKey: ['llm', 'agent-models', module],
    queryFn: () => llmApi.getAgentModels(module),
    enabled: !!module,
  })
}

export function useUpdateAgentModel() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      role,
      data,
    }: {
      role: string
      data: UpdateAgentModelRequest
    }) => llmApi.updateAgentModel(role, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['llm', 'agent'] })
      queryClient.invalidateQueries({ queryKey: ['llm', 'agent-models'] })
    },
  })
}
