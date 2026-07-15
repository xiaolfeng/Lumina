import { useState } from 'react'
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { motion } from 'motion/react'
import {
	ArrowLeft,
	BookOpen,
	Plus,
	ExternalLink,
	Copy,
	Check,
	GitBranch,
	Globe,
	KeyRound,
	Lock,
	Clock,
	FileText,
	Webhook,
	Settings2,
} from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { Card, CardContent } from '@lumina/components/ui/card'
import { Separator } from '@lumina/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@lumina/components/ui/tabs'
import { PageHeader } from '#/components/page-header'
import { StatusBadge } from '#/components/repowiki/status-badge'
import { useRepoWikiConfigByProjectId } from '#/hooks/useRepoWiki'
import { AnalyzeButton, UpdateButton, VersionList } from '#/components/repowiki/version-list'
import { WebhookTab } from '#/components/repowiki/webhook-tab'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { buildWikiReaderUrl } from '#/lib/utils'
import { toast } from 'sonner'
import type { RepoWikiConfigItem } from '#/lib/models/response/repowiki'

export const Route = createFileRoute('/console/project/$projectId/repowiki/')({
	component: RepoWikiDetailPage,
})

function RepoWikiDetailPage() {
	const { projectId } = Route.useParams()
	const navigate = useNavigate()

	const { data, isLoading } = useRepoWikiConfigByProjectId(projectId)

	const config = data?.data

	if (isLoading) {
		return (
			<motion.div className="space-y-4" initial="hidden" animate="visible" variants={staggerContainer}>
				<PageHeader title="Wiki 管理" description="加载中..." />
				<div className="grid gap-4">
					{[1, 2, 3].map((i) => (
						<motion.div key={i} variants={staggerItem}>
							<Card className="border-border bg-card">
								<CardContent className="h-32 animate-pulse bg-muted/50" />
							</Card>
						</motion.div>
					))}
				</div>
			</motion.div>
		)
	}

	// 未创建配置 → 显示空状态
	if (!config) {
		return (
			<motion.div className="space-y-4" initial="hidden" animate="visible" variants={staggerContainer}>
				<PageHeader
					title="Wiki 管理"
					description="为该项目配置代码仓库 Wiki 分析"
					action={
						<Button variant="outline" onClick={() => navigate({ to: '/console/project' })}>
							<ArrowLeft className="mr-2 size-4" />
							返回项目
						</Button>
					}
				/>
				<motion.div variants={staggerItem}>
					<Card className="border-border bg-card">
						<CardContent className="flex flex-col items-center justify-center py-16">
							<BookOpen className="mb-4 size-12 text-muted-foreground/40" />
							<h3 className="mb-2 text-lg font-semibold">尚未创建 Wiki 配置</h3>
							<p className="mb-6 max-w-md text-center text-sm text-muted-foreground">
								添加仓库地址后，Lumina 将自动分析代码结构并生成结构化 Wiki 文档
							</p>
							<Link to="/console/project/$projectId/repowiki/create" params={{ projectId }}>
								<Button className="bg-lagoon text-foam hover:bg-lagoon-deep">
									<Plus className="mr-2 size-4" />
									创建 Wiki 配置
								</Button>
							</Link>
						</CardContent>
					</Card>
				</motion.div>
			</motion.div>
		)
	}

	const hasCompletedVersion = config.latest_version?.status === 'completed'

	// 已有配置 → 概览卡 + Tabs
	return (
		<motion.div className="space-y-5" initial="hidden" animate="visible" variants={staggerContainer}>
			{/* PageHeader：所有操作按钮集中排列 */}
			<motion.div variants={staggerItem}>
				<PageHeader
					title={`Wiki 管理 — ${config.name}`}
					action={
						<div className="flex flex-wrap items-center gap-2">
							<AnalyzeButton configId={config.id} />
							<UpdateButton configId={config.id} />
							{hasCompletedVersion && (
								<Button variant="outline" asChild className="gap-2">
									<a
										href={buildWikiReaderUrl(config.id)}
										target="_blank"
										rel="noopener noreferrer"
									>
										<ExternalLink className="size-4" />
										查看 Wiki
									</a>
								</Button>
							)}
							<Button variant="ghost" onClick={() => navigate({ to: '/console/project' })}>
								<ArrowLeft className="mr-2 size-4" />
								返回项目
							</Button>
						</div>
					}
				/>
			</motion.div>

			{/* 扁平概览条：无 Card 包裹，仓库地址 + 状态 + 关键指标一行展示 */}
			<motion.div variants={staggerItem}>
				<OverviewBar config={config} />
			</motion.div>

			{/* Tabs：版本管理 / Webhook / 配置详情 */}
			<motion.div variants={staggerItem}>
				<Tabs defaultValue="versions" className="gap-3">
					<TabsList>
						<TabsTrigger value="versions" className="gap-1.5">
							<FileText className="size-3.5" />
							版本管理
						</TabsTrigger>
						<TabsTrigger value="webhook" className="gap-1.5">
							<Webhook className="size-3.5" />
							Webhook
						</TabsTrigger>
						<TabsTrigger value="config" className="gap-1.5">
							<Settings2 className="size-3.5" />
							配置详情
						</TabsTrigger>
					</TabsList>

					<TabsContent value="versions" className="mt-0">
						<div className="rounded-lg border bg-card p-4">
							<VersionList configId={config.id} selectedVersionId={config.selected_version_id} />
						</div>
					</TabsContent>

					<TabsContent value="webhook" className="mt-0">
						<div className="rounded-lg border bg-card p-4">
							<WebhookTab configId={config.id} />
						</div>
					</TabsContent>

					<TabsContent value="config" className="mt-0">
						<div className="rounded-lg border bg-card p-4">
							<ConfigDetails config={config} />
						</div>
					</TabsContent>
				</Tabs>
			</motion.div>
		</motion.div>
	)
}

// ── 扁平概览条：仓库地址 + 状态 + 关键指标（无 Card 包裹） ──

function OverviewBar({ config }: { config: RepoWikiConfigItem }) {
	const [copied, setCopied] = useState(false)

	const handleCopyUrl = async () => {
		try {
			await navigator.clipboard.writeText(config.repo_url)
			setCopied(true)
			setTimeout(() => setCopied(false), 2000)
			toast.success('仓库地址已复制')
		} catch {
			toast.error('复制失败')
		}
	}

	const lastAccessed = config.last_accessed_at
		? new Date(config.last_accessed_at).toLocaleString('zh-CN')
		: null

	return (
		<div className="flex flex-wrap items-center gap-x-3 gap-y-2 text-sm">
			{/* 仓库地址（mono + 复制） */}
			<div className="flex items-center gap-1 min-w-0 max-w-full">
				<code className="truncate rounded border bg-muted px-2 py-0.5 font-mono text-xs text-foreground">
					{config.repo_url}
				</code>
				<Button
					variant="ghost"
					size="icon"
					onClick={handleCopyUrl}
					aria-label="复制仓库地址"
					className="size-7 shrink-0"
				>
					{copied ? (
						<Check className="size-3.5 text-emerald-500" />
					) : (
						<Copy className="size-3.5" />
					)}
				</Button>
			</div>

			<Separator orientation="vertical" className="h-4 shrink-0" />

			<StatusBadge status={config.status} />

			{config.selected_version_id && (
				<span className="flex items-center gap-1 text-xs text-muted-foreground">
					当前选中
					<code className="font-mono text-foreground">#{config.selected_version_id}</code>
				</span>
			)}

			{config.latest_version && config.latest_version.duration_ms > 0 && (
				<span className="flex items-center gap-1 text-xs text-muted-foreground">
					<Clock className="size-3" />
					{(config.latest_version.duration_ms / 1000).toFixed(1)}s
				</span>
			)}

			{lastAccessed && (
				<span className="text-xs text-muted-foreground">最后访问 {lastAccessed}</span>
			)}
		</div>
	)
}

// ── 配置详情：Definition List 风格（无边框） ──

function ConfigDetails({ config }: { config: RepoWikiConfigItem }) {
	const items: Array<{ label: string; value: React.ReactNode; mono?: boolean; icon?: React.ReactNode }> = [
		{
			label: '仓库名称',
			value: config.name,
			icon: <BookOpen className="size-4 text-muted-foreground" />,
		},
		{
			label: '仓库地址',
			value: config.repo_url,
			mono: true,
			icon: <Globe className="size-4 text-muted-foreground" />,
		},
		{
			label: '默认分支',
			value: config.default_branch,
			mono: true,
			icon: <GitBranch className="size-4 text-muted-foreground" />,
		},
		{
			label: '默认语言',
			value: config.default_language,
			icon: <Globe className="size-4 text-muted-foreground" />,
		},
		{
			label: 'SSH 密钥',
			value: config.ssh_key_id ? '已关联' : '未使用',
			icon: <KeyRound className="size-4 text-muted-foreground" />,
		},
		{
			label: 'Wiki 密码',
			value: config.has_password ? '已设置' : '未设置',
			icon: <Lock className="size-4 text-muted-foreground" />,
		},
		{
			label: '当前状态',
			value: <StatusBadge status={config.status} />,
		},
		{
			label: '选中版本',
			value: config.selected_version_id ? (
				<code className="font-mono text-xs">#{config.selected_version_id}</code>
			) : (
				'未选择'
			),
		},
		{
			label: '最后访问',
			value: config.last_accessed_at
				? new Date(config.last_accessed_at).toLocaleString('zh-CN')
				: '—',
		},
		{
			label: '创建时间',
			value: new Date(config.created_at).toLocaleString('zh-CN'),
		},
	]

	return (
		<dl className="grid grid-cols-1 gap-x-8 gap-y-3 sm:grid-cols-2">
			{items.map((item) => (
				<div key={item.label} className="flex items-start justify-between gap-3 py-1.5">
					<dt className="flex items-center gap-2 shrink-0 text-sm text-muted-foreground">
						{item.icon}
						{item.label}
					</dt>
					<dd
						className={`text-right text-sm font-medium ${
							item.mono ? 'font-mono text-xs break-all' : 'text-foreground'
						}`}
					>
						{item.value ?? '—'}
					</dd>
				</div>
			))}
		</dl>
	)
}
