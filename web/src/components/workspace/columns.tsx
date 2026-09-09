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
import { formatDate } from '#/lib/format-date'
import type { WorkspaceItem } from '#/lib/models/response/workspace'

interface ColumnActions {
  onEdit: (item: WorkspaceItem) => void
  onDelete: (item: WorkspaceItem) => void
}

export function getWorkspaceColumns(
  actions: ColumnActions,
): ColumnDef<WorkspaceItem>[] {
  return [
    {
      accessorKey: 'name',
      header: '名称',
      cell: ({ row }) => (
        <span className="font-medium">{row.getValue('name')}</span>
      ),
    },
    {
      accessorKey: 'slug',
      header: '标识',
      cell: ({ row }) => (
        <span className="font-mono text-xs">{row.getValue('slug')}</span>
      ),
    },
    {
      accessorKey: 'is_default',
      header: '默认',
      cell: ({ row }) =>
        row.original.is_default ? (
          <Badge variant="outline">默认</Badge>
        ) : (
          <span className="text-muted-foreground">-</span>
        ),
    },
    {
      accessorKey: 'icon',
      header: '图标',
      cell: ({ row }) => {
        const icon = row.getValue('icon') as string
        return <span className="text-muted-foreground">{icon || '-'}</span>
      },
    },
    {
      accessorKey: 'created_at',
      header: '创建时间',
      cell: ({ row }) => formatDate(row.getValue('created_at')),
    },
    {
      id: 'actions',
      header: '操作',
      cell: ({ row }) => {
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
              {!item.is_default && (
                <DropdownMenuItem
                  onClick={() => actions.onDelete(item)}
                  className="text-destructive focus:text-destructive"
                >
                  删除
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]
}
