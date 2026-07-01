import { Loader2 } from 'lucide-react'

/**
 * 空状态 —— 收敛 index.tsx 内联的 EmptyCard。
 */
export function EmptyState({ text }: { text: string }) {
	return (
		<div className="rounded-xl border border-line bg-surface px-4 py-12 text-center shadow-[0_1px_3px_rgba(42,36,32,0.04),0_8px_24px_-8px_rgba(42,36,32,0.10)] backdrop-blur-sm">
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
		<div className="rounded-xl border border-line bg-surface px-4 py-12 shadow-[0_1px_3px_rgba(42,36,32,0.04),0_8px_24px_-8px_rgba(42,36,32,0.10)] backdrop-blur-sm">
			<div className="flex flex-col items-center justify-center gap-3">
				<Loader2 className="size-6 animate-spin text-sea-ink-soft/40" aria-hidden />
				<p className="text-xs text-sea-ink-soft/50">{text}</p>
			</div>
		</div>
	)
}
