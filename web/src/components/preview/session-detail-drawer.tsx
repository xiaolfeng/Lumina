import { useQuery } from '@tanstack/react-query'
import { ExternalLink, FileCode2, Trash2 } from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@lumina/components/ui/sheet'
import { Button } from '@lumina/components/ui/button'
import { getPreviewSessionByHash } from '#/lib/apis/preview'
import { useDeletePreviewFile } from '#/hooks/usePreviewAdmin'
import type { PreviewSessionItem } from '#/lib/models/response/preview'

interface PreviewSessionDetailDrawerProps {
  session: PreviewSessionItem | null
  onClose: () => void
}

export function PreviewSessionDetailDrawer({
  session,
  onClose,
}: PreviewSessionDetailDrawerProps) {
  const deleteFileMutation = useDeletePreviewFile()

  const { data, isLoading } = useQuery({
    queryKey: ['preview', 'session-detail', session?.hash],
    queryFn: () => getPreviewSessionByHash(session!.hash),
    enabled: !!session?.hash,
  })
  const detail = data?.data

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
          {isLoading ? (
            <div className="py-12 text-center text-muted-foreground">
              加载中...
            </div>
          ) : detail ? (
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
          ) : (
            <div className="py-12 text-center text-muted-foreground">
              会话不存在
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
