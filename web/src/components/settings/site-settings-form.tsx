import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@lumina/components/ui/card'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { Textarea } from '@lumina/components/ui/textarea'
import { useSettings, useUpdateSettings } from '#/hooks/useSettings'
import { EnvInfoCard } from './env-info-card'
import { toast } from 'sonner'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'

export function SiteSettingsForm() {
  const { data, isLoading } = useSettings('site')
  const updateMutation = useUpdateSettings()
  const [formValues, setFormValues] = useState<Record<string, string>>({})

  useEffect(() => {
    if (data?.data?.items) {
      const map: Record<string, string> = {}
      data.data.items.forEach((item) => {
        map[item.key] = item.value
      })
      setFormValues(map)
    }
  }, [data])

  const handleChange = (key: string, value: string) => {
    setFormValues((prev) => ({ ...prev, [key]: value }))
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
          toast.success('保存成功')
        },
        onError: () => {
          toast.error('保存失败')
        },
      },
    )
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="h-40 animate-pulse rounded-lg bg-muted" />
        </CardContent>
      </Card>
    )
  }

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
    >
      <motion.div variants={staggerItem}>
        <Card>
          <CardHeader>
            <CardTitle>站点设置</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="site.name">站点名称</Label>
                <Input
                  id="site.name"
                  value={formValues['site.name'] || ''}
                  onChange={(e) => handleChange('site.name', e.target.value)}
                  placeholder="输入站点名称"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="site.description">站点描述</Label>
                <Textarea
                  id="site.description"
                  value={formValues['site.description'] || ''}
                  onChange={(e) =>
                    handleChange('site.description', e.target.value)
                  }
                  placeholder="输入站点描述"
                  rows={3}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="site.logo-url">站点 Logo URL</Label>
                <Input
                  id="site.logo-url"
                  value={formValues['site.logo-url'] || ''}
                  onChange={(e) =>
                    handleChange('site.logo-url', e.target.value)
                  }
                  placeholder="https://example.com/logo.png"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="site.domain">对外访问域名</Label>
                <Input
                  id="site.domain"
                  value={formValues['site.domain'] || ''}
                  onChange={(e) =>
                    handleChange('site.domain', e.target.value)
                  }
                  placeholder="https://your-domain.com"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="site.footer-text">页脚文本</Label>
                <Input
                  id="site.footer-text"
                  value={formValues['site.footer-text'] || ''}
                  onChange={(e) =>
                    handleChange('site.footer-text', e.target.value)
                  }
                  placeholder="输入页脚文本"
                />
              </div>
              <Button type="submit" disabled={updateMutation.isPending}>
                {updateMutation.isPending ? '保存中…' : '保存'}
              </Button>
            </form>
          </CardContent>
        </Card>
      </motion.div>
      <motion.div variants={staggerItem}>
        <EnvInfoCard
          items={[
            { label: 'APP_NAME', value: 'Lumina' },
            { label: 'APP_VERSION', value: 'v0.1.0' },
          ]}
        />
      </motion.div>
    </motion.div>
  )
}
