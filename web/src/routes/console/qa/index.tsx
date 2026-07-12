import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { DataTable } from '#/components/data-table'
import { DataTablePagination } from '#/components/data-table-pagination'
import { Tabs, TabsList, TabsTrigger } from '#/components/ui/tabs'
import { useSessionList, useDeleteSession } from '#/hooks/useQaAdmin'
import { getSessionColumns } from '#/components/qa/columns'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { staggerContainer, staggerItem } from '#/lib/motion'
import { PageHeader } from '#/components/page-header'
import { SkeletonTable } from '#/components/skeleton-table'

export const Route = createFileRoute('/console/qa/')({
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
			<PageHeader title="问答管理" description="管理 Q&A 会话和问答记录" />

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
					<SkeletonTable />
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

			<ConfirmDeleteDialog
				open={!!deleteTarget}
				onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
				title="删除会话"
				description={`确定要删除会话「${items.find((s) => s.id === deleteTarget)?.title ?? ''}」吗？删除后所有问答数据将不可恢复。`}
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
