import { Badge } from '@lumina/components/ui/badge'
import React, { useMemo } from 'react'
import { LpwBlockRenderer } from './lpw-block-renderer'
import { parseLpwSource } from './lpw-parser'

export interface LpwDocumentViewerProps {
  source: string
}

export const LpwDocumentViewer: React.FC<LpwDocumentViewerProps> = ({
  source,
}) => {
  const { document: doc, error } = useMemo(
    () => parseLpwSource(source),
    [source],
  )

  if (error) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div
          data-testid="document-error"
          className="border border-destructive/40 bg-destructive/5 p-6 text-xs text-destructive shadow-sm"
        >
          <div className="flex items-center gap-2 font-mono font-semibold uppercase tracking-wider text-sm">
            <span className="inline-block w-2 h-2 rounded-full bg-destructive" />
            <span>LPW 文档解析失败</span>
            {error.path && (
              <span className="font-mono text-sea-ink-soft/70">
                [{error.path}]
              </span>
            )}
          </div>
          <p className="mt-2.5 text-sea-ink-soft font-sans leading-relaxed text-sm">
            {error.message}
          </p>
        </div>
      </div>
    )
  }

  const { meta, blocks } = doc

  return (
    <div className="max-w-4xl mx-auto my-6 bg-surface-strong border border-line shadow-sm relative text-sea-ink font-sans">
      {/* 顶部古典装订双规线 */}
      <div
        aria-hidden="true"
        className="h-1.5 border-b border-sea-ink/80 bg-surface-strong relative before:content-[''] before:absolute before:inset-x-0 before:top-[1px] before:h-[1px] before:bg-sea-ink/80"
      />

      {meta && (
        <header className="px-6 sm:px-10 pt-10 pb-8 border-b border-line bg-surface/40">
          {/* 出版系列刊头栏 */}
          <div className="flex flex-wrap justify-between items-center text-[11px] font-mono tracking-widest text-sea-ink-soft uppercase mb-5 pb-3 border-b border-dashed border-line/60 gap-2">
            <div className="flex items-center gap-2">
              <span className="inline-block w-1.5 h-1.5 rounded-full bg-lagoon" />
              <span>LUMINA MONOGRAPH // 架构白皮书</span>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-lagoon font-bold">KNOWLEDGE BLUEPRINT</span>
              <span className="text-sea-ink-soft/50">·</span>
              <span>VERIFIED CANON</span>
            </div>
          </div>

          {/* 标题 */}
          <h1 className="font-serif text-2xl sm:text-3xl lg:text-4xl font-semibold tracking-tight text-sea-ink leading-snug mb-3.5">
            {meta.title}
          </h1>

          {/* 描述与引言 */}
          {meta.description && (
            <p className="text-sm sm:text-base text-sea-ink-soft leading-relaxed italic font-serif max-w-3xl">
              {meta.description}
            </p>
          )}

          {/* 标签栏 */}
          {meta.tags && meta.tags.length > 0 && (
            <div className="flex flex-wrap gap-1.5 pt-4">
              {meta.tags.map((tag) => (
                <Badge
                  key={tag}
                  variant="secondary"
                  className="text-[11px] font-mono uppercase px-2.5 py-0.5 border border-line/60 bg-surface-muted text-sea-ink-soft hover:bg-surface-muted"
                >
                  {tag}
                </Badge>
              ))}
            </div>
          )}

          {/* 作者与版本印记 */}
          <div className="mt-6 pt-3.5 border-t border-line flex flex-wrap justify-between items-center text-xs text-sea-ink-soft/80 font-mono gap-3">
            <div className="flex items-center gap-4">
              {meta.author && (
                <span>
                  ARCHITECT:{' '}
                  <strong className="text-sea-ink font-semibold">
                    {meta.author}
                  </strong>
                </span>
              )}
              {meta.version && (
                <span className="bg-surface-muted border border-line/60 px-2 py-0.5">
                  v{meta.version}
                </span>
              )}
            </div>
            <div className="flex items-center gap-2 text-[11px]">
              <span className="inline-block w-1.5 h-1.5 rounded-full bg-kicker" />
              <span className="uppercase tracking-wider">
                IMPRIMATUR: ACTIVE SPEC
              </span>
            </div>
          </div>
        </header>
      )}

      {/* 正文区域 */}
      <div className="px-6 sm:px-10 py-8">
        {blocks.length === 0 ? (
          <div
            data-testid="document-empty"
            className="border-2 border-dashed border-line/50 p-12 text-center text-xs font-mono text-sea-ink-soft/70 bg-surface/30 my-6"
          >
            <div className="font-serif italic text-base text-sea-ink-soft mb-1">
              空白典籍 · 等待分块生息
            </div>
            空文档：等待分块写入
          </div>
        ) : (
          <main className="space-y-6">
            {blocks.map((block) => (
              <LpwBlockRenderer key={block.id} block={block} depth={1} />
            ))}
          </main>
        )}
      </div>

      {/* 底部出版物印鉴尾栏 */}
      <footer className="px-6 sm:px-10 py-4 border-t border-line bg-surface/30 flex flex-wrap justify-between items-center text-[11px] font-mono text-sea-ink-soft/70 gap-2">
        <span>LUMINA ARCHITECTURE PRESS · THE LUMINOUS EDITORIAL</span>
        <span>TYPESET IN NEW-YORK SERIF & MONO</span>
      </footer>
    </div>
  )
}
