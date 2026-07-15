import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Plus } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { DataTable } from '#/components/data-table'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useSshKeyList, useDeleteSshKey } from '#/hooks/useSshKey'
import { getColumns } from '#/components/ssh/columns'
import { CreateDialog } from '#/components/ssh/create-dialog'
import { EditDialog } from '#/components/ssh/edit-dialog'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { SkeletonTable } from '#/components/skeleton-table'
import type { SshKeyItem } from '#/lib/models/response/ssh'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { PageHeader } from '#/components/page-header'

export const Route = createFileRoute('/console/ssh')({
	component: SshPage,
})

function SshPage() {
	const [page, setPage] = useState(1)
	const [pageSize, setPageSize] = useState(10)
	const [createOpen, setCreateOpen] = useState(false)
	const [editOpen, setEditOpen] = useState(false)
	const [deleteOpen, setDeleteOpen] = useState(false)
	const [selectedItem, setSelectedItem] = useState<SshKeyItem | null>(null)

	const { data, isLoading } = useSshKeyList({ page, size: pageSize })
	const deleteMutation = useDeleteSshKey()

	const items = data?.data?.items ?? []
	const totalItems = data?.data?.total ?? 0
	const totalPages = pageSize > 0 ? Math.ceil(totalItems / pageSize) : 1

	const columns = getColumns({
		onEdit: (item) => {
			setSelectedItem(item)
			setEditOpen(true)
		},
		onDelete: (item) => {
			setSelectedItem(item)
			setDeleteOpen(true)
		},
	})

	return (
		<motion.div
			className="space-y-4"
			initial="hidden"
			animate="visible"
			variants={staggerContainer}
		>
			<PageHeader
				title="SSH 密钥管理"
				description="管理用于 Git 仓库访问的 SSH 密钥"
				action={
					<Button
						onClick={() => setCreateOpen(true)}
						className="bg-lagoon text-foam hover:bg-lagoon-deep"
					>
						<Plus className="mr-2 size-4" />
						添加密钥
					</Button>
				}
			/>

			{/* 表格区域 */}
			<motion.div variants={staggerItem}>
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
			</motion.div>

			<CreateDialog open={createOpen} onOpenChange={setCreateOpen} />
			<EditDialog
				open={editOpen}
				onOpenChange={setEditOpen}
				item={selectedItem}
			/>
			<ConfirmDeleteDialog
				open={deleteOpen}
				onOpenChange={setDeleteOpen}
				title="确定要删除此 SSH 密钥吗？"
				description={`此操作不可撤销。删除后，使用「${selectedItem?.name ?? ''}」密钥的 Git 访问将立即失效。`}
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
