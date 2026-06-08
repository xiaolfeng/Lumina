import { Link, useLocation } from '@tanstack/react-router'
import { motion } from 'motion/react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '#/components/ui/sidebar'
import { Avatar, AvatarFallback } from '#/components/ui/avatar'
import {
  LayoutDashboard,
  KeyRound,
  MessageCircle,
  Sparkles,
  ExternalLink,
} from 'lucide-react'
import { ease } from '#/lib/motion'

const navItems = [
  {
    title: '看板',
    to: '/console/dashboard',
    icon: LayoutDashboard,
  },
  {
    title: '令牌管理',
    to: '/console/apikey',
    icon: KeyRound,
  },
  {
    title: '交互问答',
    to: '/faq',
    icon: MessageCircle,
    external: true,
  },
]

export function AppSidebar() {
  const location = useLocation()

  return (
    <Sidebar variant="inset">
      <motion.div
        className="flex h-full flex-col"
        initial={{ opacity: 0, x: -24 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ duration: 0.5, ease }}
      >
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild>
                <Link to="/console/dashboard">
                  <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                    <Sparkles className="size-4" />
                  </div>
                  <div className="flex flex-col gap-0.5 leading-none">
                    <span className="font-semibold">微明 Lumina</span>
                    <span className="text-xs text-muted-foreground">
                      管理后台
                    </span>
                  </div>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupLabel>导航</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {navItems.map((item) => {
                  const isActive =
                    location.pathname === item.to ||
                    location.pathname.startsWith(item.to + '/')
                  return (
                    <SidebarMenuItem key={item.to}>
                      <SidebarMenuButton
                        asChild
                        isActive={isActive}
                        tooltip={item.title}
                      >
                        {item.external ? (
                          <a
                            href={item.to}
                            target="_blank"
                            rel="noopener noreferrer"
                          >
                            <item.icon />
                            <span>{item.title}</span>
                            <ExternalLink className="ml-auto size-3.5 text-muted-foreground" />
                          </a>
                        ) : (
                          <Link to={item.to}>
                            <item.icon />
                            <span>{item.title}</span>
                          </Link>
                        )}
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg">
                <Avatar className="size-8 rounded-lg">
                  <AvatarFallback className="rounded-lg">管</AvatarFallback>
                </Avatar>
                <div className="flex flex-col gap-0.5 leading-none">
                  <span className="text-sm font-medium">管理员</span>
                  <span className="text-xs text-muted-foreground">
                    Lumina Console
                  </span>
                </div>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarFooter>
      </motion.div>
    </Sidebar>
  )
}
