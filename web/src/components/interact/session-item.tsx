import { Bot, Check, Clock, Users } from 'lucide-react'

import type { Session } from './types'

/**
 * 会话列表项 —— 被 session-panel（桌面第三列/移动抽屉）复用。
 */
export function SessionItem({
	session,
	isActive,
	onSelect,
}: {
	session: Session
	isActive: boolean
	onSelect: (id: string) => void
}) {
	const pending = session.questions.filter((q) => !q.answered).length

	return (
		<button
			type="button"
			onClick={() => onSelect(session.id)}
			className={`group flex w-full cursor-pointer flex-col gap-1.5 rounded-lg px-3 py-2.5 text-left transition-colors duration-150 ${
				isActive
					? 'bg-lagoon/10 text-sea-ink'
					: 'text-sea-ink-soft hover:bg-lagoon/5 hover:text-sea-ink'
			}`}
			aria-label={`会话：${session.title}`}
		>
			<span className="text-sm font-medium leading-tight">{session.title}</span>
			<div className="flex items-center gap-2 text-[11px]">
				<span className="flex items-center gap-0.5">
					<Bot className="size-3" aria-hidden />
					{session.agent}
				</span>
				<span className="flex items-center gap-0.5">
					<Users className="size-3" aria-hidden />
					{session.onlineDevices}
				</span>
			</div>
			<div className="flex items-center justify-between">
				<span className="flex items-center gap-0.5 text-[10px]">
					<Clock className="size-2.5" aria-hidden />
					{session.updatedAt}
				</span>
				{pending > 0 ? (
					<span className="inline-flex size-4 items-center justify-center rounded-full bg-lagoon text-[9px] font-bold text-white">
						{pending}
					</span>
				) : (
					<Check className="size-3.5 text-lagoon" aria-hidden />
				)}
			</div>
		</button>
	)
}
