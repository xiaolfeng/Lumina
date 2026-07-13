import { useState } from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '#/components/ui/table'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { Badge } from '#/components/ui/badge'
import { Button } from '#/components/ui/button'
import { useWebhookEvents } from '#/hooks/useWebhook'
import { SkeletonTable } from '#/components/skeleton-table'
import { ChevronLeft, ChevronRight } from 'lucide-react'

interface WebhookEventsProps {
  configId: string
}

const STATUS_VARIANTS: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
  accepted: { label: '已接受', variant: 'default' as const },
  ignored: { label: '已忽略', variant: 'secondary' as const },
  failed: { label: '失败', variant: 'destructive' as const },
  pending: { label: '处理中', variant: 'outline' as const },
}

export function WebhookEvents({ configId }: WebhookEventsProps) {
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)
  const { data, isLoading, isError, error } = useWebhookEvents(configId, page, pageSize)

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Webhook 事件日志</CardTitle>
        </CardHeader>
        <CardContent>
          <SkeletonTable rows={5} />
        </CardContent>
      </Card>
    )
  }

  if (isError) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Webhook 事件日志</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-destructive">
            <p className="font-medium">加载事件日志失败</p>
            <p className="text-sm mt-1">{error.message || '请稍后重试'}</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  const events = data?.data?.items ?? []
  const total = data?.data?.total ?? 0
  const totalPages = pageSize > 0 ? Math.ceil(total / pageSize) : 1

  return (
    <Card>
      <CardHeader>
        <CardTitle>Webhook 事件日志</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {events.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">暂无 Webhook 事件记录</p>
          </div>
        ) : (
          <>
            <div className="rounded-md border overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50 hover:bg-muted/50">
                    <TableHead className="w-[160px]">时间</TableHead>
                    <TableHead className="w-[100px]">提供商</TableHead>
                    <TableHead className="w-[120px]">分支</TableHead>
                    <TableHead className="w-[100px]">状态</TableHead>
                    <TableHead>原因</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {events.map((event) => {
                    const statusConfig = STATUS_VARIANTS[event.status] ?? {
                      label: event.status,
                      variant: 'outline' as const,
                    }
                    return (
                      <TableRow key={event.id}>
                        <TableCell className="text-sm text-muted-foreground">
                          {new Date(event.received_at).toLocaleString()}
                        </TableCell>
                        <TableCell className="text-sm">{event.provider}</TableCell>
                        <TableCell className="text-sm font-mono">
                          {event.branch ?? '-'}
                        </TableCell>
                        <TableCell>
                          <Badge variant={statusConfig.variant} className="text-xs">
                            {statusConfig.label}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground max-w-[300px] truncate">
                          {event.reason ?? '-'}
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            </div>

            {/* Pagination */}
            <div className="flex items-center justify-between">
              <div className="text-sm text-muted-foreground">
                共 {total} 条记录
              </div>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                >
                  <ChevronLeft className="size-4" />
                </Button>
                <span className="text-sm text-muted-foreground">
                  第 {page} / {totalPages} 页
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages}
                >
                  <ChevronRight className="size-4" />
                </Button>
              </div>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}
