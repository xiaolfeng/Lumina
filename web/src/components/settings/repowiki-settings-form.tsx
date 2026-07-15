import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@lumina/components/ui/card'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { useSettings, useUpdateSettings } from '#/hooks/useSettings'
import { EnvInfoCard } from './env-info-card'
import { toast } from 'sonner'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'

export function RepowikiSettingsForm() {
  const { data, isLoading } = useSettings('repowiki')
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
      { category: 'repowiki', items },
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
            <CardTitle>RepoWiki 设置</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="repowiki.default_language">
                  默认语言
                </Label>
                <Input
                  id="repowiki.default_language"
                  value={formValues['repowiki.default_language'] || ''}
                  onChange={(e) =>
                    handleChange('repowiki.default_language', e.target.value)
                  }
                  placeholder="zh-CN"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="repowiki.default_branch">
                  默认分支
                </Label>
                <Input
                  id="repowiki.default_branch"
                  value={formValues['repowiki.default_branch'] || ''}
                  onChange={(e) =>
                    handleChange('repowiki.default_branch', e.target.value)
                  }
                  placeholder="main"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="repowiki.wiki_cookie_max_age">
                  Wiki Cookie 有效期（秒）
                </Label>
                <Input
                  id="repowiki.wiki_cookie_max_age"
                  type="number"
                  value={formValues['repowiki.wiki_cookie_max_age'] || ''}
                  onChange={(e) =>
                    handleChange(
                      'repowiki.wiki_cookie_max_age',
                      e.target.value,
                    )
                  }
                  placeholder="3600"
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
            { label: 'REPOWIKI_STORAGE_PATH', value: '/data/repowiki' },
            { label: 'REPOWIKI_MAX_CONCURRENT', value: '5' },
          ]}
        />
      </motion.div>
    </motion.div>
  )
}
