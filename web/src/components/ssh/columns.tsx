import type { ColumnDef } from '@tanstack/react-table'
import { Download, MoreHorizontal, Pencil, Trash2 } from 'lucide-react'
import { Badge } from '@lumina/components/ui/badge'
import { Button } from '@lumina/components/ui/button'
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from '@lumina/components/ui/dropdown-menu'
import type { SshKeyItem } from '#/lib/models/response/ssh'
import { getSshPublicKey } from '#/lib/apis/ssh'
import { toast } from 'sonner'

interface ColumnActions {
	onEdit: (item: SshKeyItem) => void
	onDelete: (item: SshKeyItem) => void
}

function handleDownloadPublicKey(item: SshKeyItem) {
	getSshPublicKey(item.id)
		.then((blob) => {
			const url = URL.createObjectURL(blob)
			const a = document.createElement('a')
			a.href = url
			a.download = `${item.name}-id_${item.key_type}.pub`
			document.body.appendChild(a)
			a.click()
			document.body.removeChild(a)
			URL.revokeObjectURL(url)
			toast.success('公钥下载成功')
		})
		.catch(() => {
			toast.error('下载公钥失败')
		})
}

export function getColumns(actions: ColumnActions): ColumnDef<SshKeyItem>[] {
	return [
		{
			accessorKey: 'name',
			header: '名称',
			cell: ({ row }) => (
				<span className="font-medium">{row.getValue('name')}</span>
			),
		},
		{
			accessorKey: 'key_type',
			header: '类型',
			cell: ({ row }) => {
				const type = row.getValue('key_type') as string
				return (
					<Badge variant="outline" className="font-mono text-xs uppercase">
						{type}
					</Badge>
				)
			},
		},
		{
			accessorKey: 'fingerprint',
			header: '指纹',
			cell: ({ row }) => {
				const fp = row.getValue('fingerprint') as string
				return (
					<span className="max-w-[280px] truncate font-mono text-xs text-muted-foreground" title={fp}>
						{fp}
					</span>
				)
			},
		},
		{
			accessorKey: 'source',
			header: '来源',
			cell: ({ row }) => {
				const source = row.getValue('source') as string
				return (
					<Badge variant={source === 'generated' ? 'default' : 'secondary'}>
						{source === 'generated' ? '生成' : '导入'}
					</Badge>
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
								<Pencil className="mr-2 size-3.5" />
								编辑
							</DropdownMenuItem>
							<DropdownMenuItem onClick={() => handleDownloadPublicKey(item)}>
								<Download className="mr-2 size-3.5" />
								下载公钥
							</DropdownMenuItem>
							<DropdownMenuItem
								onClick={() => actions.onDelete(item)}
								className="text-destructive focus:text-destructive"
							>
								<Trash2 className="mr-2 size-3.5" />
								删除
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
				)
			},
		},
	]
}
