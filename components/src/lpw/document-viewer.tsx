import { BookOpen, GitBranch, Tag, User } from "lucide-react";
import { Badge } from "../ui/badge";
import type React from "react";
import { useMemo } from "react";
import { parseLpwSource } from "./parser";
import { ensureRegistered } from "./register-all";
import { childLocation, rootLocation } from "./render-location";
import { LpwNodeRenderer } from "./renderer";
import { AnnotationProvider } from "./annotation-context";
import { AnnotationGutter } from "./annotation";

export interface LpwDocumentViewerProps {
  source: string;
}

export const LpwDocumentViewer: React.FC<LpwDocumentViewerProps> = ({
  source,
}) => {
  ensureRegistered();
  const { document: doc, error } = useMemo(
    () => parseLpwSource(source),
    [source],
  );

  if (error) {
    return (
      <div className="max-w-6xl mx-auto px-4 py-8">
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
    );
  }

  const { meta, content = [] } = doc;

  return (
    <AnnotationProvider>
      <div className="w-full max-w-7xl mx-auto my-6 px-3 sm:px-6 flex justify-center items-start gap-8 min-w-0">
        <div
          data-testid="lpw-paper"
          className="w-full max-w-5xl shrink-0 bg-surface-strong border border-line shadow-sm relative text-sea-ink font-sans min-w-0"
        >
          <div aria-hidden="true" className="h-1 bg-lagoon" />

          {meta && (
            <header className="px-6 sm:px-10 pt-10 pb-8 bg-surface/40">
              <h1 className="font-serif text-2xl sm:text-3xl lg:text-4xl font-semibold tracking-tight text-sea-ink leading-snug mb-3.5 break-words [overflow-wrap:anywhere]">
                {meta.title}
              </h1>

              {meta.description && (
                <p className="text-sm sm:text-base text-sea-ink-soft leading-relaxed italic font-serif max-w-4xl">
                  {meta.description}
                </p>
              )}

              {meta.tags && meta.tags.length > 0 && (
                <div className="flex flex-wrap items-center gap-1.5 pt-4">
                  {meta.tags.map((tag) => (
                    <Badge
                      key={tag}
                      variant="secondary"
                      className="inline-flex items-center gap-1 text-[11px] font-mono uppercase px-2.5 py-0.5 border border-line/60 bg-surface-muted text-sea-ink-soft hover:bg-surface-muted"
                    >
                      <Tag className="h-3 w-3 shrink-0 text-sea-ink-soft/70" />
                      <span>{tag}</span>
                    </Badge>
                  ))}
                </div>
              )}

              {(meta.author || meta.version) && (
                <div className="mt-6 flex flex-wrap items-center gap-3 text-xs text-sea-ink-soft/80 font-mono">
                  {meta.author && (
                    <span className="inline-flex items-center gap-1.5">
                      <User className="h-3.5 w-3.5 shrink-0 text-sea-ink-soft" />
                      <strong className="text-sea-ink font-semibold">
                        {meta.author}
                      </strong>
                    </span>
                  )}
                  {meta.version && (
                    <span className="inline-flex items-center gap-1 bg-surface-muted border border-line/60 px-2 py-0.5">
                      <GitBranch className="h-3 w-3 shrink-0 text-sea-ink-soft" />
                      <span>v{meta.version}</span>
                    </span>
                  )}
                </div>
              )}
            </header>
          )}

          <div className="px-6 sm:px-10 py-8 min-w-0 max-w-full">
            {content.length === 0 ? (
              <div
                data-testid="document-empty"
                className="border-2 border-dashed border-line/50 p-12 text-center text-xs font-mono text-sea-ink-soft/70 bg-surface/30 my-6"
              >
                <BookOpen className="mx-auto h-8 w-8 text-sea-ink-soft/60 mb-3" />
                <div className="font-serif italic text-base text-sea-ink-soft mb-1">
                  空白典籍 · 等待分块生息
                </div>
                空文档：等待节点写入
              </div>
            ) : (
              <main className="space-y-4 min-w-0 max-w-full">
                {content.map((node, i) => (
                  <LpwNodeRenderer
                    key={node.id}
                    node={node}
                    location={childLocation(rootLocation(), i, node.id)}
                  />
                ))}
              </main>
            )}
          </div>
        </div>

        {/* 纸张右侧外部批注边栏 */}
        <AnnotationGutter />
      </div>
    </AnnotationProvider>
  );
};
