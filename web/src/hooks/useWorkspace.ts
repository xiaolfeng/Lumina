import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import * as api from '#/lib/apis/workspace'
import type {
  CreateWorkspaceRequest,
  UpdateWorkspaceRequest,
  WorkspaceListParams,
} from '#/lib/models/request/workspace'

export function useWorkspaceList(params?: WorkspaceListParams) {
  return useQuery({
    queryKey: ['workspace', 'list', params],
    queryFn: () => api.getWorkspaceList(params),
  })
}

export function useCreateWorkspace() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateWorkspaceRequest) => api.createWorkspace(data),
    onSuccess: () => {
      toast.success('空间创建成功')
      queryClient.invalidateQueries({ queryKey: ['workspace'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '创建失败')
    },
  })
}

export function useUpdateWorkspace() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateWorkspaceRequest }) =>
      api.updateWorkspace(id, data),
    onSuccess: () => {
      toast.success('空间更新成功')
      queryClient.invalidateQueries({ queryKey: ['workspace'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '更新失败')
    },
  })
}

export function useDeleteWorkspace() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.deleteWorkspace,
    onSuccess: (resp) => {
      const moved = resp.data?.moved_project_count ?? 0
      toast.success(
        moved > 0 ? `空间已删除，${moved} 个项目已移至默认空间` : '空间已删除',
      )
      queryClient.invalidateQueries({ queryKey: ['workspace'] })
      queryClient.invalidateQueries({ queryKey: ['project', 'list'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '删除失败')
    },
  })
}
