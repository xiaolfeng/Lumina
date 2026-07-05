import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ArrowLeft } from 'lucide-react'
import { Button } from '#/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '#/components/ui/card'
import { SkeletonTable } from '#/components/skeleton-table'
import { useRepoWikiConfig } from '#/hooks/useRepoWiki'
import { VersionList, AnalyzeButton, UpdateButton } from '#/components/repowiki/version-list'
import { StatusBadge } from '#/components/repowiki/status-badge'
import { staggerContainer, staggerItem } from '#/lib/motion'
import { PageHeader } from '#/components/page-header'

export const Route = createFileRoute('/console/repowiki/$configId')({
	component: ConfigDetailPage,
})

function ConfigDetailPage() {
	const { configId } = Route.useParams()
	const navigate = useNavigate()

	const { data: configData, isLoading: configLoading } = useRepoWikiConfig(configId)
	const config = configData?.data

	if (configLoading) {
		return (
			<motion.div className="space-y-4" initial="hidden" animate="visible" variants={staggerContainer}>
				<PageHeader title="配置详情" />
				<motion.div variants={staggerItem}>
					<SkeletonTable />
				</motion.div>
			</motion.div>
		)
	}

	if (!config) {
		return (
			<div className="flex h-[400px] items-center justify-center">
				<p className="text-muted-foreground">配置不存在或已被删除</p>
			</div>
		)
	}

	return (
		<motion.div className="space-y-4" initial="hidden" animate="visible" variants={staggerContainer}>
			<PageHeader
				title={config.name}
				description={`仓库地址：${config.repo_url}`}
				action={
					<div className="flex gap-2">
						<Button variant="outline" onClick={() => navigate({ to: '/console/repowiki' })}>
							<ArrowLeft className="mr-2 size-4" />
							返回列表
						</Button>
					</div>
				}
			/>

			<motion.div variants={staggerItem}>
				<Card className="border-border bg-card">
					<CardHeader>
						<CardTitle>基本信息</CardTitle>
						<CardDescription>仓库的核心配置参数</CardDescription>
					</CardHeader>
					<CardContent>
						<div className="grid grid-cols-2 gap-4 md:grid-cols-3">
							<InfoItem label="状态" value={<StatusBadge status={config.status} />} />
							<InfoItem
								label="默认分支"
								value={
									<code className="rounded bg-muted px-1.5 py-0.5 text-xs">
										{config.default_branch}
									</code>
								}
							/>
							<InfoItem
								label="默认语言"
								value={
									<code className="rounded bg-muted px-1.5 py-0.5 text-xs">
										{config.default_language}
									</code>
								}
							/>
						<InfoItem
							label="最后访问时间"
							value={
								config.last_accessed_at
									? new Date(config.last_accessed_at).toLocaleString('zh-CN')
									: '-'
							}
						/>
							<InfoItem label="创建时间" value={new Date(config.created_at).toLocaleString('zh-CN')} />
							<InfoItem label="更新时间" value={new Date(config.updated_at).toLocaleString('zh-CN')} />
						</div>

						<div className="mt-6 flex gap-3 border-t pt-4">
							<AnalyzeButton configId={configId} />
							<UpdateButton configId={configId} />
						</div>
					</CardContent>
				</Card>
			</motion.div>

			<motion.div variants={staggerItem}>
				<Card className="border-border bg-card">
					<CardHeader>
						<CardTitle>版本历史</CardTitle>
						<CardDescription>已生成的 Wiki 版本记录（分析中自动刷新）</CardDescription>
					</CardHeader>
					<CardContent>
						<VersionList configId={configId} />
					</CardContent>
				</Card>
			</motion.div>
		</motion.div>
	)
}

function InfoItem({ label, value }: { label: string; value: React.ReactNode }) {
	return (
		<div className="space-y-1">
			<p className="text-sm font-medium text-muted-foreground">{label}</p>
			<div className="text-sm">{value}</div>
		</div>
	)
}
