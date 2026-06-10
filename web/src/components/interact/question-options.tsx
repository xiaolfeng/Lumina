import { Send, SkipForward, Star } from "lucide-react";
import { useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";
import { Label } from "#/components/ui/label";
import { Textarea } from "#/components/ui/textarea";
import {
	RadioGroup,
	RadioGroupItem,
} from "#/components/ui/radio-group";

import type { OptionItem } from "./types";
import type { QuestionComponentProps } from "./question-select";

interface OptionWithProsCons extends OptionItem {
	pros?: string[];
	cons?: string[];
	recommended?: boolean;
}

export function QuestionOptions({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const [selected, setSelected] = useState<string>("");
	const [feedback, setFeedback] = useState("");

	const options = (question.options ?? []) as OptionWithProsCons[];

	const handleSubmit = () => {
		if (!selected) return;
		onSubmit({
			selected,
			...(feedback.trim() ? { feedback: feedback.trim() } : {}),
		});
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

			{/* Options with pros/cons */}
			<RadioGroup
				value={selected}
				onValueChange={setSelected}
				className="space-y-3"
			>
				{options.map((opt) => (
					<Label
						key={opt.id}
						htmlFor={`options-${question.id}-${opt.id}`}
						className={`flex cursor-pointer flex-col gap-2 rounded-xl border px-4 py-3 transition-all duration-150 ${
							selected === opt.id
								? "border-[var(--lagoon)]/40 bg-[var(--lagoon)]/8 shadow-sm"
								: "border-[var(--line)] bg-[var(--foam)] hover:border-[var(--lagoon)]/25"
						}`}
					>
						<div className="flex items-start gap-3">
							<RadioGroupItem
								value={opt.id}
								id={`options-${question.id}-${opt.id}`}
								className="mt-0.5"
							/>
							<div className="min-w-0 flex-1">
								<div className="flex items-center gap-2">
									<span className="font-medium text-sm">
										{opt.label}
									</span>
									{opt.recommended && (
										<span className="inline-flex items-center gap-0.5 rounded-full bg-amber-100 px-2 py-0.5 text-[10px] font-semibold text-amber-700">
											<Star className="size-2.5" aria-hidden />
											推荐
										</span>
									)}
								</div>
								{opt.description && (
									<p className="mt-0.5 text-xs leading-relaxed text-[var(--sea-ink-soft)]">
										{opt.description}
									</p>
								)}
							</div>
						</div>

						{/* Pros & Cons */}
						{(opt.pros?.length ?? 0) > 0 ||
						(opt.cons?.length ?? 0) > 0 ? (
							<div className="ml-7 grid grid-cols-1 gap-2 sm:grid-cols-2">
								{opt.pros && opt.pros.length > 0 && (
									<div className="space-y-1">
										<p className="text-[10px] font-semibold uppercase tracking-wide text-emerald-600">
											优点
										</p>
										<ul className="space-y-0.5">
											{opt.pros.map((pro, idx) => (
												<li
													key={idx}
													className="flex items-start gap-1.5 text-xs text-emerald-700"
												>
													<span className="mt-1.5 size-1 shrink-0 rounded-full bg-emerald-500" />
													{pro}
												</li>
											))}
										</ul>
									</div>
								)}
								{opt.cons && opt.cons.length > 0 && (
									<div className="space-y-1">
										<p className="text-[10px] font-semibold uppercase tracking-wide text-red-500">
											缺点
										</p>
										<ul className="space-y-0.5">
											{opt.cons.map((con, idx) => (
												<li
													key={idx}
													className="flex items-start gap-1.5 text-xs text-red-600"
												>
													<span className="mt-1.5 size-1 shrink-0 rounded-full bg-red-400" />
													{con}
												</li>
											))}
										</ul>
									</div>
								)}
							</div>
						) : null}
					</Label>
				))}
			</RadioGroup>

			{/* Feedback */}
			<Textarea
				placeholder="可选：说明你的选择理由..."
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
						disabled={!selected}
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
