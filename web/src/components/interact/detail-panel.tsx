import { AnimatePresence, motion } from "motion/react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ScrollArea } from "#/components/ui/scroll-area";
import { ease } from "#/lib/motion";

import { MotionDemoPanel } from "./motion-demo-panel";
import type { Question } from "./types";

interface DetailPanelProps {
	visible: boolean;
	activeOption: { label: string } | null;
	isMotionDemo: boolean;
	markdownContent: string;
	onBack: () => void;
}

export function DetailPanel({
	visible,
	activeOption,
	isMotionDemo,
	markdownContent,
	onBack,
}: DetailPanelProps) {
	return (
		<AnimatePresence mode="popLayout">
			{visible && (
				<motion.main
					key="detail-panel"
					initial={{ opacity: 0, x: 60 }}
					animate={{ opacity: 1, x: 0 }}
					exit={{ opacity: 0, x: 60 }}
					transition={{ duration: 0.5, ease }}
					className="flex min-w-0 flex-1 flex-col"
				>
					<ScrollArea className="flex-1">
						<AnimatePresence mode="wait">
							{activeOption && isMotionDemo ? (
								<motion.div
									key={`motion-${activeOption.label}`}
									initial={{ opacity: 0, x: 40 }}
									animate={{ opacity: 1, x: 0 }}
									exit={{ opacity: 0, x: 40 }}
									transition={{ duration: 0.4, ease }}
								>
									<MotionDemoPanel
										selectedLabel={activeOption.label}
										onBack={onBack}
									/>
								</motion.div>
							) : (
								<motion.div
									key="markdown-content"
									initial={{ opacity: 0, x: -40 }}
									animate={{ opacity: 1, x: 0 }}
									exit={{ opacity: 0, x: -40 }}
									transition={{ duration: 0.4, ease }}
								>
									<article className="prose prose-sm prose-slate max-w-none p-8 [&_h1]:display-title [&_h1]:text-2xl [&_h1]:font-bold [&_h1]:text-[var(--sea-ink)] [&_h1]:mb-4 [&_h1]:mt-0 [&_h2]:display-title [&_h2]:text-lg [&_h2]:font-bold [&_h2]:text-[var(--sea-ink)] [&_h2]:mt-6 [&_h2]:mb-2 [&_p]:text-sm [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink-soft)] [&_strong]:text-[var(--sea-ink)] [&_strong]:font-semibold [&_code]:rounded [&_code]:bg-[var(--lagoon)]/8 [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:text-xs [&_code]:text-[var(--lagoon-deep)] [&_pre]:rounded-lg [&_pre]:border [&_pre]:border-[var(--line)] [&_pre]:bg-[var(--foam)] [&_pre]:p-4 [&_pre]:text-sm [&_pre]:text-[var(--sea-ink)] [&_blockquote]:border-l-[var(--lagoon)] [&_blockquote]:bg-[var(--lagoon)]/5 [&_blockquote]:rounded-r-lg [&_blockquote]:px-4 [&_blockquote]:py-2 [&_blockquote]:text-sm [&_blockquote]:text-[var(--sea-ink-soft)] [&_table]:w-full [&_table]:text-sm [&_th]:bg-[var(--foam)] [&_th]:px-3 [&_th]:py-2 [&_th]:text-left [&_th]:text-[var(--sea-ink)] [&_th]:font-semibold [&_td]:px-3 [&_td]:py-2 [&_td]:text-[var(--sea-ink-soft)] [&_td]:border-t [&_td]:border-[var(--line)] [&_hr]:border-[var(--line)] [&_a]:text-[var(--lagoon-deep)] [&_a]:underline [&_ul]:text-sm [&_ul]:text-[var(--sea-ink-soft)] [&_ol]:text-sm [&_ol]:text-[var(--sea-ink-soft)] [&_li]:leading-relaxed [&_em]:text-[var(--sea-ink-soft)]">
										<Markdown remarkPlugins={[remarkGfm]}>
											{markdownContent}
										</Markdown>
									</article>
								</motion.div>
							)}
						</AnimatePresence>
					</ScrollArea>
				</motion.main>
			)}
		</AnimatePresence>
	);
}

export function isMotionDemoQuestion(q: Question | undefined): boolean {
	return q?.id === "debug-motion";
}
