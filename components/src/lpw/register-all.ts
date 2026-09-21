import { CalloutBlock } from "./block/callout";
import { CardsBlock } from "./block/cards";
import { ChartBlock } from "./block/chart";
import { CodeBlock } from "./block/code";
import { ComparisonBlock } from "./block/comparison";
import { DiffBlock } from "./block/diff";
import { DividerBlock } from "./block/divider";
import { GalleryBlock } from "./block/gallery";
import { GlanceBlock } from "./block/glance";
import { HeadingBlock } from "./block/heading";
import { ImageBlock } from "./block/image";
import { ListBlock } from "./block/list";
import { MarkdownBlock } from "./block/markdown";
import { MermaidBlock } from "./block/mermaid";
import { MetricsBlock } from "./block/metrics";
import { OpenItemsBlock } from "./block/open-items";
import { PersonnelBlock } from "./block/personnel";
import { ProgressBlock } from "./block/progress";
import { QuadrantBlock } from "./block/quadrant";
import { QuoteBlock } from "./block/quote";
import { ScorecardBlock } from "./block/scorecard";
import { StepsBlock } from "./block/steps";
import { TableBlock } from "./block/table";
import { TakeawayBlock } from "./block/takeaway";
import { TimelineBlock } from "./block/timeline";
import { TreeBlock } from "./block/tree";
import { containerVariants } from "./contract";
import { DetailsContainer } from "./container/details";
import { PanelContainer } from "./container/panel";
import { SectionContainer } from "./container/section";
import { TabsContainer } from "./container/tabs";
import { LayoutNode } from "./layout/layout";
import { lpwRegistry } from "./registry";

/**
 * 注册所有已实现的 LPW 节点组件 (1 种 Layout + 4 种 Container + 26 种 Block = 31 种)
 */
export function registerAll(): void {
  // 1. Layout
  lpwRegistry.register("layout", "layout", {
    displayName: "Layout",
    Component: LayoutNode as never,
  });

  // 2. Containers (4 种)
  lpwRegistry.register("container", "section", {
    displayName: "Section",
    Component: SectionContainer as never,
    variants: containerVariants.section,
  });
  lpwRegistry.register("container", "panel", {
    displayName: "Panel",
    Component: PanelContainer as never,
    variants: containerVariants.panel,
  });
  lpwRegistry.register("container", "details", {
    displayName: "Details",
    Component: DetailsContainer as never,
    variants: containerVariants.details,
  });
  lpwRegistry.register("container", "tabs", {
    displayName: "Tabs",
    Component: TabsContainer as never,
    variants: containerVariants.tabs,
  });

  // 3. Blocks (26 种)
  const blockMap = [
    {
      type: "markdown",
      comp: MarkdownBlock,
      name: "Markdown",
      groups: ["text"],
      fields: ["content"],
    },
    {
      type: "callout",
      comp: CalloutBlock,
      name: "Callout",
      groups: ["notice"],
      fields: ["title", "content"],
    },
    {
      type: "heading",
      comp: HeadingBlock,
      name: "Heading",
      groups: ["text"],
      fields: ["content"],
    },
    {
      type: "list",
      comp: ListBlock,
      name: "List",
      groups: ["text"],
      fields: [],
    },
    {
      type: "quote",
      comp: QuoteBlock,
      name: "Quote",
      groups: ["text", "notice"],
      fields: ["content"],
    },
    {
      type: "code",
      comp: CodeBlock,
      name: "Code",
      groups: ["technical"],
      fields: [],
    },
    {
      type: "image",
      comp: ImageBlock,
      name: "Image",
      groups: ["media"],
      fields: [],
    },
    {
      type: "divider",
      comp: DividerBlock,
      name: "Divider",
      groups: [],
      fields: [],
    },
    {
      type: "cards",
      comp: CardsBlock,
      name: "Cards",
      groups: ["text"],
      fields: [],
    },
    {
      type: "metrics",
      comp: MetricsBlock,
      name: "Metrics",
      groups: ["data"],
      fields: [],
    },
    {
      type: "steps",
      comp: StepsBlock,
      name: "Steps",
      groups: ["process"],
      fields: [],
    },
    {
      type: "timeline",
      comp: TimelineBlock,
      name: "Timeline",
      groups: ["process"],
      fields: [],
    },
    {
      type: "diff",
      comp: DiffBlock,
      name: "Diff",
      groups: ["technical"],
      fields: [],
    },
    {
      type: "table",
      comp: TableBlock,
      name: "Table",
      groups: ["data", "technical"],
      fields: [],
    },
    {
      type: "mermaid",
      comp: MermaidBlock,
      name: "Mermaid",
      groups: ["media"],
      fields: [],
    },
    {
      type: "chart",
      comp: ChartBlock,
      name: "Chart",
      groups: ["data"],
      fields: [],
    },
    {
      type: "comparison",
      comp: ComparisonBlock,
      name: "Comparison",
      groups: ["decision"],
      fields: [],
    },
    {
      type: "progress",
      comp: ProgressBlock,
      name: "Progress",
      groups: ["data"],
      fields: [],
    },
    {
      type: "tree",
      comp: TreeBlock,
      name: "Tree",
      groups: ["process", "technical"],
      fields: [],
    },
    {
      type: "takeaway",
      comp: TakeawayBlock,
      name: "Takeaway",
      groups: ["decision", "notice"],
      fields: [],
    },
    {
      type: "glance",
      comp: GlanceBlock,
      name: "Glance",
      groups: [],
      fields: [],
    },
    {
      type: "open-items",
      comp: OpenItemsBlock,
      name: "OpenItems",
      groups: ["process"],
      fields: [],
    },
    {
      type: "scorecard",
      comp: ScorecardBlock,
      name: "Scorecard",
      groups: ["decision"],
      fields: [],
    },
    {
      type: "quadrant",
      comp: QuadrantBlock,
      name: "Quadrant",
      groups: ["decision"],
      fields: [],
    },
    {
      type: "personnel",
      comp: PersonnelBlock,
      name: "Personnel",
      groups: [],
      fields: [],
    },
    {
      type: "gallery",
      comp: GalleryBlock,
      name: "Gallery",
      groups: ["media"],
      fields: [],
    },
  ] as const;

  for (const b of blockMap) {
    lpwRegistry.register("block", b.type, {
      displayName: b.name,
      Component: b.comp as never,
      groups: b.groups as never,
      annotatableFields: b.fields as never,
    });
  }
}

/**
 * 确保注册表已加载，若表为空则执行全量自注册
 */
export function ensureRegistered(): void {
  if (lpwRegistry.entries().length === 0) {
    registerAll();
  }
}

// 模块加载时默认自执行
ensureRegistered();
