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
          className="border border-red-500/40 bg-red-500/5 p-4 text-xs text-red-600"
        >
          <div className="flex items-center gap-2 font-semibold">
            <span>LPW 文档解析失败</span>
            {error.path && (
              <span className="font-mono text-sea-ink-soft/70">
                [{error.path}]
              </span>
            )}
          </div>
          <p className="mt-2 text-sea-ink-soft">{error.message}</p>
        </div>
      </div>
    )
  }

  const { meta, blocks } = doc

  return (
    <div className="max-w-4xl mx-auto px-4 py-8 space-y-6">
      {meta && (
        <header className="border-b border-line pb-4 space-y-2">
          <div className="flex items-center justify-between gap-4">
            <h1 className="text-xl font-semibold text-sea-ink">{meta.title}</h1>
            <div className="flex items-center gap-2 text-xs text-sea-ink-soft/70">
              {meta.author && <span>{meta.author}</span>}
              {meta.version && (
                <span className="font-mono bg-surface-muted px-1.5 py-0.5">
                  v{meta.version}
                </span>
              )}
            </div>
          </div>
          {meta.description && (
            <p className="text-sm text-sea-ink-soft leading-relaxed">
              {meta.description}
            </p>
          )}
          {meta.tags && meta.tags.length > 0 && (
            <div className="flex flex-wrap gap-1.5 pt-1">
              {meta.tags.map((tag) => (
                <Badge key={tag} variant="secondary" className="text-xs">
                  {tag}
                </Badge>
              ))}
            </div>
          )}
        </header>
      )}

      {blocks.length === 0 ? (
        <div
          data-testid="document-empty"
          className="border border-dashed border-line/60 py-12 text-center text-xs text-sea-ink-soft/60"
        >
          空文档：等待分块写入
        </div>
      ) : (
        <main className="space-y-4">
          {blocks.map((block) => (
            <LpwBlockRenderer key={block.id} block={block} depth={1} />
          ))}
        </main>
      )}
    </div>
  )
}
