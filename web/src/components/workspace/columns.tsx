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
import { WorkspaceIcon } from '#/components/workspace-icon'

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
      cell: ({ row }) => {
        const item = row.original
        return (
          <span className="flex min-w-0 items-center gap-2">
            <WorkspaceIcon name={item.icon} label={`${item.name}的图标`} />
            <span className="truncate font-medium">{item.name}</span>
          </span>
        )
      },
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
