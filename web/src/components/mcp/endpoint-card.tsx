import { Link } from '@tanstack/react-router'
import { Button } from '@lumina/components/ui/button'
import { Label } from '@lumina/components/ui/label'
import { useCreateApikey } from '#/hooks/useApikey'
import type { McpOriginSource } from '#/hooks/useMcpEndpoint'
import {
  buildAuthorizationHeader,
  MCP_KEY_PLACEHOLDER,
} from '#/lib/mcp-connect'
import { ChannelKicker } from './channel-band'
import { CopyBlock } from './copy-block'

interface EndpointCardProps {
  mcpUrl: string
  origin: string
  source: McpOriginSource
  apiKey: string
  onApiKeyChange: (value: string) => void
}

const sourceLabel: Record<McpOriginSource, string> = {
  site: '端点来自系统设置中的对外访问域名',
  window: '未配置对外域名，暂用当前浏览器地址',
  empty: '尚无法解析访问地址',
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
        description: '由接入指南生成，完整密钥只在生成页写入配置片段',
      },
      {
        onSuccess: (res) => {
          if (res.data?.key) onApiKeyChange(res.data.key)
        },
      },
    )
  }

  const facts = [
    { label: '传输', value: 'Streamable HTTP' },
    { label: '鉴权', value: 'Authorization: Bearer' },
    { label: '端点', value: mcpUrl || '（等待解析域名）' },
    { label: '密钥', value: generated ? apiKey.trim() : MCP_KEY_PLACEHOLDER },
  ]

  return (
    <div>
      <ChannelKicker>接线端子</ChannelKicker>
      <h1 className="display-title mt-2 text-[34px] font-medium tracking-tight text-sea-ink">
        接入指南
      </h1>
      <p className="mt-2 max-w-[46em] text-[15px] leading-relaxed text-sea-ink-soft">
        Lumina 作为 Streamable HTTP MCP Server 内嵌在当前 HTTP
        服务里。导体要实心——端点、鉴权、配置都落在有底色的信道上。
      </p>
      <p className="mt-2 text-xs text-sea-ink-soft">{sourceLabel[source]}</p>

      <dl className="mt-5 grid grid-cols-1 gap-px bg-line sm:grid-cols-2">
        {facts.map((fact) => (
          <div key={fact.label} className="bg-sand px-3.5 py-3">
            <dt className="text-[11px] font-semibold uppercase tracking-[0.14em] text-sea-ink-soft">
              {fact.label}
            </dt>
            <dd className="mt-1 font-mono text-[13px] leading-snug break-all text-sea-ink">
              {fact.value}
            </dd>
          </div>
        ))}
      </dl>

      <div className="mt-4 grid gap-3 sm:grid-cols-2">
        <div className="space-y-2">
          <Label>对外域名</Label>
          <p className="flex min-h-9 items-center bg-sand px-3 font-mono text-[13px] break-all text-sea-ink">
            {origin || '解析中…'}
          </p>
          <p className="text-xs leading-relaxed text-sea-ink-soft">
            {source === 'site' ? (
              <>
                读取自{' '}
                <Link
                  to="/console/settings"
                  className="text-lagoon-deep underline-offset-2 hover:underline"
                >
                  系统设置 → 站点信息
                </Link>
                。改域名请到那里保存。
              </>
            ) : (
              <>
                Agent 若访问不到当前地址，请到{' '}
                <Link
                  to="/console/settings"
                  className="text-lagoon-deep underline-offset-2 hover:underline"
                >
                  系统设置 → 站点信息
                </Link>{' '}
                填写对外访问域名。
              </>
            )}
          </p>
        </div>
        <div className="space-y-2">
          <Label>令牌</Label>
          <Button
            type="button"
            className="w-full rounded-none bg-lagoon text-foam hover:bg-lagoon-deep"
            onClick={handleGenerate}
            disabled={createMutation.isPending}
          >
            {createMutation.isPending
              ? '生成中…'
              : generated
                ? '再生成一枚'
                : '生成令牌'}
          </Button>
          <p className="text-xs leading-relaxed text-sea-ink-soft">
            {generated ? (
              <>
                完整密钥已写入本页配置，关闭或刷新后无法再看。再生成会另建一枚，旧令牌仍留在
                <Link
                  to="/console/apikey"
                  className="text-lagoon-deep underline-offset-2 hover:underline"
                >
                  令牌管理
                </Link>
                ，那里只能看到前缀。
              </>
            ) : (
              <>
                尚未生成时，下方片段使用{' '}
                <code className="font-mono text-sea-ink">
                  {MCP_KEY_PLACEHOLDER}
                </code>{' '}
                占位。点按钮会创建一枚令牌并立刻写入本页配置。
              </>
            )}
          </p>
        </div>
      </div>

      <CopyBlock
        className="mt-5"
        filename="MCP 端点"
        code={[
          '传输：Streamable HTTP',
          `端点：${mcpUrl || '（等待解析域名）'}`,
          buildAuthorizationHeader(apiKey),
        ].join('\n')}
      />
    </div>
  )
}
