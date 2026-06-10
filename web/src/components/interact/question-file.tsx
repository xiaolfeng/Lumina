import { FileUp, Send, SkipForward, X } from "lucide-react";
import { useRef, useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";

import type { QuestionComponentProps } from "./question-select";

const MAX_SIZE_BYTES = 5 * 1024 * 1024; // 5MB

interface SelectedFile {
	filename: string;
	mimeType: string;
	content: string;
	size: number;
}

export function QuestionFile({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const fileInputRef = useRef<HTMLInputElement>(null);
	const [files, setFiles] = useState<SelectedFile[]>([]);
	const [error, setError] = useState<string>("");

	const acceptTypes = question.config?.accept as string[] | undefined;
	const acceptStr = acceptTypes?.join(",") ?? "";
	const maxFiles = (question.config?.maxFiles as number) ?? 5;

	const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
		setError("");
		const selected = Array.from(e.target.files ?? []);
		if (!selected.length) return;

		// Size check
		const oversized = selected.find((f) => f.size > MAX_SIZE_BYTES);
		if (oversized) {
			setError(`文件 "${oversized.name}" 超过 5MB 限制`);
			return;
		}

		const remaining = maxFiles - files.length;
		const toProcess = selected.slice(0, Math.max(remaining, 0));

		Promise.all(
			toProcess.map(
				(file) =>
					new Promise<SelectedFile>((resolve) => {
						const reader = new FileReader();
						reader.onload = () => {
							resolve({
								filename: file.name,
								mimeType: file.type,
								content: (reader.result as string).split(",")[1]!,
								size: file.size,
							});
						};
						reader.readAsDataURL(file);
					}),
			),
		).then((results) => {
			setFiles((prev) => [...prev, ...results]);
		});

		if (fileInputRef.current) fileInputRef.current.value = "";
	};

	const removeFile = (index: number) => {
		setFiles((prev) => prev.filter((_, i) => i !== index));
	};

	const formatSize = (bytes: number) => {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	};

	const handleSubmit = () => {
		if (files.length === 0) return;
		onSubmit({
			files: files.map(({ filename, mimeType, content }) => ({
				filename,
				mimeType,
				content,
			})),
		});
	};

	return (
		<div className="space-y-4">
			{/* Question content */}
			<div className="prose prose-sm max-w-none [&_p]:mb-1 [&_p]:text-sm [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink)]">
				<Markdown remarkPlugins={[remarkGfm]}>{question.content}</Markdown>
			</div>
			{question.description && (
				<div className="prose prose-sm max-w-none [&_p]:mt-1 [&_p]:mb-2 [&_p]:text-xs [&_p]:leading-relaxed [&_p]:text-[var(--sea-ink-soft)]">
					<Markdown remarkPlugins={[remarkGfm]}>
						{question.description}
					</Markdown>
				</div>
			)}

			{/* File input */}
			<input
				ref={fileInputRef}
				type="file"
				accept={acceptStr || undefined}
				multiple
				onChange={handleFileSelect}
				className="hidden"
				id={`file-input-${question.id}`}
			/>
			<Button
				variant="outline"
				size="sm"
				onClick={() => fileInputRef.current?.click()}
				disabled={files.length >= maxFiles}
				className="w-full border-dashed"
			>
				<FileUp className="mr-1.5 size-4" aria-hidden />
				选择文件 ({files.length}/{maxFiles})
			</Button>

			{/* Error */}
			{error && (
				<p className="text-xs font-medium text-red-500">{error}</p>
			)}

			{/* File list */}
			{files.length > 0 && (
				<div className="space-y-1.5">
					{files.map((file, idx) => (
						<div
							key={`${file.filename}-${idx}`}
							className="flex items-center gap-2 rounded-lg border border-[var(--line)] bg-[var(--foam)] px-3 py-2"
						>
							<FileUp
								className="size-4 shrink-0 text-[var(--lagoon-deep)]"
								aria-hidden
							/>
							<span className="min-w-0 flex-1 truncate text-sm">
								{file.filename}
							</span>
							<span className="shrink-0 text-xs text-[var(--sea-ink-soft)]">
								{formatSize(file.size)}
							</span>
							<button
								type="button"
								onClick={() => removeFile(idx)}
								className="shrink-0 rounded p-0.5 text-[var(--sea-ink-soft)] hover:text-red-500"
							>
								<X className="size-3.5" aria-hidden />
							</button>
						</div>
					))}
				</div>
			)}

			{/* Actions */}
			<div className="flex items-center justify-between pt-1">
				<Button
					variant="ghost"
					size="sm"
					onClick={() => onRequestSupplement([question.id])}
					className="text-xs text-[var(--sea-ink-soft)]"
				>
					请求补充信息
				</Button>
				<div className="flex gap-2">
					<Button variant="outline" size="sm" onClick={onSkip}>
						<SkipForward className="mr-1 size-3.5" aria-hidden />
						跳过
					</Button>
					<Button
						size="sm"
						onClick={handleSubmit}
						disabled={files.length === 0}
						className="rounded-lg bg-[var(--lagoon)] text-white hover:bg-[var(--lagoon)]/90"
					>
						<Send className="mr-1.5 size-3.5" aria-hidden />
						提交
					</Button>
				</div>
			</div>
		</div>
	);
}
