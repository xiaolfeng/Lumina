import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import * as api from '#/lib/apis/preview'
import type { PreviewSessionListParams } from '#/lib/apis/preview'

export function usePreviewSessionList(params?: PreviewSessionListParams) {
  return useQuery({
    queryKey: ['preview', 'sessions', params],
    queryFn: () => api.getPreviewSessions(params),
    enabled: Boolean(params?.workspace_id),
  })
}

export function useCreatePreviewSession() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.createPreviewSession,
    onSuccess: () => {
      toast.success('预览会话已创建')
      queryClient.invalidateQueries({ queryKey: ['preview', 'sessions'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '创建失败')
    },
  })
}

export function useDeletePreviewSession() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.deletePreviewSession,
    onSuccess: () => {
      toast.success('预览会话已删除')
      queryClient.invalidateQueries({ queryKey: ['preview', 'sessions'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '删除失败')
    },
  })
}

export function usePreviewSessionDetail(hash: string | null) {
  return useQuery({
    queryKey: ['preview', 'session-detail', hash],
    queryFn: () => api.getPreviewSessionDetail(hash!),
    enabled: Boolean(hash),
  })
}

export function useDeletePreviewFile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.deletePreviewFile,
    onSuccess: () => {
      toast.success('预览文件已删除')
      queryClient.invalidateQueries({ queryKey: ['preview', 'session-detail'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '删除失败')
    },
  })
}
