import type { ColumnDef } from '@tanstack/react-table'
import { MoreHorizontal } from 'lucide-react'
import { Badge } from '@lumina/components/ui/badge'
import { Button } from '@lumina/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@lumina/components/ui/dropdown-menu'
import { useProjectNameMap } from '#/hooks/useProject'
import { formatDate } from '#/lib/format-date'
import type { PinItem } from '#/lib/models/response/pin'

// ── 分类 Badge 映射 ──
const categoryMap: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
  notice: { label: '注意事项', variant: 'default' },
  dependency: { label: '依赖约束', variant: 'outline' },
  api_change: { label: '接口变更', variant: 'outline' },
}

// ── 优先级 Badge 映射 ──
const priorityMap: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
  high: { label: '高', variant: 'destructive' },
  medium: { label: '中', variant: 'outline' },
  low: { label: '低', variant: 'secondary' },
}

// ── 状态 Badge 映射 ──
const statusMap: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
  pending: { label: '待消费', variant: 'outline' },
  consumed: { label: '已消费', variant: 'default' },
}

interface ColumnActions {
  onEdit: (item: PinItem) => void
  onDelete: (item: PinItem) => void
}

export function getColumns(actions?: ColumnActions): ColumnDef<PinItem>[] {
  return [
    {
      accessorKey: 'title',
      header: '标题',
      cell: ({ row }) => (
        <span className="font-medium">{row.getValue('title')}</span>
      ),
    },
    {
      accessorKey: 'content',
      header: '内容预览',
      cell: ({ row }) => {
        const content = row.getValue('content') as string
        if (!content)
          return <span className="text-muted-foreground">-</span>
        const truncated =
          content.length > 50 ? `${content.slice(0, 50)}…` : content
        return (
          <span className="line-clamp-1 max-w-[300px] text-muted-foreground">
            {truncated}
          </span>
        )
      },
    },
    {
      accessorKey: 'category',
      header: '分类',
      cell: ({ row }) => {
        const category = row.getValue('category') as string
        const mapping = categoryMap[category]
        if (!mapping)
          return <Badge variant="secondary">其他</Badge>
        return <Badge variant={mapping.variant}>{mapping.label}</Badge>
      },
    },
    {
      accessorKey: 'priority',
      header: '优先级',
      cell: ({ row }) => {
        const priority = row.getValue('priority') as string
        const mapping = priorityMap[priority]
        if (!mapping)
          return <Badge variant="secondary">{priority}</Badge>
        return <Badge variant={mapping.variant}>{mapping.label}</Badge>
      },
    },
    {
      accessorKey: 'status',
      header: '状态',
      cell: ({ row }) => {
        const status = row.getValue('status') as string
        const mapping = statusMap[status]
        if (!mapping)
          return <Badge variant="secondary">{status}</Badge>
        return <Badge variant={mapping.variant}>{mapping.label}</Badge>
      },
    },
    {
      accessorKey: 'from_project_id',
      header: '来源项目',
      cell: ({ row }) => (
        <FromProjectCell id={row.getValue('from_project_id')} />
      ),
    },
    {
      accessorKey: 'created_at',
      header: '创建时间',
      cell: ({ row }) => {
        return formatDate(row.getValue('created_at'))
      },
    },
    ...(actions
      ? [
          {
            id: 'actions',
            header: '操作',
            cell: ({ row }: { row: { original: PinItem } }) => {
              const item = row.original
              return (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-11"
                      aria-label="打开操作菜单"
                    >
                      <MoreHorizontal className="size-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => actions.onEdit(item)}>
                      编辑
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={() => actions.onDelete(item)}
                      className="text-destructive focus:text-destructive"
                    >
                      删除
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              )
            },
          },
        ]
      : []),
  ]
}

function FromProjectCell({ id }: { id: unknown }) {
  const { names, isLoading } = useProjectNameMap()
  if (typeof id !== 'string' || !id) {
    return <span className="text-muted-foreground">-</span>
  }
  if (isLoading && !names[id]) {
    return <span className="text-muted-foreground">…</span>
  }
  return <span>{names[id] ?? id}</span>
}
