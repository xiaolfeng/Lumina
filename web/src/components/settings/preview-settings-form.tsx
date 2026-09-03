import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@lumina/components/ui/card'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { useSettings, useUpdateSettings } from '#/hooks/useSettings'
import { toast } from 'sonner'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'

export function PreviewSettingsForm() {
  const { data, isLoading } = useSettings('preview')
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
      { category: 'preview', items },
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
            <CardTitle>Preview 设置</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="preview.session.ttl">
                  会话 TTL（秒）
                </Label>
                <Input
                  id="preview.session.ttl"
                  type="number"
                  min={1}
                  value={formValues['preview.session.ttl'] || ''}
                  onChange={(e) =>
                    handleChange('preview.session.ttl', e.target.value)
                  }
                  placeholder="604800"
                />
                <p className="text-xs text-muted-foreground">
                  默认 604800（7 天）。只作用于新创建的会话，不回溯已有过期时间。
                </p>
              </div>
              <Button type="submit" disabled={updateMutation.isPending}>
                {updateMutation.isPending ? '保存中…' : '保存'}
              </Button>
            </form>
          </CardContent>
        </Card>
      </motion.div>
    </motion.div>
  )
}
