import { Check, SkipForward, X } from "lucide-react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";

import type { QuestionComponentProps } from "./question-select";

export function QuestionBoolean({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const yesLabel =
		(question.config?.yesLabel as string | undefined) || "是";
	const noLabel =
		(question.config?.noLabel as string | undefined) || "否";

	return (
		<div className="space-y-4">
			{/* Question content */}
			<div className="prose prose-sm max-w-none [&_p]:mb-1 [&_p]:text-sm [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink)]">
				<Markdown remarkPlugins={[remarkGfm]}>{question.content}</Markdown>
			</div>
			{question.description && (
				<div className="prose prose-sm max-w-none [&_p]:mt-1 [&_p]:mb-3 [&_p]:text-xs [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink-soft)]">
					<Markdown remarkPlugins={[remarkGfm]}>
						{question.description}
					</Markdown>
				</div>
			)}

			{/* Yes / No buttons */}
			<div className="flex gap-3">
				<button
					type="button"
					onClick={() => onSubmit({ choice: "yes" })}
					className="inline-flex flex-1 items-center justify-center gap-1.5 rounded-lg border border-emerald-500/25 bg-emerald-500/8 px-4 py-2.5 text-sm font-semibold text-emerald-700 transition-all duration-150 hover:border-emerald-500/45 hover:bg-emerald-500/14 active:scale-[0.98]"
				>
					<Check className="size-4" aria-hidden />
					{yesLabel}
				</button>
				<button
					type="button"
					onClick={() => onSubmit({ choice: "no" })}
					className="inline-flex flex-1 items-center justify-center gap-1.5 rounded-lg border border-red-500/25 bg-red-500/8 px-4 py-2.5 text-sm font-semibold text-red-700 transition-all duration-150 hover:border-red-500/45 hover:bg-red-500/14 active:scale-[0.98]"
				>
					<X className="size-4" aria-hidden />
					{noLabel}
				</button>
			</div>

			{/* Skip button */}
			<div className="pt-1">
				<div className="flex items-center justify-between">
					<Button
						variant="ghost"
						size="sm"
						onClick={() =>
							onRequestSupplement([question.id])
						}
						className="text-xs text-[var(--sea-ink-soft)]"
					>
						请求补充信息
					</Button>
					<Button
						variant="outline"
						size="sm"
						onClick={onSkip}
					>
						<SkipForward className="mr-1 size-3.5" aria-hidden />
						跳过
					</Button>
				</div>
			</div>
		</div>
	);
}
