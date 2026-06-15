import { Star } from 'lucide-react'
import { useState } from 'react'

import { Label } from '#/components/ui/label'
import { RadioGroup, RadioGroupItem } from '#/components/ui/radio-group'
import { Textarea } from '#/components/ui/textarea'
import { SupplementLoadingBanner } from './supplement-loading-banner'

import { OptionDetailLabel } from './option-detail-label'
import { QuestionShell  } from './question-shell'
import type {QuestionComponentProps} from './question-shell';
import type { OptionItem } from './types'

interface OptionWithProsCons extends OptionItem {
	pros?: string[]
	cons?: string[]
	recommended?: boolean
}

export function QuestionOptions({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
	onDismissSupplementLoading,
	onViewOptionDetail,
}: QuestionComponentProps) {
	const [selected, setSelected] = useState<string>('')
	const [feedback, setFeedback] = useState('')

	const options = (question.options ?? []) as OptionWithProsCons[]
	const hasSupplements = (question.supplements?.length ?? 0) > 0

	const hasOptionSupplement = (optId: string): boolean => {
		return (
			question.supplements?.some(
				(s) => s.target_type === 'option' && s.target_id === optId,
			) ?? false
		)
	}

	const handleSubmit = () => {
		if (!selected) return
		onSubmit({
			selected,
			...(feedback.trim() ? { feedback: feedback.trim() } : {}),
		})
	}

	return (
		<QuestionShell
			question={question}
			isSupplementLoading={isSupplementLoading}
			onSkip={onSkip}
			onRequestSupplement={onRequestSupplement}
			supplementButtonLabel={hasSupplements ? '重新获取详情' : '请求补充信息'}
			submitDisabled={!selected}
			onSubmit={handleSubmit}
		>
			{isSupplementLoading && (
				<SupplementLoadingBanner onDismiss={() => onDismissSupplementLoading?.()} />
			)}
			<RadioGroup
				value={selected}
				onValueChange={setSelected}
				className="space-y-3"
				disabled={isSupplementLoading}
			>
				{options.map((opt) => (
					<Label
						key={opt.id}
						htmlFor={`options-${question.id}-${opt.id}`}
						className={`flex cursor-pointer flex-col gap-2 rounded-xl border px-4 py-3 transition-all duration-150 ${
							selected === opt.id
								? 'border-lagoon/40 bg-lagoon/8 shadow-sm'
								: 'border-line bg-foam hover:border-lagoon/25'
						}`}
					>
						<div className="flex items-start gap-3">
							<RadioGroupItem
								value={opt.id}
								id={`options-${question.id}-${opt.id}`}
								className="mt-0.5"
							/>
							<div className="min-w-0 flex-1">
								<div className="flex items-start gap-2">
									<span className="min-w-0 flex-1 text-sm font-medium">
										{opt.label}
									</span>
									{opt.recommended && (
										<span className="inline-flex shrink-0 items-center gap-0.5 rounded-full bg-amber-100 px-2 py-0.5 text-[10px] font-semibold text-amber-700">
											<Star className="size-2.5" aria-hidden />
											推荐
										</span>
									)}
									{hasOptionSupplement(opt.id) && (
										<OptionDetailLabel
											optId={opt.id}
											onClick={() => onViewOptionDetail?.(opt.id)}
										/>
									)}
								</div>
								{opt.description && (
									<p className="mt-0.5 text-xs leading-relaxed text-sea-ink-soft">
										{opt.description}
									</p>
								)}
							</div>
						</div>

						{(opt.pros?.length ?? 0) > 0 || (opt.cons?.length ?? 0) > 0 ? (
							<div className="ml-7 grid grid-cols-1 gap-2 sm:grid-cols-2">
								{opt.pros && opt.pros.length > 0 && (
									<div className="space-y-1">
										<p className="text-[10px] font-semibold uppercase tracking-wide text-emerald-600">
											优点
										</p>
										<ul className="space-y-0.5">
											{opt.pros.map((pro, idx) => (
												<li
													key={idx}
													className="flex items-start gap-1.5 text-xs text-emerald-700"
												>
													<span className="mt-1.5 size-1 shrink-0 rounded-full bg-emerald-500" />
													{pro}
												</li>
											))}
										</ul>
									</div>
								)}
								{opt.cons && opt.cons.length > 0 && (
									<div className="space-y-1">
										<p className="text-[10px] font-semibold uppercase tracking-wide text-red-500">
											缺点
										</p>
										<ul className="space-y-0.5">
											{opt.cons.map((con, idx) => (
												<li
													key={idx}
													className="flex items-start gap-1.5 text-xs text-red-600"
												>
													<span className="mt-1.5 size-1 shrink-0 rounded-full bg-red-400" />
													{con}
												</li>
											))}
										</ul>
									</div>
								)}
							</div>
						) : null}
					</Label>
				))}
			</RadioGroup>

			<Textarea
				placeholder="可选：说明你的选择理由..."
				value={feedback}
				onChange={(e) => setFeedback(e.target.value)}
				className="min-h-[60px] resize-y rounded-lg border-line bg-foam text-sm"
			/>
		</QuestionShell>
	)
}
