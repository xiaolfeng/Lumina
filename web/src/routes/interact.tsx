import { createFileRoute, Outlet } from "@tanstack/react-router";
import { Sparkles } from "lucide-react";

export const Route = createFileRoute("/interact")({
	component: InteractLayout,
});

/* ─── Layout Component ─────────────────────────────────── */

function InteractLayout() {
	return (
		<div className="flex h-screen flex-col bg-[var(--bg-base)]">
			{/* 顶部品牌栏 */}
			<header className="flex h-12 shrink-0 items-center gap-2 border-b border-[var(--line)] bg-[var(--header-bg)] px-4 backdrop-blur-md">
				<Sparkles className="size-4 text-[var(--lagoon)]" aria-hidden />
				<span className="display-title text-sm font-bold tracking-tight text-[var(--sea-ink)]">
					Lumina
				</span>
				<span className="hidden items-center gap-1.5 sm:inline-flex">
					<span className="h-px w-3 bg-[var(--lagoon)]/40" />
					<span className="text-[10px] font-semibold uppercase tracking-[0.15em] text-[var(--lagoon-deep)]">
						Q&A
					</span>
				</span>
			</header>

			{/* 三栏主体 */}
			<Outlet />
		</div>
	);
}
