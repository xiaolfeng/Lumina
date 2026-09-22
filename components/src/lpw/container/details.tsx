import { ChevronRight } from "lucide-react";
import type React from "react";
import { useState } from "react";
import { containerVariants } from "../contract";
import { renderContainerBlocks } from "../renderer";
import type { LpwContainerSlotProps, LpwDetailsContainerProps } from "../types";

export const DetailsContainer: React.FC<
  LpwContainerSlotProps<LpwDetailsContainerProps>
> = ({ nodeId, blockId, props, location, children, childrenBlocks }) => {
  const actualId = nodeId || blockId || "";
  const actualChildren = children || childrenBlocks || [];
  const variant = props.variant;
  const contract = variant ? containerVariants.details[variant] : undefined;

  const [open, setOpen] = useState<boolean>(props.defaultOpen ?? false);
  const summaryId = `${actualId}-summary`;

  return (
    <section
      id={actualId}
      data-testid="details-container"
      data-variant={variant || "supplement"}
      className="border border-line/70 bg-surface/35 font-sans transition-colors"
    >
      <button
        type="button"
        id={summaryId}
        aria-expanded={open}
        aria-controls={`${actualId}-content`}
        onClick={() => setOpen((o) => !o)}
        className="flex min-h-12 w-full cursor-pointer select-none items-center justify-between gap-3 px-4 py-3 text-left transition-colors hover:bg-surface/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-lagoon/40"
      >
        <span className="min-w-0 break-words flex-1 font-serif text-sm font-semibold text-sea-ink">
          {props.summary}
        </span>
        <span className="flex shrink-0 items-center gap-2">
          <span className="hidden sm:inline text-xs text-sea-ink-soft">
            {open ? "收起内容" : "查看内容"}
          </span>
          <ChevronRight
            aria-hidden="true"
            className={`h-4 w-4 text-sea-ink-soft transition-transform duration-200 motion-reduce:transition-none ${
              open ? "rotate-90" : ""
            }`}
          />
        </span>
      </button>
      {open && (
        <div
          id={`${actualId}-content`}
          role="region"
          aria-labelledby={summaryId}
          className="flex min-w-0 flex-col gap-6 border-t border-line/50 px-4 py-5"
        >
          {renderContainerBlocks(
            actualChildren,
            location,
            contract,
            `details/${variant || "supplement"}`,
          )}
        </div>
      )}
    </section>
  );
};
