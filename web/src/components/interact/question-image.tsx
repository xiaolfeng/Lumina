import { ImagePlus, Send, SkipForward, X } from "lucide-react";
import { useRef, useState } from "react";

import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { Button } from "#/components/ui/button";

import type { QuestionComponentProps } from "./question-select";

interface ImagePreview {
	filename: string;
	mimeType: string;
	content: string;
	preview: string;
}

export function QuestionImage({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
}: QuestionComponentProps) {
	const fileInputRef = useRef<HTMLInputElement>(null);
	const [previews, setPreviews] = useState<ImagePreview[]>([]);

	const multiple = question.config?.multiple ?? false;
	const maxImages = (question.config?.maxImages as number) ?? 9;

	const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
		const files = Array.from(e.target.files ?? []);
		if (!files.length) return;

		const remaining = multiple ? maxImages - previews.length : 1;
		if (remaining <= 0) return;

		const toProcess = files.slice(0, remaining);

		Promise.all(
			toProcess.map(
				(file) =>
					new Promise<ImagePreview>((resolve) => {
						const reader = new FileReader();
						reader.onload = () => {
							resolve({
								filename: file.name,
								mimeType: file.type,
								content: (reader.result as string).split(",")[1]!,
								preview: reader.result as string,
							});
						};
						reader.readAsDataURL(file);
					}),
			),
		).then((results) => {
			setPreviews((prev) => (multiple ? [...prev, ...results] : results));
		});

		// Reset input so same file can be re-selected
		if (fileInputRef.current) fileInputRef.current.value = "";
	};

	const removePreview = (index: number) => {
		setPreviews((prev) => prev.filter((_, i) => i !== index));
	};

	const handleSubmit = () => {
		if (previews.length === 0) return;
		onSubmit({
			images: previews.map(({ filename, mimeType, content }) => ({
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
				accept="image/*"
				multiple={multiple}
				onChange={handleFileSelect}
				className="hidden"
				id={`image-input-${question.id}`}
			/>
			<Button
				variant="outline"
				size="sm"
				onClick={() => fileInputRef.current?.click()}
				disabled={multiple && previews.length >= maxImages}
				className="w-full border-dashed"
			>
				<ImagePlus className="mr-1.5 size-4" aria-hidden />
				选择图片
				{multiple && ` (${previews.length}/${maxImages})`}
			</Button>

			{/* Preview grid */}
			{previews.length > 0 && (
				<div className="grid grid-cols-3 gap-2">
					{previews.map((img, idx) => (
						<div
							key={`${img.filename}-${idx}`}
							className="group relative aspect-square overflow-hidden rounded-lg border border-[var(--line)] bg-[var(--foam)]"
						>
							<img
								src={img.preview}
								alt={img.filename}
								className="size-full object-cover"
							/>
							<button
								type="button"
								onClick={() => removePreview(idx)}
								className="absolute right-1 top-1 flex size-5 items-center justify-center rounded-full bg-black/60 text-white opacity-0 transition-opacity hover:bg-red-500 group-hover:opacity-100"
							>
								<X className="size-3" aria-hidden />
							</button>
							<p className="absolute bottom-0 left-0 right-0 truncate bg-gradient-to-t from-black/60 to-transparent px-1.5 pb-1 text-[10px] text-white">
								{img.filename}
							</p>
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
						disabled={previews.length === 0}
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
