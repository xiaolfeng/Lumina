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
      className="group relative overflow-hidden border border-sea-ink bg-sea-ink p-5 text-foam sm:p-7"
    >
      <div className="relative border-l border-lagoon/80 pl-4 sm:pl-6">
        {/* 沉金藏书票古典印信标签 */}
        <div className="mb-3 flex max-w-full items-center gap-1.5 text-xs font-semibold text-foam/70">
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
