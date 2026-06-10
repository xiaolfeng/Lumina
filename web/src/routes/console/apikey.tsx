import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Plus } from 'lucide-react'
import { Button } from '#/components/ui/button'
import { Skeleton } from '#/components/ui/skeleton'
import { DataTable } from '#/components/data-table'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useApikeyList } from '#/hooks/useApikey'
import { getColumns } from '#/components/apikey/columns'
import { CreateDialog } from '#/components/apikey/create-dialog'
import { EditDialog } from '#/components/apikey/edit-dialog'
import { DeleteDialog } from '#/components/apikey/delete-dialog'
import { ResetDialog } from '#/components/apikey/reset-dialog'
import type { ApikeyItem } from '#/lib/models/response/apikey'
import { staggerContainer, staggerItem, staggerItemLeft } from '#/lib/motion'

export const Route = createFileRoute('/console/apikey')({
  component: ApikeyPage,
})

function ApikeyPage() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [resetOpen, setResetOpen] = useState(false)
  const [selectedItem, setSelectedItem] = useState<ApikeyItem | null>(null)

  const { data, isLoading } = useApikeyList({ page, size: pageSize })

  const items = data?.data?.items ?? []
  const totalPages = data?.data?.total_pages ?? 1
  const totalItems = data?.data?.total_items ?? 0

  const columns = getColumns({
    onEdit: (item) => {
      setSelectedItem(item)
      setEditOpen(true)
    },
    onReset: (item) => {
      setSelectedItem(item)
      setResetOpen(true)
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
        className="relative flex items-center justify-between pl-1.5"
        variants={staggerItemLeft}
      >
        <div className="absolute -left-4 top-0 h-full w-1 rounded-r-full bg-gradient-to-b from-(--lagoon) to-(--palm)" />
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-(--sea-ink)">令牌管理</h1>
          <p className="text-(--sea-ink-soft)">管理你的 API 访问令牌</p>
        </div>
        <Button
          onClick={() => setCreateOpen(true)}
          className="bg-(--lagoon) text-(--foam) hover:bg-(--lagoon-deep)"
        >
          <Plus className="mr-2 size-4" />
          创建令牌
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
      <ResetDialog
        open={resetOpen}
        onOpenChange={setResetOpen}
        item={selectedItem}
      />
    </motion.div>
  )
}
