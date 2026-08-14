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
import { usePreviewWebSocket } from '#/hooks/usePreviewWebSocket'
import { useDeletePreviewFile } from '#/hooks/usePreviewAdmin'
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

  // WS 实时同步：详情与文件列表由 preview_sync 消息驱动（连接快照 / 文件变更推送）
  const { status } = usePreviewWebSocket(hash, {
    onSync: (data) => {
      if (!data?.session) return
      const syncData = data as PreviewSyncData
      setDetail(syncData)
    },
  })

  // 会话切换或抽屉关闭时重置详情缓存，避免展示上一会话的残留数据
  useEffect(() => {
    setDetail(null)
  }, [hash])

  // WS 状态驱动加载态：idle / connecting 视为加载中
  const isLoading = status === 'idle' || status === 'connecting'

  return (
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
          ) : isLoading || !detail ? (
            <div className="py-12 text-center text-muted-foreground">
              加载中...
            </div>
          ) : (
            <div className="space-y-6">
              <div className="space-y-2 text-sm">
                <div className="flex items-center justify-between">
                  <span className="text-sea-ink-soft">标题</span>
                  <span className="font-medium text-sea-ink">
                    {detail.session.title}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sea-ink-soft">状态</span>
                  <span className="font-medium text-sea-ink">
                    {detail.session.status === 'active' ? '活跃' : '已删除'}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sea-ink-soft">创建时间</span>
                  <span className="font-medium text-sea-ink">
                    {new Date(detail.session.created_at).toLocaleString()}
                  </span>
                </div>
              </div>

              <Button
                variant="outline"
                className="w-full"
                onClick={() =>
                  window.open(
                    `/preview?session=${detail.session.hash}`,
                    '_blank',
                  )
                }
              >
                <ExternalLink className="mr-2 size-4" aria-hidden />
                打开预览
              </Button>

              <div>
                <h3 className="mb-3 text-sm font-semibold">
                  文件列表 ({detail.files.length})
                </h3>
                {detail.files.length === 0 ? (
                  <p className="py-8 text-center text-muted-foreground">
                    暂无文件
                  </p>
                ) : (
                  <ul className="space-y-2">
                    {detail.files.map((f) => (
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
                          onClick={() => deleteFileMutation.mutate(f.id)}
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
  )
}
