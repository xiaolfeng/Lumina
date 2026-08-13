import { Loader2 } from 'lucide-react'

/**
 * 空状态 —— 收敛 index.tsx 内联的 EmptyCard。
 */
export function EmptyState({ text }: { text: string }) {
	return (
		<div className="border border-line bg-surface px-4 py-12 text-center">
			<p className="text-xs text-sea-ink-soft/50">{text}</p>
		</div>
	)
}

/**
 * 加载状态 —— 收敛 index.tsx 内联的 LoadingCard。
 * 用 Loader2 替代手写 spinner，统一图标体系。
 */
export function LoadingState({ text }: { text: string }) {
	return (
		<div className="border border-line bg-surface px-4 py-12">
			<div className="flex flex-col items-center justify-center gap-3">
				<Loader2 className="size-6 animate-spin text-sea-ink-soft/40" aria-hidden />
				<p className="text-xs text-sea-ink-soft/50">{text}</p>
			</div>
		</div>
	)
}
