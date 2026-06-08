import { createFileRoute, Link, Outlet, redirect } from '@tanstack/react-router'
import { useLocation } from '@tanstack/react-router'
import { motion, AnimatePresence } from 'motion/react'
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from '#/components/ui/sidebar'
import { AppSidebar } from '#/components/app-sidebar'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
} from '#/components/ui/breadcrumb'
import { Toaster } from '#/components/ui/sonner'
import Cookies from 'js-cookie'
import { ease } from '#/lib/motion'

export const Route = createFileRoute('/console')({
  beforeLoad: ({ location }) => {
    const token = Cookies.get('access_token')
    if (!token) {
      throw redirect({
        to: '/auth/login',
        search: { redirect: location.href },
      })
    }
  },
  component: ConsoleLayout,
})

/** Header 入场动画 */
const headerVariants = {
  hidden: { opacity: 0, y: -8 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4, ease, delay: 0.25 },
  },
}

function ConsoleLayout() {
  const location = useLocation()

  return (
    <SidebarProvider>
      {/* Sidebar 自身入场 — 动画在 AppSidebar 内部控制 */}
      <AppSidebar />

      {/* Main 区域 */}
      <SidebarInset>
        {/* Header 入场 */}
        <motion.div
          className="flex h-16 shrink-0 items-center gap-2 px-4"
          initial="hidden"
          animate="visible"
          variants={headerVariants}
        >
          <SidebarTrigger className="-ml-1" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbLink asChild>
                  <Link to="/console/dashboard">Console</Link>
                </BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </motion.div>

        {/* 页面切换动画 */}
        <main className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <AnimatePresence mode="wait">
            <motion.div
              key={location.pathname}
              initial={{ opacity: 0, x: 30 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -15 }}
              transition={{ duration: 0.3, ease }}
            >
              <Outlet />
            </motion.div>
          </AnimatePresence>
        </main>
      </SidebarInset>

      <Toaster />
    </SidebarProvider>
  )
}
