import { useState } from 'react'
import {
	Dialog,
	DialogTrigger,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogDescription,
	DialogFooter,
} from '@lumina/components/ui/dialog'
import { Textarea } from '@lumina/components/ui/textarea'
import { Button } from '@lumina/components/ui/button'
import { Loader2, Play, RefreshCw } from 'lucide-react'
import { useRepoWikiAnalyze, useRepoWikiUpdate } from '#/hooks/useRepoWiki'

interface AnalyzeDialogProps {
	configId: string
	mode: 'analyze' | 'update'
	trigger?: React.ReactNode
	open?: boolean
	onOpenChange?: (open: boolean) => void
}

export function AnalyzeDialog({ configId, mode, trigger, open, onOpenChange }: AnalyzeDialogProps) {
	const [extraPrompt, setExtraPrompt] = useState('')
	const [internalOpen, setInternalOpen] = useState(false)

	const isControlled = open !== undefined
	const isOpen = isControlled ? open : internalOpen
	const setOpen = isControlled ? (onOpenChange ?? (() => {})) : setInternalOpen

	const analyzeMutation = useRepoWikiAnalyze(configId)
	const updateMutation = useRepoWikiUpdate(configId)
	const mutation = mode === 'analyze' ? analyzeMutation : updateMutation
	const isPending = mutation.isPending

	const handleConfirm = () => {
		mutation.mutate(
			{ extra_prompt: extraPrompt.trim() || undefined },
			{
				onSuccess: () => {
					setOpen(false)
					setExtraPrompt('')
				},
			},
		)
	}

	const titleText = mode === 'analyze' ? '开始分析' : '增量更新'
	const descText =
		mode === 'analyze'
			? '这将触发一次完整的仓库分析流程，可能需要较长时间。'
			: '将基于已有分析结果进行增量更新。'
	const confirmText = mode === 'analyze' ? '开始分析' : '增量更新'
	const Icon = mode === 'analyze' ? Play : RefreshCw

	return (
		<Dialog open={isOpen} onOpenChange={setOpen}>
			{trigger && <DialogTrigger asChild>{trigger}</DialogTrigger>}
			<DialogContent>
				<DialogHeader>
					<DialogTitle>{titleText}</DialogTitle>
					<DialogDescription>{descText}</DialogDescription>
				</DialogHeader>
				<div className="grid gap-2 py-4">
					<label className="text-sm font-medium">额外提示词（可选）</label>
					<Textarea
						value={extraPrompt}
						onChange={(e) => setExtraPrompt(e.target.value)}
						placeholder="为本次分析添加额外指示，如「重点关注鉴权模块」..."
						rows={4}
						disabled={isPending}
					/>
					<p className="text-xs text-muted-foreground">此提示词仅在本次分析生效，不会持久化。</p>
				</div>
				<DialogFooter>
					<Button variant="outline" onClick={() => setOpen(false)} disabled={isPending}>
						取消
					</Button>
					<Button onClick={handleConfirm} disabled={isPending}>
						{isPending ? (
							<Loader2 className="mr-2 size-4 animate-spin" />
						) : (
							<Icon className="mr-2 size-4" />
						)}
						{confirmText}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	)
}
