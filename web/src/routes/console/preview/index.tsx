import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { DataTable } from '#/components/data-table'
import { DataTablePagination } from '#/components/data-table-pagination'
import {
  usePreviewSessionList,
  useDeletePreviewSession,
} from '#/hooks/usePreviewAdmin'
import { getPreviewSessionColumns } from '#/components/preview/columns'
import { PreviewSessionDetailDrawer } from '#/components/preview/session-detail-drawer'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { PageHeader } from '#/components/page-header'
import { SkeletonTable } from '#/components/skeleton-table'
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

  const { data, isLoading } = usePreviewSessionList({ page, size: pageSize })
  const deleteMutation = useDeletePreviewSession()

  const items = data?.data?.items ?? []
  const totalItems = data?.data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

  const columns = getPreviewSessionColumns(
    (session) => setDeleteTarget(session),
    (session) => setViewTarget(session),
  )

  return (
    <motion.div
      className="space-y-4"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      <PageHeader title="预览管理" description="管理前端可视化预览会话与文件" />

      <motion.div variants={staggerItem}>
        <div className="mt-4">
          {isLoading ? (
            <SkeletonTable />
          ) : (
            <>
              <DataTable columns={columns} data={items} />
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
            </>
          )}
        </div>
      </motion.div>

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

      <PreviewSessionDetailDrawer
        session={viewTarget}
        onClose={() => setViewTarget(null)}
      />
    </motion.div>
  )
}
