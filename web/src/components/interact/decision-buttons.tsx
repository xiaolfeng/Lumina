import { CheckCircle, Pencil, XCircle } from 'lucide-react'

import { Button } from '@lumina/components/ui/button'

/**
 * 决策按钮组 —— 收敛 diff/plan/review 三处重复的决策按钮。
 *
 * - variant="three"：批准(approve) / 拒绝(reject) / 编辑或修订(edit)（diff、plan）
 * - variant="two"：批准(approve) / 修改(revise)（review）
 *
 * 通过 value 受控，onChange 传出当前选中决策（或 null 表示取消）。
 */
type DecisionValue = 'approve' | 'reject' | 'edit' | 'revise'

interface DecisionButtonsProps {
	variant: 'three' | 'two'
	/** 当前选中的决策（null = 未选） */
	value: DecisionValue | null
	onChange: (value: DecisionValue | null) => void
	disabled?: boolean
	/** 三色模式下第三个按钮（edit）的文案，默认"编辑"（plan 传"修订"） */
	thirdLabel?: string
}

const activeCls = {
	approve: 'bg-emerald-600 hover:bg-emerald-700',
	reject: 'bg-red-600 hover:bg-red-700',
	edit: 'bg-amber-600 hover:bg-amber-700',
	revise: 'bg-amber-600 hover:bg-amber-700',
}
const idleCls = {
	approve: 'border-emerald-300 text-emerald-700 hover:bg-emerald-50',
	reject: 'border-red-300 text-red-700 hover:bg-red-50',
	edit: 'border-amber-300 text-amber-700 hover:bg-amber-50',
	revise: 'border-amber-300 text-amber-700 hover:bg-amber-50',
}

export function DecisionButtons({
	variant,
	value,
	onChange,
	disabled = false,
	thirdLabel = '编辑',
}: DecisionButtonsProps) {
	return (
		<div
			className={`flex flex-wrap gap-2 ${disabled ? 'pointer-events-none opacity-50' : ''}`}
		>
			<Button
				variant={value === 'approve' ? 'default' : 'outline'}
				size="sm"
				onClick={() => onChange(value === 'approve' ? null : 'approve')}
				disabled={disabled}
				className={
					value === 'approve' ? activeCls.approve : idleCls.approve
				}
			>
				<CheckCircle className="mr-1 size-3.5" aria-hidden />
				批准
			</Button>

			{variant === 'three' ? (
				<>
					<Button
						variant={value === 'reject' ? 'default' : 'outline'}
						size="sm"
						onClick={() => onChange(value === 'reject' ? null : 'reject')}
						disabled={disabled}
						className={
							value === 'reject' ? activeCls.reject : idleCls.reject
						}
					>
						<XCircle className="mr-1 size-3.5" aria-hidden />
						拒绝
					</Button>
					<Button
						variant={value === 'edit' ? 'default' : 'outline'}
						size="sm"
						onClick={() => onChange(value === 'edit' ? null : 'edit')}
						disabled={disabled}
						className={value === 'edit' ? activeCls.edit : idleCls.edit}
					>
						<Pencil className="mr-1 size-3.5" aria-hidden />
						{thirdLabel}
					</Button>
				</>
			) : (
				<Button
					variant={value === 'revise' ? 'default' : 'outline'}
					size="sm"
					onClick={() => onChange(value === 'revise' ? null : 'revise')}
					disabled={disabled}
					className={
						value === 'revise' ? activeCls.revise : idleCls.revise
					}
				>
					<Pencil className="mr-1 size-3.5" aria-hidden />
					修改
				</Button>
			)}
		</div>
	)
}
