import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from '@lumina/components/ui/select'
import type { CreateRepoWikiConfigRequest } from '#/lib/models/request/repowiki'
import { useSshKeyList } from '#/hooks/useSshKey'

interface ConfigFormProps {
	onSubmit: (data: CreateRepoWikiConfigRequest) => void
	isPending?: boolean
	onCancel?: () => void
	projectId?: string
}

export function ConfigForm({ onSubmit, isPending, onCancel, projectId }: ConfigFormProps) {
	const [name, setName] = useState('')
	const [repoUrl, setRepoUrl] = useState('')
	const [defaultBranch, setDefaultBranch] = useState('main')
	const [defaultLanguage, setDefaultLanguage] = useState('zh')
	const [sshKeyId, setSshKeyId] = useState('__none__')
	const [wikiPassword, setWikiPassword] = useState('')

	const { data: sshData } = useSshKeyList()
	const sshKeys = sshData?.data?.items ?? []

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault()
		if (!name.trim() || !repoUrl.trim()) return

		onSubmit({
			name: name.trim(),
			repo_url: repoUrl.trim(),
			default_branch: defaultBranch.trim() || undefined,
			default_language: defaultLanguage.trim() || undefined,
			ssh_key_id: sshKeyId && sshKeyId !== '__none__' ? sshKeyId : undefined,
			wiki_password: wikiPassword.trim() || undefined,
			project_id: projectId ?? '',
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

			{/* SSH 密钥选择 */}
			<div className="grid gap-2">
				<Label>SSH 密钥</Label>
				<Select value={sshKeyId} onValueChange={setSshKeyId} disabled={isPending}>
					<SelectTrigger className="w-full">
						<SelectValue placeholder="不使用 SSH 密钥" />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="__none__">不使用 SSH 密钥</SelectItem>
						{sshKeys.map((key) => (
							<SelectItem key={key.id} value={key.id}>
								{key.name}
								{key.description ? ` (${key.description})` : ''}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
				<p className="text-xs text-muted-foreground">
					用于克隆私有仓库（可选）·{' '}
					<Link to="/console/ssh" className="underline hover:text-foreground">
						管理 SSH 密钥
					</Link>
				</p>
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
