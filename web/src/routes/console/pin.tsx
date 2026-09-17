import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { DataTablePagination } from '#/components/data-table-pagination'
import { usePinList, useDeletePin } from '#/hooks/usePin'
import { useProjectNameMap } from '#/hooks/useProject'
import { useCurrentWorkspace } from '#/hooks/useCurrentWorkspace'
import { CreatePinDialog } from '#/components/pin/create-dialog'
import { EditPinDialog } from '#/components/pin/edit-dialog'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { Skeleton } from '@lumina/components/ui/skeleton'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { PageHeader } from '#/components/page-header'
import { formatDate } from '#/lib/format-date'
import type { PinItem } from '#/lib/models/response/pin'

export const Route = createFileRoute('/console/pin')({
  staticData: { crumb: 'Pin 管理' },
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

const categoryLabels: Record<string, string> = {
  notice: '注意事项',
  dependency: '依赖约束',
  api_change: '接口变更',
}

const priorityStyles: Record<string, string> = {
  high: 'border-destructive/30 bg-destructive/10 text-destructive',
  medium: 'border-chip-line bg-lagoon/10 text-lagoon-deep',
  low: 'border-line bg-chip-bg text-sea-ink-soft',
}

const priorityLabels: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低',
}

function PinBadge({
  className,
  children,
}: {
  className?: string
  children: React.ReactNode
}) {
  return (
    <span
      className={`inline-flex items-center border px-1.5 py-0.5 text-[11px] leading-5 font-semibold ${className ?? 'border-line bg-surface text-sea-ink-soft'}`}
    >
      {children}
    </span>
  )
}

function PinRow({
  item,
  onEdit,
  onDelete,
}: {
  item: PinItem
  onEdit: (item: PinItem) => void
  onDelete: (item: PinItem) => void
}) {
  const { names } = useProjectNameMap()
  const fromName = names[item.from_project_id] ?? item.from_project_id
  const toName = names[item.to_project_id] ?? item.to_project_id
  const priorityStyle = priorityStyles[item.priority] ?? priorityStyles.low
  const priorityLabel = priorityLabels[item.priority] ?? item.priority
  const categoryLabel = categoryLabels[item.category] ?? '其他'
  const isPending = item.status === 'pending'

  return (
    <div className="flex items-center gap-4 border-b border-line px-2 py-4 transition-colors hover:bg-chip-bg">
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <div className="flex flex-wrap items-center gap-2">
          <span className="truncate text-sm font-semibold text-sea-ink">
            {item.title}
          </span>
          {isPending ? (
            <PinBadge className="border-chip-line bg-lagoon/10 text-lagoon-deep">
              待消费
            </PinBadge>
          ) : (
            <PinBadge className="border-line bg-chip-bg text-sea-ink-soft">
              已消费
            </PinBadge>
          )}
          {item.priority ? (
            <PinBadge className={priorityStyle}>{priorityLabel}</PinBadge>
          ) : null}
          <PinBadge>{categoryLabel}</PinBadge>
        </div>
        {item.content ? (
          <p
            className="truncate text-[13px] text-sea-ink-soft"
            title={item.content}
          >
            {item.content}
          </p>
        ) : null}
        <div className="flex flex-wrap items-center gap-2 text-xs text-sea-ink-soft">
          <span className="font-mono text-[11.5px] text-sea-ink">
            {fromName} <span className="text-lagoon">➔</span> {toName}
          </span>
          <span>·</span>
          <span>{formatDate(item.created_at)}</span>
        </div>
      </div>
      <div className="flex shrink-0 items-center gap-1">
        <button
          type="button"
          className="grid size-9 place-items-center text-sea-ink-soft transition-colors hover:bg-foam hover:text-sea-ink"
          aria-label={`编辑 ${item.title}`}
          title="编辑约束"
          onClick={() => onEdit(item)}
        >
          <Pencil className="size-3.5" />
        </button>
        <button
          type="button"
          className="grid size-9 place-items-center text-sea-ink-soft transition-colors hover:bg-destructive/10 hover:text-destructive"
          aria-label={`删除 ${item.title}`}
          title="删除约束"
          onClick={() => onDelete(item)}
        >
          <Trash2 className="size-3.5" />
        </button>
      </div>
    </div>
  )
}

function PinPage() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [statusFilter, setStatusFilter] = useState('')
  const [categoryFilter, setCategoryFilter] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [selectedItem, setSelectedItem] = useState<PinItem | null>(null)

  const { current } = useCurrentWorkspace()
  const { data, isLoading } = usePinList({
    page,
    size: pageSize,
    workspace_id: current?.id,
    ...(statusFilter ? { status: statusFilter } : {}),
    ...(categoryFilter ? { category: categoryFilter } : {}),
  })
  const showLoading = !current?.id || isLoading
  const deleteMutation = useDeletePin()

  const items = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

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

      <motion.div variants={staggerItem} className="space-y-4">
        {/* 筛选工具栏 */}
        <div className="flex flex-wrap items-center gap-4">
          <div className="flex items-center gap-2">
            <label
              htmlFor="pin-status-filter"
              className="text-sm text-sea-ink-soft"
            >
              状态
            </label>
            <select
              id="pin-status-filter"
              value={statusFilter}
              onChange={(e) => handleStatusChange(e.target.value)}
              className="h-8 border border-line bg-sand px-2.5 text-sm text-sea-ink outline-none transition-colors focus:border-lagoon"
            >
              {statusOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
          <div className="flex items-center gap-2">
            <label
              htmlFor="pin-category-filter"
              className="text-sm text-sea-ink-soft"
            >
              分类
            </label>
            <select
              id="pin-category-filter"
              value={categoryFilter}
              onChange={(e) => handleCategoryChange(e.target.value)}
              className="h-8 border border-line bg-sand px-2.5 text-sm text-sea-ink outline-none transition-colors focus:border-lagoon"
            >
              {categoryOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* 流式约束列表 */}
        <div className="border-t border-line">
          {showLoading ? (
            <div className="flex flex-col gap-2 py-4">
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-20 w-full" />
            </div>
          ) : items.length === 0 ? (
            <div className="py-14 text-center">
              <p className="text-sm font-semibold text-sea-ink">
                暂无匹配的约束
              </p>
              <p className="mt-1 text-[13px] text-sea-ink-soft">
                当前队列空闲，或调整筛选条件
              </p>
            </div>
          ) : (
            items.map((item) => (
              <PinRow
                key={item.id}
                item={item}
                onEdit={(target) => {
                  setSelectedItem(target)
                  setEditOpen(true)
                }}
                onDelete={(target) => {
                  setSelectedItem(target)
                  setDeleteOpen(true)
                }}
              />
            ))
          )}
        </div>

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
      </motion.div>

      <CreatePinDialog open={createOpen} onOpenChange={setCreateOpen} />
      <EditPinDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        pin={selectedItem}
      />
      <ConfirmDeleteDialog
        open={deleteOpen}
        onOpenChange={(open) => {
          if (!open) setDeleteOpen(false)
        }}
        title="删除 Pin"
        description={`确定删除「${selectedItem?.title}」？目标项目将不再收到该约束提醒。`}
        onConfirm={() => {
          if (selectedItem) deleteMutation.mutate(selectedItem.id)
          setDeleteOpen(false)
        }}
        isPending={deleteMutation.isPending}
      />
    </motion.div>
  )
}
