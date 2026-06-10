import { Send, SkipForward, Star } from "lucide-react";
import { useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";

import type { QuestionComponentProps } from "./question-select";

export function QuestionRate({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const options = question.options ?? [];
	const min = (question.config?.min as number) ?? 1;
	const max = (question.config?.max as number) ?? 5;
	const step = (question.config?.step as number) ?? 1;
	const useStars = (question.config?.useStars as boolean) ?? true;

	const [ratings, setRatings] = useState<Record<string, number>>(() => {
		const initial: Record<string, number> = {};
		options.forEach((opt) => {
			initial[opt.id] = min;
		});
		return initial;
	});

	const setRating = (optionId: string, value: number) => {
		setRatings((prev) => ({ ...prev, [optionId]: value }));
	};

	const handleSubmit = () => {
		onSubmit({ ratings: { ...ratings } });
	};

	// Generate star/number buttons
	const renderRatingControl = (optionId: string, current: number) => {
		if (useStars) {
			const count = Math.round((max - min) / step) + 1;
			return (
				<div className="flex gap-1">
					{Array.from({ length: count }, (_, i) => {
						const val = min + i * step;
						const filled = val <= current;
						return (
							<button
								key={val}
								type="button"
								onClick={() => setRating(optionId, val)}
								className="group p-0.5 transition-transform hover:scale-110"
								aria-label={`${val} 分`}
							>
								<Star
									className={`size-6 transition-colors ${
										filled
											? "fill-amber-400 text-amber-400"
											: "fill-transparent text-gray-300 group-hover:text-amber-200"
									}`}
								/>
							</button>
						);
					})}
				</div>
			);
		}

		// Number input mode
		return (
			<input
				type="number"
				min={min}
				max={max}
				step={step}
				value={current}
				onChange={(e) => setRating(optionId, Number(e.target.value))}
				className="w-24 rounded-lg border border-[var(--line)] bg-[var(--foam)] px-3 py-1.5 text-center font-mono text-sm tabular-nums"
			/>
		);
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

			{/* Rating list */}
			<div className="space-y-3">
				{options.map((opt) => (
					<div
						key={opt.id}
						className="flex flex-col gap-2 rounded-lg border border-[var(--line)] bg-[var(--foam)] px-4 py-3 sm:flex-row sm:items-center sm:justify-between"
					>
						<div className="min-w-0 flex-1">
							<p className="font-medium text-sm">{opt.label}</p>
							{opt.description && (
								<p className="text-xs text-[var(--sea-ink-soft)]">
									{opt.description}
								</p>
							)}
						</div>
						<div className="flex items-center gap-3 sm:shrink-0">
							{renderRatingControl(opt.id, ratings[opt.id] ?? min)}
							<span className="w-8 text-center font-mono text-sm font-semibold text-[var(--lagoon-deep)]">
								{ratings[opt.id] ?? min}
							</span>
						</div>
					</div>
				))}
			</div>

			{/* Scale hint */}
			<p className="text-center text-[11px] text-[var(--sea-ink-soft)]">
				评分范围：{min} — {max}
				{step !== 1 && `（步长 ${step}）`}
			</p>

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
