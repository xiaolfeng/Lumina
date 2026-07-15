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

  // 获取 manifest 确定 home 路径
  const {
    data: manifest,
    isLoading: loadingManifest,
    error: manifestError,
  } = useQuery({
    queryKey: ['wiki-manifest', wikiId],
    queryFn: () => wikiReaderApi.getManifest(wikiId),
    retry: 1,
    staleTime: 5 * 60 * 1000,
  })

  // 用 manifest.home 获取首页内容
  const homePath = manifest?.home ?? ''
  const {
    data: pageData,
    isLoading: loadingPage,
    error: pageError,
  } = useQuery({
    queryKey: ['wiki-page', wikiId, homePath],
    queryFn: () => wikiReaderApi.getPage(wikiId, homePath),
    enabled: !!homePath,
    retry: 1,
    staleTime: 5 * 60 * 1000,
  })

  let body: React.ReactNode = null
  if (loadingManifest) {
    body = (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-lagoon" />
        <span className="ml-2 text-sea-ink-soft">加载中...</span>
      </div>
    )
  } else if (manifestError) {
    body = (
      <div className="flex items-center gap-2 rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
        <AlertCircle className="h-5 w-5 flex-shrink-0" />
        <p>
          {manifestError instanceof Error
            ? manifestError.message
            : '加载导航失败'}
        </p>
      </div>
    )
  } else if (loadingPage) {
    body = (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-lagoon" />
        <span className="ml-2 text-sea-ink-soft">加载首页...</span>
      </div>
    )
  } else if (pageError) {
    body = (
      <div className="flex items-center gap-2 rounded-lg border border-destructive/20 bg-destructive/10 p-4 text-destructive">
        <AlertCircle className="h-5 w-5 flex-shrink-0" />
        <p>{pageError instanceof Error ? pageError.message : '加载失败'}</p>
      </div>
    )
  }

  return (
    <PasswordGate wikiId={wikiId}>
      <WikiLayout
        wikiId={wikiId}
        currentPagePath=""
        content={pageData?.content || ''}
        title={pageData?.title || 'Wiki 首页'}
      >
        {body}
      </WikiLayout>
    </PasswordGate>
  )
}
