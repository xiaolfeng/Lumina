import { useEffect, useState } from 'react'
import { Toaster } from '@lumina/components/ui/sonner'
import { createFileRoute, Outlet } from '@tanstack/react-router'
import { PreviewBrandHeader } from '#/components/preview/brand-header'
import { PreviewHeaderContext } from '#/hooks/usePreviewHeader'

export const Route = createFileRoute('/preview')({
  component: PreviewLayout,
})

/* ─── Layout Component ─────────────────────────────────── */

function PreviewLayout() {
  const [title, setTitle] = useState<string | null>(null)

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
    <PreviewHeaderContext.Provider value={{ title, setTitle }}>
      <div className="flex h-screen flex-col bg-bg-base">
        <PreviewBrandHeader title={title} />

        {/* 主体：文件列表 + 预览区 */}
        <Outlet />

        <Toaster />
      </div>
    </PreviewHeaderContext.Provider>
  )
}
