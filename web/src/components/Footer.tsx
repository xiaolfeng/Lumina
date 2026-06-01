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
    <footer className="border-t border-[var(--line)] bg-[var(--surface-strong)] py-8">
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
                className="group relative inline-flex items-center gap-1.5 pb-0.5 text-sm font-medium text-[var(--sea-ink)] no-underline"
              >
                <Github className="h-4 w-4" aria-hidden />
                {link.label}
                <span className="absolute bottom-0 left-1/2 h-[1.5px] w-0 -translate-x-1/2 bg-[var(--lagoon)] transition-all duration-250 group-hover:w-full group-hover:left-0 group-hover:translate-x-0" />
              </a>
            ) : (
              <Link
                key={link.to}
                to={link.to}
                className="group relative pb-0.5 text-sm font-medium text-[var(--sea-ink)] no-underline"
              >
                {link.label}
                <span className="absolute bottom-0 left-1/2 h-[1.5px] w-0 -translate-x-1/2 bg-[var(--lagoon)] transition-all duration-250 group-hover:w-full group-hover:left-0 group-hover:translate-x-0" />
              </Link>
            ),
          )}
        </nav>
      </div>
    </footer>
  )
}
