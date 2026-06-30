import { motion } from 'motion/react'
import { staggerItemLeft } from '#/lib/motion'

interface PageHeaderProps {
  title: string
  description?: string
  action?: React.ReactNode
}

export function PageHeader({ title, description, action }: PageHeaderProps) {
  return (
    <motion.div
      className="relative flex items-center justify-between pl-1.5"
      variants={staggerItemLeft}
    >
      <div className="absolute -left-4 top-0 h-full w-1 rounded-r-full bg-gradient-to-b from-lagoon to-palm" />
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-sea-ink">{title}</h1>
        {description && <p className="text-sea-ink-soft">{description}</p>}
      </div>
      {action}
    </motion.div>
  )
}
