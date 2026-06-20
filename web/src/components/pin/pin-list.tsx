import { useState } from 'react'
import { motion } from 'motion/react'
import { Plus } from 'lucide-react'
import { Button } from '#/components/ui/button'
import { Skeleton } from '#/components/ui/skeleton'
import { DataTable } from '#/components/data-table'
import { DataTablePagination } from '#/components/data-table-pagination'
import { usePinList } from '#/hooks/usePin'
import { getColumns } from './columns'
import { CreatePinDialog } from './create-dialog'
import { EditPinDialog } from './edit-dialog'
import { DeletePinDialog } from './delete-dialog'
import type { PinItem } from '#/lib/models/response/pin'
import { staggerContainer, staggerItem, staggerItemLeft } from '#/lib/motion'

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

export function PinList() {
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
      {/* 标题行 */}
      <motion.div
        className="relative flex items-center justify-between pl-1.5"
        variants={staggerItemLeft}
      >
        <div className="absolute -left-4 top-0 h-full w-1 rounded-r-full bg-gradient-to-b from-lagoon to-palm" />
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-sea-ink">
            Pin 管理
          </h1>
          <p className="text-sea-ink-soft">
            跨项目依赖约束传递与点对点定向推送
          </p>
        </div>
        <Button
          onClick={() => setCreateOpen(true)}
          className="bg-lagoon text-foam hover:bg-lagoon-deep"
        >
          <Plus className="mr-2 size-4" />
          创建 Pin
        </Button>
      </motion.div>

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

      <CreatePinDialog open={createOpen} onOpenChange={setCreateOpen} />
      <EditPinDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        pin={selectedItem}
      />
      <DeletePinDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        pinId={selectedItem?.id ?? null}
      />
    </motion.div>
  )
}
