import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Plus } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@lumina/components/ui/tabs'
import { PageHeader } from '#/components/page-header'
import { DataTable } from '#/components/data-table'
import { SkeletonTable } from '#/components/skeleton-table'
import { ConfirmDeleteDialog } from '#/components/confirm-delete-dialog'
import { getProviderColumns } from '#/components/llm/provider-columns'
import { ProviderCreateDialog } from '#/components/llm/provider-create-dialog'
import { ProviderEditDialog } from '#/components/llm/provider-edit-dialog'
import { getModelColumns } from '#/components/llm/model-columns'
import { ModelCreateDialog } from '#/components/llm/model-create-dialog'
import { ModelEditDialog } from '#/components/llm/model-edit-dialog'
import { AgentModelAssignGroup } from '#/components/llm/agent-model-assign'
import { SiteSettingsForm } from '#/components/settings/site-settings-form'
import { QaSettingsForm } from '#/components/settings/qa-settings-form'
import { RepowikiSettingsForm } from '#/components/settings/repowiki-settings-form'
import { SecuritySettingsForm } from '#/components/settings/security-settings-form'
import {
  useProviders,
  useDeleteProvider,
  useModels,
  useDeleteModel,
} from '#/hooks/useLlmConfig'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import type { Provider, Model } from '#/lib/models/response/llm'

export const Route = createFileRoute('/console/settings')({
  component: SettingsPage,
})

function SettingsPage() {
  // Provider dialog state
  const [createProviderOpen, setCreateProviderOpen] = useState(false)
  const [editProviderOpen, setEditProviderOpen] = useState(false)
  const [deleteProviderOpen, setDeleteProviderOpen] = useState(false)
  const [selectedProvider, setSelectedProvider] = useState<Provider | null>(null)

  // Model dialog state
  const [createModelOpen, setCreateModelOpen] = useState(false)
  const [editModelOpen, setEditModelOpen] = useState(false)
  const [deleteModelOpen, setDeleteModelOpen] = useState(false)
  const [selectedModel, setSelectedModel] = useState<Model | null>(null)

  // Data hooks
  const { data: providersData, isLoading: providersLoading } = useProviders()
  const { data: modelsData, isLoading: modelsLoading } = useModels()
  const deleteProviderMutation = useDeleteProvider()
  const deleteModelMutation = useDeleteModel()

  const providerItems = providersData?.data?.items ?? []
  const modelItems = modelsData?.data?.items ?? []

  // Column definitions
  const providerColumns = getProviderColumns({
    onEdit: (item) => {
      setSelectedProvider(item)
      setEditProviderOpen(true)
    },
    onDelete: (item) => {
      setSelectedProvider(item)
      setDeleteProviderOpen(true)
    },
  })

  const modelColumns = getModelColumns({
    onEdit: (item) => {
      setSelectedModel(item)
      setEditModelOpen(true)
    },
    onDelete: (item) => {
      setSelectedModel(item)
      setDeleteModelOpen(true)
    },
  })

  return (
    <motion.div
      className="space-y-4"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      <PageHeader
        title="系统设置"
        description="管理 LLM 配置、站点信息、模块运行参数和安全策略"
      />

      <Tabs defaultValue="provider" className="space-y-4">
        <TabsList>
          <TabsTrigger value="provider">Provider 管理</TabsTrigger>
          <TabsTrigger value="model">模型管理</TabsTrigger>
          <TabsTrigger value="agent">Agent 分配</TabsTrigger>
          <TabsTrigger value="site">站点信息</TabsTrigger>
          <TabsTrigger value="qa">Q&A 配置</TabsTrigger>
          <TabsTrigger value="repowiki">RepoWiki</TabsTrigger>
          <TabsTrigger value="security">安全策略</TabsTrigger>
        </TabsList>

        {/* Provider 管理 */}
        <TabsContent value="provider">
          <motion.div variants={staggerItem} className="space-y-4">
            <div className="flex justify-end">
              <Button
                onClick={() => setCreateProviderOpen(true)}
                className="bg-lagoon text-foam hover:bg-lagoon-deep"
              >
                <Plus className="mr-2 size-4" />
                创建 Provider
              </Button>
            </div>
            {providersLoading ? (
              <SkeletonTable />
            ) : (
              <DataTable columns={providerColumns} data={providerItems} />
            )}
          </motion.div>
        </TabsContent>

        {/* 模型管理 */}
        <TabsContent value="model">
          <motion.div variants={staggerItem} className="space-y-4">
            <div className="flex justify-end">
              <Button
                onClick={() => setCreateModelOpen(true)}
                className="bg-lagoon text-foam hover:bg-lagoon-deep"
              >
                <Plus className="mr-2 size-4" />
                创建模型
              </Button>
            </div>
            {modelsLoading ? (
              <SkeletonTable />
            ) : (
              <DataTable columns={modelColumns} data={modelItems} />
            )}
          </motion.div>
        </TabsContent>

        {/* Agent 分配 */}
        <TabsContent value="agent">
          <motion.div variants={staggerItem}>
            <AgentModelAssignGroup module="repowiki" />
          </motion.div>
        </TabsContent>

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
      </Tabs>

      {/* Provider Dialogs */}
      <ProviderCreateDialog
        open={createProviderOpen}
        onOpenChange={setCreateProviderOpen}
      />
      <ProviderEditDialog
        item={selectedProvider}
        open={editProviderOpen}
        onOpenChange={setEditProviderOpen}
      />
      <ConfirmDeleteDialog
        open={deleteProviderOpen}
        onOpenChange={setDeleteProviderOpen}
        title="删除 Provider"
        description={`确定要删除 Provider「${selectedProvider?.name ?? ''}」吗？此操作不可撤销。`}
        onConfirm={() => {
          if (!selectedProvider) return
          deleteProviderMutation.mutate(selectedProvider.id, {
            onSuccess: () => setDeleteProviderOpen(false),
          })
        }}
        isPending={deleteProviderMutation.isPending}
      />

      {/* Model Dialogs */}
      <ModelCreateDialog
        open={createModelOpen}
        onOpenChange={setCreateModelOpen}
      />
      <ModelEditDialog
        item={selectedModel}
        open={editModelOpen}
        onOpenChange={setEditModelOpen}
      />
      <ConfirmDeleteDialog
        open={deleteModelOpen}
        onOpenChange={setDeleteModelOpen}
        title="删除模型"
        description={`确定要删除模型「${selectedModel?.display_name ?? ''}」吗？此操作不可撤销。`}
        onConfirm={() => {
          if (!selectedModel) return
          deleteModelMutation.mutate(selectedModel.id, {
            onSuccess: () => setDeleteModelOpen(false),
          })
        }}
        isPending={deleteModelMutation.isPending}
      />
    </motion.div>
  )
}
