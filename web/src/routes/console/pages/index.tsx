import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Archive, ExternalLink, Shield } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { Skeleton } from '@lumina/components/ui/skeleton'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@lumina/components/ui/dialog'
import { PageHeader } from '#/components/page-header'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useCurrentWorkspace } from '#/hooks/useCurrentWorkspace'
import {
  useArchivePage,
  usePageList,
  useUpdateAccessPolicy,
} from '#/hooks/usePages'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { formatDateTime } from '#/lib/format-date'
import type { PageItem } from '#/lib/models/response/pages'

export const Route = createFileRoute('/console/pages/')({
  staticData: { crumb: '页面' },
  component: ConsolePagesPage,
})

function ConsolePagesPage() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [tab, setTab] = useState<'published' | 'archived'>('published')
  const [archiveTarget, setArchiveTarget] = useState<PageItem | null>(null)
  const [policyTarget, setPolicyTarget] = useState<PageItem | null>(null)
  const [accessMode, setAccessMode] = useState<'public' | 'password'>('public')
  const [password, setPassword] = useState('')
  const { current } = useCurrentWorkspace()
  const { data, isLoading } = usePageList({
    page,
    size: pageSize,
    workspace_id: current?.id,
  })
  const archive = useArchivePage()
  const updatePolicy = useUpdateAccessPolicy()
  const items = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

  const publishedItems = items.filter((item) => item.status === 'published')
  const archivedItems = items.filter((item) => item.status === 'archived')
  const currentItems = tab === 'published' ? publishedItems : archivedItems

  const publishedCount = publishedItems.length
  const archivedCount = archivedItems.length
  const protectedCount = publishedItems.filter(
    (item) => item.access_mode === 'password',
  ).length

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
      className="mx-auto max-w-5xl space-y-8"
    >
      <PageHeader
        title="页面管理"
        description="管理已发布的即时页面、版本生效指针与访问策略。密码保护只在这里设置。"
      />

      {/* ─── KPI band ─── */}
      <motion.div variants={staggerItem}>
        <div className="grid grid-cols-2 md:grid-cols-4 border-b border-line">
          <Kpi
            label="线上已发布"
            value={publishedCount}
            delta="对外开放访问中"
            loading={isLoading}
          />
          <Kpi
            label="密码保护"
            value={protectedCount}
            delta="受口令安全策略保护"
            loading={isLoading}
          />
          <Kpi
            label="归档收纳"
            value={archivedCount}
            delta="已停用并收纳历史页"
            loading={isLoading}
          />
          <Kpi
            label="总记录数"
            value={totalItems}
            delta="当前空间全部页面"
            loading={isLoading}
            last
          />
        </div>
      </motion.div>

      {/* ─── 页面列表区域（方案 A：Tab 隔离收纳） ─── */}
      <motion.div variants={staggerItem} className="space-y-4">
        <div className="flex items-center justify-between border-b border-line">
          <div className="-mb-px flex items-center gap-2">
            <button
              type="button"
              onClick={() => setTab('published')}
              className={`flex items-center gap-2 border-b-2 px-4 py-3 text-sm font-semibold transition-colors ${
                tab === 'published'
                  ? 'border-lagoon text-sea-ink'
                  : 'border-transparent text-sea-ink-soft hover:text-sea-ink'
              }`}
            >
              <span>已发布页面</span>
              <span
                className={`px-1.5 py-0.5 text-xs font-mono font-bold ${
                  tab === 'published'
                    ? 'bg-lagoon text-foam'
                    : 'border border-line bg-chip-bg text-sea-ink-soft'
                }`}
              >
                {publishedCount}
              </span>
            </button>
            <button
              type="button"
              onClick={() => setTab('archived')}
              className={`flex items-center gap-2 border-b-2 px-4 py-3 text-sm font-semibold transition-colors ${
                tab === 'archived'
                  ? 'border-lagoon text-sea-ink'
                  : 'border-transparent text-sea-ink-soft hover:text-sea-ink'
              }`}
            >
              <Archive className="size-3.5" />
              <span>归档收纳箱</span>
              <span
                className={`px-1.5 py-0.5 text-xs font-mono font-bold ${
                  tab === 'archived'
                    ? 'bg-lagoon text-foam'
                    : 'border border-line bg-chip-bg text-sea-ink-soft'
                }`}
              >
                {archivedCount}
              </span>
            </button>
          </div>
          <span className="text-xs text-sea-ink-soft">
            共 {items.length} 个页面记录
          </span>
        </div>

        <div>
          {isLoading ? (
            <div className="space-y-2 py-4">
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
            </div>
          ) : currentItems.length === 0 ? (
            <div className="border border-line px-8 py-16 text-center">
              <p className="display-title text-xl font-semibold text-sea-ink">
                {tab === 'published' ? '暂无已发布页面' : '归档箱为空'}
              </p>
              <p className="mt-2.5 text-sm text-sea-ink-soft">
                {tab === 'published'
                  ? '从 Preview 工作台晋升后会出现在这里。'
                  : '已下线或不再对外展示的页面会被收纳在这里，不占用线上主列表展示。'}
              </p>
            </div>
          ) : (
            currentItems.map((item) => (
              <div
                key={item.id}
                className="flex items-center gap-4 border-b border-line px-1 py-4 transition-colors hover:bg-chip-bg"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2.5">
                    <p className="truncate text-sm font-semibold text-sea-ink">
                      {item.title}
                    </p>
                    {item.status === 'archived' ? (
                      <span className="inline-flex items-center border border-line bg-chip-bg px-1.5 py-0.5 text-[11px] font-medium text-sea-ink-soft">
                        已收纳
                      </span>
                    ) : item.access_mode === 'password' ? (
                      <span className="inline-flex items-center gap-1 border border-chip-line bg-lagoon/10 px-1.5 py-0.5 text-[11px] font-medium text-lagoon-deep">
                        <Shield className="size-3" />
                        密码保护
                      </span>
                    ) : (
                      <span className="inline-flex items-center border border-line bg-chip-bg px-1.5 py-0.5 text-[11px] font-medium text-sea-ink-soft">
                        公开访问
                      </span>
                    )}
                  </div>
                  <div className="mt-1 flex flex-wrap items-center gap-2.5 text-xs text-sea-ink-soft">
                    <span className="border border-line bg-surface px-1.5 py-0.5 font-mono text-sea-ink">
                      /{item.project_name}/{item.slug}/
                    </span>
                    <span className="font-mono font-semibold text-lagoon-deep">
                      {item.latest_version}
                    </span>
                    <span>·</span>
                    <span>更新于 {formatDateTime(item.updated_at)}</span>
                  </div>
                </div>

                {item.status === 'published' ? (
                  <>
                    {/* page_url 的域名在 site.domain 未配置时回退 localhost，跨部署不可靠；
                        改用相对裸路径，入口由展示页按 version.entry_filename 兜底解析 */}
                    <a
                      href={`/pages/${item.project_name}/${item.slug}/`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="grid size-11 place-items-center text-sea-ink-soft transition-colors hover:text-sea-ink"
                      title="在新标签页打开访问地址"
                      aria-label={`打开 ${item.title} 线上页面`}
                    >
                      <ExternalLink className="size-3.5" />
                    </a>
                    <button
                      type="button"
                      className="grid size-11 place-items-center text-sea-ink-soft transition-colors hover:text-sea-ink"
                      onClick={() => {
                        setPolicyTarget(item)
                        setAccessMode(item.access_mode)
                        setPassword('')
                      }}
                      title="配置安全与访问策略"
                      aria-label={`配置 ${item.title} 安全策略`}
                    >
                      <Shield className="size-3.5" />
                    </button>
                    <button
                      type="button"
                      onClick={() => setArchiveTarget(item)}
                      className="grid size-11 place-items-center text-sea-ink-soft transition-colors hover:text-destructive"
                      title="收纳归档此页面"
                      aria-label={`归档 ${item.title}`}
                    >
                      <Archive className="size-3.5" />
                    </button>
                  </>
                ) : (
                  <div className="flex items-center gap-2 pr-2">
                    <span className="text-xs text-sea-ink-soft">
                      已归档收纳
                    </span>
                  </div>
                )}
              </div>
            ))
          )}
        </div>

        {totalItems > 0 ? (
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
        ) : null}
      </motion.div>

      <ConfirmDeleteDialog
        open={!!archiveTarget}
        onOpenChange={(open) => {
          if (!open) setArchiveTarget(null)
        }}
        title="归档页面"
        description={`确定归档「${archiveTarget?.title}」？归档后对外路径将不可访问，页面将移入「归档收纳箱」。`}
        onConfirm={() => {
          if (archiveTarget) archive.mutate(archiveTarget.id)
          setArchiveTarget(null)
        }}
        isPending={archive.isPending}
      />

      <Dialog
        open={!!policyTarget}
        onOpenChange={(open) => {
          if (!open) setPolicyTarget(null)
        }}
      >
        <DialogContent className="border-line sm:max-w-md">
          <DialogHeader>
            <DialogTitle>页面安全设置</DialogTitle>
            <DialogDescription>
              访问策略只在管理员端口配置，不会出现在晋升弹窗里。
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3 py-2">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                checked={accessMode === 'public'}
                onChange={() => setAccessMode('public')}
              />
              公开访问
            </label>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                checked={accessMode === 'password'}
                onChange={() => setAccessMode('password')}
              />
              密码保护
            </label>
            {accessMode === 'password' ? (
              <div className="space-y-1">
                <Label>访问密码</Label>
                <Input
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                />
              </div>
            ) : null}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPolicyTarget(null)}>
              取消
            </Button>
            <Button
              disabled={
                updatePolicy.isPending ||
                (accessMode === 'password' && !password)
              }
              onClick={() => {
                if (!policyTarget) return
                updatePolicy.mutate({
                  id: policyTarget.id,
                  access_mode: accessMode,
                  password: accessMode === 'password' ? password : undefined,
                })
                setPolicyTarget(null)
              }}
            >
              保存生效
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
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
      {delta && <p className="mt-2 text-[11.5px] text-sea-ink-soft">{delta}</p>}
    </div>
  )
}
