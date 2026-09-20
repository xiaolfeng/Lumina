import { createFileRoute, Outlet } from '@tanstack/react-router'
import { Toaster } from '@lumina/components/ui/sonner'

export const Route = createFileRoute('/pages')({
  component: PagesLayout,
})

function PagesLayout() {
  return (
    <div className="min-h-screen bg-sand">
      <Outlet />
      <Toaster />
    </div>
  )
}
