import type React from "react";
import { AnnotatedText } from "../annotation";
import type { LpwBlockSlotProps, LpwQuoteProps } from "../types";

export const QuoteBlock: React.FC<LpwBlockSlotProps<LpwQuoteProps>> = ({
  nodeId = "",
  props,
  annotation,
}) => {
  const hasFooter = Boolean(props.author || props.source);

  return (
    <blockquote className="my-6 border-l-2 border-lagoon bg-lagoon/5 p-5 sm:p-6 text-sm text-sea-ink shadow-2xs relative">
      <div className="leading-relaxed font-serif text-base sm:text-lg italic text-sea-ink/90">
        <AnnotatedText
          nodeId={nodeId}
          field="content"
          value={props.content}
          annotation={annotation}
        />
      </div>
      {hasFooter && (
        <footer className="mt-3 text-xs not-italic font-mono text-sea-ink-soft tracking-wider flex items-center gap-1">
          <span>— {props.author ?? ""}</span>
          {props.author && props.source
            ? ` · ${props.source}`
            : (props.source ?? "")}
        </footer>
      )}
    </blockquote>
  );
};
