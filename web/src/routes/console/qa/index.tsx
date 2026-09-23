import { useEffect, useMemo, useState } from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Skeleton } from '@lumina/components/ui/skeleton'
import {
  Eye,
  MessageCircleQuestion,
  Search,
  Settings,
  Trash2,
} from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { useSessionList, useDeleteSession } from '#/hooks/useQaAdmin'
import { useCurrentWorkspace } from '#/hooks/useCurrentWorkspace'
import { useDashboardOverview } from '#/hooks/useDashboard'
import { SessionDetailDrawer } from '#/components/qa/session-detail-drawer'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { DataTablePagination } from '#/components/data-table-pagination'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { PageHeader } from '#/components/page-header'
import { formatDateTime } from '#/lib/format-date'
import type { SessionItem } from '#/lib/models/response/qa-admin'

export const Route = createFileRoute('/console/qa/')({
  staticData: { crumb: '问答管理' },
  component: QaPage,
})

type SessionStatusFilter = '' | 'active' | 'expired' | 'deleted'

const STATUS_FILTERS: { value: SessionStatusFilter; label: string }[] = [
  { value: '', label: '全部会话' },
  { value: 'active', label: '活跃中' },
  { value: 'expired', label: '已过期' },
]

function SessionStatusBadge({ status }: { status: string }) {
  if (status === 'active') {
    return (
      <span className="font-mono text-[10px] px-1.5 py-0.5 border border-chip-line bg-lagoon/10 text-lagoon-deep font-semibold">
        活跃中
      </span>
    )
  }
  if (status === 'expired') {
    return (
      <span className="font-mono text-[10px] px-1.5 py-0.5 border border-line bg-chip-bg text-sea-ink-soft">
        已过期
      </span>
    )
  }
  return (
    <span className="font-mono text-[10px] px-1.5 py-0.5 border border-line bg-sand text-sea-ink-soft">
      {status}
    </span>
  )
}

function QaPage() {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [statusFilter, setStatusFilter] = useState<SessionStatusFilter>('')
  const [searchQuery, setSearchQuery] = useState('')

  const [deleteTarget, setDeleteTarget] = useState<SessionItem | null>(null)
  const [viewTarget, setViewTarget] = useState<string | null>(null)

  const { current } = useCurrentWorkspace()

  // 🌟 Q-04 切换空间时重置选中态
  useEffect(() => {
    setPage(1)
    setDeleteTarget(null)
    setViewTarget(null)
  }, [current?.id])

  const { data, isLoading } = useSessionList({
    page,
    size: pageSize,
    workspace_id: current?.id,
    ...(statusFilter ? { status: statusFilter } : {}),
  })
  const showLoading = !current?.id || isLoading
  const { data: overviewData } = useDashboardOverview()
  const deleteMutation = useDeleteSession()

  const rawItems = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))
  const qaStats = overviewData?.data?.qa

  // 本地搜索过滤
  const items = useMemo(() => {
    if (!searchQuery.trim()) return rawItems
    const q = searchQuery.toLowerCase().trim()
    return rawItems.filter(
      (item) =>
        item.title.toLowerCase().includes(q) ||
        (item.agent && item.agent.toLowerCase().includes(q)) ||
        (item.project_name && item.project_name.toLowerCase().includes(q)),
    )
  }, [rawItems, searchQuery])

  const handleStatusChange = (val: SessionStatusFilter) => {
    setStatusFilter(val)
    setPage(1)
  }

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
      className="space-y-4"
    >
      <PageHeader
        title="问答管理"
        description="查看 Agent 发起的交互式富问答会话、待答队列状态与历史归档"
        action={
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              onClick={() =>
                void (navigate as any)({
                  to: '/console/settings',
                  search: { tab: 'qa', from: '/console/qa' },
                })
              }
              className="rounded-none border-line text-xs text-sea-ink hover:bg-chip-bg"
              title="跳转至系统设置配置 Q&A 参数"
            >
              <Settings className="mr-1.5 size-3.5 text-lagoon" />
              Q&A 运行参数配置
            </Button>
            <Button
              onClick={() => window.open('/interact', '_blank')}
              className="bg-lagoon text-foam hover:bg-lagoon-deep rounded-none text-xs"
            >
              <MessageCircleQuestion className="mr-1.5 size-3.5" />
              进入交互大厅
            </Button>
          </div>
        }
      />

      {/* KPI 指标带 */}
      <motion.div variants={staggerItem}>
        <div className="grid grid-cols-1 border border-line bg-sand/30 sm:grid-cols-3">
          <div className="p-4 border-b border-line sm:border-b-0 sm:border-r">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              活跃会话 (ACTIVE)
            </span>
            <span className="font-mono text-xl font-semibold text-lagoon-deep">
              {qaStats?.active ?? 0}
            </span>
            <p className="text-[11px] text-sea-ink-soft mt-0.5">
              临时会话默认保留 48h
            </p>
          </div>
          <div className="p-4 border-b border-line sm:border-b-0 sm:border-r">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              待回答问题 (PENDING)
            </span>
            <span className="font-mono text-xl font-semibold text-destructive">
              {qaStats?.pending_questions ?? 0}
            </span>
            <p className="text-[11px] text-sea-ink-soft mt-0.5">
              阻塞等待用户裁决中
            </p>
          </div>
          <div className="p-4">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              累计历史会话
            </span>
            <span className="font-mono text-xl font-semibold text-sea-ink">
              {qaStats?.total ?? totalItems}
            </span>
            <p className="text-[11px] text-sea-ink-soft mt-0.5">
              已归档 {qaStats?.expired ?? 0} 个会话
            </p>
          </div>
        </div>
      </motion.div>

      {/* 工具栏与筛选胶囊 */}
      <motion.div variants={staggerItem} className="space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line pb-3">
          {/* 状态筛选胶囊 */}
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

          {/* 搜索框 */}
          <div className="flex items-center gap-2 border border-line bg-foam px-2.5 py-1 w-64 focus-within:border-lagoon">
            <Search className="size-3.5 text-sea-ink-soft shrink-0" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="过滤当前页会话主题或 Agent..."
              className="h-6 border-0 bg-transparent p-0 text-xs shadow-none focus-visible:ring-0"
            />
          </div>
        </div>

        {/* 会话台账列表 */}
        <div className="border border-line bg-surface">
          {showLoading ? (
            <div className="space-y-2 p-4">
              <Skeleton className="h-14 w-full" />
              <Skeleton className="h-14 w-full" />
              <Skeleton className="h-14 w-full" />
            </div>
          ) : items.length > 0 ? (
            items.map((item) => (
              <div
                key={item.id}
                className="flex items-center justify-between gap-4 border-b border-line px-4 py-3.5 transition-colors hover:bg-chip-bg last:border-b-0"
              >
                <div className="min-w-0 flex-1 flex flex-col gap-1">
                  <div className="flex items-center gap-2">
                    <span
                      onClick={() => setViewTarget(item.id)}
                      className="cursor-pointer font-semibold text-sm text-sea-ink hover:text-lagoon transition-colors truncate max-w-lg"
                      title={item.title}
                    >
                      {item.title}
                    </span>
                    <SessionStatusBadge status={item.status} />
                  </div>
                  <div className="flex flex-wrap items-center gap-2 text-xs text-sea-ink-soft">
                    <span className="font-mono text-sea-ink bg-sand/60 px-1.5 py-0.2 border border-line">
                      Agent: {item.agent || '系统'}
                    </span>
                    <span>•</span>
                    <span>
                      {item.question_count} 个问题 (已答 {item.answered_count})
                    </span>
                    <span>•</span>
                    <span>更新于 {formatDateTime(item.updated_at)}</span>
                  </div>
                </div>

                <div className="flex shrink-0 items-center gap-1.5">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setViewTarget(item.id)}
                    className="h-7 text-xs rounded-none border-line px-2 text-sea-ink hover:bg-foam"
                    title="查看会话全景详情"
                  >
                    <Eye className="mr-1 size-3" />
                    详情
                  </Button>
                  <button
                    type="button"
                    onClick={() => setDeleteTarget(item)}
                    className="grid size-7 place-items-center text-sea-ink-soft transition-colors hover:bg-destructive/10 hover:text-destructive border border-transparent hover:border-destructive/30"
                    aria-label={`删除 ${item.title}`}
                    title="删除会话"
                  >
                    <Trash2 className="size-3.5" />
                  </button>
                </div>
              </div>
            ))
          ) : (
            <div className="border border-line px-8 py-16 text-center">
              <p className="text-sm font-semibold text-sea-ink">
                暂无问答会话记录
              </p>
              <p className="mt-1 text-xs text-sea-ink-soft">
                Agent 发起交互问答后，会话将在此实时同步展示。
              </p>
            </div>
          )}
        </div>

        {totalItems > 0 && (
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
        )}
      </motion.div>

      {/* 删除确认对话框 */}
      <ConfirmDeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null)
        }}
        title="确认删除问答会话"
        description={`确定要删除会话「${deleteTarget?.title ?? ''}」吗？删除后所有问答与选项记录将不可恢复。`}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget.id, {
              onSuccess: () => setDeleteTarget(null),
            })
          }
        }}
        isPending={deleteMutation.isPending}
      />

      {/* 🌟 详情全景抽屉 */}
      <SessionDetailDrawer
        sessionId={viewTarget}
        onClose={() => setViewTarget(null)}
      />
    </motion.div>
  )
}
