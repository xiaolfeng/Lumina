import { useState } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Button } from '@lumina/components/ui/button'
import { Skeleton } from '@lumina/components/ui/skeleton'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@lumina/components/ui/dialog'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@lumina/components/ui/select'
import { ExternalLink, Eye, Plus, Trash2 } from 'lucide-react'
import {
  usePreviewSessionList,
  useDeletePreviewSession,
  useCreatePreviewSession,
} from '#/hooks/usePreviewAdmin'
import { useProjectList } from '#/hooks/useProject'
import { useDashboardOverview } from '#/hooks/useDashboard'
import { PreviewSessionDetailDrawer } from '#/components/preview/session-detail-drawer'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { DataTablePagination } from '#/components/data-table-pagination'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import type { PreviewSessionItem } from '#/lib/models/response/preview'

export const Route = createFileRoute('/console/preview/')({
  component: PreviewPage,
})

function PreviewPage() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [deleteTarget, setDeleteTarget] = useState<PreviewSessionItem | null>(
    null,
  )
  const [viewTarget, setViewTarget] = useState<PreviewSessionItem | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [createTitle, setCreateTitle] = useState('')
  const [createProjectId, setCreateProjectId] = useState('')

  const { data, isLoading } = usePreviewSessionList({ page, size: pageSize })
  const { data: overviewData } = useDashboardOverview()
  const { data: projectData } = useProjectList({ page: 1, size: 50 })
  const deleteMutation = useDeletePreviewSession()
  const createMutation = useCreatePreviewSession()

  const items = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))
  const projects = projectData?.data?.items ?? []
  const previewStats = overviewData?.data?.preview

  const handleCreate = () => {
    if (!createProjectId) return
    createMutation.mutate(
      { project_id: createProjectId, title: createTitle.trim() || undefined },
      {
        onSuccess: () => {
          setCreateOpen(false)
          setCreateTitle('')
          setCreateProjectId('')
        },
      },
    )
  }

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
      className="mx-auto max-w-5xl space-y-8"
    >
      {/* ─── KPI band ─── */}
      <motion.div variants={staggerItem}>
        <div className="grid grid-cols-2 border-b border-line">
          <Kpi
            label="活跃会话"
            value={previewStats?.active}
            delta="会话以 hash 分享"
            loading={!previewStats}
          />
          <Kpi
            label="文件总数"
            value={previewStats?.files}
            delta="扁平单层 · 单文件 ≤ 256KB"
            loading={!previewStats}
            last
          />
        </div>
      </motion.div>

      {/* ─── 会话列表 ─── */}
      <motion.div variants={staggerItem}>
        <div className="mb-6 flex items-center justify-between">
          <h3 className="display-title text-lg font-semibold text-sea-ink">
            预览会话
            <span className="ml-1 text-lagoon">──</span>
          </h3>
          <Button
            onClick={() => setCreateOpen(true)}
            className="bg-sea-ink text-foam hover:bg-lagoon-deep"
          >
            <Plus className="size-4" />
            新建会话
          </Button>
        </div>

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
                    <span className="font-mono">{item.hash.slice(0, 8)}</span>
                    <span>
                      {item.file_count} 文件 · {relativeTime(item.updated_at)}
                    </span>
                    <StatusDot active={item.status === 'active'} />
                  </div>
                </div>
                <Link
                  to="/preview"
                  search={{ session: item.hash }}
                  target="_blank"
                  className="grid size-7 place-items-center text-sea-ink-soft transition-colors hover:text-sea-ink"
                  aria-label={`查看 ${item.title}`}
                >
                  <Eye className="size-3.5" />
                </Link>
                <button
                  type="button"
                  onClick={() => setViewTarget(item)}
                  className="grid size-7 place-items-center text-sea-ink-soft transition-colors hover:text-sea-ink"
                  aria-label={`打开 ${item.title} 详情`}
                >
                  <ExternalLink className="size-3.5" />
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
                暂无预览会话
              </p>
              <p className="mt-2.5 text-sm text-sea-ink-soft">
                创建第一个预览会话，可视化你的前端原型。
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

      {/* ─── 新建会话 dialog ─── */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="border-line sm:max-w-md">
          <DialogHeader>
            <DialogTitle>新建预览会话</DialogTitle>
            <DialogDescription>
              选择关联项目并填写会话标题，创建后可通过 hash 分享预览。
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label className="text-xs text-sea-ink-soft">关联项目</Label>
              <Select value={createProjectId} onValueChange={setCreateProjectId}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="选择项目" />
                </SelectTrigger>
                <SelectContent>
                  {projects.map((p) => (
                    <SelectItem key={p.id} value={p.id}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label className="text-xs text-sea-ink-soft">会话标题</Label>
              <Input
                value={createTitle}
                onChange={(e) => setCreateTitle(e.target.value)}
                placeholder="未命名预览"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              取消
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!createProjectId || createMutation.isPending}
              className="bg-sea-ink text-foam hover:bg-lagoon-deep"
            >
              {createMutation.isPending ? '创建中...' : '创建'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ─── 删除确认 ─── */}
      <ConfirmDeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null)
        }}
        title="删除预览会话"
        description={`确定要删除会话「${deleteTarget?.title ?? ''}」吗？删除后该会话下的所有前端文件将不可恢复。`}
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
      <PreviewSessionDetailDrawer
        session={viewTarget}
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
    <div className={`py-7 ${last ? '' : 'border-r border-line'} first:pl-0 md:px-6`}>
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

function StatusDot({ active }: { active: boolean }) {
  return (
    <span className="inline-flex items-center gap-1.5 text-xs font-medium">
      <span
        className={`inline-block size-[7px] rounded-full ${
          active ? 'bg-green-600' : 'bg-sea-ink-soft/40'
        }`}
      />
      <span className={active ? 'text-green-700' : 'text-sea-ink-soft'}>
        {active ? '活跃' : '已删除'}
      </span>
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
