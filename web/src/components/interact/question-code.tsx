import { Code2, Send, SkipForward } from "lucide-react";
import { useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";
import { Label } from "#/components/ui/label";
import { Textarea } from "#/components/ui/textarea";

import type { QuestionComponentProps } from "./question-select";

export function QuestionCode({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const [code, setCode] = useState("");
	const language = question.config?.language as string | undefined;
	const placeholder =
		(question.config?.placeholder as string | undefined) ||
		"输入代码...";

	const handleSubmit = () => {
		if (!code.trim()) return;
		const answer: any = { code };
		if (language) answer.language = language;
		onSubmit(answer);
	};

	return (
		<div className="space-y-4">
			{/* Question content */}
			<div className="prose prose-sm max-w-none [&_p]:mb-1 [&_p]:text-sm [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink)]">
				<Markdown remarkPlugins={[remarkGfm]}>{question.content}</Markdown>
			</div>
			{question.description && (
				<div className="prose prose-sm max-w-none [&_p]:mt-1 [&_p]:mb-2 [&_p]:text-xs [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink-soft)]">
					<Markdown remarkPlugins={[remarkGfm]}>
						{question.description}
					</Markdown>
				</div>
			)}

			{/* Language badge */}
			{language && (
				<div className="flex items-center gap-2">
					<Code2 className="size-3.5 text-[var(--lagoon-deep)]" />
					<Label className="inline-flex items-center rounded-md bg-[var(--lagoon)]/10 px-2 py-0.5 text-xs font-mono font-semibold text-[var(--lagoon-deep)]">
						{language}
					</Label>
				</div>
			)}

			{/* Code textarea */}
			<Textarea
				placeholder={placeholder}
				value={code}
				onChange={(e) => setCode(e.target.value)}
				className="min-h-[160px] resize-y rounded-lg border-[var(--line)] bg-[var(--foam)] font-mono text-sm tabular-nums leading-relaxed"
				spellCheck={false}
			/>

			{/* Actions */}
			<div className="flex items-center justify-between pt-1">
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
				<div className="flex gap-2">
					<Button
						variant="outline"
						size="sm"
						onClick={onSkip}
					>
						<SkipForward className="mr-1 size-3.5" aria-hidden />
						跳过
					</Button>
					<Button
						size="sm"
						onClick={handleSubmit}
						disabled={!code.trim()}
						className="rounded-lg bg-[var(--lagoon)] text-white hover:bg-[var(--lagoon)]/90"
					>
						<Send className="mr-1.5 size-3.5" aria-hidden />
						提交
					</Button>
				</div>
			</div>
		</div>
	);
}
