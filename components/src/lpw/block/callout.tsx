import type React from "react";
import { CheckCircle2, Info, OctagonAlert, TriangleAlert } from "lucide-react";
import { AnnotatedText } from "../annotation";
import type { LpwBlockSlotProps, LpwCalloutProps } from "../types";

const LEVEL_CONFIGS: Record<
  NonNullable<LpwCalloutProps["level"]>,
  {
    border: string;
    bg: string;
    title: string;
    tag: string;
    Icon: React.ComponentType<{ className?: string; "aria-hidden"?: boolean | "true" | "false" }>;
  }
> = {
  info: {
    border: "border-lagoon",
    bg: "bg-lagoon/5",
    title: "text-lagoon-deep",
    tag: "说明",
    Icon: Info,
  },
  success: {
    border: "border-kicker",
    bg: "bg-kicker/5",
    title: "text-kicker",
    tag: "已验证",
    Icon: CheckCircle2,
  },
  warning: {
    border: "border-palm",
    bg: "bg-palm/5",
    title: "text-palm",
    tag: "注意",
    Icon: TriangleAlert,
  },
  error: {
    border: "border-destructive",
    bg: "bg-destructive/5",
    title: "text-destructive",
    tag: "错误",
    Icon: OctagonAlert,
  },
};

export const CalloutBlock: React.FC<LpwBlockSlotProps<LpwCalloutProps>> = ({
  nodeId = "",
  blockId = "",
  props,
  annotation,
}) => {
  const actualId = nodeId || blockId;
  const levelKey = (props.level ?? "info") as string;
  const config = LEVEL_CONFIGS[levelKey as keyof typeof LEVEL_CONFIGS] as
    | (typeof LEVEL_CONFIGS)[keyof typeof LEVEL_CONFIGS]
    | undefined;
  const resolvedConfig = config ?? LEVEL_CONFIGS.info;
  const level = config ? (levelKey as keyof typeof LEVEL_CONFIGS) : "info";
  const LevelIcon = resolvedConfig.Icon;

  return (
    <div
      id={actualId}
      role={level === "error" ? "alert" : "status"}
      data-testid="callout-block"
      data-level={level}
      className={`relative border border-solid border-line/50 border-l-4 p-4 text-sm text-sea-ink sm:p-5 ${resolvedConfig.border} ${resolvedConfig.bg}`}
    >
      {/* 出版物微型分类签 */}
      <div
        className={`mb-2 flex items-center gap-1.5 text-xs font-semibold ${resolvedConfig.title}`}
      >
        <LevelIcon className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
        <span>{resolvedConfig.tag}</span>
      </div>

      {props.title && (
        <div
          className={`mb-2 font-serif text-base font-semibold tracking-tight ${resolvedConfig.title}`}
        >
          <AnnotatedText
            nodeId={nodeId}
            field="title"
            value={props.title}
            annotation={annotation}
          />
        </div>
      )}
      <div className="leading-relaxed font-sans text-sea-ink/90 break-words [overflow-wrap:anywhere]">
        <AnnotatedText
          nodeId={nodeId}
          field="content"
          value={props.content}
          annotation={annotation}
        />
      </div>
    </div>
  );
};
