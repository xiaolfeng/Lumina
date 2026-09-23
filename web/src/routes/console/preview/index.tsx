import { useEffect, useMemo, useState } from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ExternalLink, Eye, Plus, Search, Settings, Trash2 } from 'lucide-react'
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
import { PageHeader } from '#/components/page-header'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useCurrentWorkspace } from '#/hooks/useCurrentWorkspace'
import { useProjectList } from '#/hooks/useProject'
import {
  useCreatePreviewSession,
  useDeletePreviewSession,
  usePreviewSessionList,
} from '#/hooks/usePreviewAdmin'
import { useDashboardOverview } from '#/hooks/useDashboard'
import { PreviewSessionDetailDrawer } from '#/components/preview/session-detail-drawer'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { formatDateTime } from '#/lib/format-date'
import type { PreviewSessionItem } from '#/lib/models/response/preview'
import { toast } from 'sonner'

export const Route = createFileRoute('/console/preview/')({
  staticData: { crumb: '预览管理' },
  component: PreviewPage,
})

function isPreviewExpired(session: PreviewSessionItem): boolean {
  if (session.status === 'deleted') return true
  if (!session.expires_at) return false
  return new Date(session.expires_at).getTime() <= Date.now()
}

function PreviewPage() {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [statusFilter, setStatusFilter] = useState<
    'all' | 'active' | 'expired'
  >('all')
  const [searchQuery, setSearchQuery] = useState('')

  const [deleteTarget, setDeleteTarget] = useState<PreviewSessionItem | null>(
    null,
  )
  const [viewTarget, setViewTarget] = useState<PreviewSessionItem | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [createTitle, setCreateTitle] = useState('')
  const [createProjectId, setCreateProjectId] = useState('')
  const [projectSearch, setProjectSearch] = useState('')

  const { current } = useCurrentWorkspace()

  // 🌟 Q-04 & R-02 切换工作空间时重置选中态与全部创建表单草稿
  useEffect(() => {
    setPage(1)
    setDeleteTarget(null)
    setViewTarget(null)
    setCreateOpen(false)
    setCreateTitle('')
    setCreateProjectId('')
    setProjectSearch('')
  }, [current?.id])

  const { data, isLoading } = usePreviewSessionList({
    page,
    size: pageSize,
    workspace_id: current?.id,
  })
  const showLoading = !current?.id || isLoading
  const { data: overviewData } = useDashboardOverview()

  // 🌟 服务端按关键词直接检索空间项目（彻底根除 Q-05 客户端分页截断上限）
  const { data: projectData } = useProjectList({
    page: 1,
    size: 50,
    workspace_id: current?.id,
    search: projectSearch.trim() || undefined,
  })
  const deleteMutation = useDeletePreviewSession()
  const createMutation = useCreatePreviewSession()

  const rawItems = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))
  const projects = projectData?.data?.items ?? []
  const previewStats = overviewData?.data?.preview

  // 本地根据状态与关键词过滤
  const items = useMemo(() => {
    return rawItems.filter((item) => {
      const expired = isPreviewExpired(item)
      if (statusFilter === 'active' && expired) return false
      if (statusFilter === 'expired' && !expired) return false

      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase().trim()
        const titleMatch = (item.title || '').toLowerCase().includes(q)
        const hashMatch = item.hash.toLowerCase().includes(q)
        if (!titleMatch && !hashMatch) return false
      }
      return true
    })
  }, [rawItems, statusFilter, searchQuery])

  // 新建弹窗直接使用后端检索返回的候选项目列表
  const filteredProjects = projects

  const handleCreate = () => {
    if (!createProjectId) {
      toast.error('请选择当前工作空间下的有效项目')
      return
    }
    createMutation.mutate(
      { project_id: createProjectId, title: createTitle.trim() || undefined },
      {
        onSuccess: () => {
          setCreateOpen(false)
          setCreateTitle('')
          setCreateProjectId('')
          setProjectSearch('')
        },
      },
    )
  }

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
      className="space-y-4"
    >
      <PageHeader
        title="预览管理"
        description="管理 HTML/CSS/JS 与 React 前端原型沙盒会话，支持路径式寻址、实时同步与快照晋升"
        action={
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              onClick={() =>
                void (navigate as any)({
                  to: '/console/settings',
                  search: { tab: 'preview', from: '/console/preview' },
                })
              }
              className="rounded-none border-line text-xs text-sea-ink hover:bg-chip-bg"
              title="跳转至系统设置配置 Preview 沙盒配额"
            >
              <Settings className="mr-1.5 size-3.5 text-lagoon" />
              Preview 沙盒配额配置
            </Button>
            <Button
              onClick={() => setCreateOpen(true)}
              className="bg-lagoon text-foam hover:bg-lagoon-deep rounded-none text-xs"
            >
              <Plus className="mr-1.5 size-3.5" />
              新建预览会话
            </Button>
          </div>
        }
      />

      {/* KPI 指标带 */}
      <motion.div variants={staggerItem}>
        <div className="grid grid-cols-1 border border-line bg-sand/30 sm:grid-cols-3">
          <div className="p-4 border-b border-line sm:border-b-0 sm:border-r">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              活跃沙盒 (ACTIVE)
            </span>
            <span className="font-mono text-xl font-semibold text-lagoon-deep">
              {previewStats?.active ?? 0}
            </span>
            <p className="text-[11px] text-sea-ink-soft mt-0.5">
              路径式独立沙盒可实时访问
            </p>
          </div>
          <div className="p-4 border-b border-line sm:border-b-0 sm:border-r">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              承载文件总数
            </span>
            <span className="font-mono text-xl font-semibold text-sea-ink">
              {previewStats?.files ?? 0}
            </span>
            <p className="text-[11px] text-sea-ink-soft mt-0.5">
              单文件上限 256KB · 扁平组织
            </p>
          </div>
          <div className="p-4">
            <span className="text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft block">
              全部历史会话
            </span>
            <span className="font-mono text-xl font-semibold text-sea-ink">
              {totalItems}
            </span>
            <p className="text-[11px] text-sea-ink-soft mt-0.5">
              当前工作空间会话总量
            </p>
          </div>
        </div>
      </motion.div>

      {/* 工具栏与筛选胶囊 */}
      <motion.div variants={staggerItem} className="space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line pb-3">
          {/* 状态 Filter 胶囊 */}
          <div className="flex items-center border border-line bg-foam p-0.5">
            <button
              type="button"
              onClick={() => setStatusFilter('all')}
              className={`px-2.5 py-1 text-xs transition-colors ${
                statusFilter === 'all'
                  ? 'bg-lagoon text-foam font-semibold'
                  : 'text-sea-ink-soft hover:text-sea-ink'
              }`}
            >
              全部会话
            </button>
            <button
              type="button"
              onClick={() => setStatusFilter('active')}
              className={`px-2.5 py-1 text-xs transition-colors ${
                statusFilter === 'active'
                  ? 'bg-lagoon text-foam font-semibold'
                  : 'text-sea-ink-soft hover:text-sea-ink'
              }`}
            >
              活跃中
            </button>
            <button
              type="button"
              onClick={() => setStatusFilter('expired')}
              className={`px-2.5 py-1 text-xs transition-colors ${
                statusFilter === 'expired'
                  ? 'bg-sea-ink text-foam font-semibold'
                  : 'text-sea-ink-soft hover:text-sea-ink'
              }`}
            >
              已过期
            </button>
          </div>

          {/* 搜索框 */}
          <div className="flex items-center gap-2 border border-line bg-foam px-2.5 py-1 w-64 focus-within:border-lagoon">
            <Search className="size-3.5 text-sea-ink-soft shrink-0" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="过滤当前页会话标题或 Hash..."
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
            items.map((item) => {
              const expired = isPreviewExpired(item)
              return (
                <div
                  key={item.id}
                  className="flex items-center justify-between gap-4 border-b border-line px-4 py-3.5 transition-colors hover:bg-chip-bg last:border-b-0"
                >
                  <div className="min-w-0 flex-1 flex flex-col gap-1">
                    <div className="flex items-center gap-2">
                      <span
                        onClick={() => setViewTarget(item)}
                        className="cursor-pointer font-semibold text-sm text-sea-ink hover:text-lagoon transition-colors truncate max-w-lg"
                        title={item.title}
                      >
                        {item.title || '（未命名会话）'}
                      </span>
                      <span
                        className={`font-mono text-[10px] px-1.5 py-0.5 border ${
                          !expired
                            ? 'border-chip-line bg-lagoon/10 text-lagoon-deep font-semibold'
                            : 'border-line bg-chip-bg text-sea-ink-soft'
                        }`}
                      >
                        {!expired ? '活跃沙盒' : '已过期收回'}
                      </span>
                      <span className="font-mono text-[10px] bg-sand px-1.5 py-0.5 border border-line text-sea-ink-soft">
                        {item.hash.slice(0, 8)}
                      </span>
                    </div>

                    <div className="flex flex-wrap items-center gap-2 text-xs text-sea-ink-soft">
                      <span className="font-mono text-sea-ink bg-sand/60 px-1.5 py-0.2 border border-line">
                        {item.file_count} 个前端文件
                      </span>
                      <span>•</span>
                      <span>
                        {expired
                          ? `于 ${formatDateTime(item.expires_at)} 到期`
                          : `有效期至 ${formatDateTime(item.expires_at)}`}
                      </span>
                    </div>
                  </div>

                  <div className="flex shrink-0 items-center gap-1.5">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setViewTarget(item)}
                      className="h-7 text-xs rounded-none border-line px-2 text-sea-ink hover:bg-foam"
                      title="打开会话详情管理文件"
                    >
                      <Eye className="mr-1 size-3" />
                      文件详情
                    </Button>
                    {!expired ? (
                      <Button
                        size="sm"
                        onClick={() =>
                          window.open(
                            `/preview/${item.hash}/index.html`,
                            '_blank',
                          )
                        }
                        className="h-7 text-xs rounded-none bg-primary text-primary-foreground hover:bg-primary/90 px-2"
                        title="在独立窗口中打开预览工作台"
                      >
                        <ExternalLink className="mr-1 size-3" />
                        访问沙盒
                      </Button>
                    ) : null}
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
              )
            })
          ) : (
            <div className="border border-line px-8 py-16 text-center">
              <p className="text-sm font-semibold text-sea-ink">
                暂无匹配的预览会话
              </p>
              <p className="mt-1 text-xs text-sea-ink-soft">
                可通过上方「新建预览会话」按钮创建第一个前端原型沙盒。
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

      {/* 新建会话 Dialog (已升级搜索支持 Q-05) */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="border-line rounded-none sm:max-w-md bg-foam">
          <DialogHeader>
            <div className="flex items-center gap-2">
              <span className="size-2 rotate-45 bg-lagoon" />
              <DialogTitle className="text-base font-semibold">
                新建预览会话
              </DialogTitle>
            </div>
            <DialogDescription className="text-xs text-sea-ink-soft">
              选择所属项目并填写会话标题，创建后可通过 Hash
              进行路径式访问或直接上传文件。
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2 text-xs">
            <div className="space-y-1.5">
              <Label className="text-xs font-semibold text-sea-ink">
                关联项目 <span className="text-destructive">*</span>
              </Label>
              <Input
                value={projectSearch}
                onChange={(e) => setProjectSearch(e.target.value)}
                placeholder="键入关键词过滤项目..."
                className="h-7 text-xs bg-sand/30 font-mono mb-1"
              />
              <Select
                value={createProjectId}
                onValueChange={setCreateProjectId}
              >
                <SelectTrigger className="w-full text-xs bg-sand/30">
                  <SelectValue placeholder="从筛选结果中选择项目" />
                </SelectTrigger>
                <SelectContent className="max-h-48">
                  {filteredProjects.map((p) => (
                    <SelectItem key={p.id} value={p.id} className="text-xs">
                      {p.name} {p.alias_name ? `(${p.alias_name})` : ''}
                    </SelectItem>
                  ))}
                  {filteredProjects.length === 0 && (
                    <div className="p-2 text-center text-xs text-sea-ink-soft">
                      未找到匹配的项目
                    </div>
                  )}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1.5">
              <Label className="text-xs font-semibold text-sea-ink">
                会话标题
              </Label>
              <Input
                value={createTitle}
                onChange={(e) => setCreateTitle(e.target.value)}
                placeholder="如：LPW 文档大纲原型 (留空将自动生成)"
                className="text-xs bg-sand/30"
              />
            </div>
          </div>

          <DialogFooter className="border-t border-line pt-3">
            <Button
              variant="outline"
              size="sm"
              className="rounded-none text-xs"
              onClick={() => setCreateOpen(false)}
            >
              取消
            </Button>
            <Button
              size="sm"
              onClick={handleCreate}
              disabled={!createProjectId || createMutation.isPending}
              className="rounded-none bg-primary text-primary-foreground hover:bg-primary/90 text-xs"
            >
              {createMutation.isPending ? '创建中...' : '立即创建'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 删除确认对话框 */}
      <ConfirmDeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null)
        }}
        title="确认删除预览会话"
        description={`确定要删除会话「${deleteTarget?.title || deleteTarget?.hash.slice(0, 8)}」吗？删除后沙盒内的前端文件将被清空。`}
        onConfirm={() => {
          if (deleteTarget) {
            deleteMutation.mutate(deleteTarget.id, {
              onSuccess: () => setDeleteTarget(null),
            })
          }
        }}
        isPending={deleteMutation.isPending}
      />

      {/* 🌟 详情抽屉（带 REST 兜底与文件删除确认） */}
      <PreviewSessionDetailDrawer
        session={viewTarget}
        onClose={() => setViewTarget(null)}
      />
    </motion.div>
  )
}
