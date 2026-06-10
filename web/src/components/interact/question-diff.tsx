import { CheckCircle, Pencil, Send, SkipForward, XCircle } from "lucide-react";
import { useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";
import { Textarea } from "#/components/ui/textarea";

import type { QuestionComponentProps } from "./question-select";

type Decision = "approve" | "reject" | "edit";

export function QuestionDiff({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const [decision, setDecision] = useState<Decision | null>(null);
	const [editedCode, setEditedCode] = useState("");
	const [feedback, setFeedback] = useState("");

	const before = (question.config?.before as string) ?? "";
	const after = (question.config?.after as string) ?? "";
	const filePath = (question.config?.filePath as string) ?? "";
	const language = (question.config?.language as string) ?? "";

	const isEditing = decision === "edit";

	const handleSubmit = () => {
		if (!decision) return;
		onSubmit({
			decision,
			...(isEditing ? { edited: editedCode || after } : {}),
			...(feedback.trim() ? { feedback: feedback.trim() } : {}),
		});
	};

	return (
		<div className="space-y-4">
			{/* Question content */}
			<div className="prose prose-sm max-w-none [&_p]:mb-1 [&_p]:text-sm [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink)]">
				<Markdown remarkPlugins={[remarkGfm]}>{question.content}</Markdown>
			</div>

			{/* File path */}
			{filePath && (
				<p className="rounded-md bg-[var(--foam)] px-3 py-1.5 font-mono text-xs text-[var(--sea-ink-soft)]">
					📄 {filePath}
				</p>
			)}

			{/* Code blocks */}
			<div className="grid grid-cols-1 gap-3 lg:grid-cols-2">
				{/* Before */}
				<div className="space-y-1.5">
					<p className="text-xs font-semibold uppercase tracking-wide text-red-400">
						修改前 (Before)
					</p>
					<pre className="max-h-[360px] overflow-auto rounded-lg bg-gray-900 p-3 text-xs leading-relaxed text-gray-200">
						<code>{before || "(无内容)"}</code>
					</pre>
				</div>

				{/* After */}
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
							className="min-h-[200px] resize-y rounded-lg border-[var(--line)] bg-[var(--foam)] font-mono text-xs tabular-nums"
							placeholder="编辑代码..."
						/>
					) : (
						<pre className="max-h-[360px] overflow-auto rounded-lg bg-gray-900 p-3 text-xs leading-relaxed text-gray-200">
							<code>{after || "(无内容)"}</code>
						</pre>
					)}
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
					variant={decision === "reject" ? "default" : "outline"}
					size="sm"
					onClick={() => setDecision("reject")}
					className={
						decision === "reject"
							? "bg-red-600 hover:bg-red-700"
							: "border-red-300 text-red-700 hover:bg-red-50"
					}
				>
					<XCircle className="mr-1 size-3.5" aria-hidden />
					拒绝
				</Button>
				<Button
					variant={isEditing ? "default" : "outline"}
					size="sm"
					onClick={() => setDecision(isEditing ? null : "edit")}
					className={
						isEditing
							? "bg-amber-600 hover:bg-amber-700"
							: "border-amber-300 text-amber-700 hover:bg-amber-50"
					}
				>
					<Pencil className="mr-1 size-3.5" aria-hidden />
					编辑
				</Button>
			</div>

			{/* Feedback textarea */}
			<Textarea
				placeholder="可选：添加反馈意见..."
				value={feedback}
				onChange={(e) => setFeedback(e.target.value)}
				className="min-h-[60px] resize-y rounded-lg border-[var(--line)] bg-[var(--foam)] text-sm"
			/>

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
						disabled={!decision}
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
