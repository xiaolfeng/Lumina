import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@lumina/components/ui/badge'
import { Button } from '@lumina/components/ui/button'
import { ExternalLink, Eye, Trash2 } from 'lucide-react'
import type { PreviewSessionItem } from '#/lib/models/response/preview'

export function getPreviewSessionColumns(
  onDelete: (session: PreviewSessionItem) => void,
  onView: (session: PreviewSessionItem) => void,
): ColumnDef<PreviewSessionItem>[] {
  return [
    {
      accessorKey: 'title',
      header: '标题',
      cell: ({ row }) => (
        <span className="display-title font-medium">{row.original.title}</span>
      ),
    },
    {
      accessorKey: 'status',
      header: '状态',
      cell: ({ row }) => {
        const status = row.original.status
        return (
          <Badge variant={status === 'active' ? 'default' : 'secondary'}>
            {status === 'active' ? '活跃' : '已删除'}
          </Badge>
        )
      },
    },
    {
      accessorKey: 'updated_at',
      header: '更新时间',
      cell: ({ row }) => {
        const val = row.original.updated_at
        return val ? new Date(val).toLocaleString() : '—'
      },
    },
    {
      id: 'actions',
      header: '操作',
      cell: ({ row }) => {
        const session = row.original
        return (
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              title="查看详情"
              onClick={() => onView(session)}
            >
              <Eye className="size-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              title="打开预览"
              onClick={() =>
                window.open(`/preview?session=${session.hash}`, '_blank')
              }
            >
              <ExternalLink className="size-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              title="删除"
              onClick={() => onDelete(session)}
            >
              <Trash2 className="size-4 text-destructive" />
            </Button>
          </div>
        )
      },
    },
  ]
}
