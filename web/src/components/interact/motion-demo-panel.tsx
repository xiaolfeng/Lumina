import { ArrowLeft } from "lucide-react";
import { motion } from "motion/react";
import { useEffect, useState } from "react";

const motionConfig = {
	从上到下滑入: {
		title: "从上到下滑入",
		initial: { y: -40, opacity: 0 },
		animate: { y: 0, opacity: 1 },
		transition: { duration: 0.5, ease: "easeOut" as const },
		code: `{ initial: { y: -40, opacity: 0 },\n  animate: { y: 0, opacity: 1 },\n  transition: { duration: 0.5, ease: "easeOut" } }`,
		desc: "组件从视口上方滑入，模拟物理空间的「落下」感。适合列表项依次出现、通知弹窗、下拉菜单等场景。",
		tags: ["列表项", "通知弹窗", "下拉菜单"],
	},
	淡入: {
		title: "淡入",
		initial: { opacity: 0 },
		animate: { opacity: 1 },
		transition: { duration: 0.4 },
		code: `{ initial: { opacity: 0 },\n  animate: { opacity: 1 },\n  transition: { duration: 0.4 } }`,
		desc: "最温和的动画方案。不涉及位移，仅通过透明度渐变实现入场。适合内容区块、图片加载完成后的展示。",
		tags: ["内容区块", "图片展示", "Tab 切换"],
	},
	从左到右滑入: {
		title: "从左到右滑入",
		initial: { x: -40, opacity: 0 },
		animate: { x: 0, opacity: 1 },
		transition: { duration: 0.5, ease: "easeOut" as const },
		code: `{ initial: { x: -40, opacity: 0 },\n  animate: { x: 0, opacity: 1 },\n  transition: { duration: 0.5, ease: "easeOut" } }`,
		desc: "模拟阅读方向的动态入场，符合从左到右的视觉习惯。适合侧边栏、导航项、步骤条等水平方向组件。",
		tags: ["侧边栏", "导航项", "步骤条"],
	},
	缩放弹入: {
		title: "缩放弹入",
		initial: { scale: 0.8, opacity: 0 },
		animate: { scale: 1, opacity: 1 },
		transition: { duration: 0.5, ease: [0.34, 1.56, 0.64, 1] },
		code: `{ initial: { scale: 0.8, opacity: 0 },\n  animate: { scale: 1, opacity: 1 },\n  transition: { duration: 0.5, ease: [0.34, 1.56, 0.64, 1] } }`,
		desc: "带弹性缓动的缩放入场，视觉冲击力强。适合模态弹窗、悬浮卡片、重要操作反馈等需要吸引注意力的场景。",
		tags: ["模态弹窗", "悬浮卡片", "操作反馈"],
	},
} as const;

type MotionKey = keyof typeof motionConfig;

interface MotionDemoPanelProps {
	selectedLabel: string;
	onBack: () => void;
}

export function MotionDemoPanel({
	selectedLabel,
	onBack,
}: MotionDemoPanelProps) {
	const [replayKey, setReplayKey] = useState(0);

	const cfg = motionConfig[selectedLabel as MotionKey];

	useEffect(() => {
		setReplayKey((k) => k + 1);
	}, []);

	if (!cfg) return null;

	return (
		<div className="p-8">
			<div className="mb-6 flex items-center justify-between">
				<button
					type="button"
					onClick={onBack}
					className="inline-flex items-center gap-1.5 rounded-lg border border-line bg-surface px-3 py-1.5 text-xs font-medium text-sea-ink-soft shadow-sm transition-colors duration-150 hover:border-lagoon/30 hover:text-lagoon-deep cursor-pointer"
				>
					<ArrowLeft className="size-3.5" aria-hidden />
					返回
				</button>
				<button
					type="button"
					onClick={() => setReplayKey((k) => k + 1)}
					className="inline-flex items-center gap-1.5 rounded-lg border border-line bg-surface px-3 py-1.5 text-xs font-medium text-sea-ink-soft shadow-sm transition-colors duration-150 hover:border-lagoon/30 hover:text-lagoon-deep cursor-pointer"
				>
					⟳ 重播
				</button>
			</div>

			<div className="space-y-6">
				<div>
					<h2 className="display-title mb-2 text-xl font-bold text-sea-ink">
						{cfg.title}
					</h2>
					<p className="mb-4 text-sm leading-relaxed text-sea-ink-soft">
						{cfg.desc}
					</p>
					<pre className="overflow-x-auto rounded-lg border border-line bg-foam p-4 text-xs leading-relaxed text-sea-ink">
						<code>{cfg.code}</code>
					</pre>
				</div>

				<div className="rounded-xl border border-line p-5">
					<p className="mb-3 text-[11px] font-semibold uppercase tracking-[0.15em] text-sea-ink-soft">
						实际效果预览
					</p>
					<motion.div
						key={`${selectedLabel}-${replayKey}`}
						initial={cfg.initial}
						animate={cfg.animate}
						transition={cfg.transition}
						className="rounded-lg bg-foam p-4"
					>
						<div className="flex items-center gap-3">
							<div className="size-10 rounded-lg bg-lagoon/15" />
							<div className="flex-1 space-y-1.5">
								<div className="h-2.5 w-48 rounded-full bg-sea-ink/12" />
								<div className="h-2 w-32 rounded-full bg-sea-ink/8" />
							</div>
							<div className="h-7 w-16 rounded-md bg-lagoon/20" />
						</div>
					</motion.div>
				</div>

				<div className="flex flex-wrap gap-2">
					{cfg.tags.map((tag) => (
						<span
							key={tag}
							className="rounded-md border border-line bg-foam px-2.5 py-1 text-xs text-sea-ink-soft"
						>
							{tag}
						</span>
					))}
				</div>
			</div>
		</div>
	);
}
