import { createFileRoute } from "@tanstack/react-router";
import { useState, useEffect } from "react";
import type { DebugConfig } from "#/components/interact/debug-dialog";

import { DebugDialog } from "#/components/interact/debug-dialog";
import { DetailPanel } from "#/components/interact/detail-panel";
import { HistoryCard } from "#/components/interact/history-card";
import { QuestionCard } from "#/components/interact/question-card";
import { SessionSidebar } from "#/components/interact/session-sidebar";
import { useQaSession } from "#/hooks/useQaSession";
import type { Session } from "#/components/interact/types";
import { ScrollArea } from "#/components/ui/scroll-area";

export const Route = createFileRoute("/interact/")({
	component: InteractPage,
});

function InteractPage() {
	const [selectedSessionId, setSelectedSessionId] = useState<string>("");
	const [sessions, setSessions] = useState<Session[]>([]);
	const [debug, setDebug] = useState<DebugConfig>({
		hideMarkdown: false,
		forceSelect: false,
		forceMulti: false,
		forceBoolean: false,
		forceText: false,
	});

	// WebSocket 实时数据
	const {
		questions,
		activeSupplement,
		session: currentSession,
		connectionStatus,
		submitAnswer,
		skipQuestion,
		requestSupplement,
	} = useQaSession({
		sessionId: selectedSessionId || null,
	});

	// 选中 session 时同步到本地 session 信息
	useEffect(() => {
		if (currentSession) {
			setSessions((prev) => {
				const idx = prev.findIndex((s) => s.id === currentSession.id);
				if (idx >= 0) {
					const next = [...prev];
					next[idx] = currentSession;
					return next;
				}
				return [...prev, currentSession];
			});
		}
	}, [currentSession]);

	// 当前活跃的未回答问题
	const activeQuestion = questions.find((q) => q.status === "pending");
	const answeredQuestions = questions.filter(
		(q) => q.status === "answered" || q.answered,
	);

	// 按分组聚合历史记录
	const groupedHistory: Record<string, typeof answeredQuestions> = {};
	for (const q of answeredQuestions) {
		const key = q.groupLabel || "未分组";
		if (!(key in groupedHistory)) groupedHistory[key] = [];
		groupedHistory[key].push(q);
	}

	// 详情面板可见性：有补充内容且未隐藏
	const hasDetailContent =
		!debug.hideMarkdown && activeSupplement != null;

	return (
		<div className="relative flex min-h-0 flex-1 overflow-hidden">
			<aside
				className={`relative flex shrink-0 flex-col p-4 transition-[width,max-width,margin] duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] ${
					hasDetailContent ? "w-[420px]" : "w-full max-w-3xl mx-auto"
				}`}
			>
				<DebugDialog
					config={debug}
					onChange={(patch) => setDebug((d) => ({ ...d, ...patch }))}
				/>

				{selectedSessionId && (
					<div className="mb-2 text-[11px] text-[var(--sea-ink-soft)]">
						连接状态：
						<span
							className={
								connectionStatus === "connected"
									? "text-emerald-600"
									: connectionStatus === "connecting"
										? "text-amber-500"
										: "text-[var(--sea-ink-soft)]/60"
							}
						>
							{connectionStatus === "connected"
								? "已连接"
								: connectionStatus === "connecting"
									? "连接中…"
									: connectionStatus === "disconnected"
										? "已断开"
										: "未连接"}
						</span>
					</div>
				)}

				<ScrollArea className="flex-1 pt-10">
					<div className="space-y-3">
						{activeQuestion ? (
							<QuestionCard
								question={activeQuestion}
								onSubmit={(answer) =>
									submitAnswer(activeQuestion.id, answer)
								}
								onSkip={() => skipQuestion(activeQuestion.id)}
								onRequestSupplement={(targets) =>
									requestSupplement(targets)
								}
							/>
						) : (
							<div className="rounded-xl border border-[var(--line)] bg-[var(--surface)] shadow-[0_2px_12px_rgba(42,36,32,0.06)] backdrop-blur-sm">
								<div className="px-4 py-12 text-center">
									<p className="text-xs text-[var(--sea-ink-soft)]/50">
										{selectedSessionId
											? "等待问题推送…"
											: "请在右侧选择一个会话"}
									</p>
								</div>
							</div>
						)}

						<HistoryCard
							answeredQuestions={answeredQuestions}
							groupedHistory={groupedHistory}
						/>
					</div>
				</ScrollArea>
			</aside>

			<DetailPanel
				visible={hasDetailContent}
				activeOption={
					activeSupplement
						? { label: activeSupplement.content_type }
						: null
				}
				isMotionDemo={false}
				markdownContent={
					activeSupplement?.content_type === "markdown"
						? activeSupplement.content
						: ""
				}
				onBack={() => {}}
			/>

			<SessionSidebar
				sessions={sessions}
				selectedId={selectedSessionId}
				onSelect={setSelectedSessionId}
			/>
		</div>
	);
}
