import type { ColumnDef } from '@tanstack/react-table'
import { MoreHorizontal } from 'lucide-react'
import { Badge } from '#/components/ui/badge'
import { Button } from '#/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '#/components/ui/dropdown-menu'
import type { ApikeyItem } from '#/lib/models/response/apikey'

interface ColumnActions {
  onEdit: (item: ApikeyItem) => void
  onReset: (item: ApikeyItem) => void
  onDelete: (item: ApikeyItem) => void
}

export function getColumns(actions: ColumnActions): ColumnDef<ApikeyItem>[] {
  return [
    {
      accessorKey: 'name',
      header: '名称',
      cell: ({ row }) => (
        <span className="font-medium">{row.getValue('name')}</span>
      ),
    },
    {
      accessorKey: 'key_prefix',
      header: '前缀',
      cell: ({ row }) => (
        <code className="rounded bg-muted px-1.5 py-0.5 text-xs">
          {row.getValue('key_prefix')}
        </code>
      ),
    },
    {
      accessorKey: 'is_active',
      header: '状态',
      cell: ({ row }) => {
        const isActive = row.getValue('is_active') as boolean
        return (
          <Badge variant={isActive ? 'default' : 'secondary'}>
            {isActive ? '启用' : '禁用'}
          </Badge>
        )
      },
    },
    {
      accessorKey: 'description',
      header: '描述',
      cell: ({ row }) => {
        const desc = row.getValue('description') as string
        return (
          <span className="line-clamp-1 max-w-[200px] text-muted-foreground">
            {desc || '-'}
          </span>
        )
      },
    },
    {
      accessorKey: 'created_at',
      header: '创建时间',
      cell: ({ row }) => {
        const val = row.getValue('created_at') as string
        return val ? new Date(val).toLocaleDateString('zh-CN') : '-'
      },
    },
    {
      id: 'actions',
      header: '操作',
      cell: ({ row }) => {
        const item = row.original
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreHorizontal className="size-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => actions.onEdit(item)}>
                编辑
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => actions.onReset(item)}>
                重置密钥
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
}
