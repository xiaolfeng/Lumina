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
import type { Model } from '#/lib/models/response/llm'

interface ModelColumnActions {
  onEdit: (item: Model) => void
  onDelete: (item: Model) => void
}

export function getModelColumns(
  actions: ModelColumnActions,
): ColumnDef<Model>[] {
  return [
    {
      accessorKey: 'model_name',
      header: '模型标识',
      cell: ({ row }) => (
        <code className="rounded bg-muted px-1.5 py-0.5 text-xs">
          {row.getValue('model_name')}
        </code>
      ),
    },
    {
      accessorKey: 'display_name',
      header: '显示名称',
      cell: ({ row }) => (
        <span className="font-medium">
          {row.getValue('display_name')}
        </span>
      ),
    },
    {
      accessorKey: 'max_tokens',
      header: '最大输出',
      cell: ({ row }) => {
        // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
        const val = row.getValue('max_tokens') as number
        return (
          <span className="text-muted-foreground">
            {/* eslint-disable-next-line @typescript-eslint/no-unnecessary-condition */}
            {val?.toLocaleString() || '-'}
          </span>
        )
      },
    },
    {
      accessorKey: 'context_window',
      header: '上下文窗口',
      cell: ({ row }) => {
        // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
        const val = row.getValue('context_window') as number
        return (
          <span className="text-muted-foreground">
            {/* eslint-disable-next-line @typescript-eslint/no-unnecessary-condition */}
            {val?.toLocaleString() || '-'}
          </span>
        )
      },
    },
    {
      accessorKey: 'temperature',
      header: '温度',
      cell: ({ row }) => {
        // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
        const val = row.getValue('temperature') as number
        return (
          <span className="text-muted-foreground">
            {/* eslint-disable-next-line @typescript-eslint/no-unnecessary-condition */}
            {val !== undefined ? val : '-'}
          </span>
        )
      },
    },
    {
      accessorKey: 'is_active',
      header: '状态',
      cell: ({ row }) => {
        // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
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
        // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
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
        // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
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
