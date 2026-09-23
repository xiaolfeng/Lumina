import {
  createFileRoute,
  Outlet,
  redirect,
  useLocation,
  useNavigate,
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
  const isSettingsRoute =
    location.pathname === '/console/settings' ||
    location.pathname.startsWith('/console/settings/')

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

          {/* 🌟 系统设置下顶栏右侧集中唯一的退出返回按钮 */}
          {isSettingsRoute && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => void navigate({ to: '/console/dashboard' })}
              className="flex items-center gap-2 rounded-none border-line bg-foam text-xs text-sea-ink hover:border-lagoon hover:bg-chip-bg hover:text-lagoon-deep"
              title="退出系统设置并返回控制台主看板 (ESC)"
            >
              <ArrowLeft className="size-3.5" />
              <span>退出设置返回看板</span>
              <kbd className="font-mono text-[9.5px] bg-sand border border-line px-1 py-0.5 text-sea-ink-soft">
                ESC
              </kbd>
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
