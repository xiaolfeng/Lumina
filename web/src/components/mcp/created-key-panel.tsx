import { Link } from '@tanstack/react-router'
import { Button } from '@lumina/components/ui/button'
import { CopyBlock } from './copy-block'
import { ClientConfigPanel } from './client-config-panel'

interface CreatedKeyPanelProps {
  apiKey: string
  mcpUrl: string
  onDismiss: () => void
}

export function CreatedKeyPanel({
  apiKey,
  mcpUrl,
  onDismiss,
}: CreatedKeyPanelProps) {
  return (
    <div className="space-y-4">
      <div className="bg-chip-bg px-4 py-3 text-sm text-sea-ink">
        完整密钥只显示这一次。关闭后无法再查看，请立刻复制到客户端配置，不要提交到
        Git。
      </div>
      <CopyBlock filename="API Key" code={apiKey} />
      <div className="bg-chip-bg p-4">
        <p className="text-[11px] font-bold uppercase tracking-[0.18em] text-lagoon-deep">
          接入 MCP
        </p>
        <p className="mt-1 mb-3 text-xs text-sea-ink-soft">
          端点 <span className="font-mono text-sea-ink">{mcpUrl}</span>
          ，鉴权 Header 为 Authorization: Bearer。
        </p>
        <ClientConfigPanel mcpUrl={mcpUrl} apiKey={apiKey} />
      </div>
      <div className="flex flex-col gap-2 sm:flex-row">
        <Button onClick={onDismiss} className="flex-1">
          我已安全保存
        </Button>
        <Button asChild variant="outline" className="flex-1">
          <Link to="/console/connect" onClick={onDismiss}>
            打开接入指南
          </Link>
        </Button>
      </div>
    </div>
  )
}
