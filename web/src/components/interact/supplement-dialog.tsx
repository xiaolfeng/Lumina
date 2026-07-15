import { useState } from "react";

import { Button } from "@lumina/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@lumina/components/ui/dialog";
import { Label } from "@lumina/components/ui/label";
import { Switch } from "@lumina/components/ui/switch";
import { Textarea } from "@lumina/components/ui/textarea";

export interface SupplementRequestPayload {
	note: string;
	withOptions: boolean;
}

interface SupplementRequestDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** 选择题（select/multi-select）传 true 显示选项开关，其余题型传 false */
	showOptionSwitch: boolean;
	/** 选项数量，开关旁展示提示用 */
	optionCount?: number;
	onConfirm: (payload: SupplementRequestPayload) => void;
}

export function SupplementRequestDialog({
	open,
	onOpenChange,
	showOptionSwitch,
	optionCount,
	onConfirm,
}: SupplementRequestDialogProps) {
	const [note, setNote] = useState("");
	const [withOptions, setWithOptions] = useState(false);

	const handleConfirm = () => {
		onConfirm({ note: note.trim(), withOptions });
		setNote("");
		setWithOptions(false);
		onOpenChange(false);
	};

	const handleOpenChange = (next: boolean) => {
		if (!next) {
			setNote("");
			setWithOptions(false);
		}
		onOpenChange(next);
	};

	return (
		<Dialog open={open} onOpenChange={handleOpenChange}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>请求补充信息</DialogTitle>
					<DialogDescription>
						向 AI 请求对当前问题的更详细说明。可以填写备注指引补充方向。
					</DialogDescription>
				</DialogHeader>
				<div className="grid gap-4 py-4">
					<div className="grid gap-2">
						<Label htmlFor="supplement-note">补充说明（可选）</Label>
						<Textarea
							id="supplement-note"
							value={note}
							onChange={(e) => setNote(e.target.value)}
							placeholder="补充说明（可选）..."
							className="min-h-[100px] resize-y"
						/>
					</div>
					{showOptionSwitch && (
						<div className="flex items-center justify-between rounded-lg border border-line bg-foam px-3 py-2.5">
							<div className="flex flex-col gap-0.5">
								<Label
									htmlFor="supplement-options-switch"
									className="cursor-pointer"
								>
									同时为所有选项请求详细补充
								</Label>
								{optionCount !== undefined && (
									<span className="text-xs text-sea-ink-soft">
										将请求为 {optionCount} 个选项逐一补充说明
									</span>
								)}
							</div>
							<Switch
								id="supplement-options-switch"
								checked={withOptions}
								onCheckedChange={setWithOptions}
							/>
						</div>
					)}
				</div>
				<DialogFooter>
					<Button variant="outline" onClick={() => handleOpenChange(false)}>
						取消
					</Button>
					<Button onClick={handleConfirm}>确认请求</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
