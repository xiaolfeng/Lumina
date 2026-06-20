import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import * as api from '#/lib/apis/pin'
import type {
  CreatePinRequest,
  UpdatePinRequest,
  PinListParams,
} from '#/lib/models/request/pin'

export function usePinList(params?: PinListParams) {
  return useQuery({
    queryKey: ['pin', 'list', params],
    queryFn: () => api.getPinList(params),
  })
}

export function usePinDetail(id: string) {
  return useQuery({
    queryKey: ['pin', 'detail', id],
    queryFn: () => api.getPinDetail(id),
    enabled: !!id,
  })
}

export function useCreatePin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreatePinRequest) => api.createPin(data),
    onSuccess: () => {
      toast.success('Pin 创建成功')
      queryClient.invalidateQueries({ queryKey: ['pin', 'list'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '创建失败')
    },
  })
}

export function useUpdatePin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdatePinRequest }) =>
      api.updatePin(id, data),
    onSuccess: () => {
      toast.success('Pin 更新成功')
      queryClient.invalidateQueries({ queryKey: ['pin', 'list'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '更新失败')
    },
  })
}

export function useDeletePin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.deletePin,
    onSuccess: () => {
      toast.success('Pin 已删除')
      queryClient.invalidateQueries({ queryKey: ['pin', 'list'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '删除失败')
    },
  })
}
