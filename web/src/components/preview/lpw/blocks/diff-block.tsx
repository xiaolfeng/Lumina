import { Badge } from '@lumina/components/ui/badge'
import React, { useState } from 'react'
import ReactDiffViewer from 'react-diff-viewer-continued'
import type { LpwBlockSlotProps, LpwDiffProps } from '../types'

export const DiffBlock: React.FC<LpwBlockSlotProps<LpwDiffProps>> = ({
  props,
}) => {
  const [splitView, setSplitView] = useState<boolean>(props.splitView ?? true)

  const isIdentical = props.oldCode === props.newCode
  const showHeader = Boolean(props.filename || props.language || true)

  return (
    <div
      data-testid="diff-block"
      className="my-4 border border-line bg-surface text-xs"
    >
      {showHeader && (
        <div className="flex items-center justify-between border-b border-line bg-surface-muted/30 px-3 py-1.5">
          <div className="flex items-center gap-2">
            {props.filename && (
              <span className="font-mono text-sea-ink-soft">
                {props.filename}
              </span>
            )}
            {props.language && (
              <Badge
                variant="outline"
                className="font-mono text-[10px] uppercase"
              >
                {props.language}
              </Badge>
            )}
          </div>
          <button
            type="button"
            onClick={() => setSplitView((v) => !v)}
            className="cursor-pointer border border-line bg-surface px-2 py-0.5 text-[11px] text-sea-ink hover:bg-surface-muted"
          >
            {splitView ? '切换为统一视图' : '切换为并排视图'}
          </button>
        </div>
      )}

      {isIdentical ? (
        <div className="p-8 text-center text-xs text-sea-ink-soft/60">
          文件内容一致，无差异
        </div>
      ) : (
        <div className="overflow-x-auto text-xs font-mono">
          <ReactDiffViewer
            oldValue={props.oldCode}
            newValue={props.newCode}
            splitView={splitView}
            useDarkTheme={false}
            styles={{
              variables: {
                light: {
                  diffViewerBackground: 'var(--surface, #ffffff)',
                  addedBackground: 'rgba(122, 78, 26, 0.08)',
                  addedColor: 'var(--kicker, #7a4e1a)',
                  removedBackground: 'rgba(184, 112, 80, 0.08)',
                  removedColor: 'var(--palm, #b87050)',
                  wordAddedBackground: 'rgba(122, 78, 26, 0.2)',
                  wordRemovedBackground: 'rgba(184, 112, 80, 0.2)',
                  gutterBackground: 'var(--surface-muted, #faf7f1)',
                  gutterColor: 'var(--sea-ink-soft, #8a7c6e)',
                  codeFoldGutterBackground: 'var(--surface-muted, #faf7f1)',
                  codeFoldBackground: 'var(--surface-muted, #faf7f1)',
                },
              },
            }}
          />
        </div>
      )}
    </div>
  )
}
