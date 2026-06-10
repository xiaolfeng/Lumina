import { CircleDot } from "lucide-react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import type { Question } from "./types";

import { QuestionBoolean } from "./question-boolean";
import { QuestionCode } from "./question-code";
import { QuestionDiff } from "./question-diff";
import { QuestionFile } from "./question-file";
import { QuestionImage } from "./question-image";
import type { QuestionComponentProps } from "./question-select";
import { QuestionMultiSelect } from "./question-multi-select";
import { QuestionOptions } from "./question-options";
import { QuestionPlan } from "./question-plan";
import { QuestionRate } from "./question-rate";
import { QuestionRank } from "./question-rank";
import { QuestionReview } from "./question-review";
import { QuestionSelect } from "./question-select";
import { QuestionSlider } from "./question-slider";
import { QuestionText } from "./question-text";

interface QuestionCardProps {
	question: Question | undefined;
	onSubmit: (answer: any) => void;
	onSkip: () => void;
	onRequestSupplement: (targets: string[]) => void;
}

export function QuestionCard({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionCardProps) {
	if (!question) {
		return (
			<div className="rounded-xl border border-[var(--line)] bg-[var(--surface)] shadow-[0_2px_12px_rgba(42,36,32,0.06)] backdrop-blur-sm">
				<div className="px-4 py-6 text-center">
					<p className="text-sm text-[var(--sea-ink-soft)]">
						所有问题已回答完毕
					</p>
				</div>
			</div>
		);
	}

	const props: QuestionComponentProps = {
		question,
		onSubmit,
		onSkip,
		onRequestSupplement,
	};

	return (
		<div className="rounded-xl border border-[var(--line)] bg-[var(--surface)] shadow-[0_2px_12px_rgba(42,36,32,0.06)] backdrop-blur-sm">
			{/* Header */}
			<div className="flex items-center justify-between border-b border-[var(--line)]/50 px-4 py-2.5">
				<span className="text-[11px] font-semibold uppercase tracking-[0.15em] text-[var(--lagoon-deep)]">
					当前问题
				</span>
				<span className="inline-flex items-center gap-1 rounded-full bg-[var(--lagoon)]/10 px-2 py-0.5 text-[10px] font-semibold text-[var(--lagoon-deep)]">
					<CircleDot className="size-2.5" aria-hidden />
					{question.groupLabel}
				</span>
			</div>

			{/* Body — dispatch by type */}
			<div className="p-4">{renderByType(question.type, props)}</div>
		</div>
	);
}

function renderByType(
	type: Question["type"],
	props: QuestionComponentProps,
) {
	switch (type) {
		case "select":
			return <QuestionSelect {...props} />;
		case "multi-select":
			return <QuestionMultiSelect {...props} />;
		case "text":
			return <QuestionText {...props} />;
		case "boolean":
			return <QuestionBoolean {...props} />;
		case "code":
			return <QuestionCode {...props} />;
		case "image":
			return <QuestionImage {...props} />;
		case "file":
			return <QuestionFile {...props} />;
		case "diff":
			return <QuestionDiff {...props} />;
		case "plan":
			return <QuestionPlan {...props} />;
		case "options":
			return <QuestionOptions {...props} />;
		case "review":
			return <QuestionReview {...props} />;
		case "slider":
			return <QuestionSlider {...props} />;
		case "rank":
			return <QuestionRank {...props} />;
		case "rate":
			return <QuestionRate {...props} />;
		default:
			return (
				<div className="space-y-3">
					<div className="prose prose-sm max-w-none [&_p]:mb-1 [&_p]:text-sm [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink)]">
						<Markdown remarkPlugins={[remarkGfm]}>
							{props.question.content}
						</Markdown>
					</div>
					<div className="rounded-lg border border-dashed border-[var(--line)] bg-[var(--foam)]/50 p-6 text-center">
						<p className="text-xs text-[var(--sea-ink-soft)]">
							暂不支持的问题类型：
							<code className="rounded bg-[var(--line)] px-1.5 py-0.5 font-mono">
								{type}
							</code>
						</p>
						<p className="mt-1 text-[10px] text-[var(--sea-ink-soft)]/60">
							该类型将在后续版本中支持
						</p>
					</div>
				</div>
			);
	}
}
