import { FileText } from 'lucide-react'
import { useState } from 'react'
import ReactDiffViewer, { DiffMethod } from 'react-diff-viewer-continued'

import { Textarea } from '@lumina/components/ui/textarea'

import { DecisionButtons } from './decision-buttons'
import { QuestionShell } from './question-shell'
import type { QuestionComponentProps } from './question-shell'

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

			{isEditing ? (
				<div className="space-y-1.5">
					<p className="text-xs font-semibold uppercase tracking-wide text-amber-500">
						编辑模式 — 修改后内容
					</p>
					<Textarea
						value={editedCode || after}
						onChange={(e) => setEditedCode(e.target.value)}
						disabled={isSupplementLoading}
						className="min-h-[200px] resize-y rounded-lg border-line bg-foam font-mono text-xs tabular-nums disabled:opacity-50"
						placeholder="编辑代码..."
					/>
				</div>
			) : (
				<div className="max-h-[400px] overflow-auto rounded-lg border border-line">
					<ReactDiffViewer
						oldValue={before}
						newValue={after}
						splitView={false}
						compareMethod={DiffMethod.WORDS}
						useDarkTheme={false}
						hideLineNumbers={false}
						showDiffOnly={false}
						styles={{
							variables: {
								light: {
									diffViewerBackground: '#ffffff',
									diffViewerColor: '#2a2420',
									addedBackground: '#f0fdf4',
									addedColor: '#166534',
									removedBackground: '#fef2f2',
									removedColor: '#991b1b',
									wordAddedBackground: '#bbf7d0',
									wordRemovedBackground: '#fecaca',
									addedGutterBackground: '#dcfce7',
									removedGutterBackground: '#fee2e2',
									gutterBackground: '#fafaf9',
									diffViewerTitleBackground: '#f5f5f4',
									diffViewerTitleColor: '#57534e',
									diffViewerTitleBorderColor: '#e7e5e4',
								},
							},
							diffContainer: {
								minWidth: '100%',
							},
							line: {
								padding: '2px 10px',
							},
							contentText: {
								fontSize: '12px',
								fontFamily:
									'ui-monospace, SFMono-Regular, Menlo, Monaco, monospace',
							},
						}}
					/>
				</div>
			)}

			{language && (
				<p className="text-center text-[10px] text-sea-ink-soft">
					语言：{language}
				</p>
			)}

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
