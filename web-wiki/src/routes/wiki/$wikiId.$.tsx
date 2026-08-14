import { createFileRoute, lazyRouteComponent } from '@tanstack/react-router'

export const Route = createFileRoute('/wiki/$wikiId/$')({
  component: lazyRouteComponent(() => import('./$wikiId/wiki-catchall-page')),
})
