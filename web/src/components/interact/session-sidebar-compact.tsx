import { useSidebarOpen } from '#/hooks/useSidebarOpen'
import { SessionPanel } from './session-panel'
import type { Session } from './types'

/**
 * 桌面端第三列（≥ xl）—— 真 flex 子列，宽度由 sidebar open 状态过渡。
 */
export function SessionSidebarCompact({
	sessions,
	selectedId,
	onSelect,
}: {
	sessions: Session[]
	selectedId: string
	onSelect: (id: string) => void
}) {
	const { open } = useSidebarOpen()

	return (
		<aside
			className={`hidden shrink-0 flex-col overflow-hidden rounded-xl border border-line bg-surface shadow-[0_1px_3px_rgba(42,36,32,0.04),0_8px_24px_-8px_rgba(42,36,32,0.10)] transition-[width,opacity,margin] duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] xl:flex ${
				open
					? 'm-4 w-[320px] opacity-100'
					: 'm-0 w-0 rounded-none border-0 opacity-0 shadow-none pointer-events-none'
			}`}
		>
			<SessionPanel
				sessions={sessions}
				selectedId={selectedId}
				onSelect={onSelect}
			/>
		</aside>
	)
}
