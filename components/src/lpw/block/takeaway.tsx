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
      className="my-6 bg-sea-ink p-6 sm:p-7 text-foam relative shadow-md border border-sea-ink group"
    >
      <div className="border border-foam/20 p-5 sm:p-6 relative">
        {/* 沉金藏书票古典印信标签 */}
        <div className="absolute -top-3 left-6 bg-sea-ink px-3 font-mono text-[10px] tracking-[0.2em] text-foam/80 uppercase flex items-center gap-1.5">
          <span className="inline-block w-1.5 h-1.5 bg-lagoon rounded-none" />
          <span>{title}</span>
        </div>

        {/* 核心引言金句 */}
        <div className="font-serif text-base sm:text-lg lg:text-xl font-normal leading-relaxed text-foam/95">
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
