import { Code2 } from 'lucide-react'
import { useState } from 'react'

import { Label } from '#/components/ui/label'
import { Textarea } from '#/components/ui/textarea'

import { QuestionShell  } from './question-shell'
import type {QuestionComponentProps} from './question-shell';

export function QuestionCode({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
}: QuestionComponentProps) {
	const [code, setCode] = useState('')
	const language = question.config?.language as string | undefined
	const placeholder =
		(question.config?.placeholder as string | undefined) || '输入代码...'

	const handleSubmit = () => {
		if (!code.trim()) return
		const answer: { code: string; language?: string } = { code }
		if (language) answer.language = language
		onSubmit(answer)
	}

	return (
		<QuestionShell
			question={question}
			isSupplementLoading={isSupplementLoading}
			onSkip={onSkip}
			onRequestSupplement={onRequestSupplement}
			submitDisabled={!code.trim()}
			onSubmit={handleSubmit}
		>
			{language && (
				<div className="flex items-center gap-2">
					<Code2 className="size-3.5 text-lagoon-deep" />
					<Label className="inline-flex items-center rounded-md bg-lagoon/10 px-2 py-0.5 font-mono text-xs font-semibold text-lagoon-deep">
						{language}
					</Label>
				</div>
			)}

			<Textarea
				placeholder={placeholder}
				value={code}
				onChange={(e) => setCode(e.target.value)}
				disabled={isSupplementLoading}
				className="min-h-[160px] resize-y rounded-lg border-line bg-foam font-mono text-sm tabular-nums leading-relaxed"
				spellCheck={false}
			/>
		</QuestionShell>
	)
}
