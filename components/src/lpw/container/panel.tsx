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
      ? "border-l-4 border-lagoon bg-surface/35 p-5"
      : variant === "aside"
        ? "border-l-2 border-line/70 bg-surface/20 p-4 text-sm"
        : "p-2";

  return (
    <div
      id={nodeId}
      data-testid="panel-container"
      data-variant={variant}
      className={`my-3 font-sans ${variantClass}`}
    >
      {title && (
        <div className="mb-4 flex items-center gap-2 font-serif font-semibold text-sea-ink">
          {Icon && <Icon className="h-4 w-4 text-lagoon" aria-hidden="true" />}
          <span>{title}</span>
        </div>
      )}
      <div
        className={`${
          variant === "dashboard" ? "grid gap-4 md:grid-cols-2" : "space-y-4"
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
