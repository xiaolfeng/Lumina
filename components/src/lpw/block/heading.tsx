import type React from "react";
import { AnnotatedText } from "../annotation";
import { ICON_COMPONENTS, isLpwIconName } from "../icon-map";
import type { LpwBlockSlotProps, LpwHeadingProps } from "../types";

const LEVEL_SPEC = {
  1: {
    Tag: "h1" as const,
    size: "text-2xl sm:text-3xl lg:text-4xl",
    color: "text-sea-ink",
    bar: "",
    iconSize: "h-6 w-6",
    iconColor: "text-lagoon",
    margin: "mt-2",
  },
  2: {
    Tag: "h2" as const,
    size: "text-xl sm:text-2xl",
    color: "text-sea-ink/90",
    bar: "",
    iconSize: "h-5 w-5",
    iconColor: "text-palm",
    margin: "mt-1",
  },
  3: {
    Tag: "h3" as const,
    size: "text-base sm:text-lg",
    color: "text-sea-ink-soft",
    bar: "",
    iconSize: "h-4 w-4",
    iconColor: "text-sea-ink-soft/70",
    margin: "mt-1",
  },
} as const;

const DEFAULT_ICON: Record<
  1 | 2 | 3,
  "bookmark" | "panel-left" | "circle-dot"
> = {
  1: "bookmark",
  2: "panel-left",
  3: "circle-dot",
};

export const HeadingBlock: React.FC<LpwBlockSlotProps<LpwHeadingProps>> = ({
  nodeId,
  blockId,
  props,
  annotation,
}) => {
  const actualId = nodeId || blockId;
  const rawLevel = Number(props.level);
  const level: 1 | 2 | 3 = rawLevel === 1 || rawLevel === 3 ? rawLevel : 2;
  const spec = LEVEL_SPEC[level];
  const iconName = isLpwIconName(props.icon) ? props.icon : DEFAULT_ICON[level];
  const Icon = ICON_COMPONENTS[iconName];

  return (
    <spec.Tag
      id={actualId}
      data-testid={`heading-${level}`}
      className={`font-serif font-semibold tracking-tight scroll-mt-16 flex items-start gap-2.5 ${spec.bar} ${spec.size} ${spec.color} ${spec.margin}`}
    >
      <Icon
        aria-hidden="true"
        className={`shrink-0 mt-1 ${spec.iconSize} ${spec.iconColor}`}
      />
      <span className="min-w-0 flex-1 break-words [overflow-wrap:anywhere]">
        <AnnotatedText
          nodeId={actualId}
          field="content"
          value={props.content}
          annotation={annotation}
        />
      </span>
    </spec.Tag>
  );
};
