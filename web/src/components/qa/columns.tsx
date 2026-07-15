import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@lumina/components/ui/badge'
import { Button } from '@lumina/components/ui/button'
import { Trash2 } from 'lucide-react'
import type { SessionItem } from '#/lib/models/response/qa-admin'

export function getSessionColumns(onDelete: (session: SessionItem) => void): ColumnDef<SessionItem>[] {
	return [
		{
			accessorKey: 'title',
			header: '标题',
			cell: ({ row }) => (
				<span className="font-medium">{row.getValue('title')}</span>
			),
		},
		{
			accessorKey: 'agent',
			header: 'Agent',
		},
		{
			accessorKey: 'type',
			header: '类型',
			cell: ({ row }) => {
				const type = row.getValue('type') as string
				return (
					<Badge variant={type === 'permanent' ? 'default' : 'secondary'}>
						{type === 'permanent' ? '永久' : '临时'}
					</Badge>
				)
			},
		},
		{
			accessorKey: 'project_name',
			header: '关联项目',
			cell: ({ row }) => {
				const name = row.getValue('project_name') as string
				return name || <span className="text-muted-foreground">—</span>
			},
		},
		{
			accessorKey: 'status',
			header: '状态',
			cell: ({ row }) => {
				const status = row.getValue('status') as string
				const variant =
					status === 'active' ? 'default' : status === 'expired' ? 'outline' : 'destructive'
				const label = status === 'active' ? '活跃' : status === 'expired' ? '已过期' : '已删除'
				return <Badge variant={variant}>{label}</Badge>
			},
		},
		{
			accessorKey: 'online_devices',
			header: '在线设备',
		},
		{
			accessorKey: 'expires_at',
			header: '过期时间',
			cell: ({ row }) => {
				const val = row.getValue('expires_at') as string
				return val ? new Date(val).toLocaleString() : '永久有效'
			},
		},
		{
			id: 'actions',
			header: '操作',
			cell: ({ row }) => {
				const session = row.original
				return (
					<div className="flex items-center gap-2">
						<Button variant="ghost" size="icon" onClick={() => onDelete(session)}>
							<Trash2 className="size-4 text-destructive" />
						</Button>
					</div>
				)
			},
		},
	]
}
