import { useState } from 'react'
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
import { Textarea } from '#/components/ui/textarea'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs'
import { useCreateSshKey } from '#/hooks/useSshKey'
import type { CreateSshKeyResponse } from '#/lib/models/response/ssh'
import { Copy, Check, Download, KeyRound } from 'lucide-react'
import { toast } from 'sonner'
import { getSshPublicKey } from '#/lib/apis/ssh'

interface CreateDialogProps {
	open: boolean
	onOpenChange: (open: boolean) => void
}

export function CreateDialog({ open, onOpenChange }: CreateDialogProps) {
	const [activeTab, setActiveTab] = useState<'generate' | 'import'>('generate')
	const [name, setName] = useState('')
	const [description, setDescription] = useState('')
	const [privateKey, setPrivateKey] = useState('')
	const [result, setResult] = useState<CreateSshKeyResponse | null>(null)
	const [copied, setCopied] = useState(false)

	const createMutation = useCreateSshKey()

	const handleSubmit = () => {
		if (!name.trim()) return

		if (activeTab === 'generate') {
			createMutation.mutate(
				{ source: 'generated', name: name.trim(), description: description.trim() || undefined },
				{
					onSuccess: (res) => {
						if (res.data) setResult(res.data)
					},
				},
			)
		} else {
			if (!privateKey.trim()) return
			createMutation.mutate(
				{
					source: 'imported',
					name: name.trim(),
					description: description.trim() || undefined,
					private_key: privateKey.trim(),
				},
				{
					onSuccess: (res) => {
						if (res.data) setResult(res.data)
					},
				},
			)
		}
	}

	const handleCopy = async () => {
		if (!result?.public_key) return
		await navigator.clipboard.writeText(result.public_key)
		setCopied(true)
		toast.success('已复制到剪贴板')
		setTimeout(() => setCopied(false), 2000)
	}

	const handleDownload = () => {
		if (!result) return
		getSshPublicKey(result.id)
			.then((blob) => {
				const url = URL.createObjectURL(blob)
				const a = document.createElement('a')
				a.href = url
				a.download = `${result.name}-id_${result.key_type}.pub`
				document.body.appendChild(a)
				a.click()
				document.body.removeChild(a)
				URL.revokeObjectURL(url)
			})
			.catch(() => toast.error('下载失败'))
	}

	const handleClose = () => {
		setName('')
		setDescription('')
		setPrivateKey('')
		setResult(null)
		setCopied(false)
		setActiveTab('generate')
		onOpenChange(false)
	}

	const isFormValid = name.trim() && (activeTab === 'generate' || privateKey.trim())

	return (
		<Dialog open={open} onOpenChange={result ? undefined : onOpenChange}>
			<DialogContent
				className="sm:max-w-lg"
				onPointerDownOutside={result ? (e) => e.preventDefault() : undefined}
			>
				{!result ? (
					<>
						<DialogHeader>
							<DialogTitle>添加 SSH 密钥</DialogTitle>
							<DialogDescription>
								生成新密钥对或导入已有 PEM 格式私钥
							</DialogDescription>
						</DialogHeader>

						<Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'generate' | 'import')}>
							<TabsList className="grid w-full grid-cols-2">
								<TabsTrigger value="generate">生成密钥</TabsTrigger>
								<TabsTrigger value="import">导入密钥</TabsTrigger>
							</TabsList>

							<TabsContent value="generate" className="grid gap-4 py-4">
								<div className="grid gap-2">
									<Label htmlFor="gen-name">名称 *</Label>
									<Input
										id="gen-name"
										value={name}
										onChange={(e) => setName(e.target.value)}
										placeholder="输入密钥名称，如：我的 Mac"
									/>
								</div>
								<div className="grid gap-2">
									<Label htmlFor="gen-desc">描述</Label>
									<Input
										id="gen-desc"
										value={description}
										onChange={(e) => setDescription(e.target.value)}
										placeholder="输入描述（可选）"
									/>
								</div>
							</TabsContent>

							<TabsContent value="import" className="grid gap-4 py-4">
								<div className="grid gap-2">
									<Label htmlFor="imp-name">名称 *</Label>
									<Input
										id="imp-name"
										value={name}
										onChange={(e) => setName(e.target.value)}
										placeholder="输入密钥名称"
									/>
								</div>
								<div className="grid gap-2">
									<Label htmlFor="imp-desc">描述</Label>
									<Input
										id="imp-desc"
										value={description}
										onChange={(e) => setDescription(e.target.value)}
										placeholder="输入描述（可选）"
									/>
								</div>
								<div className="grid gap-2">
									<Label htmlFor="imp-key">PEM 私钥 *</Label>
									<Textarea
										id="imp-key"
										value={privateKey}
										onChange={(e) => setPrivateKey(e.target.value)}
										placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
										rows={6}
										className="font-mono text-xs"
									/>
								</div>
							</TabsContent>
						</Tabs>

						<DialogFooter>
							<Button variant="outline" onClick={handleClose}>
								取消
							</Button>
							<Button
								onClick={handleSubmit}
								disabled={!isFormValid || createMutation.isPending}
								className="bg-lagoon text-foam hover:bg-lagoon-deep"
							>
								{createMutation.isPending ? '创建中...' : '创建'}
							</Button>
						</DialogFooter>
					</>
				) : (
					<>
						<DialogHeader>
							<DialogTitle>密钥已创建</DialogTitle>
							<DialogDescription>
								请妥善保存公钥内容。私钥不会在此展示。
							</DialogDescription>
						</DialogHeader>
						<div className="space-y-4 py-4">
							<div className="flex items-center gap-2 text-sm text-muted-foreground">
								<KeyRound className="size-4" />
								<span>指纹：{result.fingerprint}</span>
							</div>
							<div className="rounded-md bg-muted p-4">
								<pre className="overflow-x-auto break-all text-sm font-mono">
									{result.public_key}
								</pre>
							</div>
							<div className="flex gap-2">
								<Button variant="outline" className="flex-1" onClick={handleCopy}>
									{copied ? <Check className="mr-2 size-4" /> : <Copy className="mr-2 size-4" />}
									{copied ? '已复制' : '复制公钥'}
								</Button>
								<Button variant="outline" className="flex-1" onClick={handleDownload}>
									<Download className="mr-2 size-4" />
									下载公钥
								</Button>
							</div>
						</div>
						<DialogFooter>
							<Button onClick={handleClose} className="w-full">
								我已保存
							</Button>
						</DialogFooter>
					</>
				)}
			</DialogContent>
		</Dialog>
	)
}
