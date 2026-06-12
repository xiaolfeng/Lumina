import {
	createFileRoute,
	useNavigate,
	useSearch,
} from "@tanstack/react-router";
import { useState, useEffect } from "react";
import { Bot, Check, Clock, FolderOpen, Loader2, Plus, Users } from "lucide-react";
import type { DebugConfig } from "#/components/interact/debug-dialog";
import type { Session } from "#/components/interact/types";

import { DebugDialog } from "#/components/interact/debug-dialog";
import { DetailPanel } from "#/components/interact/detail-panel";
import { HistoryCard } from "#/components/interact/history-card";
import { QuestionCard } from "#/components/interact/question-card";
import { useQaSession } from "#/hooks/useQaSession";
import { ScrollArea } from "#/components/ui/scroll-area";
import { Separator } from "#/components/ui/separator";
import {
	getSessionByHash,
	createSession,
	getSessionList,
} from "#/lib/apis/qa-admin";
import { getProjectList } from "#/lib/apis/project";
import type { ProjectItem } from "#/lib/models/response/project";

interface InteractSearch {
	session?: string;
}

export const Route = createFileRoute("/interact/")({
	validateSearch: (search: Record<string, unknown>): InteractSearch => {
		return {
			session:
				typeof search.session === "string" ? search.session : undefined,
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

	const search = useSearch({ from: "/interact/" });
	const navigate = useNavigate();
	const hashParam = search.session;

	const isConnected = !!sessionHash;

	// URL 有 hash → 加载会话详情
	useEffect(() => {
		if (!hashParam) return;

		setSessionHash(hashParam);
		setIsLoading(true);
		(async () => {
			try {
				const res = await getSessionByHash(hashParam);
				const sessionData = res.data?.items?.[0];
				if (sessionData) {
					setSelectedSessionId(sessionData.id);
				}
			} catch (e) {
				console.error("Failed to load session info:", e);
			} finally {
				setIsLoading(false);
			}
		})();
	}, [hashParam]);

	useEffect(() => {
		sessionStorage.setItem("qa_page_active", "1");
		return () => {
			sessionStorage.removeItem("qa_page_active");
		};
	}, []);

	useEffect(() => {
		(async () => {
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
		})();
	}, []);

	const {
		questions,
		activeSupplement,
		session: currentSession,
		connectionStatus,
		submitAnswer,
		skipQuestion,
		requestSupplement,
		isSupplementLoading,
	} = useQaSession({
		sessionHash: sessionHash || null,
		onReject: () => {
			setSessionHash("");
			setSelectedSessionId("");
			navigate({ to: "/interact", replace: true });
		},
	});

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

	const activeQuestion = questions.find((q) => q.status === "pending");
	const answeredQuestions = questions.filter(
		(q) => q.status === "answered" || q.answered,
	);

	const groupedHistory: Record<string, typeof answeredQuestions> = {};
	for (const q of answeredQuestions) {
		const key = q.groupLabel || "未分组";
		if (!(key in groupedHistory)) groupedHistory[key] = [];
		groupedHistory[key].push(q);
	}

	const hasDetailContent =
		!debug.hideMarkdown && activeSupplement != null;

	if (!isConnected) {
		return <Lobby sessions={sessions} />;
	}

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
							<LoadingCard text="正在准备会话…" />
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
								isSupplementLoading={isSupplementLoading}
							/>
						) : (
							<EmptyCard
								text={
									selectedSessionId
										? "等待问题推送…"
										: "请在右侧选择一个会话"
								}
							/>
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

			<SessionSidebarCompact
				sessions={sessions}
				selectedId={selectedSessionId}
				onSelect={setSelectedSessionId}
			/>
		</div>
	);
}

/* ─── 未连接态：项目选择 + 会话列表 ────────────────────── */

function Lobby({ sessions }: { sessions: Session[] }) {
	const navigate = useNavigate();
	const [projects, setProjects] = useState<ProjectItem[]>([]);
	const [loadingProjects, setLoadingProjects] = useState(true);
	const [creating, setCreating] = useState(false);

	useEffect(() => {
		(async () => {
			try {
				const res = await getProjectList({ size: 100 });
				setProjects(res.data?.items || []);
			} catch (e) {
				console.error("Failed to load projects:", e);
			} finally {
				setLoadingProjects(false);
			}
		})();
	}, []);

	async function handleCreate(projectId: string) {
		setCreating(true);
		try {
			const createRes = await createSession({
				project_id: projectId,
				title: `临时问答 ${new Date().toLocaleString("zh-CN")}`,
				agent: "web",
				type: "temporary",
			});
			const newSession = createRes.data;
			if (newSession?.hash) {
				navigate({
					to: "/interact",
					search: { session: newSession.hash },
					replace: true,
				});
			}
		} catch (e) {
			console.error("Failed to create session:", e);
		} finally {
			setCreating(false);
		}
	}

	return (
		<div className="flex min-h-0 flex-1 items-center justify-center gap-8 p-6">
			<div className="w-full max-w-sm space-y-4">
				<div className="rounded-xl border border-[var(--line)] bg-[var(--surface)] shadow-[0_2px_12px_rgba(42,36,32,0.06)] backdrop-blur-sm">
					<div className="border-b border-[var(--line)]/50 px-4 py-2.5">
						<span className="text-[11px] font-semibold uppercase tracking-[0.15em] text-[var(--sea-ink-soft)]">
							选择项目
						</span>
					</div>

					{loadingProjects ? (
						<LoadingCard text="加载项目中…" />
					) : projects.length === 0 ? (
						<div className="px-4 py-8 text-center">
							<FolderOpen
								className="mx-auto mb-2 size-8 text-[var(--sea-ink-soft)]/30"
								aria-hidden
							/>
							<p className="text-xs text-[var(--sea-ink-soft)]/50">
								暂无可用项目
							</p>
						</div>
					) : (
						<div className="divide-y divide-[var(--line)]/30">
							{projects.map((project) => (
								<button
									key={project.id}
									type="button"
									disabled={creating}
									onClick={() => handleCreate(project.id)}
									className="group flex w-full items-center gap-3 px-4 py-3 text-left transition-colors duration-150 hover:bg-[var(--lagoon)]/5 cursor-pointer disabled:cursor-wait"
								>
									<FolderOpen
										className="size-4 shrink-0 text-[var(--lagoon)]/70 group-hover:text-[var(--lagoon)]"
										aria-hidden
									/>
									<div className="min-w-0 flex-1">
										<p className="truncate text-sm font-medium text-[var(--sea-ink)]">
											{project.name}
										</p>
										{project.description && (
											<p className="truncate text-[11px] text-[var(--sea-ink-soft)]/60">
												{project.description}
											</p>
										)}
									</div>
									{creating ? (
										<Loader2
											className="size-4 animate-spin text-[var(--sea-ink-soft)]/40"
											aria-hidden
										/>
									) : (
										<Plus
											className="size-4 text-[var(--sea-ink-soft)]/30 group-hover:text-[var(--lagoon)]"
											aria-hidden
										/>
									)}
								</button>
							))}
						</div>
					)}
				</div>
			</div>

			<SessionListCentered sessions={sessions} />
		</div>
	);
}

/* ─── 居中会话列表（未连接态） ─────────────────────────── */

function SessionListCentered({ sessions }: { sessions: Session[] }) {
	const navigate = useNavigate();

	if (sessions.length === 0) return null;

	return (
		<div className="w-full max-w-sm">
			<div className="rounded-xl border border-[var(--line)] bg-[var(--surface)] shadow-[0_2px_12px_rgba(42,36,32,0.06)] backdrop-blur-sm">
				<div className="border-b border-[var(--line)]/50 px-4 py-2.5">
					<span className="text-[11px] font-semibold uppercase tracking-[0.15em] text-[var(--sea-ink-soft)]">
						活跃会话
					</span>
				</div>

				<ScrollArea className="max-h-[400px]">
					<div className="space-y-1 p-2">
						{sessions.map((session) => (
							<button
								key={session.id}
								type="button"
								onClick={() => {
									if (session.hash) {
										navigate({
											to: "/interact",
											search: { session: session.hash },
											replace: true,
										});
									}
								}}
								className="group flex w-full flex-col gap-1 rounded-lg px-3 py-2.5 text-left transition-colors duration-150 cursor-pointer hover:bg-[var(--lagoon)]/5"
							>
								<span className="text-sm font-medium leading-tight text-[var(--sea-ink)]">
									{session.title}
								</span>
								<div className="flex items-center gap-2 text-[11px] text-[var(--sea-ink-soft)]">
									<span className="flex items-center gap-0.5">
										<Bot className="size-3" aria-hidden />
										{session.agent}
									</span>
									<span className="flex items-center gap-0.5">
										<Users className="size-3" aria-hidden />
										{session.onlineDevices}
									</span>
									<span className="flex items-center gap-0.5">
										<Clock className="size-2.5" aria-hidden />
										{session.updatedAt}
									</span>
								</div>
							</button>
						))}
					</div>
				</ScrollArea>
			</div>
		</div>
	);
}

/* ─── 侧边栏会话列表（已连接态） ───────────────────────── */

function SessionSidebarCompact({
	sessions,
	selectedId,
	onSelect,
}: {
	sessions: Session[];
	selectedId: string;
	onSelect: (id: string) => void;
}) {
	const current = sessions.find((s) => s.id === selectedId);
	const answeredCount = current
		? current.questions.filter((q) => q.answered).length
		: 0;
	const totalCount = current?.questions.length ?? 0;
	const remainingCount = totalCount - answeredCount;

	return (
		<aside className="flex w-64 shrink-0 flex-col border-l border-[var(--line)] bg-[var(--surface)]">
			<div className="px-3 py-2">
				<span className="text-[11px] font-semibold uppercase tracking-[0.15em] text-[var(--sea-ink-soft)]">
					会话列表
				</span>
			</div>

			<Separator className="bg-[var(--line)]" />

			<ScrollArea className="flex-1 px-2">
				<div className="space-y-1 pb-2">
					{sessions.map((session) => (
						<SessionItem
							key={session.id}
							session={session}
							isActive={session.id === selectedId}
							onSelect={onSelect}
						/>
					))}
				</div>
			</ScrollArea>

			{current && (
				<>
					<Separator className="bg-[var(--line)]" />
					<div className="p-3">
						<div className="rounded-lg border border-[var(--line)] bg-[var(--foam)] p-3">
							<div className="mb-2 flex items-center justify-between">
								<span className="text-[10px] font-semibold uppercase tracking-wider text-[var(--sea-ink-soft)]">
									会话信息
								</span>
								<span className="flex items-center gap-0.5 text-[10px] text-emerald-600">
									<span className="inline-block size-1.5 rounded-full bg-emerald-500" />
									活跃
								</span>
							</div>
							<div className="space-y-1.5">
								<div className="flex items-center justify-between text-xs">
									<span className="text-[var(--sea-ink-soft)]">Agent</span>
									<span className="font-medium text-[var(--sea-ink)]">
										{current.agent}
									</span>
								</div>
								<div className="flex items-center justify-between text-xs">
									<span className="text-[var(--sea-ink-soft)]">进度</span>
									<span className="font-medium text-[var(--sea-ink)]">
										已答 {answeredCount} · 剩余 {remainingCount}
									</span>
								</div>
							</div>

							<div className="mt-2">
								<div className="h-1.5 overflow-hidden rounded-full bg-[var(--line)]">
									<div
										className="h-full rounded-full bg-gradient-to-r from-[var(--lagoon)] to-[var(--palm)] transition-all duration-300"
										style={{
											width: `${
												totalCount > 0
													? (answeredCount / totalCount) * 100
													: 0
											}%`,
										}}
									/>
								</div>
							</div>
						</div>
					</div>
				</>
			)}
		</aside>
	);
}

/* ─── 通用小组件 ──────────────────────────────────────── */

function SessionItem({
	session,
	isActive,
	onSelect,
}: {
	session: Session;
	isActive: boolean;
	onSelect: (id: string) => void;
}) {
	const pending = session.questions.filter((q) => !q.answered).length;

	return (
		<button
			type="button"
			onClick={() => onSelect(session.id)}
			className={`group flex w-full flex-col gap-1.5 rounded-lg px-3 py-2.5 text-left transition-colors duration-150 cursor-pointer ${
				isActive
					? "bg-[var(--lagoon)]/10 text-[var(--sea-ink)]"
					: "text-[var(--sea-ink-soft)] hover:bg-[var(--lagoon)]/5 hover:text-[var(--sea-ink)]"
			}`}
			aria-label={`会话：${session.title}`}
		>
			<span className="text-sm font-medium leading-tight">
				{session.title}
			</span>
			<div className="flex items-center gap-2 text-[11px]">
				<span className="flex items-center gap-0.5">
					<Bot className="size-3" aria-hidden />
					{session.agent}
				</span>
				<span className="flex items-center gap-0.5">
					<Users className="size-3" aria-hidden />
					{session.onlineDevices}
				</span>
			</div>
			<div className="flex items-center justify-between">
				<span className="flex items-center gap-0.5 text-[10px]">
					<Clock className="size-2.5" aria-hidden />
					{session.updatedAt}
				</span>
				{pending > 0 ? (
					<span className="inline-flex size-4 items-center justify-center rounded-full bg-[var(--lagoon)] text-[9px] font-bold text-white">
						{pending}
					</span>
				) : (
					<Check className="size-3.5 text-emerald-500" aria-hidden />
				)}
			</div>
		</button>
	);
}

function LoadingCard({ text }: { text: string }) {
	return (
		<div className="rounded-xl border border-[var(--line)] bg-[var(--surface)] shadow-[0_2px_12px_rgba(42,36,32,0.06)] backdrop-blur-sm">
			<div className="flex flex-col items-center justify-center gap-3 px-4 py-12">
				<div className="h-6 w-6 animate-spin rounded-full border-2 border-[var(--sea-ink-soft)]/20 border-t-[var(--sea-ink-soft)]" />
				<p className="text-xs text-[var(--sea-ink-soft)]/50">{text}</p>
			</div>
		</div>
	);
}

function EmptyCard({ text }: { text: string }) {
	return (
		<div className="rounded-xl border border-[var(--line)] bg-[var(--surface)] shadow-[0_2px_12px_rgba(42,36,32,0.06)] backdrop-blur-sm">
			<div className="px-4 py-12 text-center">
				<p className="text-xs text-[var(--sea-ink-soft)]/50">{text}</p>
			</div>
		</div>
	);
}
