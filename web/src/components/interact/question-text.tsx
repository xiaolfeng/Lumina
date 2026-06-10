import { Send, SkipForward } from "lucide-react";
import { useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import { Textarea } from "#/components/ui/textarea";

import type { QuestionComponentProps } from "./question-select";

export function QuestionText({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const [value, setValue] = useState("");

	const multiline = question.config?.multiline === true;
	const maxLength = question.config?.maxLength as number | undefined;
	const placeholder =
		(question.config?.placeholder as string | undefined) ||
		(multiline ? "输入你的回答..." : "输入你的回答...");

	const charCount = value.length;
	const isOverLimit = maxLength !== undefined && charCount > maxLength;

	const handleSubmit = () => {
		if (!value.trim()) return;
		onSubmit({ text: value });
	};

	const InputEl = multiline ? Textarea : Input;

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

			{/* Text input */}
			<InputEl
				placeholder={placeholder}
				value={value}
				onChange={(e) => setValue(e.target.value)}
				maxLength={maxLength}
				className={
					multiline
						? "min-h-[100px] resize-y rounded-lg border-[var(--line)] bg-[var(--foam)]"
						: "rounded-lg border-[var(--line)] bg-[var(--foam)]"
				}
			/>

			{/* Character count */}
			{maxLength && (
				<p
					className={`text-right text-xs ${isOverLimit ? "text-red-500 font-medium" : "text-[var(--sea-ink-soft)]"}`}
				>
					{charCount} / {maxLength}
				</p>
			)}

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
						disabled={!value.trim() || isOverLimit}
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
