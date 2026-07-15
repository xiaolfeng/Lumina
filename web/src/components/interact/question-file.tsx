import { FileUp, X } from 'lucide-react'
import { useRef, useState } from 'react'

import { Button } from '@lumina/components/ui/button'

import { QuestionShell  } from './question-shell'
import type {QuestionComponentProps} from './question-shell';

const MAX_SIZE_BYTES = 5 * 1024 * 1024 // 5MB

interface SelectedFile {
	filename: string
	mimeType: string
	content: string
	size: number
}

export function QuestionFile({
	question,
	onSubmit,
	onSkip,
	onRequestSupplement,
	isSupplementLoading = false,
}: QuestionComponentProps) {
	const fileInputRef = useRef<HTMLInputElement>(null)
	const [files, setFiles] = useState<SelectedFile[]>([])
	const [error, setError] = useState<string>('')

	const acceptTypes = question.config?.accept as string[] | undefined
	const acceptStr = acceptTypes?.join(',') ?? ''
	const maxFiles = (question.config?.maxFiles as number) ?? 5

	const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
		setError('')
		const selected = Array.from(e.target.files ?? [])
		if (!selected.length) return

		const oversized = selected.find((f) => f.size > MAX_SIZE_BYTES)
		if (oversized) {
			setError(`文件 "${oversized.name}" 超过 5MB 限制`)
			return
		}

		const remaining = maxFiles - files.length
		const toProcess = selected.slice(0, Math.max(remaining, 0))

		Promise.all(
			toProcess.map(
				(file) =>
					new Promise<SelectedFile>((resolve) => {
						const reader = new FileReader()
						reader.onload = () => {
							resolve({
								filename: file.name,
								mimeType: file.type,
								content: (reader.result as string).split(',')[1],
								size: file.size,
							})
						}
						reader.readAsDataURL(file)
					}),
			),
		).then((results) => {
			setFiles((prev) => [...prev, ...results])
		})

		if (fileInputRef.current) fileInputRef.current.value = ''
	}

	const formatSize = (bytes: number) => {
		if (bytes < 1024) return `${bytes} B`
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
	}

	const handleSubmit = () => {
		if (files.length === 0) return
		onSubmit({
			files: files.map(({ filename, mimeType, content }) => ({
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
			submitDisabled={files.length === 0}
			onSubmit={handleSubmit}
		>
			<input
				ref={fileInputRef}
				type="file"
				accept={acceptStr || undefined}
				multiple
				onChange={handleFileSelect}
				className="hidden"
				disabled={isSupplementLoading}
				id={`file-input-${question.id}`}
			/>
			<Button
				variant="outline"
				size="sm"
				onClick={() => fileInputRef.current?.click()}
				disabled={isSupplementLoading || files.length >= maxFiles}
				className="w-full border-dashed"
			>
				<FileUp className="mr-1.5 size-4" aria-hidden />
				选择文件 ({files.length}/{maxFiles})
			</Button>

			{error && <p className="text-xs font-medium text-red-500">{error}</p>}

			{files.length > 0 && (
				<div
					className={`space-y-1.5 ${isSupplementLoading ? 'pointer-events-none opacity-50' : ''}`}
				>
					{files.map((file, idx) => (
						<div
							key={`${file.filename}-${idx}`}
							className="flex items-center gap-2 rounded-lg border border-line bg-foam px-3 py-2"
						>
							<FileUp className="size-4 shrink-0 text-lagoon-deep" aria-hidden />
							<span className="min-w-0 flex-1 truncate text-sm">
								{file.filename}
							</span>
							<span className="shrink-0 text-xs text-sea-ink-soft">
								{formatSize(file.size)}
							</span>
							<button
								type="button"
								onClick={() => setFiles((prev) => prev.filter((_, i) => i !== idx))}
								className="shrink-0 rounded p-0.5 text-sea-ink-soft hover:text-red-500"
								aria-label={`移除 ${file.filename}`}
							>
								<X className="size-3.5" aria-hidden />
							</button>
						</div>
					))}
				</div>
			)}
		</QuestionShell>
	)
}
