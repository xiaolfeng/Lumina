import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Skeleton } from '@lumina/components/ui/skeleton'
import { Eye, Trash2 } from 'lucide-react'
import { useSessionList, useDeleteSession } from '#/hooks/useQaAdmin'
import { useDashboardOverview } from '#/hooks/useDashboard'
import { SessionDetailDrawer } from '#/components/qa/session-detail-drawer'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { DataTablePagination } from '#/components/data-table-pagination'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import type { SessionItem } from '#/lib/models/response/qa-admin'

export const Route = createFileRoute('/console/qa/')({
  component: QaPage,
})

function QaPage() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [deleteTarget, setDeleteTarget] = useState<SessionItem | null>(null)
  const [viewTarget, setViewTarget] = useState<string | null>(null)

  const { data, isLoading } = useSessionList({ page, size: pageSize })
  const { data: overviewData } = useDashboardOverview()
  const deleteMutation = useDeleteSession()

  const items = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))
  const qaStats = overviewData?.data?.qa

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
      className="mx-auto max-w-5xl space-y-8"
    >
      {/* ─── KPI band ─── */}
      <motion.div variants={staggerItem}>
        <div className="grid grid-cols-1 border-b border-line sm:grid-cols-3">
          <Kpi
            label="活跃会话"
            value={qaStats?.active}
            delta="临时会话 48h 过期"
            loading={!qaStats}
          />
          <Kpi
            label="待回答问题"
            value={qaStats?.pending_questions}
            delta="等待用户回答"
            loading={!qaStats}
          />
          <Kpi
            label="已归档"
            value={qaStats?.expired}
            delta={`共 ${qaStats?.total ?? 0} 个会话`}
            loading={!qaStats}
            last
          />
        </div>
      </motion.div>

      {/* ─── 会话列表 ─── */}
      <motion.div variants={staggerItem}>
        <h3 className="display-title mb-2 text-lg font-semibold text-sea-ink">
          问答会话
          <span className="ml-1 text-lagoon">──</span>
        </h3>

        <div className="border-t border-line">
          {isLoading ? (
            <div className="space-y-2 py-4">
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
            </div>
          ) : items.length > 0 ? (
            items.map((item) => (
              <div
                key={item.id}
                className="flex items-center gap-4 border-b border-line px-1 py-4 transition-colors hover:bg-chip-bg"
              >
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-semibold text-sea-ink">
                    {item.title}
                  </p>
                  <div className="mt-1 flex items-center gap-2.5 text-xs text-sea-ink-soft">
                    <SessionStatus status={item.status} />
                    <span>
                      {item.question_count} 问题 · 已答 {item.answered_count}
                    </span>
                    <span>{relativeTime(item.updated_at)}</span>
                  </div>
                </div>
                <button
                  type="button"
                  onClick={() => setViewTarget(item.id)}
                  className="grid size-7 place-items-center text-sea-ink-soft transition-colors hover:text-sea-ink"
                  aria-label={`查看 ${item.title} 详情`}
                >
                  <Eye className="size-3.5" />
                </button>
                <button
                  type="button"
                  onClick={() => setDeleteTarget(item)}
                  className="grid size-7 place-items-center text-sea-ink-soft transition-colors hover:text-destructive"
                  aria-label={`删除 ${item.title}`}
                >
                  <Trash2 className="size-3.5" />
                </button>
              </div>
            ))
          ) : (
            <div className="border border-line px-8 py-16 text-center">
              <p className="display-title text-xl font-semibold text-sea-ink">
                暂无问答会话
              </p>
              <p className="mt-2.5 text-sm text-sea-ink-soft">
                Agent 发起交互问答后，会话将在此展示。
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

      {/* ─── 删除确认 ─── */}
      <ConfirmDeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null)
        }}
        title="删除会话"
        description={`确定要删除会话「${deleteTarget?.title ?? ''}」吗？删除后所有问答数据将不可恢复。`}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget.id, {
              onSuccess: () => setDeleteTarget(null),
            })
          }
        }}
        isPending={deleteMutation.isPending}
      />

      {/* ─── 详情抽屉 ─── */}
      <SessionDetailDrawer
        sessionId={viewTarget}
        onClose={() => setViewTarget(null)}
      />
    </motion.div>
  )
}

function Kpi({
  label,
  value,
  delta,
  loading,
  last,
}: {
  label: string
  value?: number
  delta?: string
  loading?: boolean
  last?: boolean
}) {
  return (
    <div
      className={`py-7 ${last ? '' : 'border-r border-line'} first:pl-0 md:px-6`}
    >
      <p className="text-[10.5px] font-bold uppercase tracking-[0.18em] text-sea-ink-soft">
        {label}
      </p>
      {loading ? (
        <Skeleton className="mt-3 h-11 w-16" />
      ) : (
        <p className="display-title mt-3 text-5xl font-medium tracking-tight text-sea-ink">
          {value ?? 0}
        </p>
      )}
      {delta && (
        <p className="mt-2 text-[11.5px] text-sea-ink-soft">{delta}</p>
      )}
    </div>
  )
}

function SessionStatus({ status }: { status: SessionItem['status'] }) {
  const config = {
    active: { dot: 'bg-green-600', text: 'text-green-700', label: '进行中' },
    expired: {
      dot: 'bg-sea-ink-soft/40',
      text: 'text-sea-ink-soft',
      label: '已归档',
    },
    deleted: { dot: 'bg-destructive', text: 'text-destructive', label: '已删除' },
  }[status]

  return (
    <span className={`inline-flex items-center gap-1.5 text-xs font-medium ${config.text}`}>
      <span className={`inline-block size-[7px] rounded-full ${config.dot}`} />
      {config.label}
    </span>
  )
}

function relativeTime(iso: string): string {
  const time = new Date(iso).getTime()
  if (Number.isNaN(time)) return ''
  const diff = Date.now() - time
  const min = Math.floor(diff / 60000)
  if (min < 1) return '刚刚'
  if (min < 60) return `${min} 分钟前`
  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour} 小时前`
  const day = Math.floor(hour / 24)
  return `${day} 天前`
}
