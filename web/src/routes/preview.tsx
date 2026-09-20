import { useEffect, useState } from 'react'
import { Toaster } from '@lumina/components/ui/sonner'
import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'
import Cookies from 'js-cookie'
import { PreviewBrandHeader } from '#/components/preview/brand-header'
import type { PreviewHeaderControls } from '#/hooks/usePreviewHeader'
import { PreviewHeaderContext } from '#/hooks/usePreviewHeader'
import { getSafeRedirect } from '#/lib/apis/client'

export const Route = createFileRoute('/preview')({
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
  component: PreviewLayout,
})

/* ─── Layout Component ─────────────────────────────────── */

function PreviewLayout() {
  const [title, setTitle] = useState<string | null>(null)
  const [controls, setControls] = useState<PreviewHeaderControls | null>(null)

  useEffect(() => {
    const previous = document.title
    return () => {
      document.title = previous
    }
  }, [])

  useEffect(() => {
    if (title === null) {
      document.title = 'Preview | Lumina'
      return
    }
    const name = title.trim() === '' ? '未命名预览' : title
    document.title = `${name} | Lumina`
  }, [title])

  return (
    <PreviewHeaderContext.Provider
      value={{ title, setTitle, controls, setControls }}
    >
      <div className="flex h-screen flex-col bg-bg-base">
        <PreviewBrandHeader title={title} controls={controls} />

        {/* 主体：文件列表 + 预览区 */}
        <Outlet />

        <Toaster />
      </div>
    </PreviewHeaderContext.Provider>
  )
}
