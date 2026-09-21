import React, { useMemo } from "react";
import { Markdown, proseArticle } from "../../markdown";
import { useAnnotationContext } from "../annotation-context";
import { isSafeAnnotationPattern } from "../safe-pattern";
import type { LpwBlockSlotProps, LpwMarkdownProps } from "../types";

const KIND_WAVY_COLOR: Record<string, string> = {
  note: "decoration-sea-ink-soft/70",
  suggestion: "decoration-lagoon",
  todo: "decoration-kicker",
  issue: "decoration-destructive",
  approved: "decoration-palm",
  question: "decoration-purple-500",
};

const KIND_HIGHLIGHT_BG: Record<string, string> = {
  note: "bg-sea-ink-soft/10",
  suggestion: "bg-lagoon/20",
  todo: "bg-kicker/20",
  issue: "bg-destructive/20",
  approved: "bg-palm/20",
  question: "bg-purple-100",
};

function highlightText(
  text: string,
  regex: RegExp,
  renderMark: (match: string, key: string | number) => React.ReactNode,
): React.ReactNode {
  const parts: React.ReactNode[] = [];
  let lastIndex = 0;
  let match: RegExpExecArray | null;
  let iterations = 0;

  while ((match = regex.exec(text)) !== null) {
    iterations++;
    if (iterations > 1000) break;

    const matchStart = match.index;
    const matchText = match[0];

    if (matchText.length === 0) {
      regex.lastIndex++;
      continue;
    }

    if (matchStart > lastIndex) {
      parts.push(text.slice(lastIndex, matchStart));
    }

    parts.push(renderMark(matchText, `${matchStart}-${iterations}`));
    lastIndex = matchStart + matchText.length;
  }

  if (lastIndex < text.length) {
    parts.push(text.slice(lastIndex));
  }

  return parts.length > 0 ? parts : text;
}

function recursivelyHighlightNode(
  node: React.ReactNode,
  regex: RegExp,
  renderMark: (match: string, key: string | number) => React.ReactNode,
): React.ReactNode {
  if (typeof node === "string") {
    return highlightText(node, regex, renderMark);
  }
  if (Array.isArray(node)) {
    return node.map((child) =>
      recursivelyHighlightNode(child, regex, renderMark),
    );
  }
  if (React.isValidElement(node)) {
    const element = node as React.ReactElement<{ children?: React.ReactNode }>;
    if (element.props.children) {
      return React.cloneElement(element, {
        children: recursivelyHighlightNode(
          element.props.children,
          regex,
          renderMark,
        ),
      });
    }
  }
  return node;
}

export const MarkdownBlock: React.FC<LpwBlockSlotProps<LpwMarkdownProps>> = ({
  nodeId = "",
  props,
  annotation,
}) => {
  const ctx = useAnnotationContext();
  const isHovered = ctx?.hoveredAnnotationId === nodeId;
  const isActive = ctx?.activeAnnotationId === nodeId;

  const target = annotation?.targets?.find((t) => t.field === "content");
  const kind = annotation?.kind ?? "suggestion";
  const wavyColorClass = KIND_WAVY_COLOR[kind];
  const highlightBgClass = KIND_HIGHLIGHT_BG[kind];

  const customComponents = useMemo(() => {
    if (!target?.pattern || !isSafeAnnotationPattern(target.pattern)) {
      return undefined;
    }

    const hasI = target.flags?.includes("i");
    const hasU = target.flags?.includes("u");
    const flags = `g${hasI ? "i" : ""}${hasU ? "u" : ""}`;
    let regex: RegExp;
    try {
      regex = new RegExp(target.pattern, flags);
    } catch {
      return undefined;
    }

    const renderMark = (match: string, key: string | number) => (
      <mark
        key={key}
        onMouseEnter={() => ctx?.setHoveredAnnotationId(nodeId)}
        onMouseLeave={() => ctx?.setHoveredAnnotationId(null)}
        onClick={() => ctx?.scrollToNode(nodeId)}
        className={`bg-transparent text-inherit underline decoration-wavy underline-offset-4 cursor-pointer transition-all duration-200 ${wavyColorClass} ${
          isHovered || isActive
            ? `decoration-3 ${highlightBgClass} px-0.5 rounded-xs font-medium`
            : "decoration-2"
        }`}
      >
        {match}
      </mark>
    );

    const wrapWithHighlight = (
      Tag: React.ElementType,
      p: { children?: React.ReactNode; [key: string]: any },
    ) => {
      const { children, ...rest } = p;
      return (
        <Tag {...rest}>
          {recursivelyHighlightNode(children, regex, renderMark)}
        </Tag>
      );
    };

    return {
      p: (p: any) => wrapWithHighlight("p", p),
      li: (p: any) => wrapWithHighlight("li", p),
      h1: (p: any) => wrapWithHighlight("h1", p),
      h2: (p: any) => wrapWithHighlight("h2", p),
      h3: (p: any) => wrapWithHighlight("h3", p),
      h4: (p: any) => wrapWithHighlight("h4", p),
      h5: (p: any) => wrapWithHighlight("h5", p),
      h6: (p: any) => wrapWithHighlight("h6", p),
      blockquote: (p: any) => wrapWithHighlight("blockquote", p),
    };
  }, [
    target,
    isHovered,
    isActive,
    nodeId,
    ctx,
    wavyColorClass,
    highlightBgClass,
  ]);

  return (
    <div
      className={`${proseArticle} my-5 font-sans leading-relaxed text-sea-ink min-w-0 max-w-full`}
    >
      <Markdown components={customComponents}>{props.content}</Markdown>
    </div>
  );
};
