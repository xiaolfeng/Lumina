import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ArrowLeft, BookOpen, Plus, ExternalLink } from 'lucide-react'
import { Button } from '#/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { Badge } from '#/components/ui/badge'
import { PageHeader } from '#/components/page-header'
import { useRepoWikiConfigByProjectId } from '#/hooks/useRepoWiki'
import { AnalyzeButton, UpdateButton, VersionList } from '#/components/repowiki/version-list'
import { VersionSwitcher } from '#/components/repowiki/version-switcher'
import { staggerContainer, staggerItem } from '#/lib/motion'
import { buildWikiReaderUrl } from '#/lib/utils'

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

	// 已有配置 → 展示详情（当前为静态占位，后续接入真实数据）
	return (
		<motion.div className="space-y-4" initial="hidden" animate="visible" variants={staggerContainer}>
			<PageHeader
				title={`Wiki 管理 — ${config.name}`}
				description={config.repo_url}
				action={
					<Button variant="outline" onClick={() => navigate({ to: '/console/project' })}>
						<ArrowLeft className="mr-2 size-4" />
						返回项目
					</Button>
				}
			/>

			{/* 配置信息卡 */}
			<motion.div variants={staggerItem}>
				<Card className="border-border bg-card">
					<CardHeader>
						<CardTitle>配置信息</CardTitle>
						<CardDescription>仓库基本配置与状态概览</CardDescription>
					</CardHeader>
					<CardContent>
						<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
							<InfoField label="仓库名称" value={config.name} />
							<InfoField label="仓库地址" value={config.repo_url} mono />
							<InfoField label="默认分支" value={config.default_branch} mono />
							<InfoField label="默认语言" value={config.default_language} />
							<InfoField label="状态" value={
								<Badge variant="outline">{config.status}</Badge>
							} />
							<InfoField label="SSH 密钥" value={config.ssh_key_id ? '已关联' : '未使用'} />
							<InfoField label="Wiki 密码" value={config.has_password ? '已设置' : '未设置'} />
							<InfoField
								label="最后访问"
								value={config.last_accessed_at
									? new Date(config.last_accessed_at).toLocaleString('zh-CN')
									: '-'}
							/>
							<InfoField
								label="创建时间"
								value={new Date(config.created_at).toLocaleDateString('zh-CN')}
							/>
						</div>
					</CardContent>
				</Card>
			</motion.div>

			{/* 操作按钮区 */}
			<motion.div variants={staggerItem} className="flex flex-wrap gap-3">
				<AnalyzeButton configId={config.id} />
				<UpdateButton configId={config.id} />
		{config.latest_version?.status === 'completed' && (
			<Button variant="outline" asChild>
				<a href={buildWikiReaderUrl(config.id)} target="_blank" rel="noopener noreferrer">
					<ExternalLink className="mr-2 size-4" />
					查看 Wiki
				</a>
			</Button>
		)}
			</motion.div>

			{/* 选中版本切换 */}
			<motion.div variants={staggerItem}>
				<VersionSwitcher configId={config.id} selectedVersionId={config.selected_version_id} />
			</motion.div>

			{/* 版本列表 */}
			<motion.div variants={staggerItem}>
				<Card className="border-border bg-card">
					<CardHeader>
						<CardTitle>版本历史</CardTitle>
						<CardDescription>Wiki 分析版本记录</CardDescription>
					</CardHeader>
					<CardContent>
						<VersionList configId={config.id} />
					</CardContent>
				</Card>
			</motion.div>
		</motion.div>
	)
}

// ── 内部组件：信息字段展示 ──

function InfoField({
	label,
	value,
	mono,
}: {
	label: string
	value: React.ReactNode
	mono?: boolean
}) {
	return (
		<div className="rounded-lg border border-border/50 bg-muted/30 p-3">
			<p className="mb-1 text-xs font-medium text-muted-foreground">{label}</p>
			<p className={mono ? 'font-mono text-xs break-all' : 'text-sm font-medium'}>
				{value ?? '-'}
			</p>
		</div>
	)
}
