import { useEffect, useMemo, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Eye, Pencil, Plus, Search, Trash2 } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { DataTablePagination } from '#/components/data-table-pagination'
import { usePinList, useDeletePin } from '#/hooks/usePin'
import { useProjectNameMap } from '#/hooks/useProject'
import { useCurrentWorkspace } from '#/hooks/useCurrentWorkspace'
import { CreatePinDialog } from '#/components/pin/create-dialog'
import { PinDetailSheet } from '#/components/pin/pin-detail-sheet'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { Skeleton } from '@lumina/components/ui/skeleton'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { PageHeader } from '#/components/page-header'
import { formatDate } from '#/lib/format-date'
import type { PinItem } from '#/lib/models/response/pin'

export const Route = createFileRoute('/console/pin')({
  staticData: { crumb: 'Pin 约束' },
  component: PinPage,
})

const STATUS_FILTERS = [
  { value: '', label: '全部状态' },
  { value: 'pending', label: '待消费' },
  { value: 'consumed', label: '已消费' },
] as const

const CATEGORY_FILTERS = [
  { value: '', label: '全部分类' },
  { value: 'notice', label: '注意事项' },
  { value: 'dependency', label: '依赖约束' },
  { value: 'api_change', label: '接口变更' },
  { value: 'other', label: '其他' },
] as const

const categoryLabels: Record<string, string> = {
  notice: '注意事项',
  dependency: '依赖约束',
  api_change: '接口变更',
  other: '其他',
}

const priorityStyles: Record<string, string> = {
  high: 'border-destructive/40 bg-destructive/10 text-destructive',
  medium: 'border-chip-line bg-lagoon/10 text-lagoon-deep',
  low: 'border-line bg-chip-bg text-sea-ink-soft',
}

const priorityLabels: Record<string, string> = {
  high: '高优',
  medium: '中优',
  low: '低优',
}

function PinRow({
  item,
  onView,
  onEdit,
  onDelete,
}: {
  item: PinItem
  onView: (item: PinItem) => void
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
    <div className="flex items-center justify-between gap-4 border-b border-line bg-surface px-4 py-3.5 transition-colors hover:bg-chip-bg">
      <div className="flex min-w-0 flex-1 flex-col gap-1.5">
        <div className="flex flex-wrap items-center gap-2">
          <span
            onClick={() => onView(item)}
            className="cursor-pointer font-semibold text-sm text-sea-ink hover:text-lagoon transition-colors truncate max-w-md"
            title={item.title}
          >
            {item.title}
          </span>
          <span
            className={`font-mono text-[10px] px-1.5 py-0.5 border ${
              isPending
                ? 'border-chip-line bg-lagoon/10 text-lagoon-deep font-semibold'
                : 'border-line bg-chip-bg text-sea-ink-soft'
            }`}
          >
            {isPending ? '待消费' : '已消费'}
          </span>
          <span
            className={`font-mono text-[10px] px-1.5 py-0.5 border ${priorityStyle}`}
          >
            {priorityLabel}
          </span>
          <span className="font-mono text-[10px] px-1.5 py-0.5 border border-line bg-sand text-sea-ink-soft">
            {categoryLabel}
          </span>
        </div>

        {item.content ? (
          <p
            className="truncate text-xs text-sea-ink-soft max-w-2xl font-mono"
            title={item.content}
          >
            {item.content}
          </p>
        ) : null}

        <div className="flex flex-wrap items-center gap-2 text-[11px] text-sea-ink-soft">
          <span className="font-mono text-sea-ink bg-sand/60 px-1.5 py-0.2 border border-line">
            {fromName} <span className="text-lagoon font-bold">➔</span> {toName}
          </span>
          <span>•</span>
          <span>{formatDate(item.created_at)}</span>
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-1.5">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onView(item)}
          className="h-7 text-xs rounded-none border-line px-2 text-sea-ink hover:bg-foam"
          title="查看约束详情"
        >
          <Eye className="mr-1 size-3" />
          详情
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onEdit(item)}
          className="h-7 text-xs rounded-none border-line px-2 text-sea-ink hover:bg-foam"
          title="编辑约束属性"
        >
          <Pencil className="mr-1 size-3" />
          编辑
        </Button>
        <button
          type="button"
          className="grid size-7 place-items-center text-sea-ink-soft transition-colors hover:bg-destructive/10 hover:text-destructive border border-transparent hover:border-destructive/30"
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
  const [searchQuery, setSearchQuery] = useState('')

  const [createOpen, setCreateOpen] = useState(false)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailMode, setDetailMode] = useState<'view' | 'edit'>('view')
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [selectedItem, setSelectedItem] = useState<PinItem | null>(null)

  const { current } = useCurrentWorkspace()

  // 🌟 Q-04 切换空间时清空选中态
  useEffect(() => {
    setPage(1)
    setCreateOpen(false)
    setDetailOpen(false)
    setDeleteOpen(false)
    setSelectedItem(null)
  }, [current?.id])

  const { data, isLoading } = usePinList({
    page,
    size: pageSize,
    workspace_id: current?.id,
    ...(statusFilter ? { status: statusFilter } : {}),
    ...(categoryFilter ? { category: categoryFilter } : {}),
  })
  const showLoading = !current?.id || isLoading
  const deleteMutation = useDeletePin()

  const rawItems = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

  // 前端搜索过滤
  const items = useMemo(() => {
    if (!searchQuery.trim()) return rawItems
    const q = searchQuery.toLowerCase().trim()
    return rawItems.filter(
      (item) =>
        item.title.toLowerCase().includes(q) ||
        (item.content && item.content.toLowerCase().includes(q)),
    )
  }, [rawItems, searchQuery])

  // 统计概览
  const pendingCount = rawItems.filter((i) => i.status === 'pending').length
  const consumedCount = rawItems.filter((i) => i.status === 'consumed').length

  const handleStatusChange = (val: string) => {
    setStatusFilter(val)
    setPage(1)
  }

  const handleCategoryChange = (val: string) => {
    setCategoryFilter(val)
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
        title="Pin 约束管理"
        description="跨项目依赖约束传递与点对点定向推送，保障多仓库架构协同一致性"
        action={
          <Button
            onClick={() => setCreateOpen(true)}
            className="bg-lagoon text-foam hover:bg-lagoon-deep rounded-none"
          >
            <Plus className="mr-1.5 size-4" />
            推送新 Pin 约束
          </Button>
        }
      />

      {/* KPI 指标带 */}
      <motion.div variants={staggerItem}>
        <div className="grid grid-cols-3 border border-line bg-sand/30">
          <div className="p-4 border-r border-line">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              待处理约束 (当页 PENDING)
            </span>
            <span className="font-mono text-xl font-semibold text-lagoon-deep">
              {pendingCount}
            </span>
            <p className="text-[11px] text-sea-ink-soft mt-0.5">
              待下游消费的架构约束
            </p>
          </div>
          <div className="p-4 border-r border-line">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              已闭环消费 (当页 CONSUMED)
            </span>
            <span className="font-mono text-xl font-semibold text-sea-ink">
              {consumedCount}
            </span>
            <p className="text-[11px] text-sea-ink-soft mt-0.5">
              已被下游确认消费闭环
            </p>
          </div>
          <div className="p-4">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              当前条件总约束数
            </span>
            <span className="font-mono text-xl font-semibold text-sea-ink">
              {totalItems}
            </span>
            <p className="text-[11px] text-sea-ink-soft mt-0.5">
              服务端全量契约统计
            </p>
          </div>
        </div>
      </motion.div>

      {/* Filter 胶囊组合与搜索栏 */}
      <motion.div variants={staggerItem} className="space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line pb-3">
          <div className="flex flex-wrap items-center gap-2">
            {/* 状态胶囊 */}
            <div className="flex items-center border border-line bg-foam p-0.5">
              {STATUS_FILTERS.map((s) => (
                <button
                  key={s.value}
                  type="button"
                  onClick={() => handleStatusChange(s.value)}
                  className={`px-2.5 py-1 text-xs transition-colors ${
                    statusFilter === s.value
                      ? 'bg-lagoon text-foam font-semibold'
                      : 'text-sea-ink-soft hover:text-sea-ink'
                  }`}
                >
                  {s.label}
                </button>
              ))}
            </div>

            {/* 分类胶囊 */}
            <div className="flex items-center border border-line bg-foam p-0.5">
              {CATEGORY_FILTERS.map((c) => (
                <button
                  key={c.value}
                  type="button"
                  onClick={() => handleCategoryChange(c.value)}
                  className={`px-2.5 py-1 text-xs transition-colors ${
                    categoryFilter === c.value
                      ? 'bg-sea-ink text-foam font-semibold'
                      : 'text-sea-ink-soft hover:text-sea-ink'
                  }`}
                >
                  {c.label}
                </button>
              ))}
            </div>
          </div>

          {/* 搜索框 */}
          <div className="flex items-center gap-2 border border-line bg-foam px-2.5 py-1 w-64 focus-within:border-lagoon">
            <Search className="size-3.5 text-sea-ink-soft shrink-0" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="过滤当前页约束标题或正文..."
              className="h-6 border-0 bg-transparent p-0 text-xs shadow-none focus-visible:ring-0"
            />
          </div>
        </div>

        {/* 约束台账列表 */}
        <div className="border border-line bg-surface">
          {showLoading ? (
            <div className="flex flex-col gap-2 p-4">
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
            </div>
          ) : items.length === 0 ? (
            <div className="py-16 text-center">
              <p className="text-sm font-semibold text-sea-ink">
                暂无匹配的 Pin 约束契约
              </p>
              <p className="mt-1 text-xs text-sea-ink-soft">
                当前项目队列空闲，或尝试调整上方分类/状态筛选条件
              </p>
            </div>
          ) : (
            items.map((item) => (
              <PinRow
                key={item.id}
                item={item}
                onView={(target) => {
                  setSelectedItem(target)
                  setDetailMode('view')
                  setDetailOpen(true)
                }}
                onEdit={(target) => {
                  setSelectedItem(target)
                  setDetailMode('edit')
                  setDetailOpen(true)
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

      {/* 创建对话框 */}
      <CreatePinDialog open={createOpen} onOpenChange={setCreateOpen} />

      {/* 🌟 专属全景详情与在位编辑抽屉 */}
      <PinDetailSheet
        open={detailOpen}
        onOpenChange={setDetailOpen}
        item={selectedItem}
        initialMode={detailMode}
      />

      {/* 删除确认对话框 */}
      <ConfirmDeleteDialog
        open={deleteOpen}
        onOpenChange={(open) => {
          if (!open) setDeleteOpen(false)
        }}
        title="确认删除约束契约"
        description={`确定要删除约束「${selectedItem?.title ?? ''}」吗？此操作将永久废弃该跨项目约束。`}
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
