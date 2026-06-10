import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Plus } from 'lucide-react'
import { Button } from '#/components/ui/button'
import { Skeleton } from '#/components/ui/skeleton'
import { DataTable } from '#/components/data-table'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useProjectList } from '#/hooks/useProject'
import { getColumns } from '#/components/project/columns'
import { CreateDialog } from '#/components/project/create-dialog'
import { EditDialog } from '#/components/project/edit-dialog'
import { DeleteDialog } from '#/components/project/delete-dialog'
import type { ProjectItem } from '#/lib/models/response/project'
import { staggerContainer, staggerItem } from '#/lib/motion'

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
      {/* 标题行 */}
      <motion.div
        className="flex items-center justify-between"
        variants={staggerItem}
      >
        <div>
          <h1 className="text-2xl font-bold tracking-tight">项目管理</h1>
          <p className="text-muted-foreground">
            管理你的项目，用于组织 Pin 和 Q&A
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 size-4" />
          创建项目
        </Button>
      </motion.div>

      {/* 表格区域 */}
      <motion.div variants={staggerItem}>
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
      </motion.div>

      <CreateDialog open={createOpen} onOpenChange={setCreateOpen} />
      <EditDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        item={selectedItem}
      />
      <DeleteDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        item={selectedItem}
      />
    </motion.div>
  )
}
