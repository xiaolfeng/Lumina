import { createFileRoute, useNavigate, useSearch } from "@tanstack/react-router";
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
import {
	getSessionByHash,
	createSession,
	getSessionList,
} from "#/lib/apis/qa-admin";

const SESSION_VISITED_KEY = "qa_session_visited";

interface InteractSearch {
	session?: string;
}

export const Route = createFileRoute("/interact/")({
	validateSearch: (search: Record<string, unknown>): InteractSearch => {
		return {
			session: typeof search.session === "string" ? search.session : undefined,
		};
	},
	component: InteractPage,
});

function InteractPage() {
	const [selectedSessionId, setSelectedSessionId] = useState<string>("");
	const [sessionHash, setSessionHash] = useState<string>("");
	const [sessions, setSessions] = useState<Session[]>([]);
	const [isLoading, setIsLoading] = useState(false);
	const [debug, setDebug] = useState<DebugConfig>({
		hideMarkdown: false,
		forceSelect: false,
		forceMulti: false,
		forceBoolean: false,
		forceText: false,
	});

	// URL 参数读取
	const search = useSearch({ from: "/interact/" });
	const navigate = useNavigate();
	const hashParam = search.session;

	// 初始化 Session：URL 参数查询 / 自动创建 / 复用已有
	useEffect(() => {
		async function initSession() {
			const existingHash = sessionStorage.getItem(SESSION_VISITED_KEY);

			if (hashParam) {
				// 模式 1：URL 中有 hash → 通过 hash 查找 session
				setIsLoading(true);
				try {
					const res = await getSessionByHash(hashParam);
					const sessionData = res.data?.items?.[0];
					if (sessionData) {
						setSelectedSessionId(sessionData.id);
						setSessionHash(hashParam);
						sessionStorage.setItem(SESSION_VISITED_KEY, hashParam);
					}
				} catch (e) {
					console.error("Failed to load session by hash:", e);
				} finally {
					setIsLoading(false);
				}
			} else if (existingHash) {
				// 模式 2：已访问过 → 使用已有的 hash 导航回去（避免重复创建）
				navigate({
					to: "/interact",
					search: { session: existingHash },
					replace: true,
				});
			} else {
				// 模式 3：无 hash 且首次访问 → 自动创建临时 session
				setIsLoading(true);
				try {
					const listRes = await getSessionList({
						status: "active",
						type: "temporary",
						size: 5,
					});
					const existingTemp = listRes.data?.items?.[0];

					if (existingTemp) {
						setSelectedSessionId(existingTemp.id);
						const hash = existingTemp.hash;
						setSessionHash(hash);
						sessionStorage.setItem(SESSION_VISITED_KEY, hash);
						navigate({
							to: "/interact",
							search: { session: hash },
							replace: true,
						});
					} else {
						const createRes = await createSession({
							project_id: "",
							title: `临时问答 ${new Date().toLocaleString("zh-CN")}`,
							agent: "web",
							type: "temporary",
						});
						const newSession = createRes.data;
						if (newSession?.hash) {
							setSelectedSessionId(newSession.id);
							setSessionHash(newSession.hash);
							sessionStorage.setItem(
								SESSION_VISITED_KEY,
								newSession.hash,
							);
							navigate({
								to: "/interact",
								search: { session: newSession.hash },
								replace: true,
							});
						}
					}
				} catch (e) {
					console.error("Failed to create session:", e);
				} finally {
					setIsLoading(false);
				}
			}
		}

		initSession();
	}, [hashParam]);

	useEffect(() => {
		sessionStorage.setItem("qa_page_active", "1");
		return () => {
			sessionStorage.removeItem("qa_page_active");
		};
	}, []);

	useEffect(() => {
		async function loadSessions() {
			try {
				const res = await getSessionList({ status: "active", size: 50 });
				const apiSessions = res.data?.items || [];

				const mapped: Session[] = apiSessions.map((item) => ({
					id: item.id,
					hash: item.hash,
					title: item.title,
					agent: item.agent,
					type: item.type as "temporary" | "permanent",
					status: item.status as "active" | "expired" | "deleted",
					onlineDevices: item.online_devices,
					owner: "",
					questions: [],
					updatedAt: item.updated_at || item.created_at,
					expiresAt: item.expires_at || "",
				}));

				setSessions(mapped);
			} catch (e) {
				console.error("Failed to load sessions:", e);
			}
		}

		loadSessions();
	}, []);

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
		sessionHash: sessionHash || null,
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
						{isLoading ? (
							<div className="rounded-xl border border-[var(--line)] bg-[var(--surface)] shadow-[0_2px_12px_rgba(42,36,32,0.06)] backdrop-blur-sm">
								<div className="flex flex-col items-center justify-center gap-3 px-4 py-12">
									<div className="h-6 w-6 animate-spin rounded-full border-2 border-[var(--sea-ink-soft)]/20 border-t-[var(--sea-ink-soft)]" />
									<p className="text-xs text-[var(--sea-ink-soft)]/50">
										正在准备会话…
									</p>
								</div>
							</div>
						) : activeQuestion ? (
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
