import { Send, SkipForward } from "lucide-react";
import { useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";

import type { QuestionComponentProps } from "./question-select";

export function QuestionSlider({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const min = (question.config?.min as number) ?? 0;
	const max = (question.config?.max as number) ?? 100;
	const step = (question.config?.step as number) ?? 1;
	const defaultValue = (question.config?.defaultValue as number) ?? min;
	const minLabel = (question.config?.minLabel as string) ?? String(min);
	const maxLabel = (question.config?.maxLabel as string) ?? String(max);

	const [value, setValue] = useState(defaultValue);

	const percentage = ((value - min) / (max - min)) * 100;

	const handleSubmit = () => {
		onSubmit({ value });
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

			{/* Slider container */}
			<div className="space-y-3 rounded-xl border border-[var(--line)] bg-[var(--foam)] px-5 py-6">
				{/* Current value display */}
				<div className="text-center">
					<span className="inline-flex items-center justify-center rounded-full bg-[var(--lagoon)]/10 px-4 py-1.5 font-mono text-2xl font-bold text-[var(--lagoon-deep)]">
						{value}
					</span>
				</div>

				{/* Range input */}
				<input
					type="range"
					min={min}
					max={max}
					step={step}
					value={value}
					onChange={(e) => setValue(Number(e.target.value))}
					className="slider w-full accent-[var(--lagoon)]"
					style={{
						background: `linear-gradient(to right, var(--lagoon) ${percentage}%, var(--line) ${percentage}%)`,
						borderRadius: "9999px",
						height: "6px",
						appearance: "none",
						cursor: "pointer",
					}}
				/>

				{/* Labels */}
				<div className="flex items-center justify-between text-xs text-[var(--sea-ink-soft)]">
					<span>{minLabel}</span>
					<span>{maxLabel}</span>
				</div>
			</div>

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
