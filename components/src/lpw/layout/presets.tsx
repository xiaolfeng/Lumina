import type React from "react";
import type {
  LpwLayoutChild,
  LpwLayoutPlacement,
  LpwLayoutProps,
} from "../types";

export interface LayoutPresetProps {
  props: LpwLayoutProps;
  rawChildren: LpwLayoutChild[];
  renderedChildren: React.ReactNode[];
}

const GAP_MAP = {
  sm: "gap-4",
  md: "gap-6",
  lg: "gap-8",
};

const COL_SPAN_CLASS: Record<number, string> = {
  1: "",
  2: "md:col-span-2",
  3: "md:col-span-3",
  4: "md:col-span-4",
};

const ROW_SPAN_CLASS: Record<number, string> = {
  1: "",
  2: "md:row-span-2",
  3: "md:row-span-3",
};

const BENTO_COLS_CLASS: Record<number, string> = {
  2: "md:grid-cols-2",
  3: "md:grid-cols-3",
  4: "md:grid-cols-4",
};

function mobileOrderProps(
  rawNode: LpwLayoutChild | undefined,
  placements: LpwLayoutPlacement[],
  fallback: number,
): { className: string; style: React.CSSProperties } {
  const placed = rawNode
    ? placements.find((p) => p.nodeId === rawNode.id)
    : undefined;
  const order = placed?.orderOnMobile ?? fallback;
  return {
    className: "max-md:[order:var(--m-order)]",
    style: { ["--m-order" as string]: String(order) },
  };
}

function splitTemplate(
  strategy: LpwLayoutProps["strategy"],
): string {
  if (strategy?.type === "ratio" && strategy.tracks.length > 0) {
    return strategy.tracks.map((t) => `minmax(0, ${t}fr)`).join(" ");
  }
  if (strategy?.type === "fixed-fluid") {
    const sizeMap: Record<string, string> = {
      sm: "240px",
      md: "360px",
      lg: "480px",
    };
    const fixedSize = sizeMap[strategy.size] || strategy.size;
    return strategy.fixed === "start"
      ? `${fixedSize} minmax(0, 1fr)`
      : `minmax(0, 1fr) ${fixedSize}`;
  }
  return "minmax(0, 1fr) minmax(0, 1fr)";
}

export function renderLayoutPreset({
  props,
  rawChildren,
  renderedChildren,
}: LayoutPresetProps): React.ReactNode {
  const {
    pattern,
    direction = "horizontal",
    strategy,
    gap = "md",
    placements = [],
    alternateFrom = "media",
  } = props;
  const gapClass = GAP_MAP[gap] || "gap-6";

  switch (pattern) {
    case "split": {
      if (direction === "vertical") {
        return (
          <div
            data-testid="layout-split"
            className={`grid grid-cols-1 ${gapClass} w-full min-w-0`}
          >
            {renderedChildren.map((child, idx) => {
              const order = mobileOrderProps(rawChildren[idx], placements, idx + 1);
              return (
                <div
                  key={idx}
                  className={`min-w-0 w-full ${order.className}`}
                  style={order.style}
                >
                  {child}
                </div>
              );
            })}
          </div>
        );
      }

      return (
        <div
          data-testid="layout-split"
          className={`grid grid-cols-1 md:[grid-template-columns:var(--layout-cols)] ${gapClass} w-full min-w-0`}
          style={{ ["--layout-cols" as string]: splitTemplate(strategy) }}
        >
          {renderedChildren.map((child, idx) => {
            const order = mobileOrderProps(rawChildren[idx], placements, idx + 1);
            return (
              <div
                key={idx}
                className={`min-w-0 w-full ${order.className}`}
                style={order.style}
              >
                {child}
              </div>
            );
          })}
        </div>
      );
    }

    case "alternating": {
      const pairs: React.ReactNode[][] = [];
      for (let i = 0; i < renderedChildren.length; i += 2) {
        pairs.push([renderedChildren[i], renderedChildren[i + 1]]);
      }

      return (
        <div data-testid="layout-alternating" className={`space-y-8`}>
          {pairs.map((pair, pIdx) => {
            const isOddGroup = pIdx % 2 === 1;
            const shouldMirror = isOddGroup && alternateFrom === "media";

            return (
              <div
                key={pIdx}
                className={`grid grid-cols-1 md:grid-cols-2 ${gapClass} items-center`}
              >
                {pair.map((child, i) => {
                  const rawIdx = pIdx * 2 + i;
                  const order = mobileOrderProps(
                    rawChildren[rawIdx],
                    placements,
                    rawIdx + 1,
                  );
                  const mirrorClass = shouldMirror
                    ? i === 0
                      ? "md:order-2"
                      : "md:order-1"
                    : "";
                  return (
                    <div
                      key={i}
                      className={`min-w-0 w-full ${mirrorClass} ${order.className}`}
                      style={order.style}
                    >
                      {child}
                    </div>
                  );
                })}
              </div>
            );
          })}
        </div>
      );
    }

    case "grid": {
      const count = renderedChildren.length;
      const colClass =
        count <= 2
          ? "grid-cols-1 sm:grid-cols-2"
          : "grid-cols-1 sm:grid-cols-2 lg:grid-cols-3";

      return (
        <div
          data-testid="layout-grid"
          className={`grid ${colClass} ${gapClass} w-full min-w-0`}
        >
          {renderedChildren.map((child, idx) => {
            const order = mobileOrderProps(rawChildren[idx], placements, idx + 1);
            return (
              <div
                key={idx}
                className={`min-w-0 w-full ${order.className}`}
                style={order.style}
              >
                {child}
              </div>
            );
          })}
        </div>
      );
    }

    case "bento": {
      const placementMap = new Map<string, LpwLayoutPlacement>();
      for (const p of placements) {
        placementMap.set(p.nodeId, p);
      }

      const totalCols = strategy?.type === "spans" ? strategy.columns : 3;
      const colsClass = BENTO_COLS_CLASS[totalCols] || "md:grid-cols-3";

      return (
        <div
          data-testid="layout-bento"
          className={`grid grid-cols-1 ${colsClass} ${gapClass} w-full min-w-0`}
        >
          {renderedChildren.map((child, idx) => {
            const rawNode = rawChildren[idx];
            // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
            const p = rawNode ? placementMap.get(rawNode.id) : undefined;
            const colSpan = p?.colSpan ?? 1;
            const rowSpan = p?.rowSpan ?? 1;
            const order = mobileOrderProps(rawNode, placements, idx + 1);

            return (
              <div
                // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
                key={rawNode ? rawNode.id : idx}
                className={`min-w-0 w-full ${COL_SPAN_CLASS[colSpan]} ${ROW_SPAN_CLASS[rowSpan]} ${order.className}`}
                style={order.style}
              >
                {child}
              </div>
            );
          })}
        </div>
      );
    }

    case "newspaper": {
      const placementMap = new Map<string, LpwLayoutPlacement>();
      for (const p of placements) {
        placementMap.set(p.nodeId, p);
      }

      let leadNode: React.ReactNode = null;
      let bodyNode: React.ReactNode = null;
      const asideNodes: React.ReactNode[] = [];
      const fullNodes: React.ReactNode[] = [];

      renderedChildren.forEach((child, idx) => {
        const rawNode = rawChildren[idx];
        // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
        const p = rawNode ? placementMap.get(rawNode.id) : undefined;
        const role = p?.role || "body";

        if (role === "lead") {
          leadNode = child;
        } else if (role === "body") {
          bodyNode = child;
        } else if (role === "full") {
          fullNodes.push(child);
        } else {
          asideNodes.push(child);
        }
      });

      const colCount = strategy?.type === "columns" ? strategy.count : 3;
      const colClass =
        colCount === 2
          ? "columns-1 md:columns-2"
          : "columns-1 md:columns-2 lg:columns-3";

      return (
        <div data-testid="layout-newspaper" className="space-y-8 font-sans">
          {/* eslint-disable-next-line @typescript-eslint/no-unnecessary-condition */}
          {leadNode ? (
            <div className="border-b border-line pb-6">{leadNode}</div>
          ) : null}
          <div
            data-testid="layout-newspaper-flow"
            className={`${colClass} gap-8 space-y-6`}
          >
            <div data-testid="layout-newspaper-body">{bodyNode}</div>
            {asideNodes.map((fn, idx) => (
              <div
                key={idx}
                data-testid="layout-newspaper-aside"
                className="break-inside-avoid"
              >
                {fn}
              </div>
            ))}
          </div>
          {fullNodes.map((fn, idx) => (
            <div key={idx} className="w-full">
              {fn}
            </div>
          ))}
        </div>
      );
    }

    case "editorial-wrap": {
      let imgNode: React.ReactNode = null;
      let mdNode: React.ReactNode = null;
      let imgId = "";
      let mdId = "";

      rawChildren.forEach((c, idx) => {
        if (c.type === "image") {
          imgNode = renderedChildren[idx];
          imgId = c.id;
        } else if (c.type === "markdown") {
          mdNode = renderedChildren[idx];
          mdId = c.id;
        }
      });

      const wrapStrategy =
        strategy?.type === "media-wrap"
          ? strategy
          : {
              mediaPosition: "top-start" as const,
              mediaWidth: "third" as const,
              mediaShape: "square" as const,
            };

      const widthMap = {
        quarter: "25%",
        third: "33.333%",
        "two-fifths": "40%",
        half: "50%",
      };
      const aspectMap = {
        square: "1 / 1",
        portrait: "3 / 4",
        landscape: "4 / 3",
      };

      const width = widthMap[wrapStrategy.mediaWidth] || "33.333%";
      const isRight = wrapStrategy.mediaPosition === "top-end";

      const imgPlacement = placements.find((p) => p.nodeId === imgId);
      const mdPlacement = placements.find((p) => p.nodeId === mdId);
      const imgOrder = imgPlacement?.orderOnMobile ?? 1;
      const mdOrder = mdPlacement?.orderOnMobile ?? 2;

      return (
        <div
          data-testid="layout-editorial-wrap"
          className="relative my-6 font-sans"
        >
          <div className="hidden md:block">
            <div
              style={{
                float: isRight ? "right" : "left",
                width,
                aspectRatio: aspectMap[wrapStrategy.mediaShape],
                marginLeft: isRight ? "1.5rem" : "0",
                marginRight: isRight ? "0" : "1.5rem",
                marginBottom: "1rem",
              }}
            >
              {imgNode}
            </div>
            <div>{mdNode}</div>
            <div style={{ clear: "both" }} />
          </div>

          <div className="flex flex-col gap-4 md:hidden">
            {imgOrder <= mdOrder ? (
              <>
                <div>{imgNode}</div>
                <div>{mdNode}</div>
              </>
            ) : (
              <>
                <div>{mdNode}</div>
                <div>{imgNode}</div>
              </>
            )}
          </div>
        </div>
      );
    }

    case "flow":
    default: {
      if (direction === "horizontal") {
        return (
          <div
            data-testid="layout-flow"
            className={`flex flex-wrap items-center ${gapClass}`}
          >
            {renderedChildren.map((child, idx) => {
              const order = mobileOrderProps(rawChildren[idx], placements, idx + 1);
              return (
                <div key={idx} className={order.className} style={order.style}>
                  {child}
                </div>
              );
            })}
          </div>
        );
      }
      return (
        <div data-testid="layout-flow" className="space-y-6">
          {renderedChildren.map((child, idx) => {
            const order = mobileOrderProps(rawChildren[idx], placements, idx + 1);
            return (
              <div key={idx} className={order.className} style={order.style}>
                {child}
              </div>
            );
          })}
        </div>
      );
    }
  }
}
