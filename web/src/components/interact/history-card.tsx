import { Check } from "lucide-react";

import type { Question } from "./types";

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
											→ {q.answer}
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
