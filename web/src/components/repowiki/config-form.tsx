import { useState } from 'react'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Textarea } from '#/components/ui/textarea'
import type { CreateRepoWikiConfigRequest } from '#/lib/models/request/repowiki'

interface ConfigFormProps {
	onSubmit: (data: CreateRepoWikiConfigRequest) => void
	isPending?: boolean
	onCancel?: () => void
}

export function ConfigForm({ onSubmit, isPending, onCancel }: ConfigFormProps) {
	const [name, setName] = useState('')
	const [repoUrl, setRepoUrl] = useState('')
	const [defaultBranch, setDefaultBranch] = useState('main')
	const [defaultLanguage, setDefaultLanguage] = useState('zh')
	const [sshKey, setSshKey] = useState('')
	const [wikiPassword, setWikiPassword] = useState('')
	const [projectId, setProjectId] = useState('')

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault()
		if (!name.trim() || !repoUrl.trim()) return

		onSubmit({
			name: name.trim(),
			repo_url: repoUrl.trim(),
			default_branch: defaultBranch.trim() || undefined,
			default_language: defaultLanguage.trim() || undefined,
			ssh_key: sshKey.trim() || undefined,
			wiki_password: wikiPassword.trim() || undefined,
			project_id: projectId.trim() || undefined,
		})
	}

	return (
		<form onSubmit={handleSubmit} className="grid gap-4 py-4">
			{/* 仓库名称 */}
			<div className="grid gap-2">
				<Label htmlFor="rw-name">仓库名称 *</Label>
				<Input
					id="rw-name"
					value={name}
					onChange={(e) => setName(e.target.value)}
					placeholder="输入仓库名称"
					disabled={isPending}
				/>
			</div>

			{/* 仓库地址 */}
			<div className="grid gap-2">
				<Label htmlFor="rw-url">仓库地址 *</Label>
				<Input
					id="rw-url"
					value={repoUrl}
					onChange={(e) => setRepoUrl(e.target.value)}
					placeholder="https://github.com/owner/repo.git 或 git@github.com:owner/repo.git"
					disabled={isPending}
				/>
			</div>

			{/* 默认分支 & 默认语言 - 两列布局 */}
			<div className="grid grid-cols-2 gap-4">
				<div className="grid gap-2">
					<Label htmlFor="rw-branch">默认分支</Label>
					<Input
						id="rw-branch"
						value={defaultBranch}
						onChange={(e) => setDefaultBranch(e.target.value)}
						placeholder="main"
						disabled={isPending}
					/>
				</div>
				<div className="grid gap-2">
					<Label htmlFor="rw-lang">默认语言</Label>
					<Input
						id="rw-lang"
						value={defaultLanguage}
						onChange={(e) => setDefaultLanguage(e.target.value)}
						placeholder="zh"
						disabled={isPending}
					/>
				</div>
			</div>

			{/* SSH 私钥 */}
			<div className="grid gap-2">
				<Label htmlFor="rw-ssh">SSH 私钥</Label>
				<Textarea
					id="rw-ssh"
					value={sshKey}
					onChange={(e) => setSshKey(e.target.value)}
					placeholder="用于克隆私有仓库（可选）"
					className="min-h-[80px] font-mono text-xs"
					disabled={isPending}
				/>
			</div>

			{/* Wiki 密码 */}
			<div className="grid gap-2">
				<Label htmlFor="rw-password">Wiki 密码</Label>
				<Input
					id="rw-password"
					type="password"
					value={wikiPassword}
					onChange={(e) => setWikiPassword(e.target.value)}
					placeholder="访问 Wiki 的密码（可选）"
					disabled={isPending}
				/>
			</div>

			{/* 项目 ID */}
			<div className="grid gap-2">
				<Label htmlFor="rw-project">项目 ID</Label>
				<Input
					id="rw-project"
					value={projectId}
					onChange={(e) => setProjectId(e.target.value)}
					placeholder="关联的项目 ID（可选）"
					disabled={isPending}
				/>
			</div>

			{/* 操作按钮 */}
			<div className="flex justify-end gap-2 pt-2">
				{onCancel && (
					<Button type="button" variant="outline" onClick={onCancel} disabled={isPending}>
						取消
					</Button>
				)}
				<Button
					type="submit"
					disabled={!name.trim() || !repoUrl.trim() || isPending}
					className="bg-lagoon text-foam hover:bg-lagoon-deep"
				>
					{isPending ? '提交中...' : '创建配置'}
				</Button>
			</div>
		</form>
	)
}
