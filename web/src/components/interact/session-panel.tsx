import { X } from 'lucide-react'

import { ScrollArea } from '#/components/ui/scroll-area'
import { Separator } from '#/components/ui/separator'

import { Kicker } from './primitives'
import { SessionItem } from './session-item'
import type { Session } from './types'

/**
 * 会话面板内容 —— 桌面第三列与移动端抽屉共享。
 * isMobile=true 时渲染拖拽手柄与关闭按钮。
 */
export function SessionPanel({
	sessions,
	selectedId,
	onSelect,
	onClose,
}: {
	sessions: Session[]
	selectedId: string
	onSelect: (id: string) => void
	onClose?: () => void
}) {
	const current = sessions.find((s) => s.id === selectedId)
	const answeredCount = current
		? current.questions.filter((q) => q.answered).length
		: 0
	const totalCount = current?.questions.length ?? 0
	const remainingCount = totalCount - answeredCount
	const percent =
		totalCount > 0 ? Math.round((answeredCount / totalCount) * 100) : 0
	const isMobile = !!onClose

	return (
		<>
			{isMobile && (
				<div className="flex items-center justify-center pb-1 pt-2.5">
					<div className="h-1.5 w-10 rounded-full bg-line" />
				</div>
			)}

			<div
				className={`flex items-center justify-between ${isMobile ? 'px-4 py-2' : 'px-3 py-2'}`}
			>
				<Kicker>永久会话列表</Kicker>
				{isMobile && onClose && (
					<button
						type="button"
						onClick={onClose}
						className="inline-flex items-center justify-center rounded-lg p-1 text-sea-ink-soft transition-colors duration-200 hover:bg-line/30 hover:text-sea-ink"
						aria-label="关闭"
					>
						<X className="size-4" aria-hidden />
					</button>
				)}
			</div>

			<Separator className="bg-line" />

			<ScrollArea className={`flex-1 ${isMobile ? 'px-3' : 'px-2'}`}>
				<div className="space-y-1 pb-3 py-1">
					{sessions.map((session) => (
						<SessionItem
							key={session.id}
							session={session}
							isActive={session.id === selectedId}
							onSelect={onSelect}
						/>
					))}
				</div>
			</ScrollArea>

			{current && (
				<>
					<Separator className="bg-line" />
					<div className={isMobile ? 'p-4' : 'p-3'}>
						<div
							className={`rounded-xl border border-line bg-foam ${isMobile ? 'p-4' : 'p-3'}`}
						>
							<div
								className={`${isMobile ? 'mb-3' : 'mb-2'} flex items-center justify-between`}
							>
								<span className="text-[10px] font-semibold uppercase tracking-wider text-sea-ink-soft">
									会话信息
								</span>
								<span className="inline-flex items-center gap-1 rounded-full bg-lagoon/10 px-2 py-0.5 text-[10px] font-medium text-lagoon-deep">
									<span className="inline-block size-1.5 rounded-full bg-lagoon" />
									活跃
								</span>
							</div>
							<div className={isMobile ? 'space-y-2' : 'space-y-1.5'}>
								<div className="flex items-center justify-between text-xs">
									<span className="text-sea-ink-soft">Agent</span>
									<span className="font-medium text-sea-ink">{current.agent}</span>
								</div>
								<div className="flex items-center justify-between text-xs">
									<span className="text-sea-ink-soft">进度</span>
									<span className="font-medium text-sea-ink">
										已答 {answeredCount} · 剩余 {remainingCount}
									</span>
								</div>
							</div>
							<div className={isMobile ? 'mt-3' : 'mt-2'}>
								{isMobile && (
									<div className="mb-1 flex items-center justify-between">
										<span className="text-[10px] text-sea-ink-soft">完成度</span>
										<span className="text-[10px] font-medium text-sea-ink">
											{percent}%
										</span>
									</div>
								)}
								<div
									className={`${isMobile ? 'h-2' : 'h-1.5'} overflow-hidden rounded-full bg-line`}
								>
									<div
										className="h-full rounded-full bg-gradient-to-r from-lagoon to-palm transition-all duration-500"
										style={{ width: `${percent}%` }}
									/>
								</div>
							</div>
						</div>
					</div>
				</>
			)}
		</>
	)
}
