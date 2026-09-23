import { useEffect, useMemo, useState } from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { BookOpen, Eye, Pencil, Plus, Search, Trash2 } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { Skeleton } from '@lumina/components/ui/skeleton'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@lumina/components/ui/tooltip'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useProjectList, useDeleteProject } from '#/hooks/useProject'
import { useCurrentWorkspace } from '#/hooks/useCurrentWorkspace'
import { usePinList } from '#/hooks/usePin'
import { useRepoWikiConfigs } from '#/hooks/useRepoWiki'
import { usePreviewSessionList } from '#/hooks/usePreviewAdmin'
import { CreateDialog } from '#/components/project/create-dialog'
import { EditDialog } from '#/components/project/edit-dialog'
import { ProjectDetailSheet } from '#/components/project/detail-sheet'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { PageHeader } from '#/components/page-header'
import { formatDate } from '#/lib/format-date'
import type { ProjectItem } from '#/lib/models/response/project'
import type { RepoWikiConfigItem } from '#/lib/models/response/repowiki'
import type { PreviewSessionItem } from '#/lib/models/response/preview'
import type { PinItem } from '#/lib/models/response/pin'

export const Route = createFileRoute('/console/project/')({
  component: ProjectPage,
})

function ProjectRow({
  item,
  index,
  page,
  pageSize,
  wikiConfig,
  pendingPinCount,
  previewSession,
  onView,
  onEdit,
  onDelete,
  onOpenWiki,
}: {
  item: ProjectItem
  index: number
  page: number
  pageSize: number
  wikiConfig?: RepoWikiConfigItem
  pendingPinCount: number
  previewSession?: PreviewSessionItem
  onView: (item: ProjectItem) => void
  onEdit: (item: ProjectItem) => void
  onDelete: (item: ProjectItem) => void
  onOpenWiki: (item: ProjectItem) => void
}) {
  const seq = String(index + 1 + (page - 1) * pageSize).padStart(2, '0')
  // 🌟 外层展示别名；无别名时明确标记未设置别名，绝不暴露真实名称
  const outerAlias = item.alias_name || '未设置别名'
  const matchPathText =
    item.match_path && item.match_path.length > 0
      ? item.match_path.join(', ')
      : '未配置匹配路径'

  const hasWiki = Boolean(wikiConfig?.selected_version_id || wikiConfig?.latest_version)
  const isPreviewLive = Boolean(previewSession && previewSession.status === 'active')

  return (
    <tr className="border-b border-line transition-colors hover:bg-sand/70">
      {/* 序号列 */}
      <td className="w-16 px-4 py-3.5 font-mono text-xs font-bold text-lagoon-deep">
        {seq}
      </td>

      {/* 项目别名与 Tooltip：外层仅展示别名，Tooltip 浮窗展示实际全名与介绍 */}
      <td className="w-72 px-4 py-3.5">
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="inline-flex cursor-pointer items-center gap-2 border-b border-dashed border-lagoon/60 pb-0.5 transition-opacity hover:opacity-80">
              <span className="size-1.5 shrink-0 rotate-45 bg-lagoon" />
              <span
                className={`font-mono text-sm font-semibold ${
                  item.alias_name ? 'text-sea-ink' : 'text-sea-ink-soft italic'
                }`}
              >
                {outerAlias}
              </span>
            </div>
          </TooltipTrigger>
          <TooltipContent
            side="bottom"
            align="start"
            sideOffset={8}
            className="w-84 max-w-sm rounded-none border border-sea-ink border-t-[3px] border-t-lagoon bg-foam p-3.5 text-sea-ink shadow-xl"
          >
            <div className="flex flex-col gap-2">
              <div className="flex flex-col gap-0.5">
                <span className="text-[10px] font-bold uppercase tracking-wider text-lagoon-deep">
                  实际项目名称
                </span>
                <div className="flex items-center justify-between gap-2">
                  <span className="text-sm font-semibold text-sea-ink">
                    {item.name}
                  </span>
                  {item.alias_name ? (
                    <span className="font-mono text-[10px] text-sea-ink-soft bg-sand border border-line px-1.5 py-0.5">
                      别名: {item.alias_name}
                    </span>
                  ) : null}
                </div>
              </div>

              <div className="flex flex-col gap-0.5">
                <span className="text-[10px] font-bold uppercase tracking-wider text-lagoon-deep">
                  项目介绍
                </span>
                <p className="text-xs text-sea-ink-soft leading-relaxed">
                  {item.description || '暂无详细介绍'}
                </p>
              </div>

              <div className="flex flex-col gap-0.5">
                <span className="text-[10px] font-bold uppercase tracking-wider text-lagoon-deep">
                  底层路径 (MatchPath)
                </span>
                <div className="font-mono text-[11px] text-sea-ink-soft bg-sand p-1.5 border border-line break-all">
                  {matchPathText}
                </div>
              </div>
            </div>
          </TooltipContent>
        </Tooltip>
      </td>

      {/* 核心关联资产徽章 (真实关联动态呈现) */}
      <td className="px-4 py-3.5">
        <div className="flex flex-wrap items-center gap-1.5">
          <span
            className={`inline-flex items-center gap-1.5 border border-line px-2 py-0.5 font-mono text-[11px] ${
              hasWiki
                ? 'bg-[#edf6ee] border-[#2e6930]/30 text-[#245b26] font-semibold'
                : 'bg-sand text-sea-ink-soft'
            }`}
          >
            <span
              className={`size-1.5 rounded-full ${
                hasWiki ? 'bg-[#2e6930]' : 'bg-sea-ink-soft'
              }`}
            />
            {hasWiki ? 'Wiki: 已就绪' : 'Wiki: 待生成'}
          </span>
          <span
            className={`inline-flex items-center gap-1.5 border border-line px-2 py-0.5 font-mono text-[11px] ${
              pendingPinCount > 0
                ? 'bg-[#fff4e6] border-lagoon/40 text-lagoon-deep font-semibold'
                : 'bg-sand text-sea-ink-soft'
            }`}
          >
            <span
              className={`size-1.5 rounded-full ${
                pendingPinCount > 0 ? 'bg-lagoon' : 'bg-sea-ink-soft'
              }`}
            />
            {pendingPinCount > 0 ? `Pin: ${pendingPinCount} 待办` : 'Pin: 0 待办'}
          </span>
          <span
            className={`inline-flex items-center gap-1.5 border border-line px-2 py-0.5 font-mono text-[11px] ${
              isPreviewLive
                ? 'bg-[#f7f3ed] border-chip-line text-sea-ink font-semibold'
                : 'bg-sand text-sea-ink-soft'
            }`}
          >
            <span
              className={`size-1.5 rounded-full ${
                isPreviewLive ? 'bg-[#2e6930]' : 'bg-sea-ink-soft'
              }`}
            />
            {isPreviewLive ? 'Preview: LIVE' : 'Preview: 无'}
          </span>
        </div>
      </td>

      {/* 更新时间 */}
      <td className="w-36 px-4 py-3.5 whitespace-nowrap text-xs text-sea-ink-soft">
        {formatDate(item.updated_at)}
      </td>

      {/* 操作列：紧凑 Lucide 图标按钮组 */}
      <td className="w-48 px-4 py-3.5 text-right whitespace-nowrap">
        <div className="inline-flex items-center gap-1">
          <button
            type="button"
            className="grid size-8 place-items-center border border-line bg-foam text-sea-ink-soft transition-colors hover:border-lagoon hover:bg-chip-bg hover:text-lagoon-deep"
            aria-label={`查看 ${item.name} 详情`}
            title="查看项目详情概览抽屉"
            onClick={() => onView(item)}
          >
            <Eye className="size-4" />
          </button>
          <button
            type="button"
            className="grid size-8 place-items-center border border-line bg-foam text-sea-ink-soft transition-colors hover:border-lagoon hover:bg-lagoon hover:text-foam"
            aria-label={`打开 ${item.name} Wiki`}
            title="打开 RepoWiki 知识库"
            onClick={() => onOpenWiki(item)}
          >
            <BookOpen className="size-4" />
          </button>
          <button
            type="button"
            className="grid size-8 place-items-center border border-line bg-foam text-sea-ink-soft transition-colors hover:border-lagoon hover:bg-chip-bg hover:text-lagoon-deep"
            aria-label={`编辑 ${item.name}`}
            title="编辑项目配置"
            onClick={() => onEdit(item)}
          >
            <Pencil className="size-3.5" />
          </button>
          <button
            type="button"
            className="grid size-8 place-items-center border border-line bg-foam text-sea-ink-soft transition-colors hover:border-destructive hover:bg-destructive/10 hover:text-destructive"
            aria-label={`删除 ${item.name}`}
            title="删除项目"
            onClick={() => onDelete(item)}
          >
            <Trash2 className="size-3.5" />
          </button>
        </div>
      </td>
    </tr>
  )
}

function ProjectPage() {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [searchTerm, setSearchTerm] = useState('')
  const [filterMode, setFilterMode] = useState<'all' | 'wiki' | 'pin'>('all')
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [detailOpen, setDetailOpen] = useState(false)
  const [selectedItem, setSelectedItem] = useState<ProjectItem | null>(null)

  const { current } = useCurrentWorkspace()
  const { data, isLoading } = useProjectList({
    page,
    size: pageSize,
    workspace_id: current?.id,
  })
  const { data: pinData } = usePinList({
    workspace_id: current?.id,
    status: 'pending',
  })
  const { data: wikiConfigsData } = useRepoWikiConfigs()
  const { data: previewData } = usePreviewSessionList({
    workspace_id: current?.id,
  })

  const showLoading = !current?.id || isLoading
  const deleteMutation = useDeleteProject()

  const rawItems = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

  // 映射各项目的关联资产状态
  const wikiConfigMap = useMemo(() => {
    const map = new Map<string, RepoWikiConfigItem>()
    wikiConfigsData?.data?.items.forEach((c) => {
      map.set(c.project_id, c)
    })
    return map
  }, [wikiConfigsData])

  const pendingPinCountMap = useMemo(() => {
    const map = new Map<string, number>()
    pinData?.data?.items.forEach((p: PinItem) => {
      if (p.to_project_id) {
        map.set(p.to_project_id, (map.get(p.to_project_id) || 0) + 1)
      }
    })
    return map
  }, [pinData])

  const previewSessionMap = useMemo(() => {
    const map = new Map<string, PreviewSessionItem>()
    previewData?.data?.items.forEach((s) => {
      if (s.project_id && !map.has(s.project_id)) {
        map.set(s.project_id, s)
      }
    })
    return map
  }, [previewData])

  // 客户端辅助搜索与分类胶囊过滤
  const items = useMemo(() => {
    let list = rawItems
    if (filterMode === 'wiki') {
      list = list.filter((item) => {
        const c = wikiConfigMap.get(item.id)
        return Boolean(c?.selected_version_id || c?.latest_version)
      })
    } else if (filterMode === 'pin') {
      list = list.filter((item) => (pendingPinCountMap.get(item.id) || 0) > 0)
    }

    if (!searchTerm.trim()) return list
    const term = searchTerm.toLowerCase().trim()
    return list.filter(
      (item) =>
        item.name.toLowerCase().includes(term) ||
        (item.alias_name && item.alias_name.toLowerCase().includes(term)),
    )
  }, [rawItems, filterMode, searchTerm, wikiConfigMap, pendingPinCountMap])

  // 统计当前空间真实的 KPI
  const kpiProjects = totalItems
  const kpiPendingPins = pinData?.data?.total ?? 0
  const kpiActivePreviews = previewData?.data?.total ?? 0
  const kpiWikiVersions = useMemo(() => {
    let count = 0
    wikiConfigsData?.data?.items.forEach((c) => {
      if (c.selected_version_id || c.latest_version) count++
    })
    return count
  }, [wikiConfigsData])

  useEffect(() => {
    setPage(1)
    setEditOpen(false)
    setDeleteOpen(false)
    setDetailOpen(false)
    setSelectedItem(null)
  }, [current?.id])

  useEffect(() => {
    if (!showLoading && page > totalPages) setPage(totalPages)
  }, [showLoading, page, totalPages])

  return (
    <TooltipProvider delayDuration={150}>
      <motion.div
        className="space-y-5"
        initial="hidden"
        animate="visible"
        variants={staggerContainer}
      >
        <PageHeader
          title="项目管理"
          description="管理代码项目实体与 RepoWiki、Pin 约束与 Q&A 资产映射"
          action={
            <Button
              onClick={() => setCreateOpen(true)}
              className="rounded-none bg-lagoon text-foam hover:bg-lagoon-deep font-semibold"
            >
              <Plus className="mr-1.5 size-4" />
              创建项目
            </Button>
          }
        />

        {/* 顶部 4 项 KPI 概览带 */}
        <motion.div variants={staggerItem}>
          <div className="grid grid-cols-2 md:grid-cols-4 border border-line bg-foam">
            <div className="p-3.5 px-4.5 border-r border-b md:border-b-0 border-line flex flex-col gap-0.5">
              <div className="flex justify-between items-center text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft">
                <span>注册项目总数</span>
                <span className="font-mono">PRJ</span>
              </div>
              <span className="font-mono text-[22px] font-bold text-sea-ink mt-0.5">
                {kpiProjects} 个
              </span>
              <span className="text-[11px] text-[#2e6930]">
                ● 当前空间下代码实体
              </span>
            </div>
            <div className="p-3.5 px-4.5 border-b md:border-b-0 md:border-r border-line flex flex-col gap-0.5">
              <div className="flex justify-between items-center text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft">
                <span>待消费 Pin 队列</span>
                <span className="font-mono">FIFO</span>
              </div>
              <span className="font-mono text-[22px] font-bold text-sea-ink mt-0.5">
                {kpiPendingPins} 条
              </span>
              <span
                className={`text-[11px] ${
                  kpiPendingPins > 0
                    ? 'text-lagoon-deep font-semibold'
                    : 'text-sea-ink-soft'
                }`}
              >
                {kpiPendingPins > 0 ? '▲ 存在跨项目依赖阻断' : '● 无阻塞约束'}
              </span>
            </div>
            <div className="p-3.5 px-4.5 border-r border-line flex flex-col gap-0.5">
              <div className="flex justify-between items-center text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft">
                <span>Preview 活跃沙盒</span>
                <span className="font-mono">LIVE</span>
              </div>
              <span className="font-mono text-[22px] font-bold text-sea-ink mt-0.5">
                {kpiActivePreviews} 个
              </span>
              <span className="text-[11px] text-[#2e6930]">
                ● 实时 WebSocket 监听中
              </span>
            </div>
            <div className="p-3.5 px-4.5 flex flex-col gap-0.5">
              <div className="flex justify-between items-center text-[10.5px] font-bold uppercase tracking-wider text-sea-ink-soft">
                <span>Wiki 架构版本</span>
                <span className="font-mono">MDX</span>
              </div>
              <span className="font-mono text-[22px] font-bold text-sea-ink mt-0.5">
                {kpiWikiVersions} 份
              </span>
              <span className="text-[11px] text-[#2e6930]">
                ● 全量索引已就绪
              </span>
            </div>
          </div>
        </motion.div>

        {/* 过滤工具栏 */}
        <motion.div
          variants={staggerItem}
          className="flex flex-wrap items-center justify-between gap-4"
        >
          <div className="flex items-center gap-2 border border-line bg-foam px-3 h-9 w-80">
            <Search className="size-3.5 text-sea-ink-soft" />
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="按项目别名搜索过滤..."
              className="w-full bg-transparent text-xs text-sea-ink outline-none placeholder:text-sea-ink-soft"
            />
          </div>
          <div className="flex items-center gap-1.5">
            <button
              type="button"
              onClick={() => setFilterMode('all')}
              className={`px-3 py-1.5 text-xs border transition-colors ${
                filterMode === 'all'
                  ? 'bg-chip-bg border-lagoon text-lagoon-deep font-semibold'
                  : 'bg-foam border-line text-sea-ink-soft hover:text-sea-ink'
              }`}
            >
              全部 ({totalItems})
            </button>
            <button
              type="button"
              onClick={() => setFilterMode('wiki')}
              className={`px-3 py-1.5 text-xs border transition-colors ${
                filterMode === 'wiki'
                  ? 'bg-chip-bg border-lagoon text-lagoon-deep font-semibold'
                  : 'bg-foam border-line text-sea-ink-soft hover:text-sea-ink'
              }`}
            >
              带 Wiki ({kpiWikiVersions})
            </button>
            <button
              type="button"
              onClick={() => setFilterMode('pin')}
              className={`px-3 py-1.5 text-xs border transition-colors ${
                filterMode === 'pin'
                  ? 'bg-chip-bg border-lagoon text-lagoon-deep font-semibold'
                  : 'bg-foam border-line text-sea-ink-soft hover:text-sea-ink'
              }`}
            >
              带待办 Pin ({kpiPendingPins})
            </button>
          </div>
        </motion.div>

        {/* 台账表格 */}
        <motion.div variants={staggerItem} className="space-y-4">
          <div className="border border-line bg-foam overflow-x-auto">
            <table className="w-full border-collapse text-left text-xs">
              <thead>
                <tr className="border-b border-line bg-sand">
                  <th className="w-16 px-4 py-3 font-mono text-[10.5px] font-bold uppercase tracking-wider text-lagoon-deep">
                    序号
                  </th>
                  <th className="w-72 px-4 py-3 font-mono text-[10.5px] font-bold uppercase tracking-wider text-lagoon-deep">
                    项目标识 (悬浮查看实际名称与介绍)
                  </th>
                  <th className="px-4 py-3 font-mono text-[10.5px] font-bold uppercase tracking-wider text-lagoon-deep">
                    核心关联资产
                  </th>
                  <th className="w-36 px-4 py-3 font-mono text-[10.5px] font-bold uppercase tracking-wider text-lagoon-deep">
                    更新时间
                  </th>
                  <th className="w-48 px-4 py-3 text-right font-mono text-[10.5px] font-bold uppercase tracking-wider text-lagoon-deep">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody>
                {showLoading ? (
                  <tr>
                    <td colSpan={5} className="p-4">
                      <div className="flex flex-col gap-2">
                        <Skeleton className="h-10 w-full rounded-none" />
                        <Skeleton className="h-10 w-full rounded-none" />
                        <Skeleton className="h-10 w-full rounded-none" />
                      </div>
                    </td>
                  </tr>
                ) : items.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="py-14 text-center">
                      <p className="text-sm font-semibold text-sea-ink">
                        {searchTerm ? '无匹配项目' : '暂无项目'}
                      </p>
                      <p className="mt-1 text-xs text-sea-ink-soft">
                        {searchTerm
                          ? '尝试使用其他别名或名称进行检索'
                          : '创建第一个项目，开始组织 Pin 和 Q&A'}
                      </p>
                    </td>
                  </tr>
                ) : (
                  items.map((item, idx) => (
                    <ProjectRow
                      key={item.id}
                      item={item}
                      index={idx}
                      page={page}
                      pageSize={pageSize}
                      wikiConfig={wikiConfigMap.get(item.id)}
                      pendingPinCount={pendingPinCountMap.get(item.id) || 0}
                      previewSession={previewSessionMap.get(item.id)}
                      onView={(target) => {
                        setSelectedItem(target)
                        setDetailOpen(true)
                      }}
                      onEdit={(target) => {
                        setSelectedItem(target)
                        setEditOpen(true)
                      }}
                      onDelete={(target) => {
                        setSelectedItem(target)
                        setDeleteOpen(true)
                      }}
                      onOpenWiki={(target) => {
                        const config = wikiConfigMap.get(target.id)
                        if (config?.selected_version_id || config?.latest_version) {
                          window.open(`/wiki/${target.name}`, '_blank')
                        } else {
                          navigate({
                            to: '/console/project/$projectId/repowiki',
                            params: { projectId: target.id },
                          })
                        }
                      }}
                    />
                  ))
                )}
              </tbody>
            </table>
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

        {/* 弹窗与抽屉 */}
        <CreateDialog open={createOpen} onOpenChange={setCreateOpen} />
        <EditDialog
          open={editOpen}
          onOpenChange={setEditOpen}
          item={selectedItem}
        />
        <ProjectDetailSheet
          open={detailOpen}
          onOpenChange={setDetailOpen}
          item={selectedItem}
          onEdit={(target) => {
            setSelectedItem(target)
            setEditOpen(true)
          }}
        />
        <ConfirmDeleteDialog
          open={deleteOpen}
          onOpenChange={(open) => {
            if (!open) setDeleteOpen(false)
          }}
          title="删除项目"
          description={`确定删除「${selectedItem?.name}」？关联的 Pin 与 Q&A 配置将一并失效。`}
          onConfirm={() => {
            if (selectedItem) deleteMutation.mutate(selectedItem.id)
            setDeleteOpen(false)
          }}
          isPending={deleteMutation.isPending}
        />
      </motion.div>
    </TooltipProvider>
  )
}
