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
  sm: "gap-3 sm:gap-4",
  md: "gap-5 sm:gap-6",
  lg: "gap-7 sm:gap-10",
};

const ALIGN_CLASS = {
  start: "items-start",
  center: "items-center",
  stretch: "items-stretch",
};

const COL_SPAN_CLASS: Record<number, string> = {
  1: "",
  2: "@3xl/lpw:col-span-2",
  3: "@3xl/lpw:col-span-3",
  4: "@3xl/lpw:col-span-4",
};

const ROW_SPAN_CLASS: Record<number, string> = {
  1: "",
  2: "@3xl/lpw:row-span-2",
  3: "@3xl/lpw:row-span-3",
};

const BENTO_COLS_CLASS: Record<number, string> = {
  2: "@3xl/lpw:grid-cols-2",
  3: "@3xl/lpw:grid-cols-3",
  4: "@3xl/lpw:grid-cols-4",
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
    className: "@max-3xl/lpw:[order:var(--m-order)]",
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
    align,
  } = props;
  const gapClass = GAP_MAP[gap] || "gap-5 sm:gap-6";
  const resolvedAlign = align ?? (pattern === "alternating" ? "center" : "stretch");
  const alignClass = ALIGN_CLASS[resolvedAlign];

  switch (pattern) {
    case "split": {
      if (direction === "vertical") {
        return (
          <div
            data-testid="layout-split"
            className={`grid grid-cols-1 ${alignClass} ${gapClass} w-full min-w-0`}
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
          className={`grid grid-cols-1 @3xl/lpw:[grid-template-columns:var(--layout-cols)] ${alignClass} ${gapClass} w-full min-w-0`}
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
        <div data-testid="layout-alternating" className="space-y-10 sm:space-y-14">
          {pairs.map((pair, pIdx) => {
            const isOddGroup = pIdx % 2 === 1;
            const shouldMirror = isOddGroup && alternateFrom === "media";

            return (
              <div
                key={pIdx}
                className={`grid grid-cols-1 @3xl/lpw:grid-cols-2 ${gapClass} ${alignClass}`}
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
                      ? "@3xl/lpw:order-2"
                      : "@3xl/lpw:order-1"
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
          ? "grid-cols-[repeat(auto-fit,minmax(min(100%,18rem),1fr))]"
          : "grid-cols-[repeat(auto-fit,minmax(min(100%,14rem),1fr))]";

      return (
        <div
          data-testid="layout-grid"
          className={`grid ${colClass} ${alignClass} ${gapClass} w-full min-w-0`}
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
      const colsClass = BENTO_COLS_CLASS[totalCols] || "@3xl/lpw:grid-cols-3";

      return (
        <div
          data-testid="layout-bento"
          className={`grid grid-cols-1 ${colsClass} ${alignClass} ${gapClass} w-full min-w-0`}
        >
          {renderedChildren.map((child, idx) => {
            const rawNode = rawChildren[idx];
            // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
            const p = rawNode ? placementMap.get(rawNode.id) : undefined;
            const colSpan = Math.min(p?.colSpan ?? 1, totalCols);
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
        const inferredRole =
          rawNode.type === "markdown" && bodyNode === null ? "body" : "aside";
        const role = p?.role || inferredRole;

        if (role === "lead" && leadNode === null) {
          leadNode = child;
        } else if (role === "body" && bodyNode === null) {
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
          ? "columns-1 @3xl/lpw:columns-2"
          : "columns-1 @3xl/lpw:columns-2 @5xl/lpw:columns-3";

      return (
        <div data-testid="layout-newspaper" className="space-y-10 font-sans">
          {/* eslint-disable-next-line @typescript-eslint/no-unnecessary-condition */}
          {leadNode ? (
            <div className="border-b border-line pb-6">{leadNode}</div>
          ) : null}
          <div
            data-testid="layout-newspaper-flow"
            className={`${colClass} gap-8`}
          >
            <div data-testid="layout-newspaper-body" className="mb-6">
              {bodyNode}
            </div>
            {asideNodes.map((fn, idx) => (
              <div
                key={idx}
                data-testid="layout-newspaper-aside"
                className="break-inside-avoid mb-6"
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
          className="relative font-sans flex flex-col gap-5 @3xl/lpw:block"
        >
          <div
            className="@max-3xl/lpw:!mx-0 @max-3xl/lpw:!w-full @max-3xl/lpw:[order:var(--m-order)] @3xl/lpw:my-0"
            style={{
              ["--m-order" as string]: String(imgOrder),
              float: isRight ? "right" : "left",
              width: "var(--media-width)",
              aspectRatio: aspectMap[wrapStrategy.mediaShape],
              marginLeft: isRight ? "1.5rem" : "0",
              marginRight: isRight ? "0" : "1.5rem",
              marginBottom: "1rem",
              ["--media-width" as string]: width,
            }}
          >
            {imgNode}
          </div>
          <div
            className="@max-3xl/lpw:[order:var(--m-order)]"
            style={{
              ["--m-order" as string]: String(mdOrder),
            }}
          >
            {mdNode}
          </div>
          <div className="clear-both hidden @3xl/lpw:block" />
        </div>
      );
    }

    case "flow":
    default: {
      if (direction === "horizontal") {
        return (
          <div
            data-testid="layout-flow"
            className={`flex flex-wrap ${alignClass} ${gapClass}`}
          >
            {renderedChildren.map((child, idx) => {
              const order = mobileOrderProps(rawChildren[idx], placements, idx + 1);
              return (
                <div key={idx} className={`min-w-0 max-w-full flex-[0_1_auto] ${order.className}`} style={order.style}>
                  {child}
                </div>
              );
            })}
          </div>
        );
      }
      return (
        <div data-testid="layout-flow" className={`flex flex-col ${gapClass}`}>
          {renderedChildren.map((child, idx) => {
            const order = mobileOrderProps(rawChildren[idx], placements, idx + 1);
            return (
              <div key={idx} className={`min-w-0 ${order.className}`} style={order.style}>
                {child}
              </div>
            );
          })}
        </div>
      );
    }
  }
}
