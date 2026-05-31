'use client'

import { Button } from '#/components/ui/button'
import { Link } from '@tanstack/react-router'
import { Sparkles, Menu, X } from 'lucide-react'
import { useState } from 'react'

const navLinks = [
  { label: '首页', to: '/' },
  { label: '文档', to: '/docs' },
  { label: 'API', to: '/api' },
] as const

export function Navbar() {
  const [open, setOpen] = useState(false)

  return (
    <nav
      className="island-shell fixed top-4 left-4 right-4 z-50 flex items-center justify-between rounded-xl px-6 py-3 rise-in"
      aria-label="主导航"
    >
      {/* ── Brand ── */}
      <Link
        to="/"
        className="flex items-center gap-2 no-underline"
        aria-label="Lumina 首页"
      >
        <Sparkles
          className="h-5 w-5 shrink-0 text-[var(--lagoon)]"
          aria-hidden
        />
        <span className="display-title text-lg font-bold tracking-tight text-[var(--sea-ink)]">
          Lumina
        </span>
        <span className="island-kicker hidden sm:inline">微明</span>
      </Link>

      {/* ── Desktop nav links ── */}
      <div className="hidden items-center gap-8 md:flex">
        {navLinks.map((link) => (
          <Link
            key={link.to}
            to={link.to}
            className="nav-link text-sm font-medium"
            activeProps={{ className: 'is-active' }}
          >
            {link.label}
          </Link>
        ))}
      </div>

      {/* ── Desktop CTA ── */}
      <div className="hidden md:block">
        <Button asChild variant="default" size="sm">
          <Link to="/login">登录</Link>
        </Button>
      </div>

      {/* ── Mobile hamburger ── */}
      <button
        type="button"
        className="inline-flex items-center justify-center rounded-md p-2 text-[var(--sea-ink-soft)] hover:text-[var(--sea-ink)] md:hidden cursor-pointer"
        onClick={() => setOpen((v) => !v)}
        aria-label={open ? '关闭菜单' : '打开菜单'}
        aria-expanded={open}
      >
        {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
      </button>

      {/* ── Mobile menu panel ── */}
      {open && (
        <div className="absolute top-full mt-2 left-0 right-0 island-shell flex flex-col gap-3 rounded-xl p-5 md:hidden rise-in">
          {navLinks.map((link) => (
            <Link
              key={link.to}
              to={link.to}
              className="nav-link text-base font-medium"
              activeProps={{ className: 'is-active' }}
              onClick={() => setOpen(false)}
            >
              {link.label}
            </Link>
          ))}

          <Button asChild variant="default" size="sm" className="mt-1 w-full">
            <Link to="/login" onClick={() => setOpen(false)}>
              登录
            </Link>
          </Button>
        </div>
      )}
    </nav>
  )
}
