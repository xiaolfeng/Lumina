import { useState } from 'react'

import { Textarea } from '@lumina/components/ui/textarea'

import { DecisionButtons } from './decision-buttons'
import { Markdown, proseQuestion } from './primitives'
import { QuestionShell  } from './question-shell'
import type {QuestionComponentProps} from './question-shell';

type ReviewDecision = 'approve' | 'revise'

export function QuestionReview({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
}: QuestionComponentProps) {
	const [decision, setDecision] = useState<ReviewDecision | null>(null)
	const [feedback, setFeedback] = useState('')

	const content = (question.config?.content as string) ?? question.content
	const context =
		(question.config?.context as string) ?? question.description ?? ''
	const isRevising = decision === 'revise'

	const handleSubmit = () => {
		if (!decision) return
		onSubmit({
			decision,
			...(feedback.trim() ? { feedback: feedback.trim() } : {}),
		})
	}

	return (
		<QuestionShell
			question={question}
			isSupplementLoading={isSupplementLoading}
			onSkip={onSkip}
			onRequestSupplement={onRequestSupplement}
			showDescription={false}
			submitDisabled={!decision || (isRevising && !feedback.trim())}
			onSubmit={handleSubmit}
		>
			{context && (
				<div className="rounded-lg border border-amber-200 bg-amber-50/80 px-3 py-2 dark:border-amber-800/40 dark:bg-amber-900/15">
					<p className="text-[11px] font-semibold text-amber-700 dark:text-amber-400">
						上下文
					</p>
					<div className="prose prose-sm mt-1 max-w-none [&_p]:mb-0 [&_p]:text-xs [&_p]:leading-relaxed [&_p]:text-amber-600 dark:[&_p]:text-amber-300 [&_code]:rounded [&_code]:bg-lagoon/8 [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:text-xs [&_code]:text-lagoon-deep">
						<Markdown>{context}</Markdown>
					</div>
				</div>
			)}

			<div className="max-h-[420px] overflow-auto rounded-lg border border-line bg-foam p-4">
				<div className={proseQuestion}>
					<Markdown>{content}</Markdown>
				</div>
			</div>

			<DecisionButtons
				variant="two"
				value={decision}
				onChange={(v) => setDecision(v as ReviewDecision)}
				disabled={isSupplementLoading}
			/>

			{isRevising && (
				<Textarea
					placeholder="请描述需要修改的内容..."
					value={feedback}
					onChange={(e) => setFeedback(e.target.value)}
					disabled={isSupplementLoading}
					className="min-h-[80px] resize-y rounded-lg border-line bg-foam text-sm disabled:opacity-50"
				/>
			)}
		</QuestionShell>
	)
}
