/**
 * Wiki 授权状态管理 Hook
 *
 * 提供 useQuery（检查授权状态）和 useMutation（密码验证）
 * 基于 TanStack Query，自动缓存 + 失效 + 重试
 */
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { wikiReaderApi } from '#/lib/api-client'
import type { AuthCheckResponse, AuthResponse } from '#/lib/api-client'

/** 检查授权状态的查询键工厂 */
export const wikiAuthKeys = {
  check: (wikiId: string) => ['wiki', 'auth-check', wikiId] as const,
}

/**
 * 检查 Wiki 授权状态
 *
 * - staleTime: 5 分钟内不重新请求
 * - retry: 1 次（网络抖动自动恢复）
 * - refetchOnWindowFocus: false（避免切换标签重复验证）
 */
export function useWikiAuthCheck(wikiId: string) {
  return useQuery<AuthCheckResponse>({
    queryKey: wikiAuthKeys.check(wikiId),
    queryFn: () => wikiReaderApi.checkAuth(wikiId),
    staleTime: 5 * 60 * 1000,
    retry: 1,
    refetchOnWindowFocus: false,
  })
}

/**
 * 密码验证 Mutation
 *
 * 成功后自动使 auth-check 缓存失效，触发重新检查
 */
export function useWikiAuth(wikiId: string) {
  const queryClient = useQueryClient()

  return useMutation<AuthResponse, Error, string>({
    mutationFn: (password: string) => wikiReaderApi.auth(wikiId, password),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: wikiAuthKeys.check(wikiId) })
    },
  })
}
