import { useEffect, useState } from 'react'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { BookOpen, Pencil, Plus, Trash2 } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { Skeleton } from '@lumina/components/ui/skeleton'
import { DataTablePagination } from '#/components/data-table-pagination'
import { useProjectList, useDeleteProject } from '#/hooks/useProject'
import { useCurrentWorkspace } from '#/hooks/useCurrentWorkspace'
import { CreateDialog } from '#/components/project/create-dialog'
import { EditDialog } from '#/components/project/edit-dialog'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { PageHeader } from '#/components/page-header'
import { formatDate } from '#/lib/format-date'
import type { ProjectItem } from '#/lib/models/response/project'

export const Route = createFileRoute('/console/project/')({
  component: ProjectPage,
})

function ProjectRow({
  item,
  onEdit,
  onDelete,
  onOpenWiki,
}: {
  item: ProjectItem
  onEdit: (item: ProjectItem) => void
  onDelete: (item: ProjectItem) => void
  onOpenWiki: (item: ProjectItem) => void
}) {
  return (
    <div className="flex items-center gap-4 border-b border-line px-2 py-4 transition-colors hover:bg-chip-bg">
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <div className="flex flex-wrap items-center gap-2">
          <span className="truncate text-sm font-semibold text-sea-ink">
            {item.name}
          </span>
          {item.alias_name ? (
            <span className="inline-flex items-center border border-chip-line bg-lagoon/10 px-1.5 py-0.5 text-[11px] leading-5 font-semibold text-lagoon-deep">
              {item.alias_name}
            </span>
          ) : null}
        </div>
        {item.description ? (
          <p
            className="truncate text-[13px] text-sea-ink-soft"
            title={item.description}
          >
            {item.description}
          </p>
        ) : null}
        <div className="flex flex-wrap items-center gap-2 text-xs text-sea-ink-soft">
          <span>创建于 {formatDate(item.created_at)}</span>
          <span>·</span>
          <span>更新于 {formatDate(item.updated_at)}</span>
        </div>
      </div>
      <div className="flex shrink-0 items-center gap-1">
        <button
          type="button"
          className="grid size-9 place-items-center text-sea-ink-soft transition-colors hover:bg-foam hover:text-sea-ink"
          aria-label={`打开 ${item.name} Wiki`}
          title="Wiki 管理"
          onClick={() => onOpenWiki(item)}
        >
          <BookOpen className="size-4" />
        </button>
        <button
          type="button"
          className="grid size-9 place-items-center text-sea-ink-soft transition-colors hover:bg-foam hover:text-sea-ink"
          aria-label={`编辑 ${item.name}`}
          title="编辑项目"
          onClick={() => onEdit(item)}
        >
          <Pencil className="size-3.5" />
        </button>
        <button
          type="button"
          className="grid size-9 place-items-center text-sea-ink-soft transition-colors hover:bg-destructive/10 hover:text-destructive"
          aria-label={`删除 ${item.name}`}
          title="删除项目"
          onClick={() => onDelete(item)}
        >
          <Trash2 className="size-3.5" />
        </button>
      </div>
    </div>
  )
}

function ProjectPage() {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [selectedItem, setSelectedItem] = useState<ProjectItem | null>(null)

  const { current } = useCurrentWorkspace()
  const { data, isLoading } = useProjectList({
    page,
    size: pageSize,
    workspace_id: current?.id,
  })
  const showLoading = !current?.id || isLoading
  const deleteMutation = useDeleteProject()

  const items = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

  useEffect(() => {
    setPage(1)
    setEditOpen(false)
    setDeleteOpen(false)
    setSelectedItem(null)
  }, [current?.id])

  useEffect(() => {
    if (!showLoading && page > totalPages) setPage(totalPages)
  }, [showLoading, page, totalPages])

  return (
    <motion.div
      className="space-y-4"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      <PageHeader
        title="项目管理"
        description="管理你的项目，用于组织 Pin 和 Q&A"
        action={
          <Button
            onClick={() => setCreateOpen(true)}
            className="bg-lagoon text-foam hover:bg-lagoon-deep"
          >
            <Plus className="mr-2 size-4" />
            创建项目
          </Button>
        }
      />

      <motion.div variants={staggerItem} className="space-y-4">
        {/* 流式项目列表（不展示匹配路径，空间留给名称与描述） */}
        <div className="border-t border-line">
          {showLoading ? (
            <div className="flex flex-col gap-2 py-4">
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-20 w-full" />
            </div>
          ) : items.length === 0 ? (
            <div className="py-14 text-center">
              <p className="text-sm font-semibold text-sea-ink">暂无项目</p>
              <p className="mt-1 text-[13px] text-sea-ink-soft">
                创建第一个项目，开始组织 Pin 和 Q&A
              </p>
            </div>
          ) : (
            items.map((item) => (
              <ProjectRow
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
                onOpenWiki={(target) => {
                  navigate({
                    to: '/console/project/$projectId/repowiki',
                    params: { projectId: target.id },
                  })
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

      <CreateDialog open={createOpen} onOpenChange={setCreateOpen} />
      <EditDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        item={selectedItem}
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
  )
}
