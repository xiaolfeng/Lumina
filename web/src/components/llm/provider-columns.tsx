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
import type { Provider } from '#/lib/models/response/llm'

interface ProviderColumnActions {
  onEdit: (item: Provider) => void
  onDelete: (item: Provider) => void
}

export function getProviderColumns(
  actions: ProviderColumnActions,
): ColumnDef<Provider>[] {
  return [
    {
      accessorKey: 'name',
      header: '名称',
      cell: ({ row }) => (
        <span className="font-medium">{row.getValue('name')}</span>
      ),
    },
    {
      accessorKey: 'protocol',
      header: '协议',
      cell: ({ row }) => (
        <code className="rounded bg-muted px-1.5 py-0.5 text-xs">
          {row.getValue('protocol')}
        </code>
      ),
    },
    {
      accessorKey: 'base_url',
      header: 'Base URL',
      cell: ({ row }) => {
        // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
        const val = row.getValue('base_url') as string
        return (
          <span className="text-muted-foreground">
            {val || '-'}
          </span>
        )
      },
    },
    {
      accessorKey: 'has_key',
      header: '密钥状态',
      cell: ({ row }) => {
        // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
        const hasKey = row.getValue('has_key') as boolean
        return (
          <Badge variant={hasKey ? 'default' : 'secondary'}>
            {hasKey ? '已设置' : '未设置'}
          </Badge>
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
