import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs'
import { staggerContainer, staggerItem, staggerItemLeft } from '#/lib/motion'
import { ProfileTab } from '#/components/profile/profile-tab'
import { BiometricTab } from '#/components/profile/biometric-tab'
import { PasswordTab } from '#/components/profile/password-tab'

export const Route = createFileRoute('/console/profile')({
  component: ProfilePage,
})

function ProfilePage() {
  return (
    <motion.div
      className="space-y-6"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      <motion.div variants={staggerItemLeft} className="relative pl-1.5">
        <div className="absolute -left-4 top-0 h-full w-1 rounded-r-full bg-gradient-to-b from-lagoon to-palm" />
        <h1 className="text-2xl font-bold tracking-tight text-sea-ink">个人信息</h1>
        <p className="text-sea-ink-soft">管理个人资料、生物特征与密码</p>
      </motion.div>

      <motion.div variants={staggerItem}>
        <Tabs defaultValue="profile" className="w-full">
          <TabsList>
            <TabsTrigger value="profile">个人资料</TabsTrigger>
            <TabsTrigger value="biometric">生物特征</TabsTrigger>
            <TabsTrigger value="password">修改密码</TabsTrigger>
          </TabsList>
          <TabsContent value="profile">
            <ProfileTab />
          </TabsContent>
          <TabsContent value="biometric">
            <BiometricTab />
          </TabsContent>
          <TabsContent value="password">
            <PasswordTab />
          </TabsContent>
        </Tabs>
      </motion.div>
    </motion.div>
  )
}
