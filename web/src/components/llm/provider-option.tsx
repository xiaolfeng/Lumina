import type { Provider } from '#/lib/models/response/llm'

/** Provider 下拉选项：名称 + 协议/域名摘要，便于区分同名或地址各异的 Provider */
export function ProviderOption({ provider }: { provider: Provider }) {
  const host = provider.base_url.replace(/^https?:\/\//, '').split('/')[0]
  const meta = host ? `${provider.protocol} · ${host}` : provider.protocol

  return (
    <span className="flex w-full items-baseline gap-2">
      <span className="shrink-0 font-medium text-sea-ink">{provider.name}</span>
      <span className="truncate text-[10px] text-sea-ink-soft">{meta}</span>
    </span>
  )
}
