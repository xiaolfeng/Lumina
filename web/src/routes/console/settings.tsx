import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { Settings } from 'lucide-react'
import { staggerContainer, staggerItem, staggerItemLeft } from '#/lib/motion'

export const Route = createFileRoute('/console/settings')({
  component: SettingsPage,
})

function SettingsPage() {
  return (
    <motion.div
      className="space-y-6"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      <motion.div variants={staggerItemLeft} className="relative pl-1.5">
        <div className="absolute -left-4 top-0 h-full w-1 rounded-r-full bg-gradient-to-b from-lagoon to-palm" />
        <h1 className="text-2xl font-bold tracking-tight text-sea-ink">系统设置</h1>
        <p className="text-sea-ink-soft">管理 Lumina 系统配置</p>
      </motion.div>

      <motion.div variants={staggerItem}>
        <Card className="border-chip-line">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base text-sea-ink">
              <div className="flex size-7 items-center justify-center rounded-lg bg-lagoon/10 text-lagoon">
                <Settings className="size-4" />
              </div>
              通用设置
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-sea-ink-soft">
              系统设置功能正在开发中，敬请期待。
            </p>
          </CardContent>
        </Card>
      </motion.div>
    </motion.div>
  )
}
