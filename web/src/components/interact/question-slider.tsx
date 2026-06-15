import { useState } from 'react'

import { QuestionShell  } from './question-shell'
import type {QuestionComponentProps} from './question-shell';

export function QuestionSlider({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
}: QuestionComponentProps) {
	const min = (question.config?.min as number) ?? 0
	const max = (question.config?.max as number) ?? 100
	const step = (question.config?.step as number) ?? 1
	const defaultValue = (question.config?.defaultValue as number) ?? min
	const minLabel = (question.config?.minLabel as string) ?? String(min)
	const maxLabel = (question.config?.maxLabel as string) ?? String(max)

	const [value, setValue] = useState(defaultValue)
	const percentage = ((value - min) / (max - min)) * 100

	return (
		<QuestionShell
			question={question}
			isSupplementLoading={isSupplementLoading}
			onSkip={onSkip}
			onRequestSupplement={onRequestSupplement}
			onSubmit={() => onSubmit({ value })}
		>
			<div
				className={`space-y-3 rounded-xl border border-line bg-foam px-5 py-6 ${isSupplementLoading ? 'pointer-events-none opacity-50' : ''}`}
			>
				<div className="text-center">
					<span className="inline-flex items-center justify-center rounded-full bg-lagoon/10 px-4 py-1.5 font-mono text-2xl font-bold text-lagoon-deep">
						{value}
					</span>
				</div>
				<input
					type="range"
					min={min}
					max={max}
					step={step}
					value={value}
					onChange={(e) => setValue(Number(e.target.value))}
					disabled={isSupplementLoading}
					className="slider w-full accent-lagoon"
					style={{
						background: `linear-gradient(to right, var(--lagoon) ${percentage}%, var(--line) ${percentage}%)`,
						borderRadius: '9999px',
						height: '6px',
						appearance: 'none',
						cursor: 'pointer',
					}}
				/>
				<div className="flex items-center justify-between text-xs text-sea-ink-soft">
					<span>{minLabel}</span>
					<span>{maxLabel}</span>
				</div>
			</div>
		</QuestionShell>
	)
}
