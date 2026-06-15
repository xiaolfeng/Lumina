import { ArrowDown, ArrowUp, GripVertical } from 'lucide-react'
import { useState } from 'react'

import { QuestionShell  } from './question-shell'
import type {QuestionComponentProps} from './question-shell';

interface RankItem {
	id: string
	label: string
	description?: string
}

export function QuestionRank({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
}: QuestionComponentProps) {
	const initialItems: RankItem[] = (question.options ?? []).map((o) => ({
		id: o.id,
		label: o.label,
		description: o.description,
	}))

	const [items, setItems] = useState<RankItem[]>(initialItems)
	const [dragIndex, setDragIndex] = useState<number | null>(null)
	const [dragOverIndex, setDragOverIndex] = useState<number | null>(null)

	const handleDragOver = (e: React.DragEvent, index: number) => {
		e.preventDefault()
		setDragOverIndex(index)
	}

	const handleDrop = (dropIndex: number) => {
		if (dragIndex === null || dragIndex === dropIndex) return
		setItems((prev) => {
			const next = [...prev]
			const [removed] = next.splice(dragIndex, 1)
			next.splice(dropIndex, 0, removed)
			return next
		})
	}

	const moveItem = (index: number, direction: 'up' | 'down') => {
		const newIndex = direction === 'up' ? index - 1 : index + 1
		if (newIndex < 0 || newIndex >= items.length) return
		setItems((prev) => {
			const next = [...prev]
			;[next[index], next[newIndex]] = [next[newIndex], next[index]]
			return next
		})
	}

	return (
		<QuestionShell
			question={question}
			isSupplementLoading={isSupplementLoading}
			onSkip={onSkip}
			onRequestSupplement={onRequestSupplement}
			onSubmit={() => onSubmit({ ranking: items.map((item) => item.id) })}
		>
			<ul
				className={`space-y-2 ${isSupplementLoading ? 'pointer-events-none opacity-50' : ''}`}
			>
				{items.map((item, index) => (
					<li
						key={item.id}
						draggable={!isSupplementLoading}
						onDragStart={() => setDragIndex(index)}
						onDragOver={(e) => handleDragOver(e, index)}
						onDragEnd={() => {
							setDragIndex(null)
							setDragOverIndex(null)
						}}
						onDrop={() => handleDrop(index)}
						className={`group flex items-center gap-2.5 rounded-lg border-2 bg-foam px-3 py-2.5 transition-all duration-150 ${
							dragOverIndex === index
								? 'border-dashed border-lagoon bg-lagoon/5'
								: 'border-line'
						} ${dragIndex === index ? 'scale-[0.98] opacity-50' : ''} ${isSupplementLoading ? '' : 'cursor-grab active:cursor-grabbing'}`}
					>
						<GripVertical className="size-4 shrink-0 text-sea-ink-soft group-hover:text-lagoon-deep" />
						<span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-lagoon/10 text-[11px] font-bold text-lagoon-deep">
							{index + 1}
						</span>
						<div className="min-w-0 flex-1">
							<p className="text-sm font-medium">{item.label}</p>
							{item.description && (
								<p className="text-xs leading-relaxed text-sea-ink-soft">
									{item.description}
								</p>
							)}
						</div>
						<div className="flex shrink-0 gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
							<button
								type="button"
								onClick={() => moveItem(index, 'up')}
								disabled={index === 0 || isSupplementLoading}
								className="rounded p-1 text-sea-ink-soft hover:bg-line disabled:opacity-20"
								aria-label="上移"
							>
								<ArrowUp className="size-3.5" />
							</button>
							<button
								type="button"
								onClick={() => moveItem(index, 'down')}
								disabled={index === items.length - 1 || isSupplementLoading}
								className="rounded p-1 text-sea-ink-soft hover:bg-line disabled:opacity-20"
								aria-label="下移"
							>
								<ArrowDown className="size-3.5" />
							</button>
						</div>
					</li>
				))}
			</ul>
		</QuestionShell>
	)
}
