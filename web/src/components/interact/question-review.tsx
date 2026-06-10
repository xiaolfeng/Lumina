import { CheckCircle, Pencil, Send, SkipForward } from "lucide-react";
import { useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";
import { Textarea } from "#/components/ui/textarea";

import type { QuestionComponentProps } from "./question-select";

type ReviewDecision = "approve" | "revise";

export function QuestionReview({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const [decision, setDecision] = useState<ReviewDecision | null>(null);
	const [feedback, setFeedback] = useState("");

	const content = (question.config?.content as string) ?? question.content;
	const context = (question.config?.context as string) ?? question.description ?? "";
	const isRevising = decision === "revise";

	const handleSubmit = () => {
		if (!decision) return;
		onSubmit({
			decision,
			...(feedback.trim() ? { feedback: feedback.trim() } : {}),
		});
	};

	return (
		<div className="space-y-4">
			{/* Question content */}
			<div className="prose prose-sm max-w-none [&_p]:mb-1 [&_p]:text-sm [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink)]">
				<Markdown remarkPlugins={[remarkGfm]}>{question.content}</Markdown>
			</div>

			{/* Context */}
			{context && (
				<div className="rounded-lg border border-amber-200 bg-amber-50/80 px-3 py-2 dark:border-amber-800/40 dark:bg-amber-900/15">
					<p className="text-[11px] font-semibold text-amber-700 dark:text-amber-400">
						上下文
					</p>
					<div className="prose prose-sm max-w-none mt-1 [&_p]:mb-0 [&_p]:text-xs [&_p]:leading-relaxed [&_p]:text-amber-600 dark:[&_p]:text-amber-300">
						<Markdown remarkPlugins={[remarkGfm]}>{context}</Markdown>
					</div>
				</div>
			)}

			{/* Content to review */}
			<div className="max-h-[420px] overflow-auto rounded-lg border border-[var(--line)] bg-[var(--foam)] p-4">
				<div className="prose prose-sm max-w-none [&_p]:mb-1 [&_p]:text-sm [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink)]">
					<Markdown remarkPlugins={[remarkGfm]}>{content}</Markdown>
				</div>
			</div>

			{/* Decision buttons */}
			<div className="flex flex-wrap gap-2">
				<Button
					variant={decision === "approve" ? "default" : "outline"}
					size="sm"
					onClick={() => setDecision("approve")}
					className={
						decision === "approve"
							? "bg-emerald-600 hover:bg-emerald-700"
							: "border-emerald-300 text-emerald-700 hover:bg-emerald-50"
					}
				>
					<CheckCircle className="mr-1 size-3.5" aria-hidden />
					批准
				</Button>
				<Button
					variant={isRevising ? "default" : "outline"}
					size="sm"
					onClick={() => setDecision(isRevising ? null : "revise")}
					className={
						isRevising
							? "bg-amber-600 hover:bg-amber-700"
							: "border-amber-300 text-amber-700 hover:bg-amber-50"
					}
				>
					<Pencil className="mr-1 size-3.5" aria-hidden />
					修改
				</Button>
			</div>

			{/* Feedback for revision */}
			{isRevising && (
				<Textarea
					placeholder="请描述需要修改的内容..."
					value={feedback}
					onChange={(e) => setFeedback(e.target.value)}
					className="min-h-[80px] resize-y rounded-lg border-[var(--line)] bg-[var(--foam)] text-sm"
				/>
			)}

			{/* Actions */}
			<div className="flex items-center justify-between pt-1">
				<Button
					variant="ghost"
					size="sm"
					onClick={() => onRequestSupplement([question.id])}
					className="text-xs text-[var(--sea-ink-soft)]"
				>
					请求补充信息
				</Button>
				<div className="flex gap-2">
					<Button variant="outline" size="sm" onClick={onSkip}>
						<SkipForward className="mr-1 size-3.5" aria-hidden />
						跳过
					</Button>
					<Button
						size="sm"
						onClick={handleSubmit}
						disabled={!decision || (isRevising && !feedback.trim())}
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
