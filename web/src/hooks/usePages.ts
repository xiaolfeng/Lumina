import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import * as api from '#/lib/apis/pages'
import type { PageListParams } from '#/lib/apis/pages'

export function usePageList(params?: PageListParams) {
  return useQuery({
    queryKey: ['pages', 'list', params],
    queryFn: () => api.getPages(params),
    enabled: Boolean(params?.workspace_id),
  })
}

export function usePageVersions(id?: string) {
  return useQuery({
    queryKey: ['pages', 'versions', id],
    queryFn: () => api.getPageVersions(id!),
    enabled: Boolean(id),
  })
}

// 展示态版本列表走公开端点（带密码门校验），避免访客触发管理接口 401 被甩到登录页
export function usePageVersionsByProject(
  projectName?: string,
  slug?: string,
  enabled = true,
) {
  return useQuery({
    queryKey: ['pages', 'versions-public', projectName, slug],
    queryFn: () => api.getPageVersionsByProject(projectName!, slug!),
    enabled: Boolean(projectName && slug) && enabled,
  })
}

export function usePageMeta(
  projectName?: string,
  slug?: string,
  version?: string,
  enabled = true,
) {
  return useQuery({
    queryKey: ['pages', 'meta', projectName, slug, version],
    queryFn: () => api.getPageMeta(projectName!, slug!, version),
    enabled: Boolean(projectName && slug) && enabled,
  })
}

export function usePageAuthCheck(projectName?: string, slug?: string) {
  return useQuery({
    queryKey: ['pages', 'auth-check', projectName, slug],
    queryFn: () => api.checkPageAuth(projectName!, slug!),
    enabled: Boolean(projectName && slug),
  })
}

export function useUnlockPage() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      projectName,
      slug,
      password,
    }: {
      projectName: string
      slug: string
      password: string
    }) => api.unlockPage(projectName, slug, password),
    onSuccess: (_data, variables) => {
      toast.success('已解锁')
      queryClient.setQueryData(
        ['pages', 'auth-check', variables.projectName, variables.slug],
        (old: any) => ({
          code: 200,
          message: 'OK',
          ...old,
          data: {
            authenticated: true,
            password_required: true,
            ...old?.data,
          },
        }),
      )
      queryClient.invalidateQueries({
        queryKey: ['pages', 'auth-check', variables.projectName, variables.slug],
      })
      queryClient.invalidateQueries({
        queryKey: ['pages', 'meta', variables.projectName, variables.slug],
      })
      queryClient.invalidateQueries({
        queryKey: [
          'pages',
          'versions-public',
          variables.projectName,
          variables.slug,
        ],
      })
    },
    onError: (error: Error) => {
      toast.error(error.message || '密码错误')
    },
  })
}

export function useUpdateAccessPolicy() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      id,
      access_mode,
      password,
    }: {
      id: string
      access_mode: 'public' | 'password'
      password?: string
    }) => api.updatePageAccessPolicy(id, { access_mode, password }),
    onSuccess: () => {
      toast.success('访问策略已更新')
      queryClient.invalidateQueries({ queryKey: ['pages'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '更新失败')
    },
  })
}

export function useArchivePage() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: api.archivePage,
    onSuccess: () => {
      toast.success('页面已归档')
      queryClient.invalidateQueries({ queryKey: ['pages'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '归档失败')
    },
  })
}

export function useSwitchActiveVersion() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, versionId }: { id: string; versionId: string }) =>
      api.switchActiveVersion(id, versionId),
    onSuccess: () => {
      toast.success('已切换生效版本')
      queryClient.invalidateQueries({ queryKey: ['pages'] })
    },
    onError: (error: Error) => {
      toast.error(error.message || '切换失败')
    },
  })
}

export function useForkPage() {
  return useMutation({
    mutationFn: ({ id, versionId }: { id: string; versionId?: string }) =>
      api.forkPage(id, versionId),
    onError: (error: Error) => {
      toast.error(error.message || 'Fork 失败')
    },
  })
}
