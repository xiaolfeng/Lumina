import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@lumina/components/ui/card'
import { Button } from '@lumina/components/ui/button'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { useSettings, useUpdateSettings } from '#/hooks/useSettings'
import { toast } from 'sonner'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'

export function SecuritySettingsForm() {
  const { data, isLoading } = useSettings('security')
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
      { category: 'security', items },
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
            <CardTitle>安全设置</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="security.access_token_ttl">
                  Access Token TTL（秒）
                </Label>
                <Input
                  id="security.access_token_ttl"
                  type="number"
                  value={formValues['security.access_token_ttl'] || ''}
                  onChange={(e) =>
                    handleChange('security.access_token_ttl', e.target.value)
                  }
                  placeholder="3600"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="security.refresh_token_ttl">
                  Refresh Token TTL（秒）
                </Label>
                <Input
                  id="security.refresh_token_ttl"
                  type="number"
                  value={formValues['security.refresh_token_ttl'] || ''}
                  onChange={(e) =>
                    handleChange('security.refresh_token_ttl', e.target.value)
                  }
                  placeholder="604800"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="security.max_api_keys">
                  最大 API Key 数量
                </Label>
                <Input
                  id="security.max_api_keys"
                  type="number"
                  value={formValues['security.max_api_keys'] || ''}
                  onChange={(e) =>
                    handleChange('security.max_api_keys', e.target.value)
                  }
                  placeholder="10"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="security.webauthn_timeout">
                  WebAuthn 超时（毫秒）
                </Label>
                <Input
                  id="security.webauthn_timeout"
                  type="number"
                  value={formValues['security.webauthn_timeout'] || ''}
                  onChange={(e) =>
                    handleChange('security.webauthn_timeout', e.target.value)
                  }
                  placeholder="60000"
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
