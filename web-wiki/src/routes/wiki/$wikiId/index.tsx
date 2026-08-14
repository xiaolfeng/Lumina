import { createFileRoute, lazyRouteComponent } from '@tanstack/react-router'

export const Route = createFileRoute('/wiki/$wikiId/')({
  component: lazyRouteComponent(() => import('./wiki-index-page')),
})
