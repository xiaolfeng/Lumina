import {
  BadgeCheck,
  CircleCheck,
  CircleHelp,
  Lightbulb,
  MessageSquare,
  StickyNote,
  TriangleAlert,
} from "lucide-react";
import type React from "react";
import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { useAnnotationContext } from "./annotation-context";
import type { RegisteredAnnotation } from "./annotation-context";
import { isSafeAnnotationPattern } from "./safe-pattern";
import type { LpwAnnotation, LpwAnnotationKind } from "./types";

/** 正文超过该长度时跳过划线标注，限制多项式级回溯的总开销 */
const MAX_ANNOTATED_TEXT_LENGTH = 10000;

/** 垂直方向小于该间距的气泡叠成一簇 */
const CLUSTER_GAP_PX = 72;

export const KIND_ICON_MAP: Record<
  LpwAnnotationKind,
  React.ComponentType<{ className?: string }>
> = {
  note: StickyNote,
  suggestion: Lightbulb,
  todo: CircleCheck,
  issue: TriangleAlert,
  approved: BadgeCheck,
  question: CircleHelp,
};

export const KIND_COLOR_MAP: Record<LpwAnnotationKind, string> = {
  note: "text-sea-ink-soft bg-surface-muted border-line",
  suggestion: "text-lagoon-deep bg-lagoon/15 border-lagoon/40",
  todo: "text-kicker bg-kicker/15 border-kicker/40",
  issue: "text-destructive bg-destructive/15 border-destructive/40",
  approved: "text-palm bg-palm/15 border-palm/40",
  question: "text-purple-600 bg-purple-100 border-purple-300",
};

const KIND_WAVY_COLOR: Record<LpwAnnotationKind, string> = {
  note: "decoration-sea-ink-soft/70",
  suggestion: "decoration-lagoon",
  todo: "decoration-kicker",
  issue: "decoration-destructive",
  approved: "decoration-palm",
  question: "decoration-purple-500",
};

const KIND_HIGHLIGHT_BG: Record<LpwAnnotationKind, string> = {
  note: "bg-sea-ink-soft/10",
  suggestion: "bg-lagoon/20",
  todo: "bg-kicker/20",
  issue: "bg-destructive/20",
  approved: "bg-palm/20",
  question: "bg-purple-100",
};

export function InvalidPatternNotice(): React.ReactElement {
  return (
    <span
      data-testid="annotation-invalid-pattern"
      role="status"
      className="ml-2 align-middle font-mono text-[10px] text-destructive"
    >
      批注匹配器无效
    </span>
  );
}

export interface AnnotationFrameProps {
  nodeId?: string;
  annotation?: LpwAnnotation;
  children: React.ReactNode;
}

export const AnnotationFrame: React.FC<AnnotationFrameProps> = ({
  nodeId = "",
  annotation,
  children,
}) => {
  const ctx = useAnnotationContext();
  const register = ctx?.registerAnnotation;

  useEffect(() => {
    if (!annotation || !register) return;
    return register({
      nodeId,
      annotation,
      label: annotation.label || annotation.kind,
      anchorText: annotation.targets?.[0]?.pattern,
    });
  }, [annotation, nodeId, register]);

  if (!annotation) {
    return <>{children}</>;
  }

  const isHovered = ctx?.hoveredAnnotationId === nodeId;
  const isActive = ctx?.activeAnnotationId === nodeId;

  return (
    <div
      id={`frame-${nodeId}`}
      data-testid="annotation-frame"
      data-kind={annotation.kind}
      className={`group relative my-2 transition-all duration-200 ${
        isHovered || isActive ? "ring-1 ring-lagoon/40" : ""
      }`}
    >
      <div className="relative">{children}</div>
    </div>
  );
};

export interface AnnotatedTextProps {
  nodeId?: string;
  field: string;
  value: string;
  annotation?: LpwAnnotation;
}

export const AnnotatedText: React.FC<AnnotatedTextProps> = ({
  nodeId = "",
  field,
  value,
  annotation,
}) => {
  const ctx = useAnnotationContext();

  const { parts, invalid } = useMemo(() => {
    if (!annotation?.targets || annotation.targets.length === 0) {
      return { parts: null, invalid: false };
    }

    const target = annotation.targets.find((t) => t.field === field);
    if (!target || !target.pattern) {
      return { parts: null, invalid: false };
    }
    if (!isSafeAnnotationPattern(target.pattern)) {
      return { parts: null, invalid: true };
    }
    if (value.length > MAX_ANNOTATED_TEXT_LENGTH) {
      return { parts: null, invalid: false };
    }

    try {
      const hasI = target.flags?.includes("i");
      const hasU = target.flags?.includes("u");
      const flags = `g${hasI ? "i" : ""}${hasU ? "u" : ""}`;
      const regex = new RegExp(target.pattern, flags);
      const segments: Array<{ text: string; isMatch: boolean }> = [];
      let lastIndex = 0;
      let match: RegExpExecArray | null;
      let iterations = 0;
      const maxIterations = 1000;

      while ((match = regex.exec(value)) !== null) {
        iterations++;
        if (iterations > maxIterations) {
          break;
        }

        const matchStart = match.index;
        const matchText = match[0];

        if (matchText.length === 0) {
          regex.lastIndex++;
          continue;
        }

        if (matchStart > lastIndex) {
          segments.push({
            text: value.slice(lastIndex, matchStart),
            isMatch: false,
          });
        }

        segments.push({
          text: matchText,
          isMatch: true,
        });

        lastIndex = matchStart + matchText.length;
      }

      if (lastIndex < value.length) {
        segments.push({
          text: value.slice(lastIndex),
          isMatch: false,
        });
      }

      return { parts: segments.length > 0 ? segments : null, invalid: false };
    } catch {
      return { parts: null, invalid: true };
    }
  }, [field, value, annotation]);

  const kind = annotation?.kind ?? "suggestion";
  const wavyColorClass = KIND_WAVY_COLOR[kind];
  const highlightBgClass = KIND_HIGHLIGHT_BG[kind];
  const isHovered = ctx?.hoveredAnnotationId === nodeId;
  const isActive = ctx?.activeAnnotationId === nodeId;

  return (
    <>
      {parts
        ? parts.map((segment, idx) => {
            if (segment.isMatch) {
              return (
                <mark
                  key={idx}
                  onMouseEnter={() => ctx?.setHoveredAnnotationId(nodeId)}
                  onMouseLeave={() => ctx?.setHoveredAnnotationId(null)}
                  onClick={() => ctx?.scrollToNode(nodeId)}
                  className={`bg-transparent text-inherit underline decoration-wavy underline-offset-4 cursor-pointer transition-all duration-200 ${wavyColorClass} ${
                    isHovered || isActive
                      ? `decoration-3 ${highlightBgClass} px-0.5 rounded-xs`
                      : "decoration-2"
                  }`}
                >
                  {segment.text}
                </mark>
              );
            }
            return <span key={idx}>{segment.text}</span>;
          })
        : value}
      {invalid ? <InvalidPatternNotice /> : null}
    </>
  );
};

export const MessageBubble: React.FC<{
  nodeId: string;
  annotation: LpwAnnotation;
}> = ({ nodeId, annotation }) => {
  const ctx = useAnnotationContext();
  const isHovered = ctx?.hoveredAnnotationId === nodeId;
  const isActive = ctx?.activeAnnotationId === nodeId;

  const Icon = KIND_ICON_MAP[annotation.kind];
  const colorClass = KIND_COLOR_MAP[annotation.kind];

  return (
    <div
      data-testid="annotation-bubble"
      onMouseEnter={() => ctx?.setHoveredAnnotationId(nodeId)}
      onMouseLeave={() => ctx?.setHoveredAnnotationId(null)}
      onClick={() => ctx?.scrollToNode(nodeId)}
      className={`relative cursor-pointer border bg-surface p-3.5 shadow-sm transition-all duration-200 font-sans text-xs ${
        isHovered || isActive
          ? "border-lagoon shadow-md -translate-x-1 ring-1 ring-lagoon/30"
          : "border-line/70 hover:border-line hover:shadow"
      }`}
    >
      <div
        aria-hidden="true"
        className={`absolute -left-1.5 top-4 h-3 w-3 rotate-45 border-b border-l bg-surface transition-colors ${
          isHovered || isActive ? "border-lagoon" : "border-line/70"
        }`}
      />

      <div className="flex items-center justify-between gap-2 mb-2">
        <span
          className={`flex items-center gap-1 rounded-full border px-2 py-0.5 font-mono text-[10px] font-semibold uppercase tracking-wider ${colorClass}`}
        >
          <Icon className="h-3 w-3" />
          <span>{annotation.label || annotation.kind}</span>
        </span>
        {annotation.author && (
          <span className="font-mono text-[11px] text-sea-ink-soft">
            {annotation.author}
          </span>
        )}
      </div>

      <p className="font-sans text-sea-ink leading-relaxed font-medium">
        {annotation.message}
      </p>
    </div>
  );
};

function clusterAnnotations(
  annotations: RegisteredAnnotation[],
  tops: Record<string, number>,
): RegisteredAnnotation[][] {
  const sorted = [...annotations].sort(
    (a, b) => (tops[a.nodeId] ?? 0) - (tops[b.nodeId] ?? 0),
  );
  const clusters: RegisteredAnnotation[][] = [];
  for (const item of sorted) {
    const itemTop = tops[item.nodeId] ?? 0;
    if (clusters.length > 0) {
      const last = clusters[clusters.length - 1];
      const lastTop = tops[last[0].nodeId] ?? 0;
      if (Math.abs(itemTop - lastTop) < CLUSTER_GAP_PX) {
        last.push(item);
        continue;
      }
    }
    clusters.push([item]);
  }
  return clusters;
}

export const AnnotationGutter: React.FC = () => {
  const ctx = useAnnotationContext();
  const annotations = ctx?.annotations ?? [];
  const hostRef = useRef<HTMLDivElement>(null);
  const [tops, setTops] = useState<Record<string, number>>({});
  const [expandedKey, setExpandedKey] = useState<string | null>(null);
  const [mobileOpen, setMobileOpen] = useState(false);

  useLayoutEffect(() => {
    if (annotations.length === 0) return;

    const paper = hostRef.current?.closest(
      '[data-testid="lpw-paper"]',
    ) as HTMLElement | null;

    const measure = () => {
      if (typeof window !== "undefined" && window.innerWidth < 768) return;
      const origin = paper?.getBoundingClientRect();
      const next: Record<string, number> = {};
      for (const item of annotations) {
        const el =
          document.getElementById(`frame-${item.nodeId}`) ??
          document.getElementById(item.nodeId);
        if (!el) continue;
        const r = el.getBoundingClientRect();
        next[item.nodeId] = Math.max(0, r.top - (origin?.top ?? 0));
      }
      setTops((prev) => {
        const keys = Object.keys(next);
        if (
          keys.length === Object.keys(prev).length &&
          keys.every((k) => prev[k] === next[k])
        ) {
          return prev;
        }
        return next;
      });
    };

    measure();
    if (typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(measure);
    if (paper) ro.observe(paper);
    return () => {
      ro.disconnect();
    };
  }, [annotations]);

  const clusters = useMemo(
    () => clusterAnnotations(annotations, tops),
    [annotations, tops],
  );

  if (annotations.length === 0) {
    return null;
  }

  return (
    <div ref={hostRef} className="contents">
      <aside
        data-testid="annotation-gutter"
        aria-label="文档批注"
        className="relative hidden w-52 shrink-0 self-stretch py-8 pr-3 md:block"
      >
        {clusters.map((group) => {
          const key = group.map((g) => g.nodeId).join("-");
          const top = tops[group[0].nodeId] ?? 0;
          const expanded = expandedKey === key || group.length === 1;

          return (
            <div
              key={key}
              className="absolute left-0 right-1"
              style={{ top }}
            >
              {expanded ? (
                <div className="space-y-2">
                  {group.map((item) => (
                    <MessageBubble
                      key={item.nodeId}
                      nodeId={item.nodeId}
                      annotation={item.annotation}
                    />
                  ))}
                  {group.length > 1 ? (
                    <button
                      type="button"
                      onClick={() => setExpandedKey(null)}
                      className="font-mono text-[10px] text-sea-ink-soft hover:text-sea-ink"
                    >
                      收起
                    </button>
                  ) : null}
                </div>
              ) : (
                <button
                  type="button"
                  data-testid="annotation-cluster"
                  aria-label={`展开 ${group.length} 条批注`}
                  onClick={() => setExpandedKey(key)}
                  className="relative w-full cursor-pointer text-left"
                >
                  <MessageBubble
                    nodeId={group[0].nodeId}
                    annotation={group[0].annotation}
                  />
                  <span className="absolute -right-1 -top-2 flex h-5 min-w-5 items-center justify-center border border-lagoon bg-surface-strong px-1 font-mono text-[10px] font-semibold text-lagoon-deep">
                    {group.length}
                  </span>
                </button>
              )}
            </div>
          );
        })}
      </aside>

      <div
        data-testid="annotation-mobile-strip"
        className="w-full border-t border-line/60 bg-surface/40 md:hidden"
      >
        <button
          type="button"
          aria-expanded={mobileOpen}
          aria-label={`查看 ${annotations.length} 条批注`}
          onClick={() => setMobileOpen((v) => !v)}
          className="flex w-full cursor-pointer items-center justify-between px-4 py-3 font-mono text-xs text-sea-ink"
        >
          <span className="flex items-center gap-2">
            <MessageSquare className="h-4 w-4 text-lagoon" />
            批注 ({annotations.length})
          </span>
          <span className="text-sea-ink-soft">{mobileOpen ? "收起" : "展开"}</span>
        </button>
        {mobileOpen ? (
          <div className="space-y-3 px-4 pb-4">
            {annotations.map((item) => (
              <MessageBubble
                key={item.nodeId}
                nodeId={item.nodeId}
                annotation={item.annotation}
              />
            ))}
          </div>
        ) : null}
      </div>
    </div>
  );
};
