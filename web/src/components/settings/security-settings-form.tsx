import { useState, useEffect, useMemo } from 'react'
import { Input } from '@lumina/components/ui/input'
import { useSettings, useUpdateSettings } from '#/hooks/useSettings'
import { PropertyGrid, PropertyRow, SettingsCommitDock } from './property-grid'
import { toast } from 'sonner'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'

export function SecuritySettingsForm() {
  const { data, isLoading } = useSettings('security')
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

  const dirtyCount = useMemo(() => {
    let count = 0
    Object.keys(formValues).forEach((key) => {
      if ((formValues[key] || '') !== (initialValues[key] || '')) {
        count++
      }
    })
    return count
  }, [formValues, initialValues])

  const handleChange = (key: string, value: string) => {
    setFormValues((prev) => ({ ...prev, [key]: value }))
  }

  const handleReset = () => {
    setFormValues(initialValues)
    toast.info('已撤销所有未保存的修改')
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const items = Object.entries(formValues).map(([key, value]) => ({
      key,
      value,
    }))
    updateMutation.mutate(
      { category: 'security', items },
      {
        onSuccess: () => {
          setInitialValues(formValues)
          toast.success('安全策略设置已保存')
        },
        onError: () => {
          toast.error('保存失败，请检查参数')
        },
      },
    )
  }

  if (isLoading) {
    return (
      <div className="border border-line bg-foam p-6">
        <div className="h-40 animate-pulse bg-muted" />
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
              title="Access Token TTL (秒)"
              propKey="security.access-token-ttl"
              description="控制台 Bearer 访问令牌的生存有效时长。"
            >
              <Input
                id="security.access-token-ttl"
                type="number"
                value={formValues['security.access-token-ttl'] || ''}
                onChange={(e) =>
                  handleChange('security.access-token-ttl', e.target.value)
                }
                placeholder="3600"
                className="max-w-md rounded-none border-line bg-foam text-xs"
              />
            </PropertyRow>

            <PropertyRow
              title="Refresh Token TTL (秒)"
              propKey="security.refresh-token-ttl"
              description="用于刷新登录态的 Refresh Token 保持时限。"
            >
              <Input
                id="security.refresh-token-ttl"
                type="number"
                value={formValues['security.refresh-token-ttl'] || ''}
                onChange={(e) =>
                  handleChange('security.refresh-token-ttl', e.target.value)
                }
                placeholder="604800"
                className="max-w-md rounded-none border-line bg-foam text-xs"
              />
            </PropertyRow>

            <PropertyRow
              title="最大 API Key 数量配额"
              propKey="security.max-api-keys"
              description="每个工作空间或单个用户允许创建的活动 API Key 上限。"
            >
              <Input
                id="security.max-api-keys"
                type="number"
                value={formValues['security.max-api-keys'] || ''}
                onChange={(e) =>
                  handleChange('security.max-api-keys', e.target.value)
                }
                placeholder="10"
                className="max-w-md rounded-none border-line bg-foam text-xs"
              />
            </PropertyRow>

            <PropertyRow
              title="WebAuthn 超时时间 (毫秒)"
              propKey="security.webauthn-timeout"
              description="通行密钥 (Passkey) 生物认证请求的浏览器握手等待时限。"
            >
              <Input
                id="security.webauthn-timeout"
                type="number"
                value={formValues['security.webauthn-timeout'] || ''}
                onChange={(e) =>
                  handleChange('security.webauthn-timeout', e.target.value)
                }
                placeholder="60000"
                className="max-w-md rounded-none border-line bg-foam text-xs"
              />
            </PropertyRow>
          </PropertyGrid>

          <SettingsCommitDock
            dirtyCount={dirtyCount}
            isSubmitting={updateMutation.isPending}
            onReset={handleReset}
          />
        </form>
      </motion.div>
    </motion.div>
  )
}
