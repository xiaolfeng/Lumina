import {
  createFileRoute,
  Outlet,
  redirect,
  useLocation,
  useNavigate,
  useSearch,
} from '@tanstack/react-router'
import { motion, AnimatePresence } from 'motion/react'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from '@lumina/components/ui/sidebar'
import { AppSidebar } from '#/components/app-sidebar'
import { SettingsSidebar } from '#/components/settings/settings-sidebar'
import { ConsoleBreadcrumb } from '#/components/console-breadcrumb'
import { Toaster } from '@lumina/components/ui/sonner'
import Cookies from 'js-cookie'
import { ease } from '@lumina/components/motion'
import { getSafeRedirect } from '#/lib/apis/client'

export const Route = createFileRoute('/console')({
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
  component: ConsoleLayout,
})

const headerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { duration: 0.3, ease },
  },
}

function ConsoleLayout() {
  const location = useLocation()
  const navigate = useNavigate()
  const search = useSearch({ strict: false }) as unknown as { from?: string }

  const isSettingsRoute =
    location.pathname === '/console/settings' ||
    location.pathname.startsWith('/console/settings/') ||
    location.pathname === '/console/profile' ||
    location.pathname === '/console/apikey' ||
    location.pathname === '/console/ssh' ||
    location.pathname === '/console/workspace'

  const handleExitSettings = () => {
    if (search.from && search.from.startsWith('/console/')) {
      void navigate({ to: search.from })
    } else {
      void navigate({ to: '/console/dashboard' })
    }
  }

  return (
    <SidebarProvider>
      <a
        href="#console-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:bg-background focus:px-3 focus:py-2 focus:text-sm focus:text-sea-ink"
      >
        跳到主内容
      </a>
      {isSettingsRoute ? <SettingsSidebar /> : <AppSidebar />}

      <SidebarInset className="min-w-0">
        <motion.div
          className="flex h-16 shrink-0 items-center justify-between px-4"
          initial="hidden"
          animate="visible"
          variants={headerVariants}
        >
          <div className="flex items-center gap-2">
            <SidebarTrigger className="-ml-1 transition-colors hover:text-lagoon" />
            <ConsoleBreadcrumb />
          </div>

          {/* 🌟 设置中心顶栏右侧统一的退出返回按钮 */}
          {isSettingsRoute && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleExitSettings}
              className="flex items-center gap-2 rounded-none border-line bg-foam text-xs text-sea-ink hover:border-lagoon hover:bg-chip-bg hover:text-lagoon-deep"
              title="退出设置并返回"
            >
              <ArrowLeft className="size-3.5" />
              <span>{search.from ? '返回来源页面' : '退出设置返回看板'}</span>
            </Button>
          )}
        </motion.div>

        <div
          id="console-main"
          className="flex min-w-0 flex-1 flex-col gap-4 p-4 pt-0"
        >
          <AnimatePresence mode="wait">
            <motion.div
              key={location.pathname}
              className="min-w-0"
              initial={{ opacity: 0, x: 30 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -15 }}
              transition={{ duration: 0.3, ease }}
            >
              <Outlet />
            </motion.div>
          </AnimatePresence>
        </div>
      </SidebarInset>

      <Toaster />
    </SidebarProvider>
  )
}
