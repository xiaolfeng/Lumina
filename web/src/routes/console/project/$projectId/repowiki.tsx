import { createFileRoute, Outlet } from '@tanstack/react-router'

export const Route = createFileRoute('/console/project/$projectId/repowiki')({
  staticData: { crumb: 'Wiki' },
  component: RepoWikiLayout,
})

function RepoWikiLayout() {
  return <Outlet />
}
