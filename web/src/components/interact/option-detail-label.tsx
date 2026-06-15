import { useState } from 'react'
import { AnimatePresence, motion } from 'motion/react'

/**
 * 选项详情入口标签 —— hover/active 展开"查看详情"，点击触发查看选项级补充。
 * 收敛 select/multi-select/options 三处逐字重复的实现。
 */
export function OptionDetailLabel({
  optId: _optId,
  onClick,
  isActive = false,
}: {
  optId: string
  onClick: () => void
  isActive?: boolean
}) {
  const [isHovered, setIsHovered] = useState(false)
  const expanded = isActive || isHovered

  return (
    <motion.div
      className={`flex shrink-0 items-center gap-0.5 self-start cursor-pointer overflow-hidden whitespace-nowrap rounded-full px-1.5 py-0.5 text-[10px] font-medium transition-colors duration-200 ${
        isActive ? 'bg-lagoon/15 text-lagoon-deep' : 'bg-blue-100 text-blue-600'
      }`}
      animate={{ width: expanded ? 'auto' : 22 }}
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
        {expanded && (
          <motion.span
            key="label-text"
            initial={{ opacity: 0, width: 0 }}
            animate={{ opacity: 1, width: 'auto' }}
            exit={{ opacity: 0, width: 0 }}
            transition={{ duration: 0.3, ease: 'easeOut' }}
            className="overflow-hidden whitespace-nowrap"
          >
            {isActive ? '当前查看' : '查看详情'}
          </motion.span>
        )}
      </AnimatePresence>
      <span className="shrink-0">→</span>
    </motion.div>
  )
}
