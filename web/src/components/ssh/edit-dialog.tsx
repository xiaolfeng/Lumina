import { useEffect, useState } from 'react'
import { Button } from '#/components/ui/button'
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from '#/components/ui/dialog'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { useUpdateSshKey } from '#/hooks/useSshKey'
import type { SshKeyItem } from '#/lib/models/response/ssh'

interface EditDialogProps {
	open: boolean
	onOpenChange: (open: boolean) => void
	item: SshKeyItem | null
}

export function EditDialog({ open, onOpenChange, item }: EditDialogProps) {
	const [name, setName] = useState('')
	const [description, setDescription] = useState('')

	const updateMutation = useUpdateSshKey()

	useEffect(() => {
		if (open && item) {
			setName(item.name)
			setDescription(item.description)
		}
	}, [open, item])

	const handleSubmit = () => {
		if (!item || !name.trim()) return
		updateMutation.mutate(
			{
				id: item.id,
				data: {
					name: name.trim(),
					description: description.trim() || undefined,
				},
			},
			{
				onSuccess: () => onOpenChange(false),
			},
		)
	}

	const handleClose = () => {
		setName('')
		setDescription('')
		onOpenChange(false)
	}

	return (
		<Dialog open={open} onOpenChange={handleClose}>
			<DialogContent className="sm:max-w-md">
				<DialogHeader>
					<DialogTitle>编辑 SSH 密钥</DialogTitle>
					<DialogDescription>修改密钥的名称和描述信息</DialogDescription>
				</DialogHeader>
				<div className="grid gap-4 py-4">
					<div className="grid gap-2">
						<Label htmlFor="edit-name">名称 *</Label>
						<Input
							id="edit-name"
							value={name}
							onChange={(e) => setName(e.target.value)}
							placeholder="输入密钥名称"
						/>
					</div>
					<div className="grid gap-2">
						<Label htmlFor="edit-desc">描述</Label>
						<Input
							id="edit-desc"
							value={description}
							onChange={(e) => setDescription(e.target.value)}
							placeholder="输入描述（可选）"
						/>
					</div>
				</div>
				<DialogFooter>
					<Button variant="outline" onClick={handleClose}>
						取消
					</Button>
					<Button
						onClick={handleSubmit}
						disabled={!name.trim() || updateMutation.isPending}
						className="bg-lagoon text-foam hover:bg-lagoon-deep"
					>
						{updateMutation.isPending ? '保存中...' : '保存'}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	)
}
