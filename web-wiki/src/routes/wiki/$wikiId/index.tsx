import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Loader2, AlertCircle } from 'lucide-react'
import { PasswordGate } from '#/components/password-gate'
import { WikiLayout } from '#/components/wiki-layout'
import { wikiReaderApi } from '#/lib/api-client'

export const Route = createFileRoute('/wiki/$wikiId/')({
  component: WikiHomePage,
})

function WikiHomePage() {
  const { wikiId } = Route.useParams()

  const {
    data: pageData,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['wiki-page', wikiId, ''],
    queryFn: () => wikiReaderApi.getPage(wikiId, ''),
    retry: 1,
    staleTime: 5 * 60 * 1000,
  })

  return (
    <PasswordGate wikiId={wikiId}>
      <WikiLayout
        wikiId={wikiId}
        currentPagePath=""
        content={pageData?.content || ''}
        title={pageData?.title || 'Wiki 首页'}
      >
        {isLoading && (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-lagoon" />
            <span className="ml-2 text-sea-ink-soft">加载中...</span>
          </div>
        )}

        {error && (
          <div className="flex items-center gap-2 rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
            <AlertCircle className="h-5 w-5 flex-shrink-0" />
            <p>{error instanceof Error ? error.message : '加载失败'}</p>
          </div>
        )}
      </WikiLayout>
    </PasswordGate>
  )
}
