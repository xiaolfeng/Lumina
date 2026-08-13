import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@lumina/components/ui/badge'
import { Button } from '@lumina/components/ui/button'
import { Eye, Trash2 } from 'lucide-react'
import type { SessionItem } from '#/lib/models/response/qa-admin'

export function getSessionColumns(
  onDelete: (session: SessionItem) => void,
  onView: (session: SessionItem) => void,
): ColumnDef<SessionItem>[] {
	return [
		{
			accessorKey: 'title',
			header: '标题',
			cell: ({ row }) => (
				<span className="display-title font-medium">{row.getValue('title')}</span>
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
				const label = status === 'active' ? '活跃' : status === 'expired' ? '已过期' : '已删除'
				const dot = status === 'active' ? 'bg-emerald-500' : status === 'expired' ? 'bg-sea-ink-soft/40' : 'bg-destructive'
				const text = status === 'active' ? 'text-emerald-600' : status === 'expired' ? 'text-sea-ink-soft' : 'text-destructive'
				return (
					<span className={`inline-flex items-center gap-1.5 text-xs font-medium ${text}`}>
						<span className={`size-1.5 rounded-full ${dot}`} aria-hidden />
						{label}
					</span>
				)
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
						<Button variant="ghost" size="icon" title="查看详情" onClick={() => onView(session)}>
							<Eye className="size-4" />
						</Button>
						<Button variant="ghost" size="icon" onClick={() => onDelete(session)}>
							<Trash2 className="size-4 text-destructive" />
						</Button>
					</div>
				)
			},
		},
	]
}
