import { ChevronDown, ChevronRight } from 'lucide-react'
import { useState } from 'react'

import { Textarea } from '@lumina/components/ui/textarea'

import { DecisionButtons } from './decision-buttons'
import { Markdown, proseQuestion } from './primitives'
import { QuestionShell  } from './question-shell'
import { SupplementLoadingBanner } from './supplement-loading-banner'
import type {QuestionComponentProps} from './question-shell';

interface PlanSection {
	id: string
	title: string
	content: string
}

type PlanDecision = 'approve' | 'reject' | 'revise'

export function QuestionPlan({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
	onDismissSupplementLoading,
}: QuestionComponentProps) {
	const [decision, setDecision] = useState<PlanDecision | null>(null)
	const [annotations, setAnnotations] = useState<Record<string, string>>({})
	const [feedback, setFeedback] = useState('')

	const sections = (question.config?.sections as PlanSection[]) ?? []
	const markdown = (question.config?.markdown as string) ?? ''
	const isRevising = decision === 'revise'

	const [expandedSections, setExpandedSections] = useState<Set<string>>(
		() => new Set(sections.map((s) => s.id)),
	)

	const toggleSection = (id: string) => {
		setExpandedSections((prev) => {
			const next = new Set(prev)
			if (next.has(id)) next.delete(id)
			else next.add(id)
			return next
		})
	}

	const handleSubmit = () => {
		if (!decision) return
		const annotationsList = Object.entries(annotations)
			.filter(([, v]) => v.trim())
			.map(([sectionId, content]) => ({ sectionId, content }))
		onSubmit({
			decision,
			...(annotationsList.length > 0 ? { annotations: annotationsList } : {}),
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
			{isSupplementLoading && (
				<SupplementLoadingBanner onDismiss={() => onDismissSupplementLoading?.()} />
			)}
			{sections.length > 0 ? (
				<div
					className={`space-y-2 ${isSupplementLoading ? 'pointer-events-none opacity-50' : ''}`}
				>
					{sections.map((section) => {
						const isOpen = expandedSections.has(section.id)
						return (
							<div
								key={section.id}
								className="overflow-hidden rounded-lg border border-line bg-foam"
							>
								<button
									type="button"
									onClick={() => toggleSection(section.id)}
									disabled={isSupplementLoading}
									className="flex w-full items-center gap-2 px-3 py-2.5 text-left transition-colors hover:bg-line/30 disabled:cursor-not-allowed disabled:opacity-50"
								>
									{isOpen ? (
										<ChevronDown className="size-4 shrink-0 text-sea-ink-soft" />
									) : (
										<ChevronRight className="size-4 shrink-0 text-sea-ink-soft" />
									)}
									<span className="text-sm font-medium">{section.title}</span>
								</button>
								{isOpen && (
									<div className="border-t border-line/50 px-3 pb-3">
										<div className={`mt-2 ${proseQuestion}`}>
											<Markdown>{section.content}</Markdown>
										</div>
										{isRevising && (
											<Textarea
												placeholder={`对「${section.title}」的修订意见...`}
												value={annotations[section.id] ?? ''}
												onChange={(e) =>
													setAnnotations((prev) => ({
														...prev,
														[section.id]: e.target.value,
													}))
												}
												disabled={isSupplementLoading}
												className="mt-2 min-h-[50px] resize-y rounded border-line bg-white text-xs disabled:opacity-50"
											/>
										)}
									</div>
								)}
							</div>
						)
					})}
				</div>
			) : markdown ? (
				<div className="max-h-[400px] overflow-auto rounded-lg border border-line bg-foam p-4">
					<div className={proseQuestion}>
						<Markdown>{markdown}</Markdown>
					</div>
				</div>
			) : null}

			<DecisionButtons
				variant="three"
				value={decision}
				onChange={(v) => setDecision(v as PlanDecision)}
				disabled={isSupplementLoading}
				thirdLabel="修订"
			/>

			<Textarea
				placeholder="可选：整体反馈意见..."
				value={feedback}
				onChange={(e) => setFeedback(e.target.value)}
				disabled={isSupplementLoading}
				className="min-h-[60px] resize-y rounded-lg border-line bg-foam text-sm disabled:opacity-50"
			/>
		</QuestionShell>
	)
}
