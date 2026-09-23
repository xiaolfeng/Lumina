import { useState, useEffect, useMemo } from 'react'
import { Input } from '@lumina/components/ui/input'
import { useSettings, useUpdateSettings } from '#/hooks/useSettings'
import { EnvInfoCard } from './env-info-card'
import { PropertyGrid, PropertyRow, SettingsCommitDock } from './property-grid'
import { toast } from 'sonner'
import { motion } from 'motion/react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'

export function RepowikiSettingsForm() {
  const { data, isLoading } = useSettings('repowiki')
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
      { category: 'repowiki', items },
      {
        onSuccess: () => {
          setInitialValues(formValues)
          toast.success('RepoWiki 设置已保存')
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
              title="默认文档语言"
              propKey="repowiki.default-language"
              description="RepoWiki 生成架构文档与 Markdown 报告时的首选语言标识。"
            >
              <Input
                id="repowiki.default-language"
                value={formValues['repowiki.default-language'] || ''}
                onChange={(e) =>
                  handleChange('repowiki.default-language', e.target.value)
                }
                placeholder="zh-CN"
                className="max-w-md rounded-none border-line bg-foam text-xs"
              />
            </PropertyRow>

            <PropertyRow
              title="默认分析分支"
              propKey="repowiki.default-branch"
              description="代码仓库进行 5 角色 SubAgent 分析时默认检出并扫描的目标分支。"
            >
              <Input
                id="repowiki.default-branch"
                value={formValues['repowiki.default-branch'] || ''}
                onChange={(e) =>
                  handleChange('repowiki.default-branch', e.target.value)
                }
                placeholder="main"
                className="max-w-md rounded-none border-line bg-foam text-xs"
              />
            </PropertyRow>

            <PropertyRow
              title="Wiki Cookie 有效期 (秒)"
              propKey="repowiki.wiki-cookie-max-age"
              description="Wiki Reader 密码门访问认证 Token 的浏览器 Cookie 保持时长。"
            >
              <Input
                id="repowiki.wiki-cookie-max-age"
                type="number"
                value={formValues['repowiki.wiki-cookie-max-age'] || ''}
                onChange={(e) =>
                  handleChange('repowiki.wiki-cookie-max-age', e.target.value)
                }
                placeholder="3600"
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

      <motion.div variants={staggerItem}>
        <EnvInfoCard
          items={[
            {
              label: 'REPOWIKI_STORAGE_PATH',
              value: 'Wiki 文档存储根目录路径 (只读)',
            },
            {
              label: 'REPOWIKI_MAX_CONCURRENT',
              value: '最大并发分析任务数 (只读)',
            },
            {
              label: 'REPOWIKI_TASK_TIMEOUT',
              value: '单次分析任务超时时间 (只读)',
            },
          ]}
        />
      </motion.div>
    </motion.div>
  )
}
