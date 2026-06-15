import { useState } from 'react'
import { AnimatePresence, motion } from 'motion/react'

/**
 * 选项详情入口标签 —— hover 展开"查看详情"，点击触发查看选项级补充。
 * 收敛 select/multi-select/options 三处逐字重复的实现。
 */
export function OptionDetailLabel({
	optId: _optId,
	onClick,
}: {
	optId: string
	onClick: () => void
}) {
	const [isHovered, setIsHovered] = useState(false)

	return (
		<motion.div
			className="flex shrink-0 items-center gap-0.5 self-start rounded-full bg-blue-100 px-1.5 py-0.5 text-[10px] font-medium text-blue-600 cursor-pointer overflow-hidden whitespace-nowrap"
			animate={{ width: isHovered ? 'auto' : 22 }}
			initial={{ width: 22 }}
			transition={{ duration: 0.3, ease: 'easeOut' }}
			onMouseEnter={() => setIsHovered(true)}
			onMouseLeave={() => setIsHovered(false)}
			onClick={(e) => {
				e.stopPropagation()
				e.preventDefault()
				onClick()
			}}
			style={{ maxWidth: 200 }}
		>
			<AnimatePresence initial={false}>
				{isHovered && (
					<motion.span
						key="label-text"
						initial={{ opacity: 0, width: 0 }}
						animate={{ opacity: 1, width: 'auto' }}
						exit={{ opacity: 0, width: 0 }}
						transition={{ duration: 0.3, ease: 'easeOut' }}
						className="overflow-hidden whitespace-nowrap"
					>
						查看详情
					</motion.span>
				)}
			</AnimatePresence>
			<span className="shrink-0">→</span>
		</motion.div>
	)
}
