import { useParams } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Loader2, AlertCircle } from 'lucide-react'
import { PasswordGate } from '#/components/password-gate'
import { DocsPage } from '#/components/docs-page'
import { wikiReaderApi } from '#/lib/api-client'
import { buildPageTree } from '#/lib/source'
import { Markdown } from '@lumina/components/markdown'

export default function WikiCatchAllPage() {
  const { wikiId = '', _splat = '' } = useParams({ strict: false })
  const pagePath = _splat || ''

  const {
    data: manifest,
    isLoading: manifestLoading,
    error: manifestError,
  } = useQuery({
    queryKey: ['wiki-manifest', wikiId],
    queryFn: () => wikiReaderApi.getManifest(wikiId),
    retry: 1,
    staleTime: 5 * 60 * 1000,
  })

  const {
    data: pageData,
    isLoading: pageLoading,
    error: pageError,
  } = useQuery({
    queryKey: ['wiki-page', wikiId, pagePath],
    queryFn: () => wikiReaderApi.getPage(wikiId, pagePath),
    retry: 1,
    staleTime: 5 * 60 * 1000,
    enabled: !!pagePath,
  })

  const tree = manifest ? buildPageTree(manifest) : undefined

  const isLoading = manifestLoading || pageLoading
  const error = manifestError || pageError

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
        <AlertCircle className="h-5 w-5 shrink-0" />
        <p>{error instanceof Error ? error.message : '加载失败'}</p>
      </div>
    )
  } else if (pageData) {
    body = <Markdown>{pageData.content}</Markdown>
  }

  return (
    <PasswordGate wikiId={wikiId}>
      <DocsPage wikiId={wikiId} tree={tree} pageData={pageData}>
        {body}
      </DocsPage>
    </PasswordGate>
  )
}
