import { Bug } from "lucide-react";

import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "#/components/ui/dialog";
import { Label } from "#/components/ui/label";
import { Separator } from "#/components/ui/separator";
import { Switch } from "#/components/ui/switch";

export interface DebugConfig {
	hideMarkdown: boolean;
	forceSelect: boolean;
	forceMulti: boolean;
	forceBoolean: boolean;
	forceText: boolean;
}

interface DebugDialogProps {
	config: DebugConfig;
	onChange: (patch: Partial<DebugConfig>) => void;
}

export function DebugDialog({ config, onChange }: DebugDialogProps) {
	function activateOnly(key: keyof DebugConfig) {
		onChange({
			forceSelect: key === "forceSelect",
			forceMulti: key === "forceMulti",
			forceBoolean: key === "forceBoolean",
			forceText: key === "forceText",
		});
	}

	return (
		<Dialog>
			<DialogTrigger asChild>
				<button
					type="button"
					className="absolute right-6 top-4 z-20 flex size-8 items-center justify-center rounded-lg border border-[var(--line)] bg-[var(--surface)] text-[var(--sea-ink-soft)] shadow-[0_2px_8px_rgba(42,36,32,0.08)] transition-all duration-200 hover:border-[var(--lagoon)]/30 hover:bg-[var(--lagoon)]/10 hover:text-[var(--lagoon-deep)] cursor-pointer"
					aria-label="Debug 配置"
				>
					<Bug className="size-4" aria-hidden />
				</button>
			</DialogTrigger>

			<DialogContent className="sm:max-w-[420px] rounded-xl border-[var(--line)] bg-[var(--surface)]">
				<DialogHeader>
					<DialogTitle className="display-title text-lg text-[var(--sea-ink)]">
						Debug 配置
					</DialogTitle>
					<DialogDescription className="text-sm text-[var(--sea-ink-soft)]">
						模拟不同场景以预览界面效果，同一时间仅激活一种题型模拟。
					</DialogDescription>
				</DialogHeader>

				<div className="space-y-5 py-2">
					<div className="flex items-center justify-between rounded-lg border border-[var(--line)] bg-[var(--foam)] px-4 py-3">
						<div className="space-y-0.5">
							<Label className="text-sm font-medium text-[var(--sea-ink)]">
								显示渲染面板内容
							</Label>
							<p className="text-xs text-[var(--sea-ink-soft)]">
								关闭后中栏显示空白占位
							</p>
						</div>
						<Switch
							checked={!config.hideMarkdown}
							onCheckedChange={(v) => onChange({ hideMarkdown: !v })}
							aria-label="切换渲染面板显示"
						/>
					</div>

					<Separator className="bg-[var(--line)]" />

					<div className="space-y-3">
						<span className="text-[11px] font-semibold uppercase tracking-[0.15em] text-[var(--sea-ink-soft)]">
							题型模拟
						</span>

						{(
							[
								{
									key: "forceSelect",
									label: "选择题场景",
									desc: "模拟 select 类型问题",
								},
								{
									key: "forceMulti",
									label: "多选题场景",
									desc: "模拟 multi-select 类型问题",
								},
								{
									key: "forceBoolean",
									label: "判断题场景",
									desc: "模拟 boolean 类型问题",
								},
								{
									key: "forceText",
									label: "文本题场景",
									desc: "模拟 text 类型问题",
								},
							] as const
						).map((item) => (
							<div
								key={item.key}
								className="flex items-center justify-between rounded-lg border border-[var(--line)] bg-[var(--foam)] px-4 py-3"
							>
								<div className="space-y-0.5">
									<Label className="text-sm font-medium text-[var(--sea-ink)]">
										{item.label}
									</Label>
									<p className="text-xs text-[var(--sea-ink-soft)]">
										{item.desc}
									</p>
								</div>
								<Switch
									checked={config[item.key]}
									onCheckedChange={(v) => {
										if (v) activateOnly(item.key);
										else onChange({ [item.key]: false });
									}}
									aria-label={`切换${item.label}模拟`}
								/>
							</div>
						))}
					</div>
				</div>

				<DialogFooter>
					<DialogClose asChild>
						<Button
							variant="outline"
							className="rounded-lg border-[var(--line)] text-[var(--sea-ink-soft)]"
						>
							关闭
						</Button>
					</DialogClose>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
