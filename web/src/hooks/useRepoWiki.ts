import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
	getRepoWikiConfigList,
	getRepoWikiConfig,
	getConfigByProjectId,
	createRepoWikiConfig,
	deleteRepoWikiConfig as deleteRepoWikiConfigApi,
	analyzeRepoWiki,
	updateRepoWiki as updateRepoWikiApi,
	getRepoWikiVersionList,
	updateSelectedVersion,
} from '#/lib/apis/repowiki'
import type {
	CreateRepoWikiConfigRequest,
	RepoWikiConfigListParams,
} from '#/lib/models/request/repowiki'

// ── 版本状态常量 ──

export const ACTIVE_STATUSES = ['pending', 'cloning', 'scanning', 'analyzing', 'assembling'] as const
export const TERMINAL_STATUSES = ['completed', 'failed', 'cancelled'] as const

// ── 配置列表 ──

export function useRepoWikiConfigs(params?: RepoWikiConfigListParams) {
	return useQuery({
		queryKey: ['repowiki', 'list', params],
		queryFn: () => getRepoWikiConfigList(params),
	})
}

// ── 单个配置详情 ──

export function useRepoWikiConfig(id: string) {
	return useQuery({
		queryKey: ['repowiki', 'detail', id],
		queryFn: () => getRepoWikiConfig(id),
		enabled: !!id,
	})
}

// ── 按项目 ID 查询配置 ──

export function useRepoWikiConfigByProjectId(projectId: string) {
	return useQuery({
		queryKey: ['repowiki', 'by-project', projectId],
		queryFn: () => getConfigByProjectId(projectId),
		enabled: !!projectId,
	})
}

// ── 创建配置 ──

export function useCreateRepoWikiConfig() {
	const queryClient = useQueryClient()
	return useMutation({
		mutationFn: (data: CreateRepoWikiConfigRequest) => createRepoWikiConfig(data),
		onSuccess: () => {
			toast.success('配置创建成功')
			queryClient.invalidateQueries({ queryKey: ['repowiki', 'list'] })
		},
		onError: (error: Error) => {
			toast.error(error.message || '创建失败')
		},
	})
}

// ── 删除配置 ──

export function useDeleteRepoWikiConfig() {
	const queryClient = useQueryClient()
	return useMutation({
		mutationFn: deleteRepoWikiConfigApi,
		onSuccess: () => {
			toast.success('配置已删除')
			queryClient.invalidateQueries({ queryKey: ['repowiki', 'list'] })
		},
		onError: (error: Error) => {
			toast.error(error.message || '删除失败')
		},
	})
}

// ── 触发分析（绑定到特定 config） ──

export function useRepoWikiAnalyze(configId: string) {
	const queryClient = useQueryClient()
	return useMutation({
		mutationFn: (data?: Record<string, unknown>) => analyzeRepoWiki(configId, data),
		onSuccess: () => {
			toast.success('分析任务已启动')
			queryClient.invalidateQueries({ queryKey: ['repowiki', 'versions', 'list', configId] })
		},
		onError: (error: Error) => {
			toast.error(error.message || '启动分析失败')
		},
	})
}

// ── 增量更新（绑定到特定 config） ──

export function useRepoWikiUpdate(configId: string) {
	const queryClient = useQueryClient()
	return useMutation({
		mutationFn: () => updateRepoWikiApi(configId),
		onSuccess: () => {
			toast.success('增量更新已启动')
			queryClient.invalidateQueries({ queryKey: ['repowiki', 'versions', 'list', configId] })
		},
		onError: (error: Error) => {
			toast.error(error.message || '启动更新失败')
		},
	})
}

// ── 版本列表（含自动轮询） ──

export function useRepoWikiVersions(configId: string, page = 1, size = 20) {
	return useQuery({
		queryKey: ['repowiki', 'versions', 'list', configId, page, size],
		queryFn: async () => {
			const res = await getRepoWikiVersionList(configId, { page, size })
			return res.data
		},
		enabled: !!configId,
		// 如果有任何版本处于活跃状态，每 3 秒刷新
		refetchInterval: (query) => {
			const data = query.state.data
			if (!data?.items) return false
			const hasActive = data.items.some((v) =>
				ACTIVE_STATUSES.includes(v.status as (typeof ACTIVE_STATUSES)[number]),
			)
			return hasActive ? 3000 : false
		},
	})
}

// ── 切换选中版本 ──

export function useUpdateSelectedVersion() {
	const queryClient = useQueryClient()
	return useMutation({
		mutationFn: ({ configId, versionId }: { configId: number; versionId: number }) =>
			updateSelectedVersion(configId, versionId),
		onSuccess: () => {
			toast.success('已切换选中版本')
			queryClient.invalidateQueries({ queryKey: ['repowiki', 'configs'] })
			queryClient.invalidateQueries({ queryKey: ['repowiki', 'list'] })
			queryClient.invalidateQueries({ queryKey: ['repowiki', 'by-project'] })
		},
		onError: (error: Error) => {
			toast.error(error.message || '切换版本失败')
		},
	})
}
