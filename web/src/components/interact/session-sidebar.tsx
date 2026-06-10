import { Bot, Check, Clock, Plus, Users } from "lucide-react";

import { Button } from "#/components/ui/button";
import { ScrollArea } from "#/components/ui/scroll-area";
import { Separator } from "#/components/ui/separator";

import type { Session } from "./types";

interface SessionSidebarProps {
	sessions: Session[];
	selectedId: string;
	onSelect: (id: string) => void;
}

export function SessionSidebar({
	sessions,
	selectedId,
	onSelect,
}: SessionSidebarProps) {
	const current = sessions.find((s) => s.id === selectedId);
	const answeredCount = current
		? current.questions.filter((q) => q.answered).length
		: 0;
	const totalCount = current?.questions.length ?? 0;
	const remainingCount = totalCount - answeredCount;

	return (
		<aside className="flex w-64 shrink-0 flex-col border-l border-[var(--line)] bg-[var(--surface)]">
			<div className="p-3">
				<Button
					disabled
					variant="outline"
					className="w-full justify-start gap-2 rounded-lg border-dashed border-[var(--line)] text-[var(--sea-ink-soft)] opacity-60"
				>
					<Plus className="size-4" aria-hidden />
					创建会话
				</Button>
			</div>

			<Separator className="bg-[var(--line)]" />

			<div className="px-3 py-2">
				<span className="text-[11px] font-semibold uppercase tracking-[0.15em] text-[var(--sea-ink-soft)]">
					会话列表
				</span>
			</div>

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
				<SessionInfo
					session={current}
					answeredCount={answeredCount}
					totalCount={totalCount}
					remainingCount={remainingCount}
				/>
			)}
		</aside>
	);
}

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
			key={session.id}
			type="button"
			onClick={() => onSelect(session.id)}
			className={`group flex w-full flex-col gap-1.5 rounded-lg px-3 py-2.5 text-left transition-colors duration-150 cursor-pointer ${
				isActive
					? "bg-[var(--lagoon)]/10 text-[var(--sea-ink)]"
					: "text-[var(--sea-ink-soft)] hover:bg-[var(--lagoon)]/5 hover:text-[var(--sea-ink)]"
			}`}
			aria-label={`会话：${session.title}`}
		>
			<span className="text-sm font-medium leading-tight">{session.title}</span>
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

function SessionInfo({
	session,
	answeredCount,
	totalCount,
	remainingCount,
}: {
	session: Session;
	answeredCount: number;
	totalCount: number;
	remainingCount: number;
}) {
	return (
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
								{session.agent}
							</span>
						</div>
						<div className="flex items-center justify-between text-xs">
							<span className="text-[var(--sea-ink-soft)]">创建者</span>
							<span className="font-medium text-[var(--sea-ink)]">
								{session.owner}
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
										totalCount > 0 ? (answeredCount / totalCount) * 100 : 0
									}%`,
								}}
							/>
						</div>
					</div>
				</div>
			</div>
		</>
	);
}
