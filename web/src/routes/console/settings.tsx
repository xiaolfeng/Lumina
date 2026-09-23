import { useEffect } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Tabs, TabsContent } from '@lumina/components/ui/tabs'
import { PageHeader } from '#/components/page-header'
import { ProviderWorkbench } from '#/components/llm/provider-workbench'
import { ModelWorkbench } from '#/components/llm/model-workbench'
import { AgentModelAssignGroup } from '#/components/llm/agent-model-assign'
import { SiteSettingsForm } from '#/components/settings/site-settings-form'
import { QaSettingsForm } from '#/components/settings/qa-settings-form'
import { RepowikiSettingsForm } from '#/components/settings/repowiki-settings-form'
import { SecuritySettingsForm } from '#/components/settings/security-settings-form'
import { PreviewSettingsForm } from '#/components/settings/preview-settings-form'
import { staggerContainer, staggerItem } from '@lumina/components/motion'

const SETTINGS_TABS = [
  'site',
  'qa',
  'preview',
  'repowiki',
  'security',
  'provider',
  'model',
  'agent',
] as const

type SettingsTab = (typeof SETTINGS_TABS)[number]

function isSettingsTab(value: unknown): value is SettingsTab {
  return (
    typeof value === 'string' &&
    (SETTINGS_TABS as readonly string[]).includes(value)
  )
}

export const Route = createFileRoute('/console/settings')({
  staticData: { crumb: '系统设置' },
  validateSearch: (search: Record<string, unknown>): { tab?: SettingsTab } => ({
    tab: isSettingsTab(search.tab) ? search.tab : undefined,
  }),
  component: SettingsPage,
})

function SettingsPage() {
  const { tab = 'site' } = Route.useSearch()
  const navigate = Route.useNavigate()

  // 🌟 监听 ESC 键一键退出设置并返回控制台看板
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        void navigate({ to: '/console/dashboard' })
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [navigate])

  const tabLabels: Record<SettingsTab, string> = {
    site: '站点基础信息',
    qa: 'Q&A 运行配置',
    preview: 'Preview 沙盒配置',
    repowiki: 'RepoWiki 分析参数',
    security: '安全策略与访问控制',
    provider: 'Provider 供应方管理',
    model: '模型目录管理',
    agent: 'Agent 角色模型分配',
  }

  return (
    <motion.div
      className="space-y-4"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      <PageHeader
        title={tabLabels[tab] || '系统设置'}
        description="管理 LLM 配置、站点信息、模块运行参数和安全策略"
      />

      <Tabs value={tab} className="min-w-0 space-y-4">
        {/* 站点信息 */}
        <TabsContent value="site">
          <motion.div variants={staggerItem}>
            <SiteSettingsForm />
          </motion.div>
        </TabsContent>

        {/* Q&A 配置 */}
        <TabsContent value="qa">
          <motion.div variants={staggerItem}>
            <QaSettingsForm />
          </motion.div>
        </TabsContent>

        <TabsContent value="preview">
          <motion.div variants={staggerItem}>
            <PreviewSettingsForm />
          </motion.div>
        </TabsContent>

        {/* RepoWiki */}
        <TabsContent value="repowiki">
          <motion.div variants={staggerItem}>
            <RepowikiSettingsForm />
          </motion.div>
        </TabsContent>

        {/* 安全策略 */}
        <TabsContent value="security">
          <motion.div variants={staggerItem}>
            <SecuritySettingsForm />
          </motion.div>
        </TabsContent>

        {/* Provider 管理：专属 Master-Detail 工作台 */}
        <TabsContent value="provider">
          <motion.div variants={staggerItem}>
            <ProviderWorkbench />
          </motion.div>
        </TabsContent>

        {/* 模型管理：专属 Master-Detail 工作台 */}
        <TabsContent value="model">
          <motion.div variants={staggerItem}>
            <ModelWorkbench />
          </motion.div>
        </TabsContent>

        {/* Agent 分配 */}
        <TabsContent value="agent">
          <motion.div variants={staggerItem}>
            <AgentModelAssignGroup module="repowiki" />
          </motion.div>
        </TabsContent>
      </Tabs>
    </motion.div>
  )
}
