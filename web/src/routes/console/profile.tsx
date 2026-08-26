import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@lumina/components/ui/tabs'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { PageHeader } from '#/components/page-header'
import { ProfileTab } from '#/components/profile/profile-tab'
import { BiometricTab } from '#/components/profile/biometric-tab'
import { PasswordTab } from '#/components/profile/password-tab'

const PROFILE_TABS = ['profile', 'biometric', 'password'] as const

type ProfileTabId = (typeof PROFILE_TABS)[number]

function isProfileTab(value: unknown): value is ProfileTabId {
  return (
    typeof value === 'string' &&
    (PROFILE_TABS as readonly string[]).includes(value)
  )
}

export const Route = createFileRoute('/console/profile')({
  staticData: { crumb: '个人信息' },
  validateSearch: (
    search: Record<string, unknown>,
  ): { tab?: ProfileTabId } => ({
    tab: isProfileTab(search.tab) ? search.tab : undefined,
  }),
  component: ProfilePage,
})

function ProfilePage() {
  const { tab = 'profile' } = Route.useSearch()
  const navigate = Route.useNavigate()

  return (
    <motion.div
      className="space-y-6"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      <PageHeader title="个人信息" description="管理个人资料、生物特征与密码" />

      <motion.div variants={staggerItem}>
        <Tabs
          value={tab}
          onValueChange={(value) => {
            if (!isProfileTab(value)) return
            void navigate({ search: { tab: value }, replace: true })
          }}
          className="min-w-0 w-full"
        >
          <TabsList variant="line" className="w-full justify-start gap-6">
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
