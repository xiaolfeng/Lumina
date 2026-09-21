import { Sparkles } from "lucide-react";
import type React from "react";
import { AnnotatedText } from "../annotation";
import type { LpwBlockSlotProps, LpwTakeawayProps } from "../types";

export const TakeawayBlock: React.FC<LpwBlockSlotProps<LpwTakeawayProps>> = ({
  nodeId = "",
  props,
  annotation,
}) => {
  const title = props.title ?? "核心判断";

  return (
    <div
      data-testid="takeaway-block"
      className="my-6 bg-sea-ink p-4 sm:p-7 text-foam relative shadow-md border border-sea-ink group"
    >
      <div className="border border-foam/20 p-5 sm:p-6 relative">
        {/* 沉金藏书票古典印信标签 */}
        <div className="absolute -top-3 left-6 bg-sea-ink px-3 font-mono text-[10px] tracking-[0.2em] text-foam/80 uppercase flex items-center gap-1.5 max-w-[calc(100%-3rem)] truncate">
          <Sparkles className="lucide-sparkles h-3 w-3 shrink-0 text-lagoon" />
          <span className="truncate">{title}</span>
        </div>

        {/* 核心引言金句 */}
        <div className="font-serif text-base sm:text-lg lg:text-xl font-normal leading-relaxed text-foam/95 break-words">
          <AnnotatedText
            nodeId={nodeId}
            field="content"
            value={props.content}
            annotation={annotation}
          />
        </div>
      </div>
    </div>
  );
};
