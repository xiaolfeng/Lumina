import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@lumina/components/ui/card'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { Switch } from '@lumina/components/ui/switch'
import { useSettings, useUpdateSettings } from '#/hooks/useSettings'
import { toast } from 'sonner'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'

export function QaSettingsForm() {
  const { data, isLoading } = useSettings('qa')
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

  const handleSwitchChange = (key: string, checked: boolean) => {
    setFormValues((prev) => ({ ...prev, [key]: String(checked) }))
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
            <CardTitle>Q&A 设置</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="qa.session.ttl">
                  Session TTL（秒）
                </Label>
                <Input
                  id="qa.session.ttl"
                  type="number"
                  value={formValues['qa.session.ttl'] || ''}
                  onChange={(e) =>
                    handleChange('qa.session.ttl', e.target.value)
                  }
                  placeholder="604800"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="qa.get_answer.poll_slice">
                  轮询间隔（毫秒）
                </Label>
                <Input
                  id="qa.get_answer.poll_slice"
                  type="number"
                  value={formValues['qa.get_answer.poll_slice'] || ''}
                  onChange={(e) =>
                    handleChange('qa.get_answer.poll_slice', e.target.value)
                  }
                  placeholder="1000"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="qa.get_answer.max_retries">
                  最大重试次数
                </Label>
                <Input
                  id="qa.get_answer.max_retries"
                  type="number"
                  value={formValues['qa.get_answer.max_retries'] || ''}
                  onChange={(e) =>
                    handleChange('qa.get_answer.max_retries', e.target.value)
                  }
                  placeholder="30"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="runtime.domain">运行时域名</Label>
                <Input
                  id="runtime.domain"
                  value={formValues['runtime.domain'] || ''}
                  onChange={(e) =>
                    handleChange('runtime.domain', e.target.value)
                  }
                  placeholder="localhost"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="qa.max_active_sessions">
                  最大活跃会话数
                </Label>
                <Input
                  id="qa.max_active_sessions"
                  type="number"
                  value={formValues['qa.max_active_sessions'] || ''}
                  onChange={(e) =>
                    handleChange('qa.max_active_sessions', e.target.value)
                  }
                  placeholder="100"
                />
              </div>
              <div className="flex items-center justify-between rounded-lg border p-4">
                <div className="space-y-0.5">
                  <Label htmlFor="qa.enable_file_upload" className="text-base">
                    启用文件上传
                  </Label>
                  <p className="text-sm text-muted-foreground">
                    允许用户在问答中上传文件
                  </p>
                </div>
                <Switch
                  id="qa.enable_file_upload"
                  checked={formValues['qa.enable_file_upload'] === 'true'}
                  onCheckedChange={(checked) =>
                    handleSwitchChange('qa.enable_file_upload', checked)
                  }
                />
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
