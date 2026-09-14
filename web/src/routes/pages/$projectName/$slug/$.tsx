import { createFileRoute, useParams } from '@tanstack/react-router'
import { PagesPasswordGate } from '#/components/pages/password-gate'
import { ShowcaseShell } from '#/components/pages/showcase-shell'
import { usePageAuthCheck, usePageMeta, usePageVersions } from '#/hooks/usePages'

export const Route = createFileRoute('/pages/$projectName/$slug/$')({
  component: PagesShowcasePage,
})

function PagesShowcasePage() {
  const { projectName, slug } = useParams({ from: '/pages/$projectName/$slug/$' })
  const splat = useParams({ from: '/pages/$projectName/$slug/$' })
  const filepath = ((splat as { _splat?: string })._splat ?? '').replace(/^\/+/, '')
  const auth = usePageAuthCheck(projectName, slug)
  const unlocked =
    Boolean(auth.data?.data) &&
    (!auth.data?.data?.password_required || Boolean(auth.data?.data?.authenticated))
  const meta = usePageMeta(projectName, slug, undefined, unlocked)
  const versions = usePageVersions(meta.data?.data?.page.id)

  if (auth.isLoading || meta.isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-sea-ink-soft">
        加载中…
      </div>
    )
  }
  if (auth.data?.data?.password_required && !auth.data.data.authenticated) {
    return <PagesPasswordGate projectName={projectName} slug={slug} />
  }
  if (!meta.data?.data) {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-red-500">
        页面不存在
      </div>
    )
  }

  return (
    <ShowcaseShell
      projectName={projectName}
      slug={slug}
      filepath={filepath}
      page={meta.data.data.page}
      version={meta.data.data.version}
      files={meta.data.data.files}
      versions={versions.data?.data?.items ?? [meta.data.data.version]}
    />
  )
}
