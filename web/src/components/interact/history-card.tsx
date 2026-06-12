import { Check } from "lucide-react";

import type { Question } from "./types";

/** 将各题型提交的 answer 格式化为可读字符串 */
function formatAnswer(answer: unknown): string {
	if (answer == null) return "—"
	if (typeof answer === "string") return answer
	if (typeof answer === "number" || typeof answer === "boolean") return String(answer)
	if (Array.isArray(answer)) return answer.join(", ")
	if (typeof answer === "object") {
		const obj = answer as Record<string, unknown>
		// 单选/多选: { selected: string | string[] }
		if ("selected" in obj && !("text" in obj)) {
			const sel = Array.isArray(obj.selected) ? obj.selected : [obj.selected]
			return sel.join(", ")
		}
		// 文本: { text: string }
		if ("text" in obj) return String(obj.text)
		// 布尔: { choice: "yes" | "no" }
		if ("choice" in obj) return String(obj.choice)
		// 滑块: { value: number }
		if ("value" in obj) return String(obj.value)
		// 排序: { ranking: string[] }
		if ("ranking" in obj && Array.isArray(obj.ranking)) return obj.ranking.join(" → ")
		// 评分: { ratings: Record<string, number> }
		if ("ratings" in obj && typeof obj.ratings === "object") {
			return Object.entries(obj.ratings as Record<string, unknown>)
				.map(([k, v]) => `${k}: ${v}`)
				.join(", ")
		}
		// 兜底
		return JSON.stringify(obj)
	}
	return String(answer)
}

interface HistoryCardProps {
	answeredQuestions: Question[];
	groupedHistory: Record<string, Question[]>;
}

export function HistoryCard({
	groupedHistory,
	answeredQuestions,
}: HistoryCardProps) {
	return (
		<div className="rounded-xl border border-[var(--line)] bg-[var(--surface)] shadow-[0_2px_12px_rgba(42,36,32,0.06)] backdrop-blur-sm">
			<div className="border-b border-[var(--line)]/50 px-4 py-2.5">
				<span className="text-[11px] font-semibold uppercase tracking-[0.15em] text-[var(--sea-ink-soft)]">
					历史问答
				</span>
			</div>

			<div className="space-y-0 divide-y divide-[var(--line)]/30">
				{Object.entries(groupedHistory).map(([group, questions]) => (
					<div key={group} className="px-4 py-3">
						<div className="mb-2 flex items-center gap-2">
							<span className="inline-flex items-center gap-1 rounded-full bg-[var(--lagoon)]/8 px-2 py-0.5 text-[10px] font-semibold text-[var(--lagoon-deep)]">
								{group}
							</span>
							<span className="text-[10px] text-[var(--sea-ink-soft)]">
								{questions.length} 个问答
							</span>
						</div>

						<div className="space-y-2">
							{questions.map((q) => (
								<div key={q.id} className="flex items-start gap-2.5">
									<Check
										className="mt-0.5 size-3.5 shrink-0 text-emerald-500"
										aria-hidden
									/>
									<div className="min-w-0 flex-1">
										<p className="text-xs leading-relaxed text-[var(--sea-ink-soft)]">
											{q.content}
										</p>
										<p className="mt-0.5 text-xs font-medium text-[var(--sea-ink)]">
											→ {formatAnswer(q.answer)}
										</p>
									</div>
								</div>
							))}
						</div>
					</div>
				))}

				{answeredQuestions.length === 0 && (
					<div className="px-4 py-6 text-center">
						<p className="text-xs text-[var(--sea-ink-soft)]/50">
							暂无历史记录
						</p>
					</div>
				)}
			</div>
		</div>
	);
}
