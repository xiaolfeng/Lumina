/**
 * Wiki Reader 数据获取 Hooks
 *
 * 提供 useWikiPage（页面内容）和 useWikiManifest（导航清单）
 * 基于 TanStack Query，自动缓存 + 失效 + 重试
 */
import { useQuery } from '@tanstack/react-query'
import { wikiReaderApi } from '#/lib/api-client'
import type { PageResponse, ManifestResponse } from '#/lib/api-client'

/** Wiki 数据查询键工厂 */
export const wikiKeys = {
  page: (wikiId: string, path: string) =>
    ['wiki', 'page', wikiId, path] as const,
  manifest: (wikiId: string) => ['wiki', 'manifest', wikiId] as const,
}

/**
 * 获取 Wiki 页面 Markdown 内容
 *
 * - retry: false — 401 不重试（由 PasswordGate 处理）
 * - staleTime: 2 分钟内不重新请求
 * - refetchOnWindowFocus: false（避免切换标签重复加载）
 */
export function useWikiPage(wikiId: string, path: string) {
  return useQuery<PageResponse>({
    queryKey: wikiKeys.page(wikiId, path),
    queryFn: () => wikiReaderApi.getPage(wikiId, path),
    enabled: !!wikiId && !!path,
    retry: false,
    staleTime: 2 * 60 * 1000,
    refetchOnWindowFocus: false,
  })
}

/**
 * 获取 Wiki 导航清单（侧边栏）
 *
 * - staleTime: 5 分钟内不重新请求
 * - retry: 1 次（网络抖动自动恢复）
 */
export function useWikiManifest(wikiId: string) {
  return useQuery<ManifestResponse>({
    queryKey: wikiKeys.manifest(wikiId),
    queryFn: () => wikiReaderApi.getManifest(wikiId),
    enabled: !!wikiId,
    staleTime: 5 * 60 * 1000,
    retry: 1,
    refetchOnWindowFocus: false,
  })
}
