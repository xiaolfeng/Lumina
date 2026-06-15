import { useEffect, useState } from "react";
import { Loader2, X } from "lucide-react";
import { Button } from "#/components/ui/button";

const WARN_THRESHOLD = 30_000; // 30s：显示提示 + 忽略按钮
const DANGER_THRESHOLD = 120_000; // 2min：变红色 + 追加提示

interface SupplementLoadingBannerProps {
	onDismiss: () => void;
}

export function SupplementLoadingBanner({ onDismiss }: SupplementLoadingBannerProps) {
	const [elapsed, setElapsed] = useState(0);

	useEffect(() => {
		const start = Date.now();
		const timer = setInterval(() => {
			setElapsed(Date.now() - start);
		}, 1000);
		return () => clearInterval(timer);
	}, []);

	const isDanger = elapsed >= DANGER_THRESHOLD;
	const showHint = elapsed >= WARN_THRESHOLD;

	return (
		<div
			className={`flex items-center gap-2 rounded-lg p-3 text-sm transition-colors duration-300 ${
				isDanger
					? "bg-red-50 text-red-600"
					: "bg-blue-50 text-blue-600"
			}`}
		>
			<Loader2 className="size-4 shrink-0 animate-spin" aria-hidden />
			<div className="min-w-0 flex-1">
				<p>
					{isDanger
						? "已等待较长时间，若 AI Agent 行为中断请让其继续或忽略补充"
						: showHint
							? "请耐心等待或检查 AI Agent 是否出现中断"
							: "正在加载补充内容..."}
				</p>
			</div>
			{showHint && (
				<Button
					variant="ghost"
					size="sm"
					onClick={onDismiss}
					className={`shrink-0 text-xs ${
						isDanger ? "text-red-600 hover:bg-red-100" : "text-blue-600 hover:bg-blue-100"
					}`}
				>
					<X className="mr-1 size-3.5" aria-hidden />
					忽略
				</Button>
			)}
		</div>
	);
}
