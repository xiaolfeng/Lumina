import {
  createFileRoute,
  Outlet,
  redirect,
  useLocation,
} from '@tanstack/react-router'
import { motion, AnimatePresence } from 'motion/react'
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from '@lumina/components/ui/sidebar'
import { AppSidebar } from '#/components/app-sidebar'
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

  return (
    <SidebarProvider>
      <a
        href="#console-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:bg-background focus:px-3 focus:py-2 focus:text-sm focus:text-sea-ink"
      >
        跳到主内容
      </a>
      <AppSidebar />

      <SidebarInset className="min-w-0">
        <motion.div
          className="flex h-16 shrink-0 items-center gap-2 px-4"
          initial="hidden"
          animate="visible"
          variants={headerVariants}
        >
          <SidebarTrigger className="-ml-1 transition-colors hover:text-lagoon" />
          <ConsoleBreadcrumb />
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
