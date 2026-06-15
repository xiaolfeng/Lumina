import { useState  } from 'react'
import type {ReactNode} from 'react';
import { Send, SkipForward } from 'lucide-react'

import { Button } from '#/components/ui/button'
import { Markdown, proseQuestion, proseHint } from './primitives'
import { SupplementRequestDialog } from './supplement-dialog'
import type { Question } from './types'
/** 题型组件统一 props —— 所有 question-* 组件共用 */
export interface QuestionComponentProps {
	question: Question
	onSubmit: (answer: any) => void
	onSkip: () => void
	onRequestSupplement: (payload: SupplementRequestArgs) => void
	isSupplementLoading?: boolean
	onDismissSupplementLoading?: () => void
	onViewOptionDetail?: (optId: string) => void
}

export interface SupplementRequestArgs {
	questionId: string
	note: string
	withOptions: boolean
	optionIds?: string[]
}

/**
 * 补充请求配置 —— 决定 SupplementRequestDialog 的展示形态。
 * - basic：非选择题（无选项开关）
 * - options：选择题（显示"为所有选项请求"开关）
 */
interface SupplementConfig {
	kind: 'basic' | 'options'
	/** options 模式下的选项数量 */
	optionCount?: number
	/** options 模式下的选项 ID 列表 */
	optionIds?: string[]
}

interface QuestionShellProps {
	question: Question
	isSupplementLoading?: boolean
	onSkip: () => void
	onRequestSupplement: (payload: SupplementRequestArgs) => void
	/** 补充请求配置，默认 basic */
	supplement?: SupplementConfig
	/** 提交按钮是否禁用 */
	submitDisabled?: boolean
	/** 点击提交 */
	onSubmit: () => void
	/** 自定义操作栏（boolean 题用，替代默认的跳过/提交行）。
	 *  render-prop 形式：传入 openDialog 以便自定义栏也能触发补充请求 dialog。 */
	actions?: (openDialog: () => void) => ReactNode
	/** 题型独特控件区 */
	children: ReactNode
	/** 是否渲染 description（review 题不渲染，默认 true） */
	showDescription?: boolean
	/** 请求补充按钮文案（有已有补充时显示"重新获取"） */
	supplementButtonLabel?: string
}

/**
 * 题型骨架 —— 抽取 15 个题型组件共享的结构：
 * 问题 Markdown(content) + 可选 description + 控件区 + 操作栏 + SupplementRequestDialog。
 *
 * 题型组件只保留独特交互逻辑，通过 children 注入。
 */
export function QuestionShell({
	question,
	isSupplementLoading = false,
	onSkip,
	onRequestSupplement,
	supplement = { kind: 'basic' },
	submitDisabled = false,
	onSubmit,
	actions,
	children,
	showDescription = true,
	supplementButtonLabel,
}: QuestionShellProps) {
	const [supplementDialogOpen, setSupplementDialogOpen] = useState(false)
	const hasSupplements = (question.supplements?.length ?? 0) > 0

	return (
		<div className="space-y-4">
			{/* 问题主文案 */}
			<div className={proseQuestion}>
				<Markdown>{question.content}</Markdown>
			</div>

			{/* 次级描述 */}
			{showDescription && question.description && (
				<div className={proseHint}>
					<Markdown>{question.description}</Markdown>
				</div>
			)}

			{/* 题型独特控件区 */}
			{children}

			{/* 操作栏：自定义优先，否则用默认跳过/提交 */}
			{actions ? (
				actions(() => setSupplementDialogOpen(true))
			) : (
				<div className="flex items-center justify-between pt-1">
					<Button
						variant="ghost"
						size="sm"
						onClick={() => setSupplementDialogOpen(true)}
						disabled={isSupplementLoading}
						className="text-xs text-sea-ink-soft"
					>
						{supplementButtonLabel ?? (hasSupplements ? '重新获取详情' : '请求补充信息')}
					</Button>
					<div className="flex gap-2">
						<Button
							variant="outline"
							size="sm"
							onClick={onSkip}
							disabled={isSupplementLoading}
						>
							<SkipForward className="mr-1 size-3.5" aria-hidden />
							跳过
						</Button>
						<Button
							size="sm"
							onClick={onSubmit}
							disabled={submitDisabled || isSupplementLoading}
							className="rounded-lg bg-gradient-to-b from-lagoon to-lagoon-deep text-white shadow-sm hover:from-lagoon/95 hover:to-lagoon-deep/95"
						>
							<Send className="mr-1.5 size-3.5" aria-hidden />
							提交
						</Button>
					</div>
				</div>
			)}

			<SupplementRequestDialog
				open={supplementDialogOpen}
				onOpenChange={setSupplementDialogOpen}
				showOptionSwitch={supplement.kind === 'options'}
				optionCount={supplement.optionCount}
				onConfirm={({ note, withOptions }) => {
					onRequestSupplement({
						questionId: question.id,
						note,
						withOptions,
						optionIds:
							withOptions && supplement.optionIds ? supplement.optionIds : undefined,
					})
				}}
			/>
		</div>
	)
}
