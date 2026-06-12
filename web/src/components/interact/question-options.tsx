import { Loader2, Send, SkipForward, Star } from "lucide-react";
import { useRef, useState } from "react";
import { motion } from "motion/react";

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

function OptionDetailLabel({ optId: _optId, onClick }: { optId: string; onClick: () => void }) {
	const [isHovered, setIsHovered] = useState(false)
	const hoverTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

	const handleMouseEnter = () => {
		hoverTimerRef.current = setTimeout(() => setIsHovered(true), 500)
	}
	const handleMouseLeave = () => {
		if (hoverTimerRef.current) clearTimeout(hoverTimerRef.current)
		setIsHovered(false)
	}

	return (
		<motion.div
			className="absolute right-1 top-1 z-10 flex items-center gap-0.5 rounded-full bg-blue-100 px-1.5 py-0.5 text-[10px] font-medium text-blue-600 cursor-pointer overflow-hidden whitespace-nowrap"
			animate={{ width: isHovered ? "auto" : 22 }}
			initial={{ width: 22 }}
			transition={{ duration: 0.3, ease: "easeOut" }}
			onMouseEnter={handleMouseEnter}
			onMouseLeave={handleMouseLeave}
			onClick={(e) => { e.stopPropagation(); e.preventDefault(); onClick() }}
			style={{ maxWidth: 200 }}
		>
			<span className="shrink-0">→</span>
			{isHovered && (
				<motion.span
					initial={{ opacity: 0 }}
					animate={{ opacity: 1 }}
					transition={{ delay: 0.1 }}
				>
					点击查看选项详情
				</motion.span>
			)}
		</motion.div>
	)
}

export function QuestionOptions({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
	onViewOptionDetail,
}: QuestionComponentProps) {
	const [selected, setSelected] = useState<string>("");
	const [feedback, setFeedback] = useState("");

	const options = (question.options ?? []) as OptionWithProsCons[];
	const hasSupplements =
		(question.supplements?.length ?? 0) > 0;

	const hasOptionSupplement = (optId: string): boolean => {
		return question.supplements?.some(
			(s) => s.target_type === "option" && s.target_id === optId,
		) ?? false;
	};

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
			{isSupplementLoading && (
				<div className="flex items-center gap-2 rounded-lg bg-blue-50 p-3 text-sm text-blue-600">
					<Loader2 className="size-4 animate-spin" aria-hidden />
					正在加载补充内容...
				</div>
			)}
			<RadioGroup
				value={selected}
				onValueChange={setSelected}
				className="space-y-3"
				disabled={isSupplementLoading}
			>
				{options.map((opt) => (
					<Label
						key={opt.id}
						htmlFor={`options-${question.id}-${opt.id}`}
						className={`relative flex cursor-pointer flex-col gap-2 rounded-xl border px-4 py-3 transition-all duration-150 ${
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
						{hasOptionSupplement(opt.id) && (
							<OptionDetailLabel
								optId={opt.id}
								onClick={() => onViewOptionDetail?.(opt.id)}
							/>
						)}
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
				{!hasSupplements && (
					<Button
						variant="ghost"
						size="sm"
						onClick={() => onRequestSupplement([question.id])}
						disabled={isSupplementLoading}
						className="text-xs text-[var(--sea-ink-soft)]"
					>
						请求补充信息
					</Button>
				)}
				<div className="flex gap-2">
					<Button
						variant="outline"
						size="sm"
						onClick={onSkip}
						disabled={isSupplementLoading}
					>
						<SkipForward className="mr-1 size-3.5" aria-hidden />
						跳过
					</Button>
					<Button
						size="sm"
						onClick={handleSubmit}
						disabled={!selected || isSupplementLoading}
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
