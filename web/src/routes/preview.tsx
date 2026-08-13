import { Toaster } from '@lumina/components/ui/sonner'
import { createFileRoute, Outlet } from '@tanstack/react-router'
import { Eye, Sparkles } from 'lucide-react'

export const Route = createFileRoute('/preview')({
  component: PreviewLayout,
})

/* ─── Layout Component ─────────────────────────────────── */

function PreviewLayout() {
  return (
    <div className="flex h-screen flex-col bg-bg-base">
      {/* 顶部品牌栏 */}
      <header className="flex h-12 shrink-0 items-center gap-2 border-b border-line bg-header-bg px-4 backdrop-blur-md">
        <Sparkles className="size-4 text-lagoon" aria-hidden />
        <span className="display-title text-sm font-bold tracking-tight text-sea-ink">
          Lumina
        </span>
        <span className="hidden items-center gap-1.5 sm:inline-flex">
          <span className="h-px w-3 bg-lagoon/40" />
          <span className="text-[10px] font-semibold uppercase tracking-[0.15em] text-lagoon-deep">
            Preview
          </span>
        </span>

        <div className="flex-1" />

        <span className="inline-flex items-center gap-1.5 text-[11px] text-sea-ink-soft">
          <Eye className="size-3.5" aria-hidden />
          前端可视化预览
        </span>
      </header>

      {/* 主体：文件列表 + 预览区 */}
      <Outlet />

      <Toaster />
    </div>
  )
}
