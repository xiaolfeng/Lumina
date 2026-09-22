import type React from "react";
import { Quote } from "lucide-react";
import { AnnotatedText } from "../annotation";
import type { LpwBlockSlotProps, LpwQuoteProps } from "../types";

export const QuoteBlock: React.FC<LpwBlockSlotProps<LpwQuoteProps>> = ({
  nodeId = "",
  blockId = "",
  props,
  annotation,
}) => {
  const actualId = nodeId || blockId;
  const hasFooter = Boolean(props.author || props.source);

  return (
    <blockquote
      id={actualId}
      className="relative border-l-2 border-lagoon bg-lagoon/5 p-4 text-sm text-sea-ink sm:p-6"
    >
      <div className="leading-relaxed font-serif text-base sm:text-lg italic text-sea-ink/90 break-words [overflow-wrap:anywhere]">
        <Quote className="h-4 w-4 text-lagoon/60 inline mr-2" aria-hidden="true" />
        <AnnotatedText
          nodeId={nodeId}
          field="content"
          value={props.content}
          annotation={annotation}
        />
      </div>
      {hasFooter && (
        <footer className="mt-3 text-xs not-italic text-sea-ink-soft flex flex-wrap items-baseline gap-1.5">
          {props.author && <span>— {props.author}</span>}
          {props.author && props.source && <span>·</span>}
          {props.source && <cite className="not-italic">{props.source}</cite>}
        </footer>
      )}
    </blockquote>
  );
};
