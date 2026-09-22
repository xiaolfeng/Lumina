import { useState } from 'react'
import {
  BookOpen,
  Boxes,
  Calendar,
  ChevronDown,
  ChevronUp,
  CircleDot,
  Code,
  Compass,
  Cpu,
  Database,
  FileCode,
  FileText,
  Folder,
  Layers,
  Lightbulb,
  Network,
  Settings,
  Shield,
  ShieldCheck,
  Sparkles,
  Tag,
  Terminal,
  Workflow,
} from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import type { MdxFrontmatter } from './frontmatter'

const ICON_MAP: Record<string, LucideIcon> = {
  filetext: FileText,
  bookopen: BookOpen,
  shield: Shield,
  shieldcheck: ShieldCheck,
  code: Code,
  filecode: FileCode,
  cpu: Cpu,
  database: Database,
  folder: Folder,
  layers: Layers,
  lightbulb: Lightbulb,
  network: Network,
  settings: Settings,
  terminal: Terminal,
  compass: Compass,
  sparkles: Sparkles,
  workflow: Workflow,
  boxes: Boxes,
  circledot: CircleDot,
}

function resolveIcon(iconName?: string): LucideIcon {
  if (!iconName) return FileText
  const normalized = iconName.toLowerCase().replace(/[-_]/g, '')
  return ICON_MAP[normalized] ?? FileText
}

export interface MdxHeaderProps {
  frontmatter: MdxFrontmatter | null
  fallbackTitle: string
  className?: string
}

export function MdxHeader({
  frontmatter,
  fallbackTitle,
  className = '',
}: MdxHeaderProps) {
  const [showMeta, setShowMeta] = useState(false)

  const title = frontmatter?.title || fallbackTitle
  const description = frontmatter?.description
  const tags = frontmatter?.tags
  const date = frontmatter?.date || frontmatter?.last_updated
  const IconComponent = resolveIcon(frontmatter?.icon)

  // 提取自定义的额外键值对
  const customMeta: Record<string, unknown> = {}
  if (frontmatter) {
    const knownKeys = new Set([
      'title',
      'description',
      'icon',
      'tags',
      'date',
      'last_updated',
    ])
    for (const [key, value] of Object.entries(frontmatter)) {
      if (!knownKeys.has(key) && value !== undefined && value !== null) {
        customMeta[key] = value
      }
    }
  }

  const hasCustomMeta = Object.keys(customMeta).length > 0

  return (
    <header
      className={`border-b border-line pb-6 mb-8 text-sea-ink ${className}`}
    >
      <div className="flex items-start gap-3.5 mb-3">
        <div className="flex size-10 shrink-0 items-center justify-center rounded-none border border-line bg-surface-muted/50 text-lagoon-deep shadow-xs">
          <IconComponent className="size-5" aria-hidden />
        </div>
        <div className="min-w-0 flex-1">
          <h1 className="display-title text-2xl font-semibold tracking-tight sm:text-3xl text-sea-ink break-words">
            {title}
          </h1>
        </div>
      </div>

      {description && (
        <p className="text-base text-sea-ink-soft leading-relaxed mb-4 max-w-3xl">
          {description}
        </p>
      )}

      <div className="flex flex-wrap items-center gap-3 text-xs text-sea-ink-soft/70">
        {date && (
          <div className="flex items-center gap-1.5">
            <Calendar className="size-3.5" aria-hidden />
            <span>{String(date)}</span>
          </div>
        )}

        {Array.isArray(tags) && tags.length > 0 && (
          <div className="flex flex-wrap items-center gap-1.5">
            <Tag className="size-3.5 text-sea-ink-soft/50" aria-hidden />
            {tags.map((tag, idx) => (
              <span
                key={`${tag}-${idx}`}
                className="inline-flex items-center border border-line bg-sand/60 px-2 py-0.5 text-[11px] font-medium text-sea-ink-soft"
              >
                {tag}
              </span>
            ))}
          </div>
        )}

        {hasCustomMeta && (
          <button
            type="button"
            onClick={() => setShowMeta((prev) => !prev)}
            className="flex items-center gap-1 text-[11px] font-medium text-lagoon-deep hover:underline cursor-pointer ml-auto"
          >
            <span>{showMeta ? '收起元数据' : '查看元数据'}</span>
            {showMeta ? (
              <ChevronUp className="size-3" />
            ) : (
              <ChevronDown className="size-3" />
            )}
          </button>
        )}
      </div>

      {hasCustomMeta && showMeta && (
        <div className="mt-4 border border-line bg-surface-muted/30 p-3 text-xs font-mono">
          <p className="text-[10px] uppercase font-semibold text-sea-ink-soft/60 mb-1.5">
            Frontmatter 属性
          </p>
          <div className="space-y-1">
            {Object.entries(customMeta).map(([k, v]) => (
              <div key={k} className="flex gap-2">
                <span className="text-lagoon-deep select-all">{k}:</span>
                <span className="text-sea-ink select-all">
                  {typeof v === 'object' ? JSON.stringify(v) : String(v)}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </header>
  )
}
