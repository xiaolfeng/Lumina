import { useEffect, useMemo, useState } from 'react'
import { useSettings } from '#/hooks/useSettings'
import { buildMcpUrl, resolveMcpOrigin } from '#/lib/mcp-connect'

export type McpOriginSource = 'site' | 'window' | 'empty'

export function useMcpEndpoint() {
  const { data, isPending, isError } = useSettings('site')
  const siteDomain =
    data?.data?.items.find((item) => item.key === 'site.domain')?.value ?? ''

  const [windowOrigin, setWindowOrigin] = useState('')
  useEffect(() => {
    setWindowOrigin(window.location.origin)
  }, [])

  const settingsSettled = !isPending || isError
  const origin = useMemo(() => {
    const fromSite = resolveMcpOrigin({ siteDomain })
    if (fromSite) return fromSite
    if (!settingsSettled) return ''
    return resolveMcpOrigin({ windowOrigin })
  }, [siteDomain, windowOrigin, settingsSettled])

  const source: McpOriginSource = trimPresent(siteDomain)
    ? 'site'
    : settingsSettled && trimPresent(windowOrigin)
      ? 'window'
      : 'empty'

  return {
    origin,
    mcpUrl: origin ? buildMcpUrl(origin) : '',
    siteDomain,
    windowOrigin,
    source,
    isPending: !origin && !settingsSettled,
  }
}

function trimPresent(value: string): boolean {
  return value.trim().length > 0
}
