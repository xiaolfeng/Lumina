import { useState } from 'react'

import { Input } from '#/components/ui/input'
import { Textarea } from '#/components/ui/textarea'

import { QuestionShell  } from './question-shell'
import type {QuestionComponentProps} from './question-shell';

export function QuestionText({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
}: QuestionComponentProps) {
	const [value, setValue] = useState('')

	const multiline = question.config?.multiline === true
	const maxLength = question.config?.maxLength as number | undefined
	const placeholder =
		(question.config?.placeholder as string | undefined) || '输入你的回答...'

	const charCount = value.length
	const isOverLimit = maxLength !== undefined && charCount > maxLength

	const InputEl = multiline ? Textarea : Input

	return (
		<QuestionShell
			question={question}
			isSupplementLoading={isSupplementLoading}
			onSkip={onSkip}
			onRequestSupplement={onRequestSupplement}
			submitDisabled={!value.trim() || isOverLimit}
			onSubmit={() => value.trim() && onSubmit({ text: value })}
		>
			<InputEl
				placeholder={placeholder}
				value={value}
				onChange={(e) => setValue(e.target.value)}
				maxLength={maxLength}
				disabled={isSupplementLoading}
				className={
					multiline
						? 'min-h-[100px] resize-y rounded-lg border-line bg-foam'
						: 'rounded-lg border-line bg-foam'
				}
			/>

			{maxLength && (
				<p
					className={`text-right text-xs ${isOverLimit ? 'font-medium text-red-500' : 'text-sea-ink-soft'}`}
				>
					{charCount} / {maxLength}
				</p>
			)}
		</QuestionShell>
	)
}
