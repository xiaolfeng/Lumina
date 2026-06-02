import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import * as api from '#/lib/apis/apikey'
import type {
  ApikeyUpdateRequest,
  ApikeyListParams,
} from '#/lib/models/request/apikey'

export function useApikeyList(params?: ApikeyListParams) {
  return useQuery({
    queryKey: ['apikey', 'list', params],
    queryFn: () => api.getApikeyList(params),
  })
}

export function useApikeyDetail(id: string) {
  return useQuery({
    queryKey: ['apikey', 'detail', id],
    queryFn: () => api.getApikeyDetail(id),
    enabled: !!id,
  })
}

export function useCreateApikey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.createApikey,
    onSuccess: () => {
      toast.success('令牌创建成功')
      queryClient.invalidateQueries({ queryKey: ['apikey', 'list'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '创建失败')
    },
  })
}

export function useUpdateApikey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: ApikeyUpdateRequest }) =>
      api.updateApikey(id, data),
    onSuccess: () => {
      toast.success('令牌更新成功')
      queryClient.invalidateQueries({ queryKey: ['apikey', 'list'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '更新失败')
    },
  })
}

export function useDeleteApikey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.deleteApikey,
    onSuccess: () => {
      toast.success('令牌已删除')
      queryClient.invalidateQueries({ queryKey: ['apikey', 'list'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '删除失败')
    },
  })
}

export function useResetApikey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.resetApikey,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['apikey', 'list'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '重置失败')
    },
  })
}
