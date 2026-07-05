import { useState } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Plus, ExternalLink, Trash2 } from 'lucide-react'
import { Button } from '#/components/ui/button'
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from '#/components/ui/table'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useRepoWikiConfigs, useDeleteRepoWikiConfig } from '#/hooks/useRepoWiki'
import { StatusBadge } from '#/components/repowiki/status-badge'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import type { RepoWikiConfigItem } from '#/lib/models/response/repowiki'
import { staggerContainer, staggerItem } from '#/lib/motion'
import { PageHeader } from '#/components/page-header'
import { SkeletonTable } from '#/components/skeleton-table'

export const Route = createFileRoute('/console/repowiki/')({
	component: RepoWikiListPage,
})

function RepoWikiListPage() {
	const [page, setPage] = useState(1)
	const [pageSize, setPageSize] = useState(10)
	const [deleteOpen, setDeleteOpen] = useState(false)
	const [selectedItem, setSelectedItem] = useState<RepoWikiConfigItem | null>(null)

	const { data, isLoading } = useRepoWikiConfigs({ page, size: pageSize })
	const deleteMutation = useDeleteRepoWikiConfig()

	const items = data?.data?.items ?? []
	const totalItems = data?.data?.total ?? 0
	const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

	return (
		<motion.div
			className="space-y-4"
			initial="hidden"
			animate="visible"
			variants={staggerContainer}
		>
			<PageHeader
				title="RepoWiki 配置"
				description="管理代码仓库的 Wiki 分析配置"
				action={
					<Link to="/console/repowiki/create">
						<Button className="bg-lagoon text-foam hover:bg-lagoon-deep">
							<Plus className="mr-2 size-4" />
							创建配置
						</Button>
					</Link>
				}
			/>

			{/* 表格区域 */}
			<motion.div variants={staggerItem}>
				{isLoading ? (
					<SkeletonTable />
				) : (
					<>
						<div className="rounded-md border border-border bg-card">
							<Table>
								<TableHeader>
									<TableRow>
										<TableHead>仓库名称</TableHead>
										<TableHead>仓库地址</TableHead>
										<TableHead>默认分支</TableHead>
										<TableHead>状态</TableHead>
										<TableHead>最后访问时间</TableHead>
										<TableHead>创建时间</TableHead>
										<TableHead className="text-right">操作</TableHead>
									</TableRow>
								</TableHeader>
								<TableBody>
									{items.length === 0 ? (
										<TableRow>
											<TableCell colSpan={7} className="h-24 text-center text-muted-foreground">
												暂无配置数据，点击「创建配置」添加第一个仓库
											</TableCell>
										</TableRow>
									) : (
										items.map((item) => (
											<TableRow key={item.id}>
												<TableCell>
													<span className="font-medium">{item.name}</span>
												</TableCell>
												<TableCell>
													<span className="max-w-[300px] line-clamp-1 font-mono text-xs">
														{item.repo_url}
													</span>
												</TableCell>
												<TableCell>
													<code className="rounded bg-muted px-1.5 py-0.5 text-xs">
														{item.default_branch}
													</code>
												</TableCell>
												<TableCell>
													<StatusBadge status={item.status} />
												</TableCell>
												<TableCell>
													<span className="text-sm text-muted-foreground">
														{item.last_accessed_at
															? new Date(item.last_accessed_at).toLocaleString('zh-CN')
															: '-'}
													</span>
												</TableCell>
												<TableCell>
													<span className="text-sm text-muted-foreground">
														{new Date(item.created_at).toLocaleDateString('zh-CN')}
													</span>
												</TableCell>
												<TableCell className="text-right">
													<div className="flex items-center justify-end gap-2">
														<Link to="/console/repowiki/$configId" params={{ configId: item.id }}>
															<Button variant="ghost" size="sm">
																<ExternalLink className="mr-1 size-3" />
																详情
															</Button>
														</Link>
														<Button
															variant="ghost"
															size="icon"
															onClick={() => {
																setSelectedItem(item)
																setDeleteOpen(true)
															}}
														>
															<Trash2 className="size-4 text-destructive" />
														</Button>
													</div>
												</TableCell>
											</TableRow>
										))
									)}
								</TableBody>
							</Table>
						</div>

						<DataTablePagination
							currentPage={page}
							totalPages={totalPages}
							totalItems={totalItems}
							pageSize={pageSize}
							onPageChange={setPage}
							onPageSizeChange={(size) => {
								setPageSize(size)
								setPage(1)
							}}
						/>
					</>
				)}
			</motion.div>

			{/* 删除确认对话框 */}
			<ConfirmDeleteDialog
				open={deleteOpen}
				onOpenChange={setDeleteOpen}
				description={`确定要删除配置「${selectedItem?.name ?? ''}」吗？此操作将同时删除所有已生成的 Wiki 版本数据，且不可撤销。`}
				onConfirm={() => {
					if (!selectedItem) return
					deleteMutation.mutate(selectedItem.id, {
						onSuccess: () => setDeleteOpen(false),
					})
				}}
				isPending={deleteMutation.isPending}
			/>
		</motion.div>
	)
}
