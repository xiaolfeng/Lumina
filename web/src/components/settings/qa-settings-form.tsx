import { useState, useEffect, useMemo } from 'react'
import { Input } from '@lumina/components/ui/input'
import { Switch } from '@lumina/components/ui/switch'
import { useSettings, useUpdateSettings } from '#/hooks/useSettings'
import { PropertyGrid, PropertyRow, SettingsCommitDock } from './property-grid'
import { toast } from 'sonner'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'

export function QaSettingsForm() {
  const { data, isLoading } = useSettings('qa')
  const updateMutation = useUpdateSettings()
  const [initialValues, setInitialValues] = useState<Record<string, string>>({})
  const [formValues, setFormValues] = useState<Record<string, string>>({})

  useEffect(() => {
    if (data?.data?.items) {
      const map: Record<string, string> = {}
      data.data.items.forEach((item) => {
        map[item.key] = item.value
      })
      setInitialValues(map)
      setFormValues(map)
    }
  }, [data])

  const dirtyKeys = useMemo(() => {
    const keys: string[] = []
    for (const key of Object.keys(formValues)) {
      if ((formValues[key] || '') !== (initialValues[key] || '')) {
        keys.push(key)
      }
    }
    return keys
  }, [formValues, initialValues])

  const handleChange = (key: string, value: string) => {
    setFormValues((prev) => ({ ...prev, [key]: value }))
  }

  const handleSwitchChange = (key: string, checked: boolean) => {
    setFormValues((prev) => ({ ...prev, [key]: String(checked) }))
  }

  const handleReset = () => {
    setFormValues(initialValues)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const items = Object.entries(formValues).map(([key, value]) => ({
      key,
      value,
    }))
    updateMutation.mutate(
      { category: 'qa', items },
      {
        onSuccess: () => {
          toast.success('Q&A 设置已保存')
          setInitialValues(formValues)
        },
        onError: () => {
          toast.error('保存失败，请重试')
        },
      },
    )
  }

  if (isLoading) {
    return (
      <div className="border border-line bg-foam p-8 space-y-4">
        <div className="h-6 w-48 bg-muted animate-pulse rounded-none" />
        <div className="h-10 w-full bg-muted animate-pulse rounded-none" />
        <div className="h-10 w-full bg-muted animate-pulse rounded-none" />
      </div>
    )
  }

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-6"
    >
      <motion.div variants={staggerItem}>
        <form onSubmit={handleSubmit}>
          <PropertyGrid>
            <PropertyRow
              title="会话生存周期 (Session TTL)"
              propKey="qa.session.ttl"
              description="问答会话在未发生交互后的存活秒数，超时后自动归档（默认 7 天）。"
            >
              <Input
                id="qa.session.ttl"
                type="number"
                value={formValues['qa.session.ttl'] || ''}
                onChange={(e) => handleChange('qa.session.ttl', e.target.value)}
                placeholder="604800"
                className="max-w-xs rounded-none border-line text-xs font-mono focus:border-lagoon"
              />
            </PropertyRow>

            <PropertyRow
              title="队列轮询片节拍 (Poll Slice)"
              propKey="qa.get-answer.poll-slice"
              description="Agent 通过 MCP 等待回答时，后台队列轮询的时间步长（毫秒）。"
            >
              <Input
                id="qa.get-answer.poll-slice"
                type="number"
                value={formValues['qa.get-answer.poll-slice'] || ''}
                onChange={(e) =>
                  handleChange('qa.get-answer.poll-slice', e.target.value)
                }
                placeholder="1000"
                className="max-w-xs rounded-none border-line text-xs font-mono focus:border-lagoon"
              />
            </PropertyRow>

            <PropertyRow
              title="最大重试轮次 (Max Retries)"
              propKey="qa.get-answer.max-retries"
              description="等待回答超时的最大轮询次数上限，超过后向 Agent 返回超时。"
            >
              <Input
                id="qa.get-answer.max-retries"
                type="number"
                value={formValues['qa.get-answer.max-retries'] || ''}
                onChange={(e) =>
                  handleChange('qa.get-answer.max-retries', e.target.value)
                }
                placeholder="30"
                className="max-w-xs rounded-none border-line text-xs font-mono focus:border-lagoon"
              />
            </PropertyRow>

            <PropertyRow
              title="最大活跃会话并发限制"
              propKey="qa.max-active-sessions"
              description="当前工作空间允许同时处于等待或应答状态的最大会话数。"
            >
              <Input
                id="qa.max-active-sessions"
                type="number"
                value={formValues['qa.max-active-sessions'] || ''}
                onChange={(e) =>
                  handleChange('qa.max-active-sessions', e.target.value)
                }
                placeholder="100"
                className="max-w-xs rounded-none border-line text-xs font-mono focus:border-lagoon"
              />
            </PropertyRow>

            <PropertyRow
              title="交互式文件附件上传"
              propKey="qa.enable-file-upload"
              description="允许用户在问答交互界面提交图片与本地辅助文件附件。"
            >
              <div className="flex items-center gap-3">
                <Switch
                  id="qa.enable-file-upload"
                  checked={formValues['qa.enable-file-upload'] === 'true'}
                  onCheckedChange={(checked) =>
                    handleSwitchChange('qa.enable-file-upload', checked)
                  }
                  className="rounded-none data-[state=checked]:bg-lagoon"
                />
                <span className="font-mono text-xs text-sea-ink-soft">
                  {formValues['qa.enable-file-upload'] === 'true'
                    ? '已启用 (ENABLED)'
                    : '已禁用 (DISABLED)'}
                </span>
              </div>
            </PropertyRow>
          </PropertyGrid>

          <SettingsCommitDock
            dirtyCount={dirtyKeys.length}
            isSubmitting={updateMutation.isPending}
            onReset={handleReset}
          />
        </form>
      </motion.div>
    </motion.div>
  )
}
