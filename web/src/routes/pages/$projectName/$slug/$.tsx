import { createFileRoute, useParams } from '@tanstack/react-router'
import { PagesShowcasePage } from '#/components/pages/showcase-page'

export const Route = createFileRoute('/pages/$projectName/$slug/$')({
  component: PagesSplatRoute,
})

function PagesSplatRoute() {
  const params = useParams({ from: '/pages/$projectName/$slug/$' })
  const filepath = ((params as { _splat?: string })._splat ?? '').replace(/^\/+/, '')
  return (
    <PagesShowcasePage
      projectName={params.projectName}
      slug={params.slug}
      filepath={filepath}
    />
  )
}
