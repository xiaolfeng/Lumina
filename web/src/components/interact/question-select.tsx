import { Pencil, Send, SkipForward } from "lucide-react";
import { useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import {
	RadioGroup,
	RadioGroupItem,
} from "#/components/ui/radio-group";

import type { Question } from "./types";

export interface QuestionComponentProps {
	question: Question;
	onSubmit: (answer: any) => void;
	onSkip: () => void;
	onRequestSupplement: (targets: string[]) => void;
}

export function QuestionSelect({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const [selected, setSelected] = useState<string>("");
	const [otherText, setOtherText] = useState("");
	const [showOtherInput, setShowOtherInput] = useState(false);

	const isOther = selected === "__other__";
	const options = question.options ?? [];

	const handleSubmit = () => {
		if (!selected) return;
		if (isOther) {
			onSubmit({ selected: "__other__", other: otherText });
		} else {
			onSubmit({ selected });
		}
	};

	const handleRadioChange = (value: string) => {
		setSelected(value);
		setShowOtherInput(value === "__other__");
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

			{/* Options */}
			<RadioGroup
				value={selected}
				onValueChange={handleRadioChange}
				className="space-y-2"
			>
				{options.map((opt) => (
					<Label
						key={opt.id}
						htmlFor={`select-${question.id}-${opt.id}`}
						className="flex cursor-pointer items-start gap-3 rounded-lg border border-[var(--line)] bg-[var(--foam)] px-3 py-2.5 transition-colors duration-150 has-[checked]:border-[var(--lagoon)]/30 has-[checked]:bg-[var(--lagoon)]/10 hover:border-[var(--lagoon)]/30"
					>
						<RadioGroupItem
							value={opt.id}
							id={`select-${question.id}-${opt.id}`}
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
					<Label
						htmlFor={`select-${question.id}-other`}
						className="flex cursor-pointer items-start gap-3 rounded-lg border border-[var(--line)] bg-[var(--foam)] px-3 py-2.5 transition-colors duration-150 has-[checked]:border-[var(--lagoon)]/30 has-[checked]:bg-[var(--lagoon)]/10 hover:border-[var(--lagoon)]/30"
					>
						<RadioGroupItem
							value="__other__"
							id={`select-${question.id}-other`}
							className="mt-0.5"
						/>
						<Pencil className="size-3.5 shrink-0 text-[var(--lagoon-deep)]" />
						<span className="font-medium text-sm">其他</span>
					</Label>
				)}
			</RadioGroup>

			{/* Other input */}
			{isOther && showOtherInput && (
				<Input
					placeholder="输入自定义选项..."
					value={otherText}
					onChange={(e) => setOtherText(e.target.value)}
					className="rounded-lg border-[var(--line)] bg-[var(--foam)]"
				/>
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
						disabled={!selected || (isOther && !otherText.trim())}
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
