import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Loader2, AlertCircle } from 'lucide-react'
import { PasswordGate } from '#/components/password-gate'
import { WikiLayout } from '#/components/wiki-layout'
import { wikiReaderApi } from '#/lib/api-client'

export const Route = createFileRoute('/wiki/$wikiId/$')({
  component: WikiCatchAllPage,
})

function WikiCatchAllPage() {
  const { wikiId, _splat } = Route.useParams()
  const pagePath = _splat || ''

  const {
    data: pageData,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['wiki-page', wikiId, pagePath],
    queryFn: () => wikiReaderApi.getPage(wikiId, pagePath),
    retry: 1,
    staleTime: 5 * 60 * 1000,
    enabled: !!pagePath,
  })

  let body: React.ReactNode = null
  if (isLoading) {
    body = (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-lagoon" />
        <span className="ml-2 text-sea-ink-soft">加载中...</span>
      </div>
    )
  } else if (error) {
    body = (
      <div className="flex items-center gap-2 rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
        <AlertCircle className="h-5 w-5 flex-shrink-0" />
        <p>{error instanceof Error ? error.message : '加载失败'}</p>
      </div>
    )
  }

  return (
    <PasswordGate wikiId={wikiId}>
      <WikiLayout
        wikiId={wikiId}
        currentPagePath={pagePath}
        content={pageData?.content || ''}
        title={pageData?.title || pagePath || 'Wiki 页面'}
      >
        {body}
      </WikiLayout>
    </PasswordGate>
  )
}
