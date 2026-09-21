import type React from "react";
import { useState } from "react";
import { containerVariants } from "../contract";
import { ICON_COMPONENTS } from "../icon-map";
import { renderContainerBlocks } from "../renderer";
import type { LpwContainerSlotProps, LpwSectionContainerProps } from "../types";

export const SectionContainer: React.FC<
  LpwContainerSlotProps<LpwSectionContainerProps>
> = ({ nodeId, blockId, props, location, children, childrenBlocks }) => {
  const actualId = nodeId || blockId || "";
  const actualChildren = children || childrenBlocks || [];
  const variant = props.variant;
  const contract = variant ? containerVariants.section[variant] : undefined;
  const collapsible = Boolean(props.collapsible);
  const [open, setOpen] = useState<boolean>(props.defaultOpen ?? true);
  const icon = props.icon;
  const Icon = icon ? ICON_COMPONENTS[icon] : undefined;

  return (
    <section
      id={actualId}
      data-testid="section-container"
      data-variant={variant || "article"}
      className="my-3 font-sans"
    >
      <div className="mb-3 border-l-[3px] border-palm/70 pl-3">
        {collapsible ? (
          <button
            type="button"
            aria-expanded={open}
            aria-controls={`${actualId}-content`}
            onClick={() => setOpen((o) => !o)}
            className="group flex w-full cursor-pointer select-none items-center justify-between text-left"
          >
            <h2
              id={`${actualId}-title`}
              className="flex items-center font-serif text-xl sm:text-2xl font-semibold tracking-tight text-sea-ink transition-colors group-hover:text-lagoon"
            >
              {Icon && (
                <Icon
                  className="h-5 w-5 text-palm shrink-0 mr-2"
                  aria-hidden="true"
                />
              )}
              {props.title}
            </h2>
            <span className="border border-line/60 bg-surface/50 px-2 py-0.5 font-mono text-xs text-sea-ink-soft/70">
              {open ? "收起 ▲" : "展开 ▼"}
            </span>
          </button>
        ) : (
          <h2
            id={`${actualId}-title`}
            className="flex items-center font-serif text-xl sm:text-2xl font-semibold tracking-tight text-sea-ink"
          >
            {Icon && (
              <Icon
                className="h-5 w-5 text-palm shrink-0 mr-2"
                aria-hidden="true"
              />
            )}
            {props.title}
          </h2>
        )}
      </div>

      {(!collapsible || open) && (
        <div
          id={`${actualId}-content`}
          role="region"
          aria-labelledby={`${actualId}-title`}
          data-testid="section-content"
          className="space-y-6"
        >
          {renderContainerBlocks(
            actualChildren,
            location,
            contract,
            `section/${variant || "article"}`,
          )}
        </div>
      )}
    </section>
  );
};
