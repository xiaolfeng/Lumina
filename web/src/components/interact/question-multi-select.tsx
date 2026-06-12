import { Loader2, Pencil, Send, SkipForward } from "lucide-react";
import { useRef, useState } from "react";
import { motion } from "motion/react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";
import { Checkbox } from "#/components/ui/checkbox";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";

import type { QuestionComponentProps } from "./question-select";

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

export function QuestionMultiSelect({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
	onViewOptionDetail,
}: QuestionComponentProps) {
	const [selected, setSelected] = useState<Set<string>>(new Set());
	const [otherChecked, setOtherChecked] = useState(false);
	const [otherText, setOtherText] = useState("");

	const options = question.options ?? [];
	const minSelect = question.config?.min ?? 0;
	const maxSelect = question.config?.max ?? options.length;
	const hasSupplements =
		(question.supplements?.length ?? 0) > 0;

	const hasOptionSupplement = (optId: string): boolean => {
		return question.supplements?.some(
			(s) => s.target_type === "option" && s.target_id === optId,
		) ?? false;
	};

	const toggleOption = (id: string) => {
	 setSelected((prev) => {
			const next = new Set(prev);
			if (next.has(id)) {
				next.delete(id);
			} else if (!maxSelect || next.size < maxSelect) {
				next.add(id);
			}
			return next;
		});
	};

	const handleSubmit = () => {
		const result: any = { selected: Array.from(selected) };
		if (otherChecked && otherText.trim()) {
			result.other = [otherText.trim()];
		}
		onSubmit(result);
	};

	const canSubmit =
		selected.size >= minSelect &&
		(!otherChecked || !!otherText.trim());

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

			{/* Selection counter */}
			{(minSelect > 0 || maxSelect < options.length) && (
				<p className="text-xs text-[var(--sea-ink-soft)]">
					已选 {selected.size} / {maxSelect > 0 ? maxSelect : "无限制"}
					{minSelect > 0 && `（至少 ${minSelect} 项）`}
				</p>
			)}

			{/* Options */}
			{isSupplementLoading && (
				<div className="flex items-center gap-2 rounded-lg bg-blue-50 p-3 text-sm text-blue-600">
					<Loader2 className="size-4 animate-spin" aria-hidden />
					正在加载补充内容...
				</div>
			)}
			<div
				className={`space-y-2 ${isSupplementLoading ? "pointer-events-none opacity-50" : ""}`}
			>
				{options.map((opt) => (
					<Label
						key={opt.id}
						htmlFor={`multi-${question.id}-${opt.id}`}
						className="relative flex cursor-pointer items-start gap-3 rounded-lg border border-[var(--line)] bg-[var(--foam)] px-3 py-2.5 transition-colors duration-150 has-[data-state=checked]:border-[var(--lagoon)]/30 has-[data-state=checked]:bg-[var(--lagoon)]/10 hover:border-[var(--lagoon)]/30"
					>
						<Checkbox
							id={`multi-${question.id}-${opt.id}`}
							checked={selected.has(opt.id)}
							onCheckedChange={() => toggleOption(opt.id)}
							className="mt-0.5"
						/>
						<div className="min-w-0 flex-1">
							<p className="font-medium text-sm leading-snug">
								{opt.label}
							</p>
							{opt.description && (
								<p className="mt-0.5 text-xs leading-relaxed text-[var(--sea-ink-soft)]">
									{opt.description}
								</p>
							)}
						</div>
						{hasOptionSupplement(opt.id) && (
							<OptionDetailLabel
								optId={opt.id}
								onClick={() => onViewOptionDetail?.(opt.id)}
							/>
						)}
					</Label>
				))}

				{/* Allow Other */}
				{question.allowOther && (
					<>
						<Label
							htmlFor={`multi-${question.id}-other`}
							className="flex cursor-pointer items-start gap-3 rounded-lg border border-[var(--line)] bg-[var(--foam)] px-3 py-2.5 transition-colors duration-150 has-[data-state=checked]:border-[var(--lagoon)]/30 has-[data-state=checked]:bg-[var(--lagoon)]/10 hover:border-[var(--lagoon)]/30"
						>
							<Checkbox
								id={`multi-${question.id}-other`}
								checked={otherChecked}
								onCheckedChange={(v) =>
									setOtherChecked(!!v)
								}
								className="mt-0.5"
							/>
							<Pencil className="size-3.5 shrink-0 text-[var(--lagoon-deep)]" />
							<span className="font-medium text-sm">其他</span>
						</Label>
						{otherChecked && (
							<Input
								placeholder="输入自定义选项..."
								value={otherText}
								onChange={(e) =>
									setOtherText(e.target.value)
								}
								className="rounded-lg border-[var(--line)] bg-[var(--foam)]"
							/>
						)}
					</>
				)}
			</div>

			{/* Actions */}
			<div className="flex items-center justify-between pt-1">
				{!hasSupplements && (
					<Button
						variant="ghost"
						size="sm"
						onClick={() =>
							onRequestSupplement([question.id])
						}
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
						disabled={!canSubmit || isSupplementLoading}
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
