import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Plus } from 'lucide-react'
import { Button } from '#/components/ui/button'
import { DataTable } from '#/components/data-table'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useProjectList, useDeleteProject } from '#/hooks/useProject'
import { getColumns } from '#/components/project/columns'
import { CreateDialog } from '#/components/project/create-dialog'
import { EditDialog } from '#/components/project/edit-dialog'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import type { ProjectItem } from '#/lib/models/response/project'
import { staggerContainer, staggerItem } from '#/lib/motion'
import { PageHeader } from '#/components/page-header'
import { SkeletonTable } from '#/components/skeleton-table'

export const Route = createFileRoute('/console/project')({
  component: ProjectPage,
})

function ProjectPage() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [selectedItem, setSelectedItem] = useState<ProjectItem | null>(null)

  const { data, isLoading } = useProjectList({ page, size: pageSize })
  const deleteMutation = useDeleteProject()

  const items = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

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
        title="项目管理"
        description="管理你的项目，用于组织 Pin 和 Q&A"
        action={
          <Button
            onClick={() => setCreateOpen(true)}
            className="bg-lagoon text-foam hover:bg-lagoon-deep"
          >
            <Plus className="mr-2 size-4" />
            创建项目
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
        description={`确定要删除项目「${selectedItem?.name ?? ''}」吗？此操作不可撤销。`}
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
