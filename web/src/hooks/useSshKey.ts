import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import * as api from '#/lib/apis/ssh'
import type {
	UpdateSshKeyRequest,
	SshKeyListParams,
} from '#/lib/models/request/ssh'

export function useSshKeyList(params?: SshKeyListParams) {
	return useQuery({
		queryKey: ['ssh', 'list', params],
		queryFn: () => api.listSshKeys(params),
	})
}

export function useSshKeyDetail(id: string) {
	return useQuery({
		queryKey: ['ssh', 'detail', id],
		queryFn: () => api.getSshKey(id),
		enabled: !!id,
	})
}

export function useCreateSshKey() {
	const queryClient = useQueryClient()
	return useMutation({
		mutationFn: api.createSshKey,
		onSuccess: () => {
			toast.success('SSH 密钥创建成功')
			queryClient.invalidateQueries({ queryKey: ['ssh', 'list'] })
		},
		onError: (error: Error) => {
			toast.error(error.message || '创建失败')
		},
	})
}

export function useUpdateSshKey() {
	const queryClient = useQueryClient()
	return useMutation({
		mutationFn: ({ id, data }: { id: string; data: UpdateSshKeyRequest }) =>
			api.updateSshKey(id, data),
		onSuccess: () => {
			toast.success('SSH 密钥更新成功')
			queryClient.invalidateQueries({ queryKey: ['ssh', 'list'] })
		},
		onError: (error: Error) => {
			toast.error(error.message || '更新失败')
		},
	})
}

export function useDeleteSshKey() {
	const queryClient = useQueryClient()
	return useMutation({
		mutationFn: api.deleteSshKey,
		onSuccess: () => {
			toast.success('SSH 密钥已删除')
			queryClient.invalidateQueries({ queryKey: ['ssh', 'list'] })
		},
		onError: (error: Error) => {
			toast.error(error.message || '删除失败')
		},
	})
}
