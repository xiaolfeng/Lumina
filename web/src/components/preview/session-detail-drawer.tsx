import { useEffect, useState } from 'react'
import { ExternalLink, FileCode2, Trash2 } from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@lumina/components/ui/sheet'
import { Button } from '@lumina/components/ui/button'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { usePreviewWebSocket } from '#/hooks/usePreviewWebSocket'
import {
  useDeletePreviewFile,
  usePreviewSessionDetail,
} from '#/hooks/usePreviewAdmin'
import { formatDateTime } from '#/lib/format-date'
import type {
  PreviewFileItem,
  PreviewSessionItem,
} from '#/lib/models/response/preview'

interface PreviewSessionDetailDrawerProps {
  session: PreviewSessionItem | null
  onClose: () => void
}

/** preview_sync 消息的 data 结构（{ session, files }，字段 snake_case） */
interface PreviewSyncData {
  session: PreviewSessionItem
  files: PreviewFileItem[]
}

export function PreviewSessionDetailDrawer({
  session,
  onClose,
}: PreviewSessionDetailDrawerProps) {
  const deleteFileMutation = useDeletePreviewFile()

  const hash = session?.hash ?? null
  const [detail, setDetail] = useState<PreviewSyncData | null>(null)
  const [fileToDelete, setFileToDelete] = useState<PreviewFileItem | null>(null)

  // 1. REST API 作为静态快照基础兜底
  const {
    data: restDetailData,
    isLoading: restLoading,
    error: restError,
    refetch: restRefetch,
  } = usePreviewSessionDetail(session ? hash : null)

  // 会话切换或抽屉关闭时首先重置详情缓存，避免展示上一会话的残留数据
  useEffect(() => {
    setDetail(null)
    setFileToDelete(null)
  }, [hash])

  useEffect(() => {
    const payload = restDetailData?.data
    if (payload && payload.session.hash === hash) {
      setDetail((prev) => {
        // 如果 WebSocket 已经推过当前 hash 的最新快照，则保留 WS 数据
        if (prev && prev.session.hash === hash) return prev
        return {
          session: payload.session,
          files: payload.files,
        }
      })
    }
  }, [restDetailData, hash])

  // 2. WS 实时同步：详情与文件列表由 preview_sync 消息驱动（连接快照 / 文件变更推送）
  const { status } = usePreviewWebSocket(hash, {
    onSync: (data) => {
      if (!data?.session) return
      const syncData = data as PreviewSyncData
      if (syncData.session.hash === hash) {
        setDetail(syncData)
      }
    },
  })

  // 🌟 R-03 严格校验 detail 必须完全属于当前展开会话的 hash，彻底消除切换会话时的瞬时旧数据残留
  const currentDetail = detail?.session.hash === hash ? detail : null
  const isLoading = !currentDetail && restLoading

  return (
    <>
      <Sheet
        open={!!session}
        onOpenChange={(open) => {
          if (!open) onClose()
        }}
      >
        <SheetContent
          side="right"
          className="w-[min(560px,92vw)] gap-0 border-line sm:max-w-[560px]"
        >
          <SheetHeader className="border-b border-line px-6 py-4">
            <SheetTitle>预览会话详情</SheetTitle>
            <SheetDescription>查看该会话的前端文件并管理</SheetDescription>
          </SheetHeader>

          <div className="flex-1 overflow-y-auto px-6 py-5">
            {status === 'rejected' ? (
              <div className="py-12 text-center text-muted-foreground">
                会话不存在
              </div>
            ) : restError && !currentDetail ? (
              <div className="py-12 text-center space-y-3">
                <p className="text-xs text-destructive">
                  获取预览会话详情失败：{restError.message}
                </p>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => void restRefetch()}
                  className="rounded-none text-xs"
                >
                  重新加载
                </Button>
              </div>
            ) : isLoading || !currentDetail ? (
              <div className="py-12 text-center text-muted-foreground">
                加载中…
              </div>
            ) : (
              <div className="space-y-6">
                <div className="rounded-md border border-line bg-surface p-4 text-sm">
                  <div className="flex justify-between py-1">
                    <span className="text-sea-ink-soft">标题</span>
                    <span className="font-medium text-sea-ink">
                      {currentDetail.session.title || '（未命名）'}
                    </span>
                  </div>
                  <div className="flex justify-between py-1">
                    <span className="text-sea-ink-soft">Hash</span>
                    <span className="font-mono text-xs text-sea-ink">
                      {currentDetail.session.hash}
                    </span>
                  </div>
                  <div className="flex justify-between py-1">
                    <span className="text-sea-ink-soft">状态</span>
                    <span className="font-medium text-sea-ink">
                      {currentDetail.session.status}
                    </span>
                  </div>
                  <div className="flex justify-between py-1">
                    <span className="text-sea-ink-soft">过期时间</span>
                    <span className="text-xs text-sea-ink-soft">
                      {formatDateTime(currentDetail.session.expires_at)}
                    </span>
                  </div>
                </div>

                {(() => {
                  const isExpired =
                    new Date(currentDetail.session.expires_at).getTime() <=
                    Date.now()
                  return (
                    <Button
                      variant="outline"
                      className="w-full"
                      disabled={isExpired}
                      onClick={() =>
                        window.open(
                          `/preview/${currentDetail.session.hash}/index.html`,
                          '_blank',
                        )
                      }
                    >
                      <ExternalLink className="mr-2 size-4" aria-hidden />
                      {isExpired ? '预览已过期（链接已收回）' : '打开预览'}
                    </Button>
                  )
                })()}

                <div>
                  <h3 className="mb-3 text-sm font-semibold">
                    文件列表 ({currentDetail.files.length})
                  </h3>
                  {currentDetail.files.length === 0 ? (
                    <p className="py-8 text-center text-muted-foreground">
                      暂无文件
                    </p>
                  ) : (
                    <ul className="space-y-2">
                      {currentDetail.files.map((f) => (
                        <li
                          key={f.id}
                          className="flex items-center justify-between rounded-md border border-line px-3 py-2"
                        >
                          <div className="flex min-w-0 items-center gap-2">
                            <FileCode2
                              className="size-4 shrink-0 text-sea-ink-soft"
                              aria-hidden
                            />
                            <span className="truncate text-sm text-sea-ink">
                              {f.filename}
                            </span>
                            <span className="text-xs text-sea-ink-soft/60">
                              {f.size} B
                            </span>
                          </div>
                          <Button
                            variant="ghost"
                            size="icon"
                            title="删除文件"
                            onClick={() => setFileToDelete(f)}
                          >
                            <Trash2 className="size-4 text-destructive" />
                          </Button>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              </div>
            )}
          </div>
        </SheetContent>
      </Sheet>

      {/* 🌟 Q-08 文件删除确认弹窗保护 */}
      <ConfirmDeleteDialog
        open={!!fileToDelete}
        onOpenChange={(open) => {
          if (!open) setFileToDelete(null)
        }}
        title="确认删除预览文件"
        description={`确定要从该会话中删除文件「${fileToDelete?.filename ?? ''}」吗？此操作将立即生效并同步刷新工作台。`}
        onConfirm={() => {
          if (!fileToDelete) return
          deleteFileMutation.mutate(fileToDelete.id, {
            onSuccess: () => setFileToDelete(null),
          })
        }}
        isPending={deleteFileMutation.isPending}
      />
    </>
  )
}
