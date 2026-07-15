import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Plus } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { DataTable } from '#/components/data-table'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useApikeyList, useDeleteApikey } from '#/hooks/useApikey'
import { getColumns } from '#/components/apikey/columns'
import { CreateDialog } from '#/components/apikey/create-dialog'
import { EditDialog } from '#/components/apikey/edit-dialog'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { ResetDialog } from '#/components/apikey/reset-dialog'
import { SkeletonTable } from '#/components/skeleton-table'
import type { ApikeyItem } from '#/lib/models/response/apikey'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { PageHeader } from '#/components/page-header'

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
  const deleteMutation = useDeleteApikey()

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
      <PageHeader
        title="令牌管理"
        description="管理你的 API 访问令牌"
        action={
          <Button
            onClick={() => setCreateOpen(true)}
            className="bg-lagoon text-foam hover:bg-lagoon-deep"
          >
            <Plus className="mr-2 size-4" />
            创建令牌
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
        title="确定要删除此令牌吗？"
        description="此操作不可撤销。删除后，使用该令牌的所有 API 请求将立即失效。"
        onConfirm={() => {
          if (!selectedItem) return
          deleteMutation.mutate(selectedItem.id, {
            onSuccess: () => setDeleteOpen(false),
          })
        }}
        isPending={deleteMutation.isPending}
      />
      <ResetDialog
        open={resetOpen}
        onOpenChange={setResetOpen}
        item={selectedItem}
      />
    </motion.div>
  )
}
