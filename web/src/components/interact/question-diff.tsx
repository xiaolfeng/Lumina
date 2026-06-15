import { FileText } from 'lucide-react'
import { useState } from 'react'

import { Textarea } from '#/components/ui/textarea'

import { DecisionButtons } from './decision-buttons'
import { QuestionShell  } from './question-shell'
import type {QuestionComponentProps} from './question-shell';

type Decision = 'approve' | 'reject' | 'edit'

export function QuestionDiff({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
}: QuestionComponentProps) {
	const [decision, setDecision] = useState<Decision | null>(null)
	const [editedCode, setEditedCode] = useState('')
	const [feedback, setFeedback] = useState('')

	const before = (question.config?.before as string) ?? ''
	const after = (question.config?.after as string) ?? ''
	const filePath = (question.config?.filePath as string) ?? ''
	const language = (question.config?.language as string) ?? ''

	const isEditing = decision === 'edit'

	const handleSubmit = () => {
		if (!decision) return
		onSubmit({
			decision,
			...(isEditing ? { edited: editedCode || after } : {}),
			...(feedback.trim() ? { feedback: feedback.trim() } : {}),
		})
	}

	return (
		<QuestionShell
			question={question}
			isSupplementLoading={isSupplementLoading}
			onSkip={onSkip}
			onRequestSupplement={onRequestSupplement}
			submitDisabled={!decision}
			onSubmit={handleSubmit}
		>
			{filePath && (
				<p className="inline-flex items-center gap-1.5 rounded-md bg-foam px-3 py-1.5 font-mono text-xs text-sea-ink-soft">
					<FileText className="size-3.5" aria-hidden />
					{filePath}
				</p>
			)}

			<div className="grid grid-cols-1 gap-3 lg:grid-cols-2">
				<div className="space-y-1.5">
					<p className="text-xs font-semibold uppercase tracking-wide text-red-400">
						修改前 (Before)
					</p>
					<pre className="max-h-[360px] overflow-auto rounded-lg bg-gray-900 p-3 text-xs leading-relaxed text-gray-200">
						<code>{before || '(无内容)'}</code>
					</pre>
				</div>

				<div className="space-y-1.5">
					<p className="text-xs font-semibold uppercase tracking-wide text-emerald-400">
						修改后 (After)
						{language && (
							<span className="ml-1.5 font-normal text-gray-400">
								({language})
							</span>
						)}
					</p>
					{isEditing ? (
						<Textarea
							value={editedCode || after}
							onChange={(e) => setEditedCode(e.target.value)}
							disabled={isSupplementLoading}
							className="min-h-[200px] resize-y rounded-lg border-line bg-foam font-mono text-xs tabular-nums disabled:opacity-50"
							placeholder="编辑代码..."
						/>
					) : (
						<pre className="max-h-[360px] overflow-auto rounded-lg bg-gray-900 p-3 text-xs leading-relaxed text-gray-200">
							<code>{after || '(无内容)'}</code>
						</pre>
					)}
				</div>
			</div>

			<DecisionButtons
				variant="three"
				value={decision}
				onChange={(v) => setDecision(v as Decision)}
				disabled={isSupplementLoading}
			/>

			<Textarea
				placeholder="可选：添加反馈意见..."
				value={feedback}
				onChange={(e) => setFeedback(e.target.value)}
				disabled={isSupplementLoading}
				className="min-h-[60px] resize-y rounded-lg border-line bg-foam text-sm disabled:opacity-50"
			/>
		</QuestionShell>
	)
}
