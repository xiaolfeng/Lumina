import { useEffect, useMemo, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Archive, ExternalLink, Eye, Search, Shield } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Skeleton } from '@lumina/components/ui/skeleton'
import { PageHeader } from '#/components/page-header'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useCurrentWorkspace } from '#/hooks/useCurrentWorkspace'
import { useArchivePage, usePageList } from '#/hooks/usePages'
import { PageDetailSheet } from '#/components/pages/page-detail-sheet'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { formatDateTime } from '#/lib/format-date'
import type { PageItem } from '#/lib/models/response/pages'

export const Route = createFileRoute('/console/pages/')({
  staticData: { crumb: '页面管理' },
  component: ConsolePagesPage,
})

function ConsolePagesPage() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [tab, setTab] = useState<'published' | 'archived'>('published')
  const [searchQuery, setSearchQuery] = useState('')

  const [archiveTarget, setArchiveTarget] = useState<PageItem | null>(null)
  const [detailTarget, setDetailTarget] = useState<PageItem | null>(null)
  const [detailMode, setDetailMode] = useState<'view' | 'policy'>('view')

  const { current } = useCurrentWorkspace()

  // 🌟 Q-04 切换工作空间时重置选中态
  useEffect(() => {
    setPage(1)
    setArchiveTarget(null)
    setDetailTarget(null)
  }, [current?.id])

  const { data, isLoading } = usePageList({
    page,
    size: pageSize,
    workspace_id: current?.id,
    status: tab,
  })
  const archive = useArchivePage()

  const rawItems = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

  // 客户端检索过滤
  const items = useMemo(() => {
    if (!searchQuery.trim()) return rawItems
    const q = searchQuery.toLowerCase().trim()
    return rawItems.filter(
      (item) =>
        (item.title && item.title.toLowerCase().includes(q)) ||
        item.slug.toLowerCase().includes(q) ||
        (item.project_name && item.project_name.toLowerCase().includes(q)),
    )
  }, [rawItems, searchQuery])

  const handleTabChange = (nextTab: 'published' | 'archived') => {
    setTab(nextTab)
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
        title="即时页面管理"
        description="管理从前端原型沙盒晋升的不可变即时页面快照、版本生效指针与安全访问策略"
      />

      {/* ─── KPI 指标带 ─── */}
      <motion.div variants={staggerItem}>
        <div className="grid grid-cols-2 md:grid-cols-4 border border-line bg-sand/30">
          <div className="p-4 border-r border-line">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              当前视图记录
            </span>
            <span className="font-mono text-xl font-semibold text-lagoon-deep">
              {totalItems}
            </span>
            <p className="text-[11px] text-sea-ink-soft mt-0.5">
              {tab === 'published' ? '对外开放访问中' : '历史收纳归档箱'}
            </p>
          </div>
          <div className="p-4 border-r border-line">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              受密码门保护 (当页)
            </span>
            <span className="font-mono text-xl font-semibold text-sea-ink">
              {items.filter((i) => i.access_mode === 'password').length}
            </span>
            <p className="text-[11px] text-sea-ink-soft mt-0.5">
              受 HMAC 安全口令守护
            </p>
          </div>
          <div className="p-4 border-r border-line">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              当前筛选分类
            </span>
            <span className="text-sm font-semibold text-sea-ink block mt-1">
              {tab === 'published' ? '线上已发布页面' : '归档收纳存储'}
            </span>
          </div>
          <div className="p-4">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              每页分页容量
            </span>
            <span className="font-mono text-xl font-semibold text-sea-ink">
              {pageSize}
            </span>
            <p className="text-[11px] text-sea-ink-soft mt-0.5">条记录 / 页</p>
          </div>
        </div>
      </motion.div>

      {/* ─── 工具栏与 Tab 分类 ─── */}
      <motion.div variants={staggerItem} className="space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line pb-3">
          <div className="flex items-center border border-line bg-foam p-0.5">
            <button
              type="button"
              onClick={() => handleTabChange('published')}
              className={`px-3 py-1 text-xs transition-colors flex items-center gap-1.5 ${
                tab === 'published'
                  ? 'bg-lagoon text-foam font-semibold'
                  : 'text-sea-ink-soft hover:text-sea-ink'
              }`}
            >
              <span>已发布页面</span>
              <span className="font-mono text-[10px] opacity-80">
                {tab === 'published' ? totalItems : ''}
              </span>
            </button>
            <button
              type="button"
              onClick={() => handleTabChange('archived')}
              className={`px-3 py-1 text-xs transition-colors flex items-center gap-1.5 ${
                tab === 'archived'
                  ? 'bg-sea-ink text-foam font-semibold'
                  : 'text-sea-ink-soft hover:text-sea-ink'
              }`}
            >
              <Archive className="size-3" />
              <span>归档收纳箱</span>
              <span className="font-mono text-[10px] opacity-80">
                {tab === 'archived' ? totalItems : ''}
              </span>
            </button>
          </div>

          <div className="flex items-center gap-2 border border-line bg-foam px-2.5 py-1 w-64 focus-within:border-lagoon">
            <Search className="size-3.5 text-sea-ink-soft shrink-0" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="过滤当前页 Slug 或项目..."
              className="h-6 border-0 bg-transparent p-0 text-xs shadow-none focus-visible:ring-0"
            />
          </div>
        </div>

        {/* ─── 页面台账列表 ─── */}
        <div className="border border-line bg-surface">
          {isLoading ? (
            <div className="space-y-2 p-4">
              <Skeleton className="h-14 w-full" />
              <Skeleton className="h-14 w-full" />
              <Skeleton className="h-14 w-full" />
            </div>
          ) : items.length === 0 ? (
            <div className="border-b border-line px-8 py-16 text-center">
              <p className="text-sm font-semibold text-sea-ink">
                {tab === 'published'
                  ? '暂无已发布的即时页面'
                  : '归档收纳箱为空'}
              </p>
              <p className="mt-1 text-xs text-sea-ink-soft">
                {tab === 'published'
                  ? '从 Preview 预览工作台核验通过后晋升即可呈现在此。'
                  : '已下线的历史页面将被安全收纳于此，不影响线上服务。'}
              </p>
            </div>
          ) : (
            items.map((item) => {
              const isPassword = item.access_mode === 'password'
              const publicUrl = `/pages/${item.project_name || item.project_id}/${item.slug}`

              return (
                <div
                  key={item.id}
                  className="flex items-center justify-between gap-4 border-b border-line px-4 py-3.5 transition-colors hover:bg-chip-bg last:border-b-0"
                >
                  <div className="min-w-0 flex-1 flex flex-col gap-1">
                    <div className="flex items-center gap-2">
                      <span
                        onClick={() => {
                          setDetailTarget(item)
                          setDetailMode('view')
                        }}
                        className="cursor-pointer font-semibold text-sm text-sea-ink hover:text-lagoon transition-colors truncate max-w-lg"
                        title={item.title || item.slug}
                      >
                        {item.title || item.slug}
                      </span>
                      <span
                        className={`font-mono text-[10px] px-1.5 py-0.5 border ${
                          item.status === 'published'
                            ? 'border-chip-line bg-lagoon/10 text-lagoon-deep font-semibold'
                            : 'border-line bg-chip-bg text-sea-ink-soft'
                        }`}
                      >
                        {item.status === 'published' ? '已上线' : '已归档'}
                      </span>
                      <span
                        className={`font-mono text-[10px] px-1.5 py-0.5 border flex items-center gap-1 ${
                          isPassword
                            ? 'border-destructive/30 bg-destructive/10 text-destructive'
                            : 'border-line bg-sand text-sea-ink-soft'
                        }`}
                      >
                        {isPassword && <Shield className="size-2.5" />}
                        {isPassword ? '密码门' : '公开'}
                      </span>
                    </div>

                    <div className="flex flex-wrap items-center gap-2 text-xs text-sea-ink-soft">
                      <span className="font-mono text-sea-ink bg-sand/60 px-1.5 py-0.2 border border-line">
                        /{item.project_name || item.project_id}/{item.slug}/
                      </span>
                      <span>•</span>
                      <span className="font-mono text-lagoon-deep font-medium">
                        版本 {item.latest_version || 'v1'}
                      </span>
                      <span>•</span>
                      <span>更新于 {formatDateTime(item.updated_at)}</span>
                    </div>
                  </div>

                  <div className="flex shrink-0 items-center gap-1.5">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setDetailTarget(item)
                        setDetailMode('view')
                      }}
                      className="h-7 text-xs rounded-none border-line px-2 text-sea-ink hover:bg-foam"
                      title="打开页面全景抽屉查看版本与资产"
                    >
                      <Eye className="mr-1 size-3" />
                      全景详情
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setDetailTarget(item)
                        setDetailMode('policy')
                      }}
                      className="h-7 text-xs rounded-none border-line px-2 text-sea-ink hover:bg-foam"
                      title="就地调整安全与密码门策略"
                    >
                      <Shield className="mr-1 size-3 text-lagoon" />
                      策略
                    </Button>
                    {item.status === 'published' ? (
                      <>
                        <Button
                          size="sm"
                          onClick={() => window.open(publicUrl, '_blank')}
                          className="h-7 text-xs rounded-none bg-primary text-primary-foreground hover:bg-primary/90 px-2"
                          title="在新窗口中访问页面"
                        >
                          <ExternalLink className="mr-1 size-3" />
                          访问
                        </Button>
                        <button
                          type="button"
                          onClick={() => setArchiveTarget(item)}
                          className="grid size-7 place-items-center text-sea-ink-soft transition-colors hover:bg-destructive/10 hover:text-destructive border border-transparent hover:border-destructive/30"
                          title="收纳归档此页面"
                        >
                          <Archive className="size-3.5" />
                        </button>
                      </>
                    ) : null}
                  </div>
                </div>
              )
            })
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

      {/* 归档确认弹窗 */}
      <ConfirmDeleteDialog
        open={!!archiveTarget}
        onOpenChange={(open) => {
          if (!open) setArchiveTarget(null)
        }}
        title="确认归档页面"
        description={`确定归档「${archiveTarget?.title || archiveTarget?.slug}」吗？归档后对外公开路径将关闭，页面将移入「归档收纳箱」。`}
        onConfirm={() => {
          if (archiveTarget) {
            archive.mutate(archiveTarget.id, {
              onSuccess: () => setArchiveTarget(null),
            })
          }
        }}
        isPending={archive.isPending}
      />

      {/* 🌟 即时页面全景资产抽屉 (包含版本快照时间线、在位策略编辑与 Fork) */}
      <PageDetailSheet
        open={!!detailTarget}
        onOpenChange={(open) => {
          if (!open) setDetailTarget(null)
        }}
        item={detailTarget}
        initialMode={detailMode}
      />
    </motion.div>
  )
}
