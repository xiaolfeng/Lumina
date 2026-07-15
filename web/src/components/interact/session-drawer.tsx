import { useState, useEffect } from 'react'

import { Sheet, SheetContent } from '@lumina/components/ui/sheet'

import { useSidebarOpen } from '#/hooks/useSidebarOpen'
import { SessionPanel } from './session-panel'
import type { Session } from './types'

/**
 * 移动端会话抽屉（< xl）—— Sheet 容器。
 * 桌面端（≥ xl）自动隐藏。
 */
export function MobileSessionDrawer({
	sessions,
	selectedId,
	onSelect,
}: {
	sessions: Session[]
	selectedId: string
	onSelect: (id: string) => void
}) {
	const { open, setOpen } = useSidebarOpen()
	const [isDesktop, setIsDesktop] = useState(false)

	useEffect(() => {
		const mq = window.matchMedia('(min-width: 1280px)')
		setIsDesktop(mq.matches)
		const handler = (e: MediaQueryListEvent) => setIsDesktop(e.matches)
		mq.addEventListener('change', handler)
		return () => mq.removeEventListener('change', handler)
	}, [])

	if (isDesktop) return null

	return (
		<Sheet open={open} onOpenChange={setOpen}>
			<SheetContent
				side="right"
				showCloseButton={false}
				overlayClassName="bg-black/20 backdrop-blur-sm"
				className="m-3 h-[calc(100%-1.5rem)] w-[340px] rounded-2xl border border-line bg-surface-strong shadow-[0_24px_80px_rgba(0,0,0,0.18)] data-[state=open]:duration-300 data-[state=closed]:duration-200"
			>
				<SessionPanel
					sessions={sessions}
					selectedId={selectedId}
					onSelect={(id) => {
						setOpen(false)
						onSelect(id)
					}}
					onClose={() => setOpen(false)}
				/>
			</SheetContent>
		</Sheet>
	)
}
