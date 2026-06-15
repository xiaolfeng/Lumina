import { ImagePlus, X } from 'lucide-react'
import { useRef, useState } from 'react'

import { Button } from '#/components/ui/button'

import { QuestionShell  } from './question-shell'
import type {QuestionComponentProps} from './question-shell';

interface ImagePreview {
	filename: string
	mimeType: string
	content: string
	preview: string
}

export function QuestionImage({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
}: QuestionComponentProps) {
	const fileInputRef = useRef<HTMLInputElement>(null)
	const [previews, setPreviews] = useState<ImagePreview[]>([])

	const multiple = question.config?.multiple ?? false
	const maxImages = (question.config?.maxImages as number) ?? 9

	const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
		const files = Array.from(e.target.files ?? [])
		if (!files.length) return

		const remaining = multiple ? maxImages - previews.length : 1
		if (remaining <= 0) return

		const toProcess = files.slice(0, remaining)

		Promise.all(
			toProcess.map(
				(file) =>
					new Promise<ImagePreview>((resolve) => {
						const reader = new FileReader()
						reader.onload = () => {
							resolve({
								filename: file.name,
								mimeType: file.type,
								content: (reader.result as string).split(',')[1],
								preview: reader.result as string,
							})
						}
						reader.readAsDataURL(file)
					}),
			),
		).then((results) => {
			setPreviews((prev) => (multiple ? [...prev, ...results] : results))
		})

		if (fileInputRef.current) fileInputRef.current.value = ''
	}

	const handleSubmit = () => {
		if (previews.length === 0) return
		onSubmit({
			images: previews.map(({ filename, mimeType, content }) => ({
				filename,
				mimeType,
				content,
			})),
		})
	}

	return (
		<QuestionShell
			question={question}
			isSupplementLoading={isSupplementLoading}
			onSkip={onSkip}
			onRequestSupplement={onRequestSupplement}
			submitDisabled={previews.length === 0}
			onSubmit={handleSubmit}
		>
			<input
				ref={fileInputRef}
				type="file"
				accept="image/*"
				multiple={multiple}
				onChange={handleFileSelect}
				className="hidden"
				disabled={isSupplementLoading}
				id={`image-input-${question.id}`}
			/>
			<Button
				variant="outline"
				size="sm"
				onClick={() => fileInputRef.current?.click()}
				disabled={isSupplementLoading || (multiple && previews.length >= maxImages)}
				className="w-full border-dashed"
			>
				<ImagePlus className="mr-1.5 size-4" aria-hidden />
				选择图片
				{multiple && ` (${previews.length}/${maxImages})`}
			</Button>

			{previews.length > 0 && (
				<div
					className={`grid grid-cols-3 gap-2 ${isSupplementLoading ? 'pointer-events-none opacity-50' : ''}`}
				>
					{previews.map((img, idx) => (
						<div
							key={`${img.filename}-${idx}`}
							className="group relative aspect-square overflow-hidden rounded-lg border border-line bg-foam"
						>
							<img
								src={img.preview}
								alt={img.filename}
								className="size-full object-cover"
							/>
							<button
								type="button"
								onClick={() => setPreviews((prev) => prev.filter((_, i) => i !== idx))}
								className="absolute right-1 top-1 flex size-5 items-center justify-center rounded-full bg-black/60 text-white opacity-0 transition-opacity hover:bg-red-500 group-hover:opacity-100"
								aria-label={`移除 ${img.filename}`}
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
		</QuestionShell>
	)
}
