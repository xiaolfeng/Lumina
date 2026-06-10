import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { DataTable } from '#/components/data-table'
import { DataTablePagination } from '#/components/data-table-pagination'
import { Tabs, TabsList, TabsTrigger } from '#/components/ui/tabs'
import { Skeleton } from '#/components/ui/skeleton'
import { useSessionList, useDeleteSession } from '#/hooks/useQaAdmin'
import { getSessionColumns } from '#/components/qa/columns'
import { DeleteSessionDialog } from '#/components/qa/delete-dialog'
import { staggerContainer, staggerItem, staggerItemLeft } from '#/lib/motion'

export const Route = createFileRoute('/console/qa')({
	component: QaPage,
})

function QaPage() {
	const [page, setPage] = useState(1)
	const [pageSize, setPageSize] = useState(10)
	const [statusFilter, setStatusFilter] = useState<string>('')
	const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

	const { data, isLoading } = useSessionList({ page, size: pageSize, status: statusFilter as any })
	const deleteMutation = useDeleteSession()

	const items = data?.data?.items ?? []
	const totalItems = data?.data?.total ?? 0
	const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

	const columns = getSessionColumns((session) => setDeleteTarget(session.id))

	return (
		<motion.div
			className="space-y-4"
			initial="hidden"
			animate="visible"
			variants={staggerContainer}
		>
			{/* 标题行 */}
			<motion.div
				className="relative flex items-center justify-between pl-1.5"
				variants={staggerItemLeft}
			>
				<div className="absolute -left-4 top-0 h-full w-1 rounded-r-full bg-gradient-to-b from-(--lagoon) to-(--palm)" />
				<div>
					<h1 className="text-2xl font-bold tracking-tight text-(--sea-ink)">问答管理</h1>
					<p className="text-(--sea-ink-soft)">管理 Q&A 会话和问答记录</p>
				</div>
			</motion.div>

			{/* 状态筛选 + 表格区域 */}
			<motion.div variants={staggerItem}>
				<Tabs value={statusFilter} onValueChange={(v) => { setStatusFilter(v); setPage(1) }}>
					<TabsList>
						<TabsTrigger value="">全部</TabsTrigger>
						<TabsTrigger value="active">活跃</TabsTrigger>
						<TabsTrigger value="expired">已过期</TabsTrigger>
						<TabsTrigger value="deleted">已删除</TabsTrigger>
					</TabsList>
				</Tabs>

				<div className="mt-4">
					{isLoading ? (
						<div className="space-y-3">
							<Skeleton className="h-8 w-full" />
							<Skeleton className="h-8 w-full" />
							<Skeleton className="h-8 w-full" />
						</div>
					) : (
						<>
							<DataTable columns={columns} data={items} />
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
				</div>
			</motion.div>

			<DeleteSessionDialog
				open={!!deleteTarget}
				onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
				session={items.find((s) => s.id === deleteTarget) ?? null}
				onConfirm={() => {
					if (deleteTarget) {
						deleteMutation.mutate(deleteTarget, {
							onSuccess: () => setDeleteTarget(null),
						})
					}
				}}
				isPending={deleteMutation.isPending}
			/>
		</motion.div>
	)
}
