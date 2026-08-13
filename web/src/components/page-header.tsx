import { motion } from 'motion/react'
import { staggerItemLeft } from '@lumina/components/motion'

interface PageHeaderProps {
  title: string
  description?: string
  action?: React.ReactNode
}

export function PageHeader({ title, description, action }: PageHeaderProps) {
  return (
    <motion.div
      className="flex items-center justify-between pl-1"
      variants={staggerItemLeft}
    >
      <div>
        <h1 className="display-title text-2xl font-semibold tracking-tight text-sea-ink">
          {title}
        </h1>
        {description && (
          <p className="mt-1.5 text-sm text-sea-ink-soft">{description}</p>
        )}
      </div>
      {action}
    </motion.div>
  )
}
