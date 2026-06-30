import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card'
import { Settings } from 'lucide-react'
import { staggerContainer, staggerItem } from '#/lib/motion'
import { PageHeader } from '#/components/page-header'

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
      <PageHeader title="系统设置" description="管理 Lumina 系统配置" />

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
