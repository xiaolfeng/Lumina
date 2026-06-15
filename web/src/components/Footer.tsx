import { Github } from 'lucide-react'

const GITHUB_URL = 'https://github.com/xiaolfeng/Lumina'

export function Footer() {
  return (
    <footer className="border-t border-line bg-surface-strong py-8">
      <div className="page-wrap flex flex-col items-center gap-4 text-center">
        <p className="text-sm text-sea-ink-soft">
          <span className="display-title font-semibold text-sea-ink">
            Lumina · 微明
          </span>
          {' — © 2026 Xiao Lfeng (筱锋)'}
        </p>

        <nav aria-label="页脚导航" className="flex items-center gap-5">
          <a
            href={GITHUB_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="group relative inline-flex items-center gap-1.5 pb-0.5 text-sm font-medium text-sea-ink no-underline"
          >
            <Github className="h-4 w-4" aria-hidden />
            GitHub
            <span className="absolute bottom-0 left-1/2 h-[1.5px] w-0 -translate-x-1/2 bg-lagoon transition-all duration-250 group-hover:w-full group-hover:left-0 group-hover:translate-x-0" />
          </a>
        </nav>
      </div>
    </footer>
  )
}
