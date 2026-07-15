import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Loader2 } from 'lucide-react'
import { PasswordGate } from '#/components/password-gate'
import { wikiReaderApi } from '#/lib/api-client'

export const Route = createFileRoute('/wiki/$wikiId/')({
  component: WikiIndexRedirect,
})

/**
 * 根路由 → 重定向到 manifest.home 对应的 splat 路由。
 *
 * 不在此处渲染页面内容，确保所有页面渲染统一走 splat 路由（$.tsx），
 * 避免"根路由空 path"与"splat 路由有 path"两种状态导致 motion 动画全量重载。
 */
function WikiIndexRedirect() {
  const { wikiId } = Route.useParams()
  const navigate = useNavigate()

  const { data: manifest } = useQuery({
    queryKey: ['wiki-manifest', wikiId],
    queryFn: () => wikiReaderApi.getManifest(wikiId),
    retry: 1,
    staleTime: 5 * 60 * 1000,
  })

  useEffect(() => {
    if (manifest?.home) {
      navigate({
        to: '/wiki/$wikiId/$',
        params: { wikiId, _splat: manifest.home },
        replace: true,
      })
    }
  }, [manifest, wikiId, navigate])

  return (
    <PasswordGate wikiId={wikiId}>
      <div className="flex min-h-svh items-center justify-center">
        <div className="flex items-center gap-2 text-sm text-sea-ink-soft">
          <Loader2 className="size-4 animate-spin text-lagoon" />
          <span>正在跳转...</span>
        </div>
      </div>
    </PasswordGate>
  )
}
