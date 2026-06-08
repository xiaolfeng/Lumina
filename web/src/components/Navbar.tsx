'use client'

import { Button } from '#/components/ui/button'
import { Link, useRouter } from '@tanstack/react-router'
import { Github, LayoutDashboard, LogOut, Menu, Sparkles, X } from 'lucide-react'
import { motion } from 'motion/react'
import { useEffect, useState } from 'react'
import { getCookie, useAuth } from '#/hooks/useAuth'

const navLinks = [
  { label: '首页', to: '/' },
  { label: '开始', to: '/start' },
] as const

export function Navbar() {
  const [open, setOpen] = useState(false)
  const [scrolled, setScrolled] = useState(false)
  const { isAuthenticated, logout } = useAuth()
  const router = useRouter()

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 16)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  return (
    <motion.header
      className="fixed top-0 left-0 right-0 z-50"
      aria-label="主导航"
      initial={{ opacity: 0, y: -8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: 'easeOut' }}
    >
      <nav
        className={
          'mx-auto grid max-w-6xl grid-cols-12 items-center gap-4 border transition-all duration-300 ' +
          (scrolled
            ? 'mt-4 rounded-xl border-[var(--line)] bg-[var(--surface)] px-6 py-3 shadow-[0_4px_24px_rgba(42,36,32,0.06)] backdrop-blur-md'
            : 'border-transparent bg-transparent px-4 py-4 md:px-6 md:py-5')
        }
      >
        {/* ── Brand (col-3) ── */}
        <div className="col-span-6 md:col-span-3">
          <Link
            to="/"
            className="inline-flex items-center gap-2 no-underline"
            aria-label="Lumina 首页"
          >
            <Sparkles
              className="h-5 w-5 shrink-0 text-[var(--lagoon)]"
              aria-hidden
            />
            <span className="display-title text-lg font-bold tracking-tight text-[var(--sea-ink)]">
              Lumina
            </span>
            <span className="hidden items-center gap-1.5 sm:inline-flex">
              <span className="h-px w-3 bg-[var(--lagoon)]/40" />
              <span className="text-[10px] font-semibold uppercase tracking-[0.15em] text-[var(--lagoon-deep)]">
                微明
              </span>
            </span>
          </Link>
        </div>

        {/* ── Desktop nav links (col-6, centered) ── */}
        <div className="col-span-6 hidden items-center justify-center gap-8 md:flex">
          {navLinks.map((link) => (
            <Link
              key={link.to}
              to={link.to}
              className="group relative pb-0.5 text-sm font-medium text-[var(--sea-ink)] no-underline"
              activeProps={{ className: 'is-active' }}
            >
              {link.label}
              <span className="absolute bottom-0 left-1/2 h-[1.5px] w-0 -translate-x-1/2 bg-[var(--lagoon)] transition-all duration-250 group-hover:w-full group-hover:left-0 group-hover:translate-x-0 [.is-active_&]:w-full [.is-active_&]:left-0 [.is-active_&]:translate-x-0" />
            </Link>
          ))}
        </div>

        {/* ── Desktop CTA (col-3, right-aligned) ── */}
        <div className="col-span-3 hidden items-center justify-end gap-3 md:flex">
          <a
            href="https://github.com/xiaolfeng/Lumina"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex h-9 w-9 items-center justify-center rounded-lg text-[var(--sea-ink-soft)] transition-colors duration-200 hover:bg-[var(--lagoon)]/10 hover:text-[var(--lagoon-deep)] cursor-pointer"
            aria-label="GitHub 仓库"
          >
            <Github className="h-[1.15rem] w-[1.15rem]" aria-hidden />
          </a>
          {isAuthenticated ? (
            <>
              <Button asChild variant="outline" size="sm">
                <Link to="/console/dashboard">
                  <LayoutDashboard className="mr-2 h-4 w-4" />
                  控制台
                </Link>
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="text-[var(--sea-ink-soft)]"
                onClick={() =>
                  logout.mutate(
                    { refresh_token: getCookie('refresh_token') || '' },
                    {
                      onSuccess: () => router.navigate({ to: '/auth/login' }),
                    },
                  )
                }
              >
                <LogOut className="mr-2 h-4 w-4" />
                登出
              </Button>
            </>
          ) : (
            <Button asChild variant="default" size="sm" className="!text-white">
              <Link to="/auth/login">登录</Link>
            </Button>
          )}
        </div>

        {/* ── Mobile hamburger (col-6, right-aligned) ── */}
        <div className="col-span-6 flex items-center justify-end md:hidden">
          <button
            type="button"
            className="inline-flex items-center justify-center rounded-md p-2 text-[var(--sea-ink-soft)] hover:text-[var(--sea-ink)] cursor-pointer"
            onClick={() => setOpen((v) => !v)}
            aria-label={open ? '关闭菜单' : '打开菜单'}
            aria-expanded={open}
          >
            {open ? (
              <X className="h-5 w-5" aria-hidden />
            ) : (
              <Menu className="h-5 w-5" aria-hidden />
            )}
          </button>
        </div>

        {/* ── Mobile menu panel ── */}
        {open && (
          <motion.div
            className="col-span-12 flex flex-col gap-3 rounded-xl border border-[var(--line)] bg-[var(--surface)] p-5 shadow-[0_4px_24px_rgba(42,36,32,0.06)] backdrop-blur-md md:hidden"
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, ease: 'easeOut' }}
          >
            {navLinks.map((link) => (
              <Link
                key={link.to}
                to={link.to}
                className="text-base font-medium text-[var(--sea-ink)] no-underline"
                activeProps={{ className: 'is-active' }}
                onClick={() => setOpen(false)}
              >
                {link.label}
              </Link>
            ))}

            <a
              href="https://github.com/xiaolfeng/Lumina"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 text-base font-medium text-[var(--sea-ink)] no-underline"
              onClick={() => setOpen(false)}
            >
              <Github className="h-[1.15rem] w-[1.15rem]" aria-hidden />
              GitHub
            </a>

            {isAuthenticated ? (
              <>
                <Button asChild variant="outline" size="sm" className="w-full">
                  <Link to="/console/dashboard" onClick={() => setOpen(false)}>
                    <LayoutDashboard className="mr-2 h-4 w-4" />
                    控制台
                  </Link>
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full text-[var(--sea-ink-soft)]"
                  onClick={() =>
                    logout.mutate(
                      { refresh_token: getCookie('refresh_token') || '' },
                      {
                        onSuccess: () => {
                          setOpen(false)
                          router.navigate({ to: '/auth/login' })
                        },
                      },
                    )
                  }
                >
                  <LogOut className="mr-2 h-4 w-4" />
                  登出
                </Button>
              </>
            ) : (
              <Button
                asChild
                variant="default"
                size="sm"
                className="mt-1 w-full !text-white"
              >
                <Link to="/auth/login" onClick={() => setOpen(false)}>
                  登录
                </Link>
              </Button>
            )}
          </motion.div>
        )}
      </nav>
    </motion.header>
  )
}
