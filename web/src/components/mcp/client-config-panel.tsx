import { MCP_CLIENTS } from '#/lib/mcp-connect'
import type { McpConnectContext } from '#/lib/mcp-connect'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@lumina/components/ui/tabs'
import { CopyBlock } from './copy-block'

interface ClientConfigPanelProps {
  mcpUrl: string
  apiKey?: string | null
  defaultClient?: string
}

export function ClientConfigPanel({
  mcpUrl,
  apiKey,
  defaultClient = 'cursor',
}: ClientConfigPanelProps) {
  const ctx: McpConnectContext = { mcpUrl, apiKey }

  return (
    <Tabs defaultValue={defaultClient} className="gap-3">
      <TabsList
        variant="line"
        className="h-auto w-full flex-wrap justify-start gap-0 rounded-none bg-transparent p-0"
      >
        {MCP_CLIENTS.map((client) => (
          <TabsTrigger
            key={client.id}
            value={client.id}
            className="rounded-none px-3 py-2 text-xs data-[state=active]:bg-foam data-[state=active]:text-sea-ink"
          >
            {client.name}
          </TabsTrigger>
        ))}
      </TabsList>
      {MCP_CLIENTS.map((client) => (
        <TabsContent key={client.id} value={client.id} className="space-y-3">
          <p className="text-[13px] leading-relaxed text-sea-ink-soft">
            <span className="font-medium text-sea-ink">{client.fileHint}</span>
            <span className="mx-1.5 text-line">·</span>
            {client.note}
          </p>
          <CopyBlock code={client.build(ctx)} filename={client.fileHint} />
        </TabsContent>
      ))}
    </Tabs>
  )
}
