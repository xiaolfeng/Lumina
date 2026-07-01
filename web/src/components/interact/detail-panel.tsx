import { ArrowLeft, RotateCw } from "lucide-react";
import { useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { Button } from "#/components/ui/button";
import { ScrollArea } from "#/components/ui/scroll-area";
import { ease } from "#/lib/motion";

import { Markdown, proseArticle, ShadowHtml } from "./primitives";
import { MotionDemoPanel } from "./motion-demo-panel";
import type { Question } from "./types";

interface DetailPanelProps {
	visible: boolean;
	activeOption: { label: string } | null;
	isMotionDemo: boolean;
	/** 底层（问题级）补充内容 */
	questionContent: string;
	/** 底层内容类型：markdown / html */
	questionContentType: string;
	/** 上层（选项级）补充内容，为空时不渲染上层 */
	optionContent: string;
	/** 上层内容类型：markdown / html */
	optionContentType: string;
	/** 当前选项级 supplement 的 ID（用于内容切换动画 key） */
	optionId: string;
	onBack: () => void;
}

/** 渲染补充内容：markdown 用统一 Markdown 组件（含高亮/katex/mermaid），html 走 Shadow DOM 沙盒隔离 */
function SupplementContent({ content, contentType }: { content: string; contentType: string }) {
	if (contentType === "html") {
		return <ShadowHtml content={content} className={proseArticle} />;
	}
	return (
		<article className={proseArticle}>
			<Markdown>{content}</Markdown>
		</article>
	);
}

export function DetailPanel({
	visible,
	activeOption,
	isMotionDemo,
	questionContent,
	questionContentType,
	optionContent,
	optionContentType,
	optionId,
	onBack,
}: DetailPanelProps) {
	const [renderKey, setRenderKey] = useState(0);
	const hasOption = optionContent.length > 0;

	return (
		<AnimatePresence mode="popLayout">
			{visible && (
				<motion.main
					key="detail-panel"
					initial={{ opacity: 0, x: 60 }}
					animate={{ opacity: 1, x: 0 }}
					exit={{ opacity: 0, x: 60 }}
					transition={{ duration: 0.5, ease }}
					className="relative flex min-w-0 flex-1 flex-col overflow-hidden min-h-0 max-w-3xl"
				>
					{/* 底层：问题级补充内容（始终渲染） */}
					<ScrollArea className="flex-1 min-h-0 pt-2" hideScrollbar>
						<div className="p-4">
							{activeOption && isMotionDemo ? (
								<MotionDemoPanel
									selectedLabel={activeOption.label}
									onBack={onBack}
								/>
							) : (
								<SupplementContent
									content={questionContent}
									contentType={questionContentType}
								/>
							)}
						</div>
					</ScrollArea>

					{/* 上层：选项级补充内容（从右侧滑入覆盖，返回从左向右滑出移开） */}
					<AnimatePresence>
						{hasOption && (
							<motion.div
								key="option-overlay"
								initial={{ x: "100%" }}
								animate={{ x: 0 }}
								exit={{ x: "100%" }}
								transition={{ duration: 0.4, ease }}
								className="absolute inset-0 z-10 flex flex-col bg-bg-base"
							>
								{/* 工具栏：返回 + 重新渲染 */}
								<div className="flex items-center justify-between border-b border-line/50 p-4 pt-6 pb-2">
									<Button
										variant="ghost"
										size="sm"
										onClick={onBack}
										className="text-xs text-sea-ink-soft"
									>
										<ArrowLeft className="mr-1 size-3.5" aria-hidden />
										返回
									</Button>
									{optionContentType === "html" && (
										<Button
											variant="ghost"
											size="sm"
											onClick={() => setRenderKey((k) => k + 1)}
											className="text-xs text-sea-ink-soft"
										>
											<RotateCw className="mr-1 size-3.5" aria-hidden />
											重新渲染
										</Button>
									)}
								</div>
								<ScrollArea className="flex-1 min-h-0" hideScrollbar>
									<div className="px-4 pb-4 pt-3">
										<AnimatePresence mode="wait">
											<motion.div
												key={`${optionId}-${renderKey}`}
												initial={{ opacity: 0 }}
												animate={{ opacity: 1 }}
												exit={{ opacity: 0 }}
												transition={{ duration: 0.2, ease }}
												>
													<SupplementContent
														content={optionContent}
														contentType={optionContentType}
													/>
												</motion.div>
										</AnimatePresence>
									</div>
								</ScrollArea>
							</motion.div>
						)}
					</AnimatePresence>
				</motion.main>
			)}
		</AnimatePresence>
	);
}

export function isMotionDemoQuestion(q: Question | undefined): boolean {
	return q?.id === "debug-motion";
}
