import { useState, useEffect, useMemo } from 'react'
import { Input } from '@lumina/components/ui/input'
import { useSettings, useUpdateSettings } from '#/hooks/useSettings'
import { PropertyGrid, PropertyRow, SettingsCommitDock } from './property-grid'
import { toast } from 'sonner'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'

export function PreviewSettingsForm() {
  const { data, isLoading } = useSettings('preview')
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
      { category: 'preview', items },
      {
        onSuccess: () => {
          toast.success('Preview 设置已保存')
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
              propKey="preview.session.ttl"
              description="默认 604800 秒（7 天）。只作用于新创建的预览沙盒会话，不回溯已有过期时间。"
            >
              <Input
                id="preview.session.ttl"
                type="number"
                min={1}
                value={formValues['preview.session.ttl'] || ''}
                onChange={(e) =>
                  handleChange('preview.session.ttl', e.target.value)
                }
                placeholder="604800"
                className="max-w-xs rounded-none border-line text-xs font-mono focus:border-lagoon"
              />
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
