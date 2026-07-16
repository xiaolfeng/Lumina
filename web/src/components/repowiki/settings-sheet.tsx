import { useState, useEffect } from 'react'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from '@lumina/components/ui/sheet'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { Textarea } from '@lumina/components/ui/textarea'
import { Separator } from '@lumina/components/ui/separator'
import { Loader2 } from 'lucide-react'
import { useUpdateRepoWikiConfig, useDeleteRepoWikiConfig } from '#/hooks/useRepoWiki'
import type { RepoWikiConfigItem } from '#/lib/models/response/repowiki'
import { AnalyzeDialog } from '#/components/repowiki/analyze-dialog'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'

interface SettingsSheetProps {
	configId: string
	config: RepoWikiConfigItem
	open: boolean
	onOpenChange: (open: boolean) => void
}

export function SettingsSheet({ configId, config, open, onOpenChange }: SettingsSheetProps) {
	const [repoUrl, setRepoUrl] = useState('')
	const [defaultBranch, setDefaultBranch] = useState('')
	const [defaultLanguage, setDefaultLanguage] = useState('')
	const [wikiPassword, setWikiPassword] = useState('')
	const [customPrompt, setCustomPrompt] = useState('')

	const [deleteOpen, setDeleteOpen] = useState(false)

	const updateMutation = useUpdateRepoWikiConfig()
	const deleteMutation = useDeleteRepoWikiConfig()

	useEffect(() => {
		if (open) {
			setRepoUrl(config.repo_url)
			setDefaultBranch(config.default_branch)
			setDefaultLanguage(config.default_language)
			setWikiPassword('')
			setCustomPrompt(config.custom_prompt ?? '')
		}
	}, [open, config])

	const handleSave = () => {
		updateMutation.mutate(
			{
				id: configId,
				data: {
					repo_url: repoUrl.trim() || undefined,
					default_branch: defaultBranch.trim() || undefined,
					default_language: defaultLanguage.trim() || undefined,
					wiki_password: wikiPassword.trim() || undefined,
					custom_prompt: customPrompt.trim(),
				},
			},
			{
				onSuccess: () => {
					onOpenChange(false)
				},
			},
		)
	}

	const handleDelete = () => {
		deleteMutation.mutate(configId, {
			onSuccess: () => {
				setDeleteOpen(false)
				onOpenChange(false)
			},
		})
	}

	const isPending = updateMutation.isPending

	return (
		<Sheet open={open} onOpenChange={onOpenChange}>
			<SheetContent className="overflow-y-auto sm:max-w-md">
				<SheetHeader>
					<SheetTitle>配置设置</SheetTitle>
					<SheetDescription>修改仓库 Wiki 的基础配置参数</SheetDescription>
				</SheetHeader>

				<div className="grid gap-4 py-4">
					<div className="grid gap-2">
						<Label htmlFor="settings-repo-url">仓库地址</Label>
						<Input
							id="settings-repo-url"
							value={repoUrl}
							onChange={(e) => setRepoUrl(e.target.value)}
							placeholder="https://github.com/owner/repo.git"
							disabled={isPending}
						/>
					</div>

					<div className="grid grid-cols-2 gap-4">
						<div className="grid gap-2">
							<Label htmlFor="settings-branch">默认分支</Label>
							<Input
								id="settings-branch"
								value={defaultBranch}
								onChange={(e) => setDefaultBranch(e.target.value)}
								placeholder="main"
								disabled={isPending}
							/>
						</div>
						<div className="grid gap-2">
							<Label htmlFor="settings-lang">默认语言</Label>
							<Input
								id="settings-lang"
								value={defaultLanguage}
								onChange={(e) => setDefaultLanguage(e.target.value)}
								placeholder="zh"
								disabled={isPending}
							/>
						</div>
					</div>

					<div className="grid gap-2">
						<Label htmlFor="settings-password">Wiki 密码</Label>
						<Input
							id="settings-password"
							type="password"
							value={wikiPassword}
							onChange={(e) => setWikiPassword(e.target.value)}
							placeholder="留空则不修改"
							disabled={isPending}
						/>
						<p className="text-xs text-muted-foreground">留空表示不修改当前密码</p>
					</div>

					<div className="grid gap-2">
						<Label htmlFor="settings-prompt">自定义提示词（L2）</Label>
						<Textarea
							id="settings-prompt"
							value={customPrompt}
							onChange={(e) => setCustomPrompt(e.target.value)}
							placeholder="为所有分析阶段添加全局自定义指示..."
							rows={3}
							disabled={isPending}
						/>
						<p className="text-xs text-muted-foreground">
							此提示词会持久化并应用于每次分析
						</p>
					</div>

					<Button onClick={handleSave} disabled={isPending || !repoUrl.trim()} className="w-full">
						{isPending ? (
							<>
								<Loader2 className="mr-2 size-4 animate-spin" />
								保存中...
							</>
						) : (
							'保存配置'
						)}
					</Button>
				</div>

				<Separator />

				<div className="space-y-3">
					<h4 className="text-sm font-medium text-destructive">危险操作区</h4>
					<div className="flex flex-wrap gap-2">
						<AnalyzeDialog
							configId={configId}
							mode="analyze"
							trigger={<Button variant="destructive">重置分析</Button>}
						/>
						<Button variant="destructive" onClick={() => setDeleteOpen(true)}>
							删除配置
						</Button>
					</div>
				</div>

				<ConfirmDeleteDialog
					open={deleteOpen}
					onOpenChange={setDeleteOpen}
					title="删除配置"
					description="确定要删除此 Wiki 配置吗？此操作不可恢复，关联的版本数据也将被清除。"
					onConfirm={handleDelete}
					isPending={deleteMutation.isPending}
				/>
			</SheetContent>
		</Sheet>
	)
}
