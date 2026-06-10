import { CheckCircle, Pencil, Send, SkipForward, XCircle, ChevronDown, ChevronRight } from "lucide-react";
import	{ useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";
import { Textarea } from "#/components/ui/textarea";

import type { QuestionComponentProps } from "./question-select";

interface PlanSection {
	id: string;
	title: string;
	content: string;
}

type PlanDecision = "approve" | "reject" | "revise";

export function QuestionPlan({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const [decision, setDecision] = useState<PlanDecision | null>(null);
	const [annotations, setAnnotations] = useState<Record<string, string>>({});
	const [feedback, setFeedback] = useState("");
	const [expandedSections, setExpandedSections] = useState<Set<string>>(
		new Set(),
	);

	const sections = (question.config?.sections as PlanSection[]) ?? [];
	const markdown = (question.config?.markdown as string) ?? "";
	const isRevising = decision === "revise";

	// Expand all by default on mount
	if (sections.length > 0 && expandedSections.size === 0) {
		setTimeout(() => {
			setExpandedSections(new Set(sections.map((s) => s.id)));
		}, 0);
	}

	const toggleSection = (id: string) => {
		setExpandedSections((prev) => {
			const next = new Set(prev);
			if (next.has(id)) next.delete(id);
			else next.add(id);
			return next;
		});
	};

	const updateAnnotation = (id: string, value: string) => {
		setAnnotations((prev) => ({ ...prev, [id]: value }));
	};

	const handleSubmit = () => {
		if (!decision) return;
		const annotationsList = Object.entries(annotations)
			.filter(([, v]) => v.trim())
			.map(([sectionId, content]) => ({ sectionId, content }));

		onSubmit({
			decision,
			...(annotationsList.length > 0 ? { annotations: annotationsList } : {}),
			...(feedback.trim() ? { feedback: feedback.trim() } : {}),
		});
	};

	return (
		<div className="space-y-4">
			{/* Question content */}
			<div className="prose prose-sm max-w-none [&_p]:mb-1 [&_p]:text-sm [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink)]">
				<Markdown remarkPlugins={[remarkGfm]}>{question.content}</Markdown>
			</div>

			{/* Plan sections or markdown */}
			{sections.length > 0 ? (
				<div className="space-y-2">
					{sections.map((section) => {
						const isOpen = expandedSections.has(section.id);
						return (
							<div
								key={section.id}
								className="overflow-hidden rounded-lg border border-[var(--line)] bg-[var(--foam)]"
							>
								<button
									type="button"
									onClick={() => toggleSection(section.id)}
									className="flex w-full items-center gap-2 px-3 py-2.5 text-left transition-colors hover:bg-[var(--line)]/30"
								>
									{isOpen ? (
										<ChevronDown className="size-4 shrink-0 text-[var(--sea-ink-soft)]" />
									) : (
										<ChevronRight className="size-4 shrink-0 text-[var(--sea-ink-soft)]" />
									)}
									<span className="font-medium text-sm">
										{section.title}
									</span>
								</button>
								{isOpen && (
									<div className="border-t border-[var(--line)]/50 px-3 pb-3">
										<div className="prose prose-sm max-w-none mt-2 [&_p]:mb-1 [&_p]:text-xs [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink)]">
											<Markdown remarkPlugins={[remarkGfm]}>
												{section.content}
											</Markdown>
										</div>
										{isRevising && (
											<Textarea
												placeholder={`对「${section.title}」的修订意见...`}
												value={annotations[section.id] ?? ""}
												onChange={(e) =>
													updateAnnotation(section.id, e.target.value)
												}
												className="mt-2 min-h-[50px] resize-y rounded border-[var(--line)] bg-white text-xs"
											/>
										)}
									</div>
								)}
							</div>
						);
					})}
				</div>
			) : markdown ? (
				<div className="max-h-[400px] overflow-auto rounded-lg border border-[var(--line)] bg-[var(--foam)] p-4">
					<div className="prose prose-sm max-w-none [&_p]:mb-1 [&_p]:text-sm [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink)]">
						<Markdown remarkPlugins={[remarkGfm]}>{markdown}</Markdown>
					</div>
				</div>
			) : null}

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
					variant={isRevising ? "default" : "outline"}
					size="sm"
					onClick={() => setDecision(isRevising ? null : "revise")}
					className={
						isRevising
							? "bg-amber-600 hover:bg-amber-700"
							: "border-amber-300 text-amber-700 hover:bg-amber-50"
					}
				>
					<Pencil className="mr-1 size-3.5" aria-hidden />
					修订
				</Button>
			</div>

			{/* General feedback */}
			<Textarea
				placeholder="可选：整体反馈意见..."
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
