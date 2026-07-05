import { Badge } from '#/components/ui/badge'
import { Loader2 } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

interface StatusConfig {
	label: string
	variant: 'default' | 'secondary' | 'destructive' | 'outline'
	className: string
	icon?: LucideIcon
	spinning?: boolean
}

const STATUS_MAP: Record<string, StatusConfig> = {
	pending: {
		label: '等待中',
		variant: 'outline',
		className: 'border-muted-foreground/30 text-muted-foreground',
	},
	cloning: {
		label: '克隆中',
		variant: 'default',
		className: 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950 dark:text-blue-300 dark:border-blue-800',
		icon: Loader2,
		spinning: true,
	},
	scanning: {
		label: '扫描中',
		variant: 'default',
		className: 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950 dark:text-blue-300 dark:border-blue-800',
		icon: Loader2,
		spinning: true,
	},
	analyzing: {
		label: '分析中',
		variant: 'default',
		className: 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950 dark:text-blue-300 dark:border-blue-800',
		icon: Loader2,
		spinning: true,
	},
	assembling: {
		label: '组装中',
		variant: 'default',
		className: 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950 dark:text-blue-300 dark:border-blue-800',
		icon: Loader2,
		spinning: true,
	},
	completed: {
		label: '已完成',
		variant: 'default',
		className: 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950 dark:text-emerald-300 dark:border-emerald-800',
	},
	failed: {
		label: '失败',
		variant: 'destructive',
		className: '',
	},
	cancelled: {
		label: '已取消',
		variant: 'outline',
		className: 'border-muted-foreground/30 text-muted-foreground',
	},
}

interface StatusBadgeProps {
	status: string
	className?: string
}

export function StatusBadge({ status, className }: StatusBadgeProps) {
	const config = STATUS_MAP[status] ?? {
		label: status,
		variant: 'outline' as const,
		className: '',
	}

	const Icon = config.icon

	return (
		<Badge
			variant={config.variant}
			className={`gap-1.5 ${config.className} ${className ?? ''}`}
		>
			{Icon && config.spinning && (
				<Icon className="size-3 animate-spin" />
			)}
			{config.label}
		</Badge>
	)
}
