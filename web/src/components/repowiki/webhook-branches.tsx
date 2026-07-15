import { useState } from 'react'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Badge } from '@lumina/components/ui/badge'
import { useWebhookConfig, useUpdateWebhookBranches } from '#/hooks/useWebhook'
import { X, Plus } from 'lucide-react'
import { toast } from 'sonner'

interface WebhookBranchesProps {
	configId: string
}

/**
 * Webhook 监听分支管理（section 风格，无外层 Card 包裹）。
 * 由 WebhookTab 组合使用。
 */
export function WebhookBranches({ configId }: WebhookBranchesProps) {
	const { data, isLoading } = useWebhookConfig(configId)
	const updateMutation = useUpdateWebhookBranches(configId)
	const [newBranch, setNewBranch] = useState('')

	const branches = data?.data?.branches ?? []

	const handleAdd = () => {
		const trimmed = newBranch.trim()
		if (!trimmed) return
		if (branches.includes(trimmed)) {
			toast.error('该分支已存在')
			return
		}
		const updated = [...branches, trimmed]
		updateMutation.mutate(updated, {
			onSuccess: () => {
				setNewBranch('')
				toast.success('分支已添加')
			},
			onError: (error: Error) => {
				toast.error(error.message || '添加失败')
			},
		})
	}

	const handleRemove = (branch: string) => {
		const updated = branches.filter((b) => b !== branch)
		updateMutation.mutate(updated, {
			onSuccess: () => {
				toast.success('分支已移除')
			},
			onError: (error: Error) => {
				toast.error(error.message || '移除失败')
			},
		})
	}

	const handleKeyDown = (e: React.KeyboardEvent) => {
		if (e.key === 'Enter') {
			e.preventDefault()
			handleAdd()
		}
	}

	if (isLoading) {
		return (
			<div className="space-y-3">
				<div className="h-8 w-24 animate-pulse rounded bg-muted" />
				<div className="h-9 animate-pulse rounded bg-muted" />
			</div>
		)
	}

	return (
		<div className="space-y-3">
			{/* Add branch */}
			<div className="flex gap-2">
				<Input
					value={newBranch}
					onChange={(e) => setNewBranch(e.target.value)}
					onKeyDown={handleKeyDown}
					placeholder="输入分支名称（如 main），回车添加"
					disabled={updateMutation.isPending}
				/>
				<Button
					onClick={handleAdd}
					disabled={!newBranch.trim() || updateMutation.isPending}
					className="gap-2"
				>
					<Plus className="size-4" />
					添加
				</Button>
			</div>

			{/* Branch list */}
			{branches.length === 0 ? (
				<div className="rounded-md border border-dashed p-4 text-center">
					<p className="text-muted-foreground text-sm">
						未配置监听分支，Webhook 不会触发分析
					</p>
				</div>
			) : (
				<div className="flex flex-wrap gap-2">
					{branches.map((branch) => (
						<Badge
							key={branch}
							variant="secondary"
							className="gap-1 px-2.5 py-1 text-sm font-mono"
						>
							{branch}
							<button
								onClick={() => handleRemove(branch)}
								disabled={updateMutation.isPending}
								className="ml-1 rounded-full p-0.5 hover:bg-muted-foreground/20 transition-colors disabled:opacity-50 cursor-pointer"
								aria-label={`移除分支 ${branch}`}
							>
								<X className="size-3" />
							</button>
						</Badge>
					))}
				</div>
			)}
		</div>
	)
}
