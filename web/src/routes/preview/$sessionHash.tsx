import {
  createFileRoute,
  Outlet,
  redirect,
  useMatches,
  useParams,
} from '@tanstack/react-router'
import Cookies from 'js-cookie'
import { PreviewWorkbenchPage } from '#/components/preview/workbench-page'
import { getSafeRedirect } from '#/lib/apis/client'

export const Route = createFileRoute('/preview/$sessionHash')({
  beforeLoad: ({ location }) => {
    const token = Cookies.get('access_token')
    const refreshToken = Cookies.get('refresh_token')
    if (!token && !refreshToken) {
      throw redirect({
        to: '/auth/login',
        search: {
          redirect: getSafeRedirect(location.href, '/console/dashboard'),
        },
      })
    }
  },
  component: PreviewHashRoute,
})

function PreviewHashRoute() {
  const { sessionHash } = useParams({ from: '/preview/$sessionHash' })
  const matches = useMatches()
  if (matches.some((match) => match.routeId === '/preview/$sessionHash/$')) {
    return <Outlet />
  }
  return <PreviewWorkbenchPage sessionHash={sessionHash} requestedFile="" />
}
