import {
  createFileRoute,
  Outlet,
  useMatches,
  useParams,
} from '@tanstack/react-router'
import { PagesShowcasePage } from '#/components/pages/showcase-page'

export const Route = createFileRoute('/pages/$projectName/$slug')({
  // 无文件路径时不再重定向到 index.html：直接渲染展示页，入口由 version.entry_filename 兜底
  component: PagesSlugRoute,
})

function PagesSlugRoute() {
  const { projectName, slug } = useParams({ from: '/pages/$projectName/$slug' })
  const matches = useMatches()
  if (
    matches.some((match) => match.routeId === '/pages/$projectName/$slug/$')
  ) {
    return <Outlet />
  }
  return <PagesShowcasePage projectName={projectName} slug={slug} filepath="" />
}
