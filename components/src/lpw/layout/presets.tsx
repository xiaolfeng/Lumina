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
      let gridColsStyle: React.CSSProperties | undefined;
      const colsClass = "grid-cols-1 md:grid";

      if (strategy?.type === "ratio" && strategy.tracks.length > 0) {
        gridColsStyle = {
          gridTemplateColumns: strategy.tracks
            .map((t) => `minmax(0, ${t}fr)`)
            .join(" "),
        };
      } else if (strategy?.type === "fixed-fluid") {
        const sizeMap: Record<string, string> = {
          sm: "240px",
          md: "360px",
          lg: "480px",
        };
        const fixedSize = sizeMap[strategy.size] || strategy.size;
        gridColsStyle = {
          gridTemplateColumns:
            strategy.fixed === "start"
              ? `${fixedSize} minmax(0, 1fr)`
              : `minmax(0, 1fr) ${fixedSize}`,
        };
      } else {
        gridColsStyle = {
          gridTemplateColumns: "minmax(0, 1fr) minmax(0, 1fr)",
        };
      }

      if (direction === "vertical") {
        return (
          <div
            data-testid="layout-split"
            className={`grid grid-cols-1 ${gapClass} w-full min-w-0`}
          >
            {renderedChildren.map((child, idx) => (
              <div key={idx} className="min-w-0 w-full">
                {child}
              </div>
            ))}
          </div>
        );
      }

      return (
        <div
          data-testid="layout-split"
          className={`grid ${colsClass} ${gapClass} w-full min-w-0`}
          style={gridColsStyle}
        >
          {renderedChildren.map((child, idx) => (
            <div key={idx} className="min-w-0 w-full">
              {child}
            </div>
          ))}
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
            const items =
              isOddGroup && alternateFrom === "media"
                ? [pair[1], pair[0]]
                : pair;

            return (
              <div
                key={pIdx}
                className={`grid grid-cols-1 md:grid-cols-2 ${gapClass} items-center`}
              >
                {items}
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
          {renderedChildren.map((child, idx) => (
            <div key={idx} className="min-w-0 w-full">
              {child}
            </div>
          ))}
        </div>
      );
    }

    case "bento": {
      const placementMap = new Map<string, LpwLayoutPlacement>();
      for (const p of placements) {
        placementMap.set(p.nodeId, p);
      }

      const totalCols = strategy?.type === "spans" ? strategy.columns : 3;

      return (
        <div
          data-testid="layout-bento"
          className={`grid grid-cols-1 md:grid ${gapClass} w-full min-w-0`}
          style={{
            gridTemplateColumns: `repeat(${totalCols}, minmax(0, 1fr))`,
          }}
        >
          {renderedChildren.map((child, idx) => {
            const rawNode = rawChildren[idx];
            // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
            const p = rawNode ? placementMap.get(rawNode.id) : undefined;
            const colSpan = p?.colSpan ?? 1;
            const rowSpan = p?.rowSpan ?? 1;

            return (
              <div
                // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
                key={rawNode ? rawNode.id : idx}
                className="min-w-0 w-full"
                style={{
                  gridColumn: `span ${colSpan}`,
                  gridRow: `span ${rowSpan}`,
                }}
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

      return (
        <div data-testid="layout-newspaper" className="space-y-8 font-sans">
          {/* eslint-disable-next-line @typescript-eslint/no-unnecessary-condition */}
          {leadNode ? (
            <div className="border-b border-line pb-6">{leadNode}</div>
          ) : null}
          <div className="columns-1 md:columns-2 lg:columns-3 gap-8 [&>*]:break-inside-avoid space-y-6">
            {bodyNode}
            {asideNodes}
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

      // placement 移动端排序
      const imgPlacement = placements.find((p) => p.nodeId === imgId);
      const mdPlacement = placements.find((p) => p.nodeId === mdId);
      const imgOrder = imgPlacement?.orderOnMobile ?? 1;
      const mdOrder = mdPlacement?.orderOnMobile ?? 2;

      return (
        <div
          data-testid="layout-editorial-wrap"
          className="relative my-6 font-sans"
        >
          {/* 桌面端浮动模式 */}
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

          {/* 移动端流式上下堆叠 */}
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
            {renderedChildren}
          </div>
        );
      }
      return (
        <div data-testid="layout-flow" className="space-y-6">
          {renderedChildren}
        </div>
      );
    }
  }
}
