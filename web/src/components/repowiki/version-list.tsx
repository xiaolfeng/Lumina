import { useState } from 'react'
import { Button } from '#/components/ui/button'
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from '#/components/ui/table'
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from '#/components/ui/alert-dialog'
import { SkeletonTable } from '#/components/skeleton-table'
import { StatusBadge } from '#/components/repowiki/status-badge'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useRepoWikiVersions, useRepoWikiAnalyze, useRepoWikiUpdate, ACTIVE_STATUSES } from '#/hooks/useRepoWiki'
import { buildWikiReaderUrl } from '#/lib/utils'
import { Play, RefreshCw, ChevronDown, ChevronRight, Clock, Loader2, ExternalLink } from 'lucide-react'

// ── 阶段配置 ──

const STAGE_LABELS: Record<string, string> = {
	scan: '扫描',
	pass1: '分析 Pass 1',
	pass2: '分析 Pass 2',
	pass3: '分析 Pass 3',
	pass4: '分析 Pass 4',
	assemble: '组装',
}

const STAGE_ORDER = ['scan', 'pass1', 'pass2', 'pass3', 'pass4', 'assemble']

// ── 阶段进度指示器 ──

function StageProgress({ currentStage, durationMs }: { currentStage?: string; durationMs?: number }) {
	if (!currentStage && !durationMs) return null

	const currentIndex = currentStage ? STAGE_ORDER.indexOf(currentStage) : -1
	const totalStages = STAGE_ORDER.length
	const progress = currentIndex >= 0 ? ((currentIndex + 1) / totalStages) * 100 : 0

	return (
		<div className="flex items-center gap-3 text-xs text-muted-foreground">
			{currentStage && (
				<div className="flex items-center gap-1.5">
					<span className="font-medium text-foreground">当前阶段:</span>
					<span className="text-lagoon font-medium">{STAGE_LABELS[currentStage] || currentStage}</span>
				</div>
			)}
			{durationMs !== undefined && durationMs > 0 && (
				<div className="flex items-center gap-1">
					<Clock className="size-3" />
					<span>{(durationMs / 1000).toFixed(1)}s</span>
				</div>
			)}
			{currentIndex >= 0 && (
				<div className="flex items-center gap-2 flex-1 max-w-[120px]">
					<div className="h-1.5 flex-1 rounded-full bg-muted overflow-hidden">
						<div
							className="h-full bg-lagoon rounded-full transition-all duration-500 ease-out"
							style={{ width: `${progress}%` }}
						/>
					</div>
					<span className="text-[10px] tabular-nums">{Math.round(progress)}%</span>
				</div>
			)}
		</div>
	)
}

// ── 错误信息（可折叠） ──

function ErrorMessage({ message }: { message: string }) {
	const [expanded, setExpanded] = useState(false)

	return (
		<div className="mt-1.5">
			<button
				type="button"
				onClick={() => setExpanded(!expanded)}
				className="flex items-center gap-1 text-xs text-destructive hover:text-destructive/80 transition-colors cursor-pointer"
			>
				{expanded ? <ChevronDown className="size-3" /> : <ChevronRight className="size-3" />}
				<span>{expanded ? '收起错误详情' : '展开错误详情'}</span>
			</button>
			{expanded && (
				<div className="mt-1.5 p-2.5 rounded-md bg-destructive/10 border border-destructive/20 text-xs text-destructive font-mono leading-relaxed break-all max-h-[120px] overflow-y-auto">
					{message}
				</div>
			)}
		</div>
	)
}

// ── 分析按钮 + 确认对话框 ──

export function AnalyzeButton({ configId }: { configId: string }) {
	const [dialogOpen, setDialogOpen] = useState(false)
	const analyzeMutation = useRepoWikiAnalyze(configId)

	const handleConfirm = () => {
		analyzeMutation.mutate(undefined)
		setDialogOpen(false)
	}

	return (
		<>
			<Button
				onClick={() => setDialogOpen(true)}
				disabled={analyzeMutation.isPending}
				className="gap-2"
			>
				<Play className="size-4" />
				开始分析
			</Button>

			<AlertDialog open={dialogOpen} onOpenChange={setDialogOpen}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>确认启动分析？</AlertDialogTitle>
						<AlertDialogDescription>
							这将触发一次完整的仓库克隆和分析流程，可能需要较长时间。确定要继续吗？
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel disabled={analyzeMutation.isPending}>取消</AlertDialogCancel>
						<AlertDialogAction onClick={handleConfirm} disabled={analyzeMutation.isPending}>
							{analyzeMutation.isPending ? (
								<span className="flex items-center gap-2">
									<Loader2 className="size-4 animate-spin" />
									启动中...
								</span>
							) : (
								'确认启动'
							)}
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</>
	)
}

// ── 增量更新按钮 + 确认对话框 ──

export function UpdateButton({ configId }: { configId: string }) {
	const [dialogOpen, setDialogOpen] = useState(false)
	const updateMutation = useRepoWikiUpdate(configId)

	const handleConfirm = () => {
		updateMutation.mutate()
		setDialogOpen(false)
	}

	return (
		<>
			<Button
				variant="outline"
				onClick={() => setDialogOpen(true)}
				disabled={updateMutation.isPending}
				className="gap-2"
			>
				<RefreshCw className={`size-4 ${updateMutation.isPending ? 'animate-spin' : ''}`} />
				增量更新
			</Button>

			<AlertDialog open={dialogOpen} onOpenChange={setDialogOpen}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>确认增量更新？</AlertDialogTitle>
						<AlertDialogDescription>
							这将拉取最新代码并执行增量分析。相比全量分析更快，但仅适用于已有历史版本的仓库。
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel disabled={updateMutation.isPending}>取消</AlertDialogCancel>
						<AlertDialogAction onClick={handleConfirm} disabled={updateMutation.isPending}>
							{updateMutation.isPending ? (
								<span className="flex items-center gap-2">
									<Loader2 className="size-4 animate-spin" />
									启动中...
								</span>
							) : (
								'确认更新'
							)}
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</>
	)
}

// ── 版本列表主组件 ──

interface VersionListProps {
	configId: string
}

export function VersionList({ configId }: VersionListProps) {
	const [page, setPage] = useState(1)
	const [pageSize, setPageSize] = useState(20)
	const { data, isLoading, isError, error } = useRepoWikiVersions(configId, page, pageSize)

	if (isLoading) return <SkeletonTable rows={5} />

	if (isError) {
		return (
			<div className="text-center py-12 text-destructive">
				<p className="font-medium">加载版本列表失败</p>
				<p className="text-sm mt-1">{error.message || '请稍后重试'}</p>
			</div>
		)
	}

	const versions = data?.items ?? []
	const total = data?.total ?? 0
	const totalPages = pageSize > 0 ? Math.ceil(total / pageSize) : 1

	if (versions.length === 0) {
		return (
			<div className="text-center py-12 text-muted-foreground">
				<p className="text-lg font-medium mb-1">暂无版本记录</p>
				<p className="text-sm">点击「开始分析」创建第一个版本</p>
			</div>
		)
	}

	return (
		<div className="space-y-4">
			<div className="rounded-lg border bg-card overflow-hidden">
				<Table>
					<TableHeader>
						<TableRow className="bg-muted/50 hover:bg-muted/50">
							<TableHead className="w-[80px]">版本号</TableHead>
							<TableHead className="w-[120px]">提交哈希</TableHead>
							<TableHead className="w-[110px]">状态</TableHead>
							<TableHead className="min-w-[280px]">进度</TableHead>
							<TableHead className="w-[160px]">创建时间</TableHead>
							<TableHead className="w-[60px]" />
						</TableRow>
					</TableHeader>
					<TableBody>
						{versions.map((version) => {
							const isActive = ACTIVE_STATUSES.includes(
								version.status as (typeof ACTIVE_STATUSES)[number],
							)

							return (
								<TableRow
									key={version.id}
									className={`group transition-colors ${
										isActive
											? 'bg-lagoon/5 hover:bg-lagoon/10 border-l-2 border-l-lagoon'
											: ''
									}`}
								>
								<TableCell className="font-mono text-sm font-medium">
									{version.id}
								</TableCell>
									<TableCell>
										<code className="text-xs bg-muted px-1.5 py-0.5 rounded font-mono">
											{version.commit_hash.slice(0, 7)}
										</code>
									</TableCell>
									<TableCell>
										<StatusBadge status={version.status} />
									</TableCell>
									<TableCell>
										<div className="space-y-1">
											<StageProgress
												currentStage={version.current_stage}
												durationMs={version.duration_ms}
											/>
											{version.status === 'failed' && version.error_msg && (
												<ErrorMessage message={version.error_msg} />
											)}
										</div>
									</TableCell>
									<TableCell className="text-sm text-muted-foreground">
										{new Date(version.created_at).toLocaleString()}
									</TableCell>
								<TableCell className="text-right">
									{version.status === 'completed' && (
										<Button variant="ghost" size="sm" asChild>
											<a href={buildWikiReaderUrl(configId)} target="_blank" rel="noopener noreferrer">
												<ExternalLink className="size-3.5" />
												查看 Wiki
											</a>
										</Button>
									)}
								</TableCell>
								</TableRow>
							)
						})}
					</TableBody>
				</Table>
			</div>

			<DataTablePagination
				currentPage={page}
				totalPages={totalPages}
				totalItems={total}
				pageSize={pageSize}
				onPageChange={(newPage) => setPage(newPage)}
				onPageSizeChange={(newSize) => {
					setPageSize(newSize)
					setPage(1)
				}}
			/>
		</div>
	)
}
