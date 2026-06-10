import { Pencil, Send, SkipForward } from "lucide-react";
import { useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";
import { Checkbox } from "#/components/ui/checkbox";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";

import type { QuestionComponentProps } from "./question-select";

export function QuestionMultiSelect({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const [selected, setSelected] = useState<Set<string>>(new Set());
	const [otherChecked, setOtherChecked] = useState(false);
	const [otherText, setOtherText] = useState("");

	const options = question.options ?? [];
	const minSelect = question.config?.min ?? 0;
	const maxSelect = question.config?.max ?? options.length;

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
			<div className="space-y-2">
				{options.map((opt) => (
					<Label
						key={opt.id}
						htmlFor={`multi-${question.id}-${opt.id}`}
						className="flex cursor-pointer items-start gap-3 rounded-lg border border-[var(--line)] bg-[var(--foam)] px-3 py-2.5 transition-colors duration-150 has-[data-state=checked]:border-[var(--lagoon)]/30 has-[data-state=checked]:bg-[var(--lagoon)]/10 hover:border-[var(--lagoon)]/30"
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
						disabled={!canSubmit}
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
