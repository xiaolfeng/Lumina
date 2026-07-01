import type { ReactNode } from 'react'

import { cn } from '#/lib/utils'

/**
 * 卡片容器 —— 收敛 interact 模块内重复的卡片样式。
 *
 * 双层阴影 + 半透明白底 + 微毛玻璃，比单调单层阴影更有层次。
 * 支持 header / footer slot（带分隔线）。
 */
interface PanelCardProps {
	children: ReactNode
	/** 头部内容（可选），渲染时带底部分隔线 */
	header?: ReactNode
	/** 底部内容（可选），渲染时带顶部分隔线 */
	footer?: ReactNode
	/** 头部是否紧凑（无 padding，由调用方自定） */
	flushHeader?: boolean
	className?: string
	bodyClassName?: string
}

export function PanelCard({
	children,
	header,
	footer,
	flushHeader,
	className,
	bodyClassName,
}: PanelCardProps) {
	return (
		<section
			className={cn(
				'overflow-hidden rounded-xl border border-line bg-surface shadow-[0_1px_3px_rgba(42,36,32,0.04),0_8px_24px_-8px_rgba(42,36,32,0.10)] backdrop-blur-sm',
				className,
			)}
		>
			{header && (
				<div
					className={cn(
						!flushHeader && 'px-4 py-2.5',
						footer || children ? 'border-b border-line/50' : '',
					)}
				>
					{header}
				</div>
			)}
			<div className={cn('p-4', bodyClassName)}>{children}</div>
			{footer && <div className="border-t border-line/50">{footer}</div>}
		</section>
	)
}
