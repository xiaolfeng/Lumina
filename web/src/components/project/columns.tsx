import type { ColumnDef } from '@tanstack/react-table'
import { BookOpen, MoreHorizontal } from 'lucide-react'
import { Badge } from '@lumina/components/ui/badge'
import { Button } from '@lumina/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@lumina/components/ui/dropdown-menu'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@lumina/components/ui/tooltip'
import type { ProjectItem } from '#/lib/models/response/project'

interface ColumnActions {
  onEdit: (item: ProjectItem) => void
  onDelete: (item: ProjectItem) => void
  onOpenWiki: (item: ProjectItem) => void
}

export function getColumns(actions: ColumnActions): ColumnDef<ProjectItem>[] {
  return [
    {
      accessorKey: 'name',
      header: '项目名称',
      cell: ({ row }) => (
        <span className="font-medium">{row.getValue('name')}</span>
      ),
    },
    {
      accessorKey: 'alias_name',
      header: '别名',
      cell: ({ row }) => {
        const alias = row.getValue('alias_name') as string
        return (
          <span className={alias ? '' : 'text-muted-foreground'}>
            {alias || '-'}
          </span>
        )
      },
    },
    {
      accessorKey: 'match_path',
      header: '匹配路径',
      cell: ({ row }) => {
        const paths = row.getValue('match_path') as string[]
        if (!paths || paths.length === 0)
          return <span className="text-muted-foreground">-</span>
        return (
          <div className="flex flex-wrap gap-1">
            {paths.map((path) => (
              <Badge key={path} variant="outline" className="font-mono text-xs">
                {path}
              </Badge>
            ))}
          </div>
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
          <div className="flex items-center gap-1">
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => actions.onOpenWiki(item)}
                >
                  <BookOpen className="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Wiki 管理</TooltipContent>
            </Tooltip>
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
          </div>
        )
      },
    },
  ]
}
