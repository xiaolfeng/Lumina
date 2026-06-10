import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import * as api from '#/lib/apis/qa-admin'
import type {
  SessionListParams,
  UpdateQaConfigRequest,
} from '#/lib/models/request/qa-admin'

export function useSessionList(params?: SessionListParams) {
  return useQuery({
    queryKey: ['qa', 'sessions', params],
    queryFn: () => api.getSessionList(params),
  })
}

export function useSessionDetail(id: string) {
  return useQuery({
    queryKey: ['qa', 'session', id],
    queryFn: () => api.getSessionDetail(id),
    enabled: !!id,
  })
}

export function useQuestionDetail(sessionId: string, questionId: string) {
  return useQuery({
    queryKey: ['qa', 'session', sessionId, 'question', questionId],
    queryFn: () => api.getQuestionDetail(sessionId, questionId),
    enabled: !!sessionId && !!questionId,
  })
}

export function useDeleteSession() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.deleteSession,
    onSuccess: () => {
      toast.success('会话已删除')
      queryClient.invalidateQueries({ queryKey: ['qa', 'sessions'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '删除失败')
    },
  })
}

export function useQaConfig() {
  return useQuery({
    queryKey: ['qa', 'config'],
    queryFn: api.getQaConfig,
  })
}

export function useUpdateQaConfig() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: UpdateQaConfigRequest) => api.updateQaConfig(data),
    onSuccess: () => {
      toast.success('配置已更新')
      queryClient.invalidateQueries({ queryKey: ['qa', 'config'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '更新失败')
    },
  })
}
