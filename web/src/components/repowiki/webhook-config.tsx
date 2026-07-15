import { useState } from 'react'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from '@lumina/components/ui/dialog'
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from '@lumina/components/ui/select'
import { useWebhookConfig, useRegenerateWebhook } from '#/hooks/useWebhook'
import { Copy, Check, AlertTriangle, RefreshCw, Shield } from 'lucide-react'
import { toast } from 'sonner'

interface WebhookConfigProps {
	configId: string
}

/**
 * Webhook 凭据与提供商配置（section 风格，无外层 Card 包裹）。
 * 由 WebhookTab 组合使用。
 */
export function WebhookConfig({ configId }: WebhookConfigProps) {
	const { data, isLoading } = useWebhookConfig(configId)
	const regenerateMutation = useRegenerateWebhook(configId)
	const [confirmOpen, setConfirmOpen] = useState(false)
	const [resultOpen, setResultOpen] = useState(false)
	const [copied, setCopied] = useState<string | null>(null)
	const [selectedProvider, setSelectedProvider] = useState('')

	const config = data?.data

	const handleCopy = async (text: string, label: string) => {
		try {
			await navigator.clipboard.writeText(text)
			setCopied(label)
			setTimeout(() => setCopied(null), 2000)
			toast.success(`${label} 已复制到剪贴板`)
		} catch {
			toast.error('复制失败')
		}
	}

	const handleRegenerate = () => {
		regenerateMutation.mutate(undefined, {
			onSuccess: () => {
				setConfirmOpen(false)
				setResultOpen(true)
			},
			onError: (error: Error) => {
				toast.error(error.message || '重新生成失败')
			},
		})
	}

	if (isLoading) {
		return (
			<div className="space-y-3">
				<div className="h-9 animate-pulse rounded bg-muted" />
				<div className="h-9 animate-pulse rounded bg-muted" />
				<div className="h-9 animate-pulse rounded bg-muted" />
			</div>
		)
	}

	if (!config) {
		return <p className="text-sm text-muted-foreground">加载 Webhook 配置失败</p>
	}

	const providers = Object.entries(config.provider_guide ?? {})
	const guideText =
		selectedProvider && config.provider_guide ? config.provider_guide[selectedProvider] : ''

	return (
		<>
			<div className="space-y-5">
				{/* Webhook URL */}
				<div className="space-y-1.5">
					<Label className="text-xs text-muted-foreground">Webhook URL</Label>
					<div className="flex gap-2">
						<Input value={config.url} readOnly className="font-mono text-sm" />
						<Button
							variant="outline"
							size="icon"
							onClick={() => handleCopy(config.url, 'URL')}
							aria-label="复制 Webhook URL"
						>
							{copied === 'URL' ? (
								<Check className="size-4 text-emerald-500" />
							) : (
								<Copy className="size-4" />
							)}
						</Button>
					</div>
				</div>

				{/* Token */}
				<div className="space-y-1.5">
					<Label className="text-xs text-muted-foreground">Token</Label>
					<div className="flex items-center gap-2">
						<code className="flex-1 rounded border bg-muted px-2.5 py-1.5 font-mono text-sm truncate">
							{config.token}
						</code>
						<Button
							variant="outline"
							size="icon"
							onClick={() => handleCopy(config.token, 'Token')}
							aria-label="复制 Token"
						>
							{copied === 'Token' ? (
								<Check className="size-4 text-emerald-500" />
							) : (
								<Copy className="size-4" />
							)}
						</Button>
					</div>
				</div>

				{/* Secret status */}
				<div className="flex items-center gap-2 text-sm">
					<Shield
						className={`size-4 ${config.has_secret ? 'text-emerald-500' : 'text-muted-foreground'}`}
					/>
					<span className={config.has_secret ? 'text-foreground' : 'text-muted-foreground'}>
						{config.has_secret ? '已设置 Secret（HMAC 签名校验已启用）' : '未设置 Secret（建议配置以启用签名校验）'}
					</span>
				</div>

				{/* Provider Guide */}
				{providers.length > 0 && (
					<div className="space-y-1.5">
						<Label className="text-xs text-muted-foreground">提供商配置指南</Label>
						<Select value={selectedProvider} onValueChange={setSelectedProvider}>
							<SelectTrigger>
								<SelectValue placeholder="选择 Git 提供商查看配置示例" />
							</SelectTrigger>
							<SelectContent>
								{providers.map(([key]) => (
									<SelectItem key={key} value={key}>
										{key}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
						{guideText && (
							<div className="rounded-md border bg-muted/50 p-3">
								<pre className="whitespace-pre-wrap text-xs leading-relaxed font-mono">{guideText}</pre>
							</div>
						)}
					</div>
				)}

				{/* Regenerate */}
				<div className="pt-1">
					<Button variant="outline" onClick={() => setConfirmOpen(true)} className="gap-2">
						<RefreshCw className="size-4" />
						重新生成凭据
					</Button>
				</div>
			</div>

			{/* Confirm dialog */}
			<Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle className="flex items-center gap-2">
							<AlertTriangle className="size-5 text-amber-500" />
							确认重新生成？
						</DialogTitle>
						<DialogDescription>
							重新生成后，旧的 Token 和 Secret 将立即失效。所有使用旧凭据的 Webhook 请求都会被拒绝。
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<Button variant="outline" onClick={() => setConfirmOpen(false)}>
							取消
						</Button>
						<Button onClick={handleRegenerate} disabled={regenerateMutation.isPending} className="gap-2">
							{regenerateMutation.isPending ? (
								<>
									<RefreshCw className="size-4 animate-spin" />
									生成中...
								</>
							) : (
								'确认重新生成'
							)}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			{/* Result dialog */}
			<Dialog open={resultOpen} onOpenChange={setResultOpen}>
				<DialogContent className="max-w-lg">
					<DialogHeader>
						<DialogTitle>凭据已重新生成</DialogTitle>
						<DialogDescription>
							请立即保存以下信息，关闭后将无法再次查看完整凭据。
						</DialogDescription>
					</DialogHeader>
					<div className="space-y-4">
						<div className="space-y-1.5">
							<Label className="text-xs text-muted-foreground">Token</Label>
							<div className="flex gap-2">
								<Input
									value={regenerateMutation.data?.data?.token ?? ''}
									readOnly
									className="font-mono text-sm"
								/>
								<Button
									variant="outline"
									size="icon"
									onClick={() => handleCopy(regenerateMutation.data?.data?.token ?? '', '新 Token')}
									aria-label="复制新 Token"
								>
									{copied === '新 Token' ? (
										<Check className="size-4 text-emerald-500" />
									) : (
										<Copy className="size-4" />
									)}
								</Button>
							</div>
						</div>
						<div className="space-y-1.5">
							<Label className="text-xs text-muted-foreground">Secret</Label>
							<div className="flex gap-2">
								<Input
									value={regenerateMutation.data?.data?.secret ?? ''}
									readOnly
									className="font-mono text-sm"
								/>
								<Button
									variant="outline"
									size="icon"
									onClick={() => handleCopy(regenerateMutation.data?.data?.secret ?? '', '新 Secret')}
									aria-label="复制新 Secret"
								>
									{copied === '新 Secret' ? (
										<Check className="size-4 text-emerald-500" />
									) : (
										<Copy className="size-4" />
									)}
								</Button>
							</div>
						</div>
						<div className="space-y-1.5">
							<Label className="text-xs text-muted-foreground">Webhook URL</Label>
							<div className="flex gap-2">
								<Input
									value={regenerateMutation.data?.data?.url ?? ''}
									readOnly
									className="font-mono text-sm"
								/>
								<Button
									variant="outline"
									size="icon"
									onClick={() => handleCopy(regenerateMutation.data?.data?.url ?? '', '新 URL')}
									aria-label="复制新 URL"
								>
									{copied === '新 URL' ? (
										<Check className="size-4 text-emerald-500" />
									) : (
										<Copy className="size-4" />
									)}
								</Button>
							</div>
						</div>
					</div>
					<DialogFooter>
						<Button onClick={() => setResultOpen(false)}>关闭</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</>
	)
}
