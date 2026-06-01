'use client'

import { Button } from '#/components/ui/button'
import { Link } from '@tanstack/react-router'
import { Sparkles, Menu, X } from 'lucide-react'
import { motion } from 'motion/react'
import { useState } from 'react'

const navLinks = [
  { label: '首页', to: '/' },
  { label: '文档', to: '/docs' },
  { label: 'API', to: '/api' },
] as const

export function Navbar() {
  const [open, setOpen] = useState(false)

  return (
    <motion.nav
      className="fixed top-4 left-4 right-4 z-50 flex items-center justify-between rounded-xl border border-[var(--line)] bg-[var(--surface)] px-6 py-3 shadow-[0_4px_24px_rgba(42,36,32,0.06)] backdrop-blur-md"
      aria-label="主导航"
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: 'easeOut' }}
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
        <span className="hidden items-center rounded-full border border-[var(--chip-line)] px-[0.7em] py-[0.25em] text-xs font-bold uppercase tracking-widest text-[var(--lagoon-deep)] sm:inline-flex"
          style={{ background: 'linear-gradient(to right, transparent, var(--chip-bg), transparent)' }}
        >
          微明
        </span>
      </Link>

      {/* ── Desktop nav links ── */}
      <div className="hidden items-center gap-8 md:flex">
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

      {/* ── Desktop CTA ── */}
      <div className="hidden md:block">
        <Button asChild variant="default" size="sm" className="!text-white">
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
        {open ? <X className="h-5 w-5" aria-hidden /> : <Menu className="h-5 w-5" aria-hidden />}
      </button>

      {/* ── Mobile menu panel ── */}
      {open && (
        <motion.div
          className="absolute top-full mt-2 left-0 right-0 flex flex-col gap-3 rounded-xl border border-[var(--line)] bg-[var(--surface)] p-5 shadow-[0_4px_24px_rgba(42,36,32,0.06)] backdrop-blur-md md:hidden"
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

          <Button asChild variant="default" size="sm" className="mt-1 w-full !text-white">
            <Link to="/login" onClick={() => setOpen(false)}>
              登录
            </Link>
          </Button>
        </motion.div>
      )}
    </motion.nav>
  )
}
