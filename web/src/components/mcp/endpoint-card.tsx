import { Link } from '@tanstack/react-router'
import { Button } from '@lumina/components/ui/button'
import { Label } from '@lumina/components/ui/label'
import { useCreateApikey } from '#/hooks/useApikey'
import type { McpOriginSource } from '#/hooks/useMcpEndpoint'
import { MCP_KEY_PLACEHOLDER } from '#/lib/mcp-connect'
import { ChannelKicker } from './channel-band'

interface EndpointCardProps {
  mcpUrl: string
  origin: string
  source: McpOriginSource
  apiKey: string
  onApiKeyChange: (value: string) => void
}

const sourceLabel: Record<McpOriginSource, string> = {
  site: '当前地址来自“系统设置 → 站点信息”。',
  window: '站点信息还没有填写对外域名，目前使用浏览器中的地址。',
  empty: 'Lumina 还在解析当前站点地址。',
}

export function EndpointCard({
  mcpUrl,
  origin,
  source,
  apiKey,
  onApiKeyChange,
}: EndpointCardProps) {
  const createMutation = useCreateApikey()
  const generated = apiKey.trim().length > 0

  const handleGenerate = () => {
    const stamp = new Date()
    const pad = (value: number) => String(value).padStart(2, '0')
    const name = `MCP 接入 ${stamp.getFullYear()}-${pad(stamp.getMonth() + 1)}-${pad(stamp.getDate())} ${pad(stamp.getHours())}:${pad(stamp.getMinutes())}`
    createMutation.mutate(
      {
        name,
        description: '由接入指南生成，用于插件或独立 MCP 客户端连接',
      },
      {
        onSuccess: (res) => {
          if (res.data?.key) onApiKeyChange(res.data.key)
        },
      },
    )
  }

  return (
    <div>
      <ChannelKicker>开始之前</ChannelKicker>
      <h2 className="display-title mt-2 text-[22px] font-medium text-sea-ink">
        确认地址，再生成一枚接入令牌
      </h2>
      <p className="mt-2 max-w-[48em] text-[13px] leading-relaxed text-sea-ink-soft">
        插件和独立 MCP
        客户端都需要这两项信息。这里生成的令牌只会完整显示一次，本页会自动把它带入后面的安装命令与配置片段。
      </p>

      <div className="mt-5 grid gap-px bg-line sm:grid-cols-2">
        <div className="space-y-2 bg-sand p-4">
          <Label htmlFor="lumina-mcp-origin">Lumina 地址</Label>
          <p
            id="lumina-mcp-origin"
            className="flex min-h-9 items-center bg-foam px-3 font-mono text-[13px] break-all text-sea-ink"
          >
            {origin || '正在解析…'}
          </p>
          <p className="text-xs leading-relaxed text-sea-ink-soft">
            {sourceLabel[source]}{' '}
            <Link
              to="/console/settings"
              className="text-lagoon-deep underline-offset-2 hover:underline"
            >
              修改对外访问域名
            </Link>
          </p>
          <p className="font-mono text-[11px] break-all text-sea-ink-soft">
            MCP 端点：{mcpUrl || '正在解析…'}
          </p>
        </div>

        <div className="space-y-2 bg-chip-bg p-4">
          <Label>接入令牌</Label>
          <Button
            type="button"
            className="w-full rounded-none bg-lagoon text-foam hover:bg-lagoon-deep"
            onClick={handleGenerate}
            disabled={createMutation.isPending}
          >
            {createMutation.isPending
              ? '正在生成…'
              : generated
                ? '再生成一枚令牌'
                : '生成接入令牌'}
          </Button>
          <p className="text-xs leading-relaxed text-sea-ink-soft">
            {generated ? (
              <>
                令牌已经填入本页。刷新页面后无法再次查看；旧令牌可在{' '}
                <Link
                  to="/console/apikey"
                  className="text-lagoon-deep underline-offset-2 hover:underline"
                >
                  令牌管理
                </Link>{' '}
                中停用。
              </>
            ) : (
              <>
                生成前，示例会显示{' '}
                <code className="font-mono text-sea-ink">
                  {MCP_KEY_PLACEHOLDER}
                </code>
                。点击后，本页会立即替换为真实令牌。
              </>
            )}
          </p>
        </div>
      </div>
    </div>
  )
}
