import { useState } from 'react'
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
