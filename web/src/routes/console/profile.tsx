import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@lumina/components/ui/tabs'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { PageHeader } from '#/components/page-header'
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
      <PageHeader title="个人信息" description="管理个人资料、生物特征与密码" />

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
