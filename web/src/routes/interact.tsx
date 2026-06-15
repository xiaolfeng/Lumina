import { useState } from "react";
import { Toaster } from "#/components/ui/sonner";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import { PanelRightOpen, Sparkles } from "lucide-react";
import { SidebarOpenContext, type SessionProgress } from "#/hooks/useSidebarOpen";

export const Route = createFileRoute("/interact")({
	component: InteractLayout,
});

/* ─── Layout Component ─────────────────────────────────── */

function InteractLayout() {
		const [sidebarOpen, setSidebarOpen] = useState(() => {
			if (typeof window !== "undefined") {
				return window.matchMedia("(min-width: 1280px)").matches;
			}
			return false;
		});
		const [progress, setProgress] = useState<SessionProgress | null>(null);
	return (
		<SidebarOpenContext.Provider
			value={{ open: sidebarOpen, setOpen: setSidebarOpen, progress, setProgress }}
		>
			<div className="flex h-screen flex-col bg-bg-base">
				{/* 顶部品牌栏 */}
				<header className="flex h-12 shrink-0 items-center gap-2 border-b border-line bg-header-bg px-4 backdrop-blur-md">
					<Sparkles className="size-4 text-lagoon" aria-hidden />
					<span className="display-title text-sm font-bold tracking-tight text-sea-ink">
						Lumina
					</span>
					<span className="hidden items-center gap-1.5 sm:inline-flex">
						<span className="h-px w-3 bg-lagoon/40" />
						<span className="text-[10px] font-semibold uppercase tracking-[0.15em] text-lagoon-deep">
							Q&A
						</span>
					</span>

					<div className="flex-1" />

					{!sidebarOpen && progress && (
						<div className="flex items-center gap-1.5 text-[11px] text-sea-ink-soft mr-2">
							已答 <span className="font-semibold text-sea-ink">{progress.answered}</span>
							<span className="text-line">·</span>
							剩余 <span className="font-semibold text-lagoon">{progress.remaining}</span>
						</div>
					)}

					{progress && (
						<span className="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-[10px] font-medium text-emerald-600 mr-2">
							<span className="inline-block size-1.5 rounded-full bg-emerald-500" />
							活跃
						</span>
					)}

					<button
						type="button"
						onClick={() => setSidebarOpen(!sidebarOpen)}
						className="inline-flex items-center justify-center rounded-lg p-1.5 text-sea-ink-soft transition-colors hover:bg-line/30"
						aria-label={sidebarOpen ? "关闭会话列表" : "打开会话列表"}
						aria-pressed={sidebarOpen}
					>
						<PanelRightOpen className={`size-4 transition-transform duration-200 ${sidebarOpen ? "rotate-180" : ""}`} aria-hidden />
					</button>
				</header>

				{/* 三栏主体 */}
				<Outlet />

				<Toaster />
			</div>
		</SidebarOpenContext.Provider>
	);
}
