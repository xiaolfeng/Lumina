import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Plus } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { DataTable } from '#/components/data-table'
import { DataTablePagination } from '#/components/data-table-pagination'
import { PageHeader } from '#/components/page-header'
import { SkeletonTable } from '#/components/skeleton-table'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { CreateWorkspaceDialog } from '#/components/workspace/create-dialog'
import { EditWorkspaceDialog } from '#/components/workspace/edit-dialog'
import { getWorkspaceColumns } from '#/components/workspace/columns'
import { useWorkspaceList, useDeleteWorkspace } from '#/hooks/useWorkspace'
import type { WorkspaceItem } from '#/lib/models/response/workspace'
import { staggerContainer, staggerItem } from '@lumina/components/motion'

export const Route = createFileRoute('/console/workspace')({
  staticData: { crumb: '空间管理' },
  component: WorkspacePage,
})

function WorkspacePage() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [selectedItem, setSelectedItem] = useState<WorkspaceItem | null>(null)

  const { data, isLoading } = useWorkspaceList({ page, size: pageSize })
  const deleteMutation = useDeleteWorkspace()
  const items = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

  const columns = getWorkspaceColumns({
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
        title="空间管理"
        description="用空间拆开生活与工作。删除非默认空间时，其中的项目会移到默认空间。"
        action={
          <Button
            onClick={() => setCreateOpen(true)}
            className="bg-lagoon text-foam hover:bg-lagoon-deep"
          >
            <Plus className="mr-2 size-4" />
            新建空间
          </Button>
        }
      />

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

      <CreateWorkspaceDialog open={createOpen} onOpenChange={setCreateOpen} />
      <EditWorkspaceDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        item={selectedItem}
      />
      <ConfirmDeleteDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        description={`删除空间「${selectedItem?.name ?? ''}」后，其中的项目将移至默认空间，空间本身不可恢复。`}
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
