import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'
import Cookies from 'js-cookie'
import { getSafeRedirect } from '#/lib/apis/client'

export const Route = createFileRoute('/preview/$sessionHash')({
  beforeLoad: ({ location, params }) => {
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
    const rest = location.pathname.replace(`/preview/${params.sessionHash}`, '')
    if (rest === '' || rest === '/') {
      throw redirect({
        to: '/preview/$sessionHash/$',
        params: { sessionHash: params.sessionHash, _splat: 'index.html' },
      })
    }
  },
  component: () => <Outlet />,
})
