import { Check, SkipForward, X } from 'lucide-react'

import { Button } from '@lumina/components/ui/button'
import { QuestionShell  } from './question-shell'
import type {QuestionComponentProps} from './question-shell';

export function QuestionBoolean({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
}: QuestionComponentProps) {
	const yesLabel = (question.config?.yesLabel as string | undefined) || '是'
	const noLabel = (question.config?.noLabel as string | undefined) || '否'

	// boolean 题用 yes/no 直接提交，无标准提交按钮 —— 自定义 actions（补充 + 跳过）
	return (
		<QuestionShell
			question={question}
			isSupplementLoading={isSupplementLoading}
			onSkip={onSkip}
			onRequestSupplement={onRequestSupplement}
			onSubmit={() => {}}
			actions={(openDialog) => (
				<div className="flex items-center justify-between pt-1">
					<Button
						variant="ghost"
						size="sm"
						onClick={openDialog}
						disabled={isSupplementLoading}
						className="text-xs text-sea-ink-soft"
					>
						请求补充信息
					</Button>
					<Button
						variant="outline"
						size="sm"
						onClick={onSkip}
						disabled={isSupplementLoading}
					>
						<SkipForward className="mr-1 size-3.5" aria-hidden />
						跳过
					</Button>
				</div>
			)}
		>
			<div
				className={`flex gap-3 ${isSupplementLoading ? 'pointer-events-none opacity-50' : ''}`}
			>
				<button
					type="button"
					onClick={() => onSubmit({ choice: 'yes' })}
					disabled={isSupplementLoading}
					className="inline-flex flex-1 items-center justify-center gap-1.5 rounded-lg border border-emerald-500/25 bg-emerald-500/8 px-4 py-2.5 text-sm font-semibold text-emerald-700 transition-all duration-150 hover:border-emerald-500/45 hover:bg-emerald-500/14 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50"
				>
					<Check className="size-4" aria-hidden />
					{yesLabel}
				</button>
				<button
					type="button"
					onClick={() => onSubmit({ choice: 'no' })}
					disabled={isSupplementLoading}
					className="inline-flex flex-1 items-center justify-center gap-1.5 rounded-lg border border-red-500/25 bg-red-500/8 px-4 py-2.5 text-sm font-semibold text-red-700 transition-all duration-150 hover:border-red-500/45 hover:bg-red-500/14 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50"
				>
					<X className="size-4" aria-hidden />
					{noLabel}
				</button>
			</div>
		</QuestionShell>
	)
}
