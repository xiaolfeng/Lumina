import { Star } from 'lucide-react'
import { useState } from 'react'

import { QuestionShell  } from './question-shell'
import type {QuestionComponentProps} from './question-shell';

export function QuestionRate({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
}: QuestionComponentProps) {
	const options = question.options ?? []
	const min = (question.config?.min as number) ?? 1
	const max = (question.config?.max as number) ?? 5
	const step = (question.config?.step as number) ?? 1
	const useStars = (question.config?.useStars as boolean) ?? true

	const [ratings, setRatings] = useState<Record<string, number>>(() => {
		const initial: Record<string, number> = {}
		options.forEach((opt) => {
			initial[opt.id] = min
		})
		return initial
	})

	const setRating = (optionId: string, value: number) => {
		setRatings((prev) => ({ ...prev, [optionId]: value }))
	}

	const renderRatingControl = (optionId: string, current: number) => {
		if (useStars) {
			const count = Math.round((max - min) / step) + 1
			return (
				<div className="flex gap-1">
					{Array.from({ length: count }, (_, i) => {
						const val = min + i * step
						const filled = val <= current
						return (
							<button
								key={val}
								type="button"
								onClick={() => setRating(optionId, val)}
								disabled={isSupplementLoading}
								className="group p-0.5 transition-transform hover:scale-110 disabled:cursor-not-allowed disabled:opacity-50"
								aria-label={`${val} 分`}
							>
								<Star
									className={`size-6 transition-colors ${
										filled
											? 'fill-amber-400 text-amber-400'
											: 'fill-transparent text-gray-300 group-hover:text-amber-200'
									}`}
								/>
							</button>
						)
					})}
				</div>
			)
		}
		return (
			<input
				type="number"
				min={min}
				max={max}
				step={step}
				value={current}
				onChange={(e) => setRating(optionId, Number(e.target.value))}
				disabled={isSupplementLoading}
				className="w-24 rounded-lg border border-line bg-foam px-3 py-1.5 text-center font-mono text-sm tabular-nums disabled:opacity-50"
			/>
		)
	}

	return (
		<QuestionShell
			question={question}
			isSupplementLoading={isSupplementLoading}
			onSkip={onSkip}
			onRequestSupplement={onRequestSupplement}
			onSubmit={() => onSubmit({ ratings: { ...ratings } })}
		>
			<div
				className={`space-y-3 ${isSupplementLoading ? 'pointer-events-none opacity-50' : ''}`}
			>
				{options.map((opt) => (
					<div
						key={opt.id}
						className="flex flex-col gap-2 rounded-lg border border-line bg-foam px-4 py-3 sm:flex-row sm:items-center sm:justify-between"
					>
						<div className="min-w-0 flex-1">
							<p className="text-sm font-medium">{opt.label}</p>
							{opt.description && (
								<p className="text-xs text-sea-ink-soft">{opt.description}</p>
							)}
						</div>
						<div className="flex items-center gap-3 sm:shrink-0">
							{renderRatingControl(opt.id, ratings[opt.id] ?? min)}
							<span className="w-8 text-center font-mono text-sm font-semibold text-lagoon-deep">
								{ratings[opt.id] ?? min}
							</span>
						</div>
					</div>
				))}
			</div>

			<p className="text-center text-[11px] text-sea-ink-soft">
				评分范围：{min} — {max}
				{step !== 1 && `（步长 ${step}）`}
			</p>
		</QuestionShell>
	)
}
