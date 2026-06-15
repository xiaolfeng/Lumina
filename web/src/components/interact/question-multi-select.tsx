import { Pencil } from 'lucide-react'
import { useState } from 'react'

import { Checkbox } from '#/components/ui/checkbox'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { SupplementLoadingBanner } from './supplement-loading-banner'

import { OptionDetailLabel } from './option-detail-label'
import { QuestionShell  } from './question-shell'
import type {QuestionComponentProps} from './question-shell';

export function QuestionMultiSelect({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
	onDismissSupplementLoading,
	onViewOptionDetail,
}: QuestionComponentProps) {
	const [selected, setSelected] = useState<Set<string>>(new Set())
	const [otherChecked, setOtherChecked] = useState(false)
	const [otherText, setOtherText] = useState('')

	const options = question.options ?? []
	const minSelect = question.config?.min ?? 0
	const maxSelect = question.config?.max ?? options.length
	const hasSupplements = (question.supplements?.length ?? 0) > 0

	const hasOptionSupplement = (optId: string): boolean => {
		return (
			question.supplements?.some(
				(s) => s.target_type === 'option' && s.target_id === optId,
			) ?? false
		)
	}

	const toggleOption = (id: string) => {
		setSelected((prev) => {
			const next = new Set(prev)
			if (next.has(id)) {
				next.delete(id)
			} else if (!maxSelect || next.size < maxSelect) {
				next.add(id)
			}
			return next
		})
	}

	const handleSubmit = () => {
		const result: { selected: string[]; other?: string[] } = {
			selected: Array.from(selected),
		}
		if (otherChecked && otherText.trim()) {
			result.other = [otherText.trim()]
		}
		onSubmit(result)
	}

	const canSubmit =
		selected.size >= minSelect && (!otherChecked || !!otherText.trim())

	return (
		<QuestionShell
			question={question}
			isSupplementLoading={isSupplementLoading}
			onSkip={onSkip}
			onRequestSupplement={onRequestSupplement}
			supplement={{ kind: 'options', optionCount: options.length, optionIds: options.map((o) => o.id) }}
			supplementButtonLabel={hasSupplements ? '重新获取详情' : '请求补充信息'}
			submitDisabled={!canSubmit}
			onSubmit={handleSubmit}
		>
			{(minSelect > 0 || maxSelect < options.length) && (
				<p className="text-xs text-sea-ink-soft">
					已选 {selected.size} / {maxSelect > 0 ? maxSelect : '无限制'}
					{minSelect > 0 && `（至少 ${minSelect} 项）`}
				</p>
			)}

			{isSupplementLoading && (
				<SupplementLoadingBanner onDismiss={() => onDismissSupplementLoading?.()} />
			)}
		<div
			className="space-y-2"
		>
				{options.map((opt) => (
					<Label
						key={opt.id}
						htmlFor={`multi-${question.id}-${opt.id}`}
						className="flex cursor-pointer items-start gap-3 rounded-lg border border-line bg-foam px-3 py-2.5 transition-colors duration-150 has-[data-state=checked]:border-lagoon/30 has-[data-state=checked]:bg-lagoon/10 hover:border-lagoon/30"
					>
						<Checkbox
							id={`multi-${question.id}-${opt.id}`}
							checked={selected.has(opt.id)}
							onCheckedChange={() => toggleOption(opt.id)}
							className="mt-0.5"
						/>
						<div className="min-w-0 flex-1">
							<div className="flex items-start gap-2">
								<p className="min-w-0 flex-1 text-sm font-medium leading-snug">
									{opt.label}
								</p>
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
					</Label>
				))}

				{question.allowOther && (
					<>
						<Label
							htmlFor={`multi-${question.id}-other`}
							className="flex cursor-pointer items-start gap-3 rounded-lg border border-line bg-foam px-3 py-2.5 transition-colors duration-150 has-[data-state=checked]:border-lagoon/30 has-[data-state=checked]:bg-lagoon/10 hover:border-lagoon/30"
						>
							<Checkbox
								id={`multi-${question.id}-other`}
								checked={otherChecked}
								onCheckedChange={(v) => setOtherChecked(!!v)}
								className="mt-0.5"
							/>
							<Pencil className="size-3.5 shrink-0 text-lagoon-deep" />
							<span className="text-sm font-medium">其他</span>
						</Label>
						{otherChecked && (
							<Input
								placeholder="输入自定义选项..."
								value={otherText}
								onChange={(e) => setOtherText(e.target.value)}
								className="rounded-lg border-line bg-foam"
							/>
						)}
					</>
				)}
			</div>
		</QuestionShell>
	)
}
