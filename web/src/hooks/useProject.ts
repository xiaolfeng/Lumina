import { useMemo } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import * as api from '#/lib/apis/project'
import type {
  CreateProjectRequest,
  UpdateProjectRequest,
  ProjectListParams,
} from '#/lib/models/request/project'

const projectNameListParams = { page: 1, size: 200 } as const

export function useProjectList(params?: ProjectListParams) {
  return useQuery({
    queryKey: ['project', 'list', params],
    queryFn: () => api.getProjectList(params),
  })
}

export function useProjectNameMap(options?: { enabled?: boolean }) {
  const { data, isLoading } = useQuery({
    queryKey: ['project', 'list', projectNameListParams],
    queryFn: () => api.getProjectList(projectNameListParams),
    enabled: options?.enabled ?? true,
    staleTime: 60_000,
  })

  const names = useMemo(() => {
    const map: Record<string, string> = {}
    for (const project of data?.data?.items ?? []) {
      map[project.id] = project.name
    }
    return map
  }, [data])

  return { names, isLoading }
}

export function useCreateProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateProjectRequest) => api.createProject(data),
    onSuccess: () => {
      toast.success('项目创建成功')
      queryClient.invalidateQueries({ queryKey: ['project', 'list'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '创建失败')
    },
  })
}

export function useUpdateProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateProjectRequest }) =>
      api.updateProject(id, data),
    onSuccess: () => {
      toast.success('项目更新成功')
      queryClient.invalidateQueries({ queryKey: ['project', 'list'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '更新失败')
    },
  })
}

export function useDeleteProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.deleteProject,
    onSuccess: () => {
      toast.success('项目已删除')
      queryClient.invalidateQueries({ queryKey: ['project', 'list'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '删除失败')
    },
  })
}
