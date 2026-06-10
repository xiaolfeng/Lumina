import { GripVertical, ArrowDown, ArrowUp, Send, SkipForward } from "lucide-react";
import { useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";

import type { QuestionComponentProps } from "./question-select";

interface RankItem {
	id: string;
	label: string;
	description?: string;
}

export function QuestionRank({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const initialItems: RankItem[] =
		(question.options ?? []).map((o) => ({
			id: o.id,
			label: o.label,
			description: o.description,
		})) ?? [];

	const [items, setItems] = useState<RankItem[]>(initialItems);
	const [dragIndex, setDragIndex] = useState<number | null>(null);
	const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);

	const handleDragStart = (index: number) => {
		setDragIndex(index);
	};

	const handleDragOver = (e: React.DragEvent, index: number) => {
		e.preventDefault();
		setDragOverIndex(index);
	};

	const handleDragEnd = () => {
		setDragIndex(null);
		setDragOverIndex(null);
	};

	const handleDrop = (dropIndex: number) => {
		if (dragIndex === null || dragIndex === dropIndex) return;

		setItems((prev) => {
			const next = [...prev];
			const [removed] = next.splice(dragIndex!, 1);
			next.splice(dropIndex, 0, removed!);
			return next;
		});
	};

	const moveItem = (index: number, direction: "up" | "down") => {
		const newIndex = direction === "up" ? index - 1 : index + 1;
		if (newIndex < 0 || newIndex >= items.length) return;

		setItems((prev) => {
			const next = [...prev];
			[next[index], next[newIndex]] = [next[newIndex], next[index]];
			return next;
		});
	};

	const handleSubmit = () => {
		onSubmit({ ranking: items.map((item) => item.id) });
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

			{/* Rank list */}
			<ul className="space-y-2">
				{items.map((item, index) => (
					<li
						key={item.id}
						draggable
						onDragStart={() => handleDragStart(index)}
						onDragOver={(e) => handleDragOver(e, index)}
						onDragEnd={handleDragEnd}
						onDrop={() => handleDrop(index)}
						className={`group flex items-center gap-2.5 rounded-lg border-2 bg-[var(--foam)] px-3 py-2.5 transition-all duration-150 ${
							dragOverIndex === index
								? "border-dashed border-[var(--lagoon)] bg-[var(--lagoon)]/5"
								: "border-[var(--line)]"
						} ${dragIndex === index ? "opacity-50 scale-[0.98]" : ""} cursor-grab active:cursor-grabbing`}
					>
						{/* Drag handle */}
						<GripVertical className="size-4 shrink-0 text-[var(--sea-ink-soft)] group-hover:text-[var(--lagoon-deep)]" />

						{/* Rank number */}
						<span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-[var(--lagoon)]/10 text-[11px] font-bold text-[var(--lagoon-deep)]">
							{index + 1}
						</span>

						{/* Content */}
						<div className="min-w-0 flex-1">
							<p className="font-medium text-sm">{item.label}</p>
							{item.description && (
								<p className="text-xs leading-relaxed text-[var(--sea-ink-soft)]">
									{item.description}
								</p>
							)}
						</div>

						{/* Arrow buttons */}
						<div className="flex shrink-0 gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
							<button
								type="button"
								onClick={() => moveItem(index, "up")}
								disabled={index === 0}
								className="rounded p-1 text-[var(--sea-ink-soft)] hover:bg-[var(--line)] disabled:opacity-20"
								aria-label="上移"
							>
								<ArrowUp className="size-3.5" />
							</button>
							<button
								type="button"
								onClick={() => moveItem(index, "down")}
								disabled={index === items.length - 1}
								className="rounded p-1 text-[var(--sea-ink-soft)] hover:bg-[var(--line)] disabled:opacity-20"
								aria-label="下移"
							>
								<ArrowDown className="size-3.5" />
							</button>
						</div>
					</li>
				))}
			</ul>

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
