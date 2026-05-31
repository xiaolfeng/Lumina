import { Link } from '@tanstack/react-router'
import { Github } from 'lucide-react'

const footerLinks = [
  {
    label: 'GitHub',
    href: 'https://github.com/xiao-lfeng/Lumina',
    external: true,
  },
  { label: '文档', to: '/docs' },
  { label: 'API', to: '/api' },
] as const

export function Footer() {
  return (
    <footer className="site-footer py-8">
      <div className="page-wrap flex flex-col items-center gap-4 text-center">
        {/* ── Brand + copyright ── */}
        <p className="text-sm text-[var(--sea-ink-soft)]">
          <span className="display-title font-semibold text-[var(--sea-ink)]">
            Lumina · 微明
          </span>
          {' — © 2026 Xiao Lfeng (筱锋)'}
        </p>

        {/* ── Links row ── */}
        <nav aria-label="页脚导航" className="flex items-center gap-5">
          {footerLinks.map((link) =>
            'href' in link ? (
              <a
                key={link.href}
                href={link.href}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 nav-link text-sm font-medium"
              >
                <Github className="h-4 w-4" aria-hidden />
                {link.label}
              </a>
            ) : (
              <Link
                key={link.to}
                to={link.to}
                className="nav-link text-sm font-medium"
              >
                {link.label}
              </Link>
            ),
          )}
        </nav>
      </div>
    </footer>
  )
}
