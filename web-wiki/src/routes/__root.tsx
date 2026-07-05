import { Link, Outlet, createRootRoute } from '@tanstack/react-router'
import { BookOpen } from 'lucide-react'

export const Route = createRootRoute({
  component: RootComponent,
})

function RootComponent() {
  return (
    <div className="min-h-screen bg-bg-base">
      {/* 文档站顶部导航 */}
      <header className="sticky top-0 z-50 border-b border-line bg-header-bg/80 backdrop-blur-md">
        <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4">
          <Link
            to="/wiki/$wikiId"
            params={{ wikiId: 'default' }}
            className="flex items-center gap-2 font-semibold"
          >
            <BookOpen className="h-5 w-5 text-lagoon" />
            <span className="display-title text-lg">Lumina Wiki</span>
          </Link>
        </div>
      </header>

      {/* 页面内容 */}
      <main>
        <Outlet />
      </main>
    </div>
  )
}
