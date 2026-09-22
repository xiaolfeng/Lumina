import type React from "react";
import { ICON_COMPONENTS } from "../icon-map";
import type { LpwContainerSlotProps, LpwPanelContainerProps } from "../types";
import { renderContainerBlocks } from "../renderer";
import { containerVariants } from "../contract";

export const PanelContainer: React.FC<
  LpwContainerSlotProps<LpwPanelContainerProps>
> = ({ nodeId, props, location, children }) => {
  const { variant = "summary", title, icon } = props;
  const Icon = icon ? ICON_COMPONENTS[icon] : undefined;
  const contract = containerVariants.panel[variant];

  const variantClass =
    variant === "summary"
      ? "border border-lagoon/30 bg-lagoon/5 p-5 sm:p-6"
      : variant === "aside"
        ? "border border-line/70 bg-surface/35 p-4 sm:p-5 text-sm"
        : "border-y border-line/70 py-5";

  return (
    <div
      id={nodeId}
      data-testid="panel-container"
      data-variant={variant}
      className={`min-w-0 font-sans ${variantClass}`}
    >
      {title && (
        <div className="mb-5 flex items-center gap-2.5 border-b border-line/60 pb-3 font-serif text-base font-semibold text-sea-ink">
          {Icon && <Icon className="h-4 w-4 text-lagoon" aria-hidden="true" />}
          <span>{title}</span>
        </div>
      )}
      <div
        className={`${
          variant === "dashboard"
            ? "grid grid-cols-[repeat(auto-fit,minmax(min(100%,16rem),1fr))] gap-5"
            : "flex flex-col gap-5"
        } [&>[data-testid$='-block']]:my-2 [&>[data-testid$='-block']:first-child]:mt-0 [&>[data-testid$='-block']:last-child]:mb-0`}
      >
        {renderContainerBlocks(
          children || [],
          location,
          contract,
          `panel/${variant}`,
        )}
      </div>
    </div>
  );
};
