import {
  BadgeCheck,
  CircleCheck,
  CircleHelp,
  Lightbulb,
  MessageSquare,
  StickyNote,
  TriangleAlert,
  X,
} from "lucide-react";
import type React from "react";
import { useEffect, useMemo, useState } from "react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../ui/tooltip";
import { useAnnotationContext } from "./annotation-context";
import { isSafeAnnotationPattern } from "./safe-pattern";
import type { LpwAnnotation, LpwAnnotationKind } from "./types";

/** 正文超过该长度时跳过划线标注，限制多项式级回溯的总开销 */
const MAX_ANNOTATED_TEXT_LENGTH = 10000;

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

  // 注册批注到全局 Context
  useEffect(() => {
    if (!annotation || !ctx) return;
    const unregister = ctx.registerAnnotation({
      nodeId,
      annotation,
      label: annotation.label || annotation.kind,
      anchorText: annotation.targets?.[0]?.pattern,
    });
    return unregister;
  }, [annotation, ctx, nodeId]);

  if (!annotation) {
    return <>{children}</>;
  }

  const Icon = KIND_ICON_MAP[annotation.kind];
  const colorClass = KIND_COLOR_MAP[annotation.kind];
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
      {/* 块右上角悬浮小圆点徽章（支持直接聚焦与 Tooltip） */}
      <div className="absolute -top-3 right-2 z-20">
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                type="button"
                aria-label={`批注：${annotation.kind}`}
                onMouseEnter={() => ctx?.setHoveredAnnotationId(nodeId)}
                onMouseLeave={() => ctx?.setHoveredAnnotationId(null)}
                onClick={() => ctx?.setActiveAnnotationId(nodeId)}
                className={`flex cursor-pointer items-center gap-1 rounded-full border px-2 py-0.5 font-mono text-[10px] shadow-2xs transition-transform hover:scale-105 ${colorClass}`}
              >
                <Icon className="h-3 w-3" />
                <span className="font-semibold uppercase tracking-wider">
                  {annotation.label || annotation.kind}
                </span>
              </button>
            </TooltipTrigger>
            <TooltipContent
              side="top"
              className="max-w-xs border border-line bg-surface-strong p-3 text-xs text-sea-ink shadow-md"
            >
              <p className="font-sans font-medium leading-relaxed">
                {annotation.message}
              </p>
              {annotation.author && (
                <p className="mt-1 font-mono text-[10px] text-sea-ink-soft">
                  — {annotation.author}
                </p>
              )}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>

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

  const parts = useMemo(() => {
    if (!annotation?.targets || annotation.targets.length === 0) {
      return null;
    }

    const target = annotation.targets.find((t) => t.field === field);
    if (!target || !target.pattern) {
      return null;
    }
    // ReDoS 防线：回溯型引擎执行前先过静态语法子集检查，危险模式回退原文渲染
    if (!isSafeAnnotationPattern(target.pattern)) {
      return null;
    }
    if (value.length > MAX_ANNOTATED_TEXT_LENGTH) {
      return null;
    }

    try {
      // Q-10 修复：对齐 Schema 枚举 [i, u, iu]，透传 u/iu 标志以支持 Unicode 属性与码点转义
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

      return segments.length > 0 ? segments : null;
    } catch {
      return null;
    }
  }, [field, value, annotation]);

  if (!parts) {
    return <>{value}</>;
  }

  const kind = annotation?.kind ?? "suggestion";
  const wavyColorClass = KIND_WAVY_COLOR[kind];
  const highlightBgClass = KIND_HIGHLIGHT_BG[kind];
  const isHovered = ctx?.hoveredAnnotationId === nodeId;
  const isActive = ctx?.activeAnnotationId === nodeId;

  return (
    <>
      {parts.map((segment, idx) => {
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
      })}
    </>
  );
};

/**
 * Word 式批注 Messages 气泡卡片
 */
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
      onMouseEnter={() => ctx?.setHoveredAnnotationId(nodeId)}
      onMouseLeave={() => ctx?.setHoveredAnnotationId(null)}
      onClick={() => ctx?.scrollToNode(nodeId)}
      className={`relative cursor-pointer border bg-surface p-3.5 shadow-sm transition-all duration-200 font-sans text-xs ${
        isHovered || isActive
          ? "border-lagoon shadow-md -translate-x-1 ring-1 ring-lagoon/30"
          : "border-line/70 hover:border-line hover:shadow"
      }`}
    >
      {/* 气泡左侧小三角形指示角 */}
      <div
        aria-hidden="true"
        className={`absolute -left-1.5 top-4 h-3 w-3 rotate-45 border-b border-l bg-surface transition-colors ${
          isHovered || isActive ? "border-lagoon" : "border-line/70"
        }`}
      />

      {/* 头部：类型徽章与作者 */}
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

      {/* 批注具体消息 */}
      <p className="font-sans text-sea-ink leading-relaxed font-medium">
        {annotation.message}
      </p>

      {/* 关联的目标字段与关键词线索 */}
      {annotation.targets && annotation.targets.length > 0 && (
        <div className="mt-2.5 pt-2 border-t border-line/40 flex flex-wrap gap-1.5 items-center">
          <span className="font-mono text-[10px] text-sea-ink-soft/70 uppercase">
            TARGET:
          </span>
          {annotation.targets.map((t, i) => (
            <span
              key={i}
              className="font-mono text-[10px] bg-surface-muted px-1.5 py-0.5 border border-line/50 text-sea-ink-soft"
            >
              &ldquo;{t.pattern}&rdquo;
            </span>
          ))}
        </div>
      )}
    </div>
  );
};

/**
 * 批注消息右侧边栏 (Annotation Gutter)
 */
export const AnnotationGutter: React.FC = () => {
  const ctx = useAnnotationContext();
  const annotations = ctx?.annotations ?? [];
  const [expanded, setExpanded] = useState(false);

  if (annotations.length === 0) {
    return null;
  }

  return (
    <>
      {/* 桌面超宽屏常驻边栏 */}
      <aside
        aria-label="文档批注列表"
        className="hidden 2xl:block w-72 shrink-0 py-8 space-y-4 select-none"
      >
        <div className="flex items-center justify-between border-b border-line/60 pb-2 mb-3">
          <div className="flex items-center gap-2 font-mono text-xs font-semibold text-sea-ink">
            <MessageSquare className="h-4 w-4 text-lagoon" />
            <span>MESSAGES // 批注 ({annotations.length})</span>
          </div>
        </div>
        <div className="space-y-3.5">
          {annotations.map((item) => (
            <MessageBubble
              key={item.nodeId}
              nodeId={item.nodeId}
              annotation={item.annotation}
            />
          ))}
        </div>
      </aside>

      {/* 响应式抽屉或浮动按钮（当屏幕未达 2xl 时提供悬浮交互入口） */}
      <div className="2xl:hidden">
        <button
          type="button"
          aria-label={`查看 ${annotations.length} 条批注`}
          onClick={() => setExpanded((v) => !v)}
          className="fixed bottom-6 right-6 z-40 flex items-center gap-2 border border-lagoon bg-surface-strong px-3.5 py-2 text-xs font-mono font-semibold text-sea-ink shadow-lg hover:bg-surface cursor-pointer rounded-full"
        >
          <MessageSquare className="h-4 w-4 text-lagoon" />
          <span>批注 ({annotations.length})</span>
        </button>

        {expanded && (
          <div className="fixed inset-y-0 right-0 z-50 w-80 max-w-full bg-surface-strong border-l border-line p-5 shadow-2xl overflow-y-auto space-y-4">
            <div className="flex items-center justify-between border-b border-line pb-3">
              <div className="flex items-center gap-2 font-mono text-xs font-semibold text-sea-ink">
                <MessageSquare className="h-4 w-4 text-lagoon" />
                <span>MESSAGES 批注列表 ({annotations.length})</span>
              </div>
              <button
                type="button"
                aria-label="关闭批注面板"
                onClick={() => setExpanded(false)}
                className="p-1 text-sea-ink-soft hover:text-sea-ink cursor-pointer"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
            <div className="space-y-3 pt-2">
              {annotations.map((item) => (
                <MessageBubble
                  key={item.nodeId}
                  nodeId={item.nodeId}
                  annotation={item.annotation}
                />
              ))}
            </div>
          </div>
        )}
      </div>
    </>
  );
};
