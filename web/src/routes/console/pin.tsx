import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Plus } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { DataTable } from '#/components/data-table'
import { DataTablePagination } from '#/components/data-table-pagination'
import { usePinList, useDeletePin } from '#/hooks/usePin'
import { getColumns } from '#/components/pin/columns'
import { CreatePinDialog } from '#/components/pin/create-dialog'
import { EditPinDialog } from '#/components/pin/edit-dialog'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { SkeletonTable } from '#/components/skeleton-table'
import type { PinItem } from '#/lib/models/response/pin'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { PageHeader } from '#/components/page-header'

export const Route = createFileRoute('/console/pin')({
  component: PinPage,
})

// ── 筛选选项常量 ──
const statusOptions = [
  { value: '', label: '全部' },
  { value: 'pending', label: '待消费' },
  { value: 'consumed', label: '已消费' },
]

const categoryOptions = [
  { value: '', label: '全部' },
  { value: 'notice', label: '注意事项' },
  { value: 'dependency', label: '依赖约束' },
  { value: 'api_change', label: '接口变更' },
  { value: 'other', label: '其他' },
]

function PinPage() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [statusFilter, setStatusFilter] = useState('')
  const [categoryFilter, setCategoryFilter] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [selectedItem, setSelectedItem] = useState<PinItem | null>(null)

  const { data, isLoading } = usePinList({
    page,
    size: pageSize,
    ...(statusFilter ? { status: statusFilter } : {}),
    ...(categoryFilter ? { category: categoryFilter } : {}),
  })
  const deleteMutation = useDeletePin()

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

  // 筛选变化时重置到第一页
  const handleStatusChange = (value: string) => {
    setStatusFilter(value)
    setPage(1)
  }

  const handleCategoryChange = (value: string) => {
    setCategoryFilter(value)
    setPage(1)
  }

  return (
    <motion.div
      className="space-y-4"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      <PageHeader
        title="Pin 管理"
        description="跨项目依赖约束传递与点对点定向推送"
        action={
          <Button
            onClick={() => setCreateOpen(true)}
            className="bg-lagoon text-foam hover:bg-lagoon-deep"
          >
            <Plus className="mr-2 size-4" />
            创建 Pin
          </Button>
        }
      />

      {/* 表格区域 */}
      <motion.div variants={staggerItem}>
        {/* 筛选工具栏 */}
        <div className="mb-3 flex flex-wrap gap-3">
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">状态</span>
            <select
              value={statusFilter}
              onChange={(e) => handleStatusChange(e.target.value)}
              className="rounded-md border border-input bg-background px-3 py-1.5 text-sm"
            >
              {statusOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">分类</span>
            <select
              value={categoryFilter}
              onChange={(e) => handleCategoryChange(e.target.value)}
              className="rounded-md border border-input bg-background px-3 py-1.5 text-sm"
            >
              {categoryOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        </div>

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

      <CreatePinDialog open={createOpen} onOpenChange={setCreateOpen} />
      <EditPinDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        pin={selectedItem}
      />
      <ConfirmDeleteDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        description="确认删除此 Pin 约束？此操作不可撤销。"
        onConfirm={() => {
          if (!selectedItem?.id) return
          deleteMutation.mutate(selectedItem.id, {
            onSuccess: () => setDeleteOpen(false),
          })
        }}
        isPending={deleteMutation.isPending}
      />
    </motion.div>
  )
}
