import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Loader2, AlertCircle } from 'lucide-react'
import { PasswordGate } from '#/components/password-gate'
import { DocsPage } from '#/components/docs-page'
import { wikiReaderApi } from '#/lib/api-client'
import { buildPageTree, getIcon } from '#/lib/source'
import { Markdown } from '@lumina/components/markdown'

export const Route = createFileRoute('/wiki/$wikiId/')({
  component: WikiIndexPage,
})

function WikiIndexPage() {
  const { wikiId } = Route.useParams()

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

  const homePath = manifest?.home
  const {
    data: homePageData,
    isLoading: homePageLoading,
    error: homePageError,
  } = useQuery({
    queryKey: ['wiki-page', wikiId, homePath],
    queryFn: () => wikiReaderApi.getPage(wikiId, homePath!),
    retry: 1,
    staleTime: 5 * 60 * 1000,
    enabled: !!homePath,
  })

  const tree = manifest ? buildPageTree(manifest) : undefined

  const isLoading = manifestLoading || (homePath ? homePageLoading : false)
  const error = manifestError || (homePath ? homePageError : null)

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
  } else if (homePageData) {
    body = <Markdown>{homePageData.content}</Markdown>
  } else if (tree) {
    body = (
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {tree.root.children?.map((node) => {
          if (node.separator) return null
          const Icon = getIcon(node.icon)
          return (
            <Link
              key={node.path}
              to="/wiki/$wikiId/$"
              params={{ wikiId, _splat: node.path }}
              className="group block rounded-lg border border-line bg-surface p-4 transition-colors hover:border-lagoon/30 hover:bg-surface-strong"
            >
              <div className="flex items-center gap-3">
                <Icon className="h-5 w-5 text-sea-ink-soft group-hover:text-lagoon" />
                <h3 className="text-lg font-semibold text-sea-ink group-hover:text-lagoon">
                  {node.title}
                </h3>
              </div>
              {node.description && (
                <p className="mt-2 text-sm text-sea-ink-soft">
                  {node.description}
                </p>
              )}
            </Link>
          )
        })}
      </div>
    )
  }

  return (
    <PasswordGate wikiId={wikiId}>
      <DocsPage wikiId={wikiId} tree={tree} pageData={homePageData}>
        {body}
      </DocsPage>
    </PasswordGate>
  )
}
