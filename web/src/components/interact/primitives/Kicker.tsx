import type { ReactNode } from 'react'

import { cn } from '#/lib/utils'

/**
 * 面板头部小标签（kicker）。
 *
 * 收敛 interact 模块内重复的 `text-[11px] font-semibold uppercase tracking-[0.15em]`。
 * 配一条 4px 琥珀金左竖线装饰，增强面板头部识别度。
 */
type KickerTone = 'kicker' | 'sea-ink-soft' | 'lagoon-deep'

const toneClass: Record<KickerTone, string> = {
	kicker: 'text-kicker',
	'sea-ink-soft': 'text-sea-ink-soft',
	'lagoon-deep': 'text-lagoon-deep',
}

interface KickerProps {
	children: ReactNode
	tone?: KickerTone
	/** 是否显示左侧琥珀竖线装饰（默认 true） */
	accent?: boolean
	className?: string
}

export function Kicker({
	children,
	tone = 'kicker',
	accent = true,
	className,
}: KickerProps) {
	return (
		<span
			className={cn(
				'inline-flex items-center gap-1.5 text-[11px] font-semibold uppercase tracking-[0.15em]',
				toneClass[tone],
				accent && 'before:h-3 before:w-0.5 before:rounded-full before:bg-lagoon',
				className,
			)}
		>
			{children}
		</span>
	)
}
