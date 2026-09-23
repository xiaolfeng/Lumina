import { useState, useEffect, useMemo } from 'react'
import { Input } from '@lumina/components/ui/input'
import { Textarea } from '@lumina/components/ui/textarea'
import { useSettings, useUpdateSettings } from '#/hooks/useSettings'
import { EnvInfoCard } from './env-info-card'
import { PropertyGrid, PropertyRow, SettingsCommitDock } from './property-grid'
import { toast } from 'sonner'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'

export function SiteSettingsForm() {
  const { data, isLoading } = useSettings('site')
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
      { category: 'site', items },
      {
        onSuccess: () => {
          toast.success('站点设置已保存')
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
              title="站点全称 (Site Name)"
              propKey="site.name"
              description="控制台顶部导航栏与浏览器标签页显示的显式名称。"
            >
              <Input
                id="site.name"
                value={formValues['site.name'] || ''}
                onChange={(e) => handleChange('site.name', e.target.value)}
                placeholder="例如：Lumina · 微明"
                className="max-w-lg rounded-none border-line text-xs focus:border-lagoon"
              />
            </PropertyRow>

            <PropertyRow
              title="站点描述 (Description)"
              propKey="site.description"
              description="简短的项目定位说明，用于公开页与元数据标签。"
            >
              <Textarea
                id="site.description"
                value={formValues['site.description'] || ''}
                onChange={(e) =>
                  handleChange('site.description', e.target.value)
                }
                placeholder="输入站点描述"
                rows={3}
                className="max-w-lg rounded-none border-line text-xs focus:border-lagoon"
              />
            </PropertyRow>

            <PropertyRow
              title="站点 Logo 资源路径"
              propKey="site.logo-url"
              description="站点左上角展示的图标矢量图或 PNG 图片网络 URL。"
            >
              <Input
                id="site.logo-url"
                value={formValues['site.logo-url'] || ''}
                onChange={(e) => handleChange('site.logo-url', e.target.value)}
                placeholder="https://example.com/logo.png"
                className="max-w-lg rounded-none border-line text-xs focus:border-lagoon"
              />
            </PropertyRow>

            <PropertyRow
              title="对外访问域名 (Domain)"
              propKey="site.domain"
              description="用于拼接 MCP 接入端点、OAuth 回调以及外部分享深链根端点。"
            >
              <Input
                id="site.domain"
                value={formValues['site.domain'] || ''}
                onChange={(e) => handleChange('site.domain', e.target.value)}
                placeholder="https://your-domain.com"
                className="max-w-lg rounded-none border-line text-xs focus:border-lagoon"
              />
            </PropertyRow>

            <PropertyRow
              title="页脚展示文本 (Footer Text)"
              propKey="site.footer-text"
              description="展示在控制台底部与公开展示态页面的版权与说明文本。"
            >
              <Input
                id="site.footer-text"
                value={formValues['site.footer-text'] || ''}
                onChange={(e) =>
                  handleChange('site.footer-text', e.target.value)
                }
                placeholder="输入页脚文本"
                className="max-w-lg rounded-none border-line text-xs focus:border-lagoon"
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

      <motion.div variants={staggerItem}>
        <EnvInfoCard
          items={[
            { label: 'APP_NAME', value: 'Lumina' },
            { label: 'APP_VERSION', value: 'v1.0.0' },
          ]}
        />
      </motion.div>
    </motion.div>
  )
}
