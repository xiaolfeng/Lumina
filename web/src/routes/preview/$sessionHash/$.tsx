import { createFileRoute, redirect, useParams } from '@tanstack/react-router'
import Cookies from 'js-cookie'

import { PreviewWorkbenchPage } from '#/components/preview/workbench-page'
import { getSafeRedirect } from '#/lib/apis/client'

export const Route = createFileRoute('/preview/$sessionHash/$')({
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
  component: PreviewSplatPage,
})

function PreviewSplatPage() {
  const { sessionHash } = useParams({ from: '/preview/$sessionHash/$' })
  const splat = useParams({ from: '/preview/$sessionHash/$' })
  const requestedFile = ((splat as { _splat?: string })._splat ?? '').replace(
    /^\/+/,
    '',
  )
  return (
    <PreviewWorkbenchPage
      sessionHash={sessionHash}
      requestedFile={requestedFile}
    />
  )
}
