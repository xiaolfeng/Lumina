/**
 * remark-fenced-blocks.ts
 *
 * remark plugin that scans for `:::type{props}` ... `:::` blocks
 * and converts them to custom mdast nodes.
 *
 * Custom mdast node format:
 *   { type: 'callout', data: { hName: 'callout', hProperties: { type: 'info' } }, children: [...] }
 *
 * remark-rehype will convert the custom node to a hast element
 * (tagName='callout'), and react-markdown's `components` prop
 * maps it to a React component.
 *
 * Supported types: callout, card, steps, step
 * Nesting: :::steps can contain multiple :::step blocks.
 * Unclosed fences are left as plain text (no error).
 * Adjacency: fences may be written without blank lines around the content
 * (markdown merges adjacent lines into one paragraph); such paragraphs are
 * split by line into standalone fence paragraphs before matching.
 */

export const SUPPORTED_FENCE_TYPES = [
  "callout",
  "card",
  "steps",
  "step",
] as const;
export type SupportedFenceType = (typeof SUPPORTED_FENCE_TYPES)[number];

const SUPPORTED_FENCE_SET = new Set<string>(SUPPORTED_FENCE_TYPES);

/**
 * 统一严格开栏正则：仅匹配白名单中的合法组件名，支持可选属性语法。
 * 格式示例：`:::callout` 或 `:::callout{type="info" title="说明"}`。
 * 属性部分使用贪婪匹配至行末闭合花括号，支持属性值内包含单引号、双引号与嵌套花括号。
 * 非白名单或格式不合法（如未闭合的花括号）不作为开栏，回退为普通文本。
 */
const OPENING_FENCE_REGEX = /^:::(callout|card|steps|step)(?:\{(.*)\})?\s*$/;

/** 统一关栏正则：仅由 `:::` 及周围空白字符构成。 */
const CLOSING_FENCE_REGEX = /^:::\s*$/;

interface MdastNode {
  type: string;
  [key: string]: unknown;
}

interface MdastParent extends MdastNode {
  children: MdastNode[];
}

function extractNodeText(node: MdastNode): string {
  if (node.type === "text") {
    return (node as unknown as { value: string }).value;
  }
  if (node.type === "link") {
    const parent = node as MdastParent;
    if (Array.isArray(parent.children) && parent.children.length > 0) {
      return parent.children.map(extractNodeText).join("");
    }
    if ("url" in node && typeof node.url === "string") {
      return node.url;
    }
  }
  return "";
}

function getOpeningFenceText(node: MdastNode): string | null {
  if (node.type !== "paragraph") return null;
  const parent = node as MdastParent;
  if (parent.children.length === 0) return null;

  // 单子节点且为文本（快速路径）
  if (parent.children.length === 1 && parent.children[0].type === "text") {
    const val = (
      parent.children[0] as unknown as { value: string }
    ).value.trim();
    return OPENING_FENCE_REGEX.test(val) ? val : null;
  }

  // 多子节点：支持还原被 autolink literal 拆分的开栏属性段落
  let text = "";
  for (const child of parent.children) {
    if (child.type === "text") {
      text += (child as unknown as { value: string }).value;
    } else if (child.type === "link") {
      text += extractNodeText(child);
    } else {
      return null;
    }
  }

  const trimmed = text.trim();
  return OPENING_FENCE_REGEX.test(trimmed) ? trimmed : null;
}

export function isOpeningFence(node: MdastNode): boolean {
  return getOpeningFenceText(node) !== null;
}

export function parseOpeningFence(
  node: MdastNode,
): { type: SupportedFenceType; props: string } | null {
  const text = getOpeningFenceText(node);
  if (!text) return null;
  const match = text.match(OPENING_FENCE_REGEX);
  if (!match) return null;
  const type = match[1] as SupportedFenceType;
  if (!SUPPORTED_FENCE_SET.has(type)) return null;
  return { type, props: match[2] || "" };
}

export function isClosingFence(node: MdastNode): boolean {
  if (node.type !== "paragraph") return false;
  const parent = node as MdastParent;
  if (parent.children.length !== 1) return false;
  const firstChild = parent.children[0];
  if (firstChild.type !== "text") return false;
  const value = (firstChild as unknown as { value: string }).value;
  return CLOSING_FENCE_REGEX.test(value.trim());
}

function parseProps(propsStr: string): Record<string, string> {
  const props: Record<string, string> = {};
  if (!propsStr) return props;

  // 支持属性名 [a-zA-Z0-9_-]+，双引号值（含单引号/花括号/空值）、单引号值（含双引号/花括号/空值）及无引号值
  const regex =
    /(?:^|\s+)([a-zA-Z0-9_-]+)=(?:"((?:[^"\\]|\\.)*)"|'((?:[^'\\]|\\.)*)'|([^\s"'=]+))/g;
  let match: RegExpExecArray | null;
  while ((match = regex.exec(propsStr)) !== null) {
    const key = match[1];
    const val = match[2] || match[3] || match[4] || "";
    props[key] = val.replace(/\\(["'\\])/g, "$1");
  }

  return props;
}

function collectFencedBlock(
  children: MdastNode[],
  startIndex: number,
): { contentNodes: MdastNode[]; endIndex: number } | null {
  const contentNodes: MdastNode[] = [];
  let depth = 1;

  for (let i = startIndex; i < children.length; i++) {
    const node = children[i];

    if (isOpeningFence(node)) {
      depth++;
      contentNodes.push(node);
    } else if (isClosingFence(node)) {
      depth--;
      if (depth === 0) {
        return { contentNodes, endIndex: i };
      }
      contentNodes.push(node);
    } else {
      contentNodes.push(node);
    }
  }

  // 未找到对应的关栏
  return null;
}

/**
 * 将「围栏行与内容同段」的段落按行切分归一化。
 *
 * Markdown 解析器（remark-parse）会将相邻行合并为一个段落。
 * 当 `:::callout` / `:::steps` 等开栏行或 `:::` 关栏行前后没有空行时，
 * 围栏行会混在段落 AST 节点的文本流中。
 *
 * 本函数逐行切分，将开栏行与关栏行剥离为独立的单行段落节点；
 * 连续的普通内容行保留换行符重新组合为普通段落。
 * 所有产出均为纯净的新节点，杜绝任何对原始 AST 节点属性的原地变异（in-place mutation）。
 */
export function splitFenceParagraph(node: MdastNode): MdastNode[] {
  if (node.type !== "paragraph") return [node];
  const parent = node as MdastParent;
  if (parent.children.length === 0) return [node];

  // 快速路径：若不含围栏特征字符串，无需按行切分
  let containsFenceSignal = false;
  for (const child of parent.children) {
    if (
      child.type === "text" &&
      (child as unknown as { value: string }).value.includes(":::")
    ) {
      containsFenceSignal = true;
      break;
    }
  }
  if (!containsFenceSignal) return [node];

  // 按换行符切分为行
  interface LineInfo {
    nodes: MdastNode[];
    isPureText: boolean;
    textValue: string;
  }

  const lines: LineInfo[] = [];
  let curNodes: MdastNode[] = [];
  let curText = "";
  let curPureText = true;

  for (const child of parent.children) {
    if (child.type === "text") {
      const val = (child as unknown as { value: string }).value;
      const parts = val.split(/\r?\n/);
      for (let p = 0; p < parts.length; p++) {
        const part = parts[p];
        if (p > 0) {
          lines.push({
            nodes: curNodes,
            isPureText: curPureText,
            textValue: curText,
          });
          curNodes = [];
          curText = "";
          curPureText = true;
        }
        if (part.length > 0) {
          curNodes.push({ type: "text", value: part });
          curText += part;
        }
      }
    } else if (child.type === "link" && curText.startsWith(":::")) {
      // 还原开栏语法内部被 remark-gfm autolink literal 误拆的 URL 为纯文本
      const val = extractNodeText(child);
      curNodes.push({ type: "text", value: val });
      curText += val;
    } else {
      curPureText = false;
      curNodes.push({ ...child });
    }
  }

  lines.push({
    nodes: curNodes,
    isPureText: curPureText,
    textValue: curText,
  });

  const result: MdastNode[] = [];
  let pendingContentNodes: MdastNode[] = [];

  function flushPending() {
    if (pendingContentNodes.length > 0) {
      result.push({
        type: "paragraph",
        children: pendingContentNodes,
      });
      pendingContentNodes = [];
    }
  }

  for (const line of lines) {
    const trimmed = line.textValue.trim();
    const isOpening = line.isPureText && OPENING_FENCE_REGEX.test(trimmed);
    const isClosing = line.isPureText && CLOSING_FENCE_REGEX.test(trimmed);

    if (isOpening) {
      flushPending();
      result.push({
        type: "paragraph",
        children: [{ type: "text", value: trimmed }],
      });
    } else if (isClosing) {
      flushPending();
      result.push({
        type: "paragraph",
        children: [{ type: "text", value: ":::" }],
      });
    } else {
      if (line.nodes.length > 0) {
        if (pendingContentNodes.length > 0) {
          pendingContentNodes.push({ type: "text", value: "\n" });
        }
        pendingContentNodes.push(...line.nodes);
      }
    }
  }

  flushPending();
  return result.length > 0 ? result : [node];
}

/**
 * 递归提取并提升由于 CommonMark 惰性延续（lazy continuation）
 * 误吞入列表项（listItem）或引用块（blockquote）中的外层关栏标记及后续内容节点。
 *
 * 识别规则：扫描容器内部子节点，当遇到 depth === 0 时的第一个孤立关栏 `:::`，
 * 说明该关栏及后续所有子节点均属于外层文档流，将其从当前容器中剥离并提升返回。
 */
function liftLazyContinuationNodes(node: MdastNode): MdastNode[] {
  if (node.type === "blockquote") {
    const parent = node as MdastParent;
    let depth = 0;
    for (let idx = 0; idx < parent.children.length; idx++) {
      const child = parent.children[idx];
      if (isOpeningFence(child)) {
        depth++;
      } else if (isClosingFence(child)) {
        if (depth > 0) {
          depth--;
        } else {
          return parent.children.splice(idx);
        }
      }
    }
  } else if (node.type === "list") {
    const listParent = node as MdastParent;
    for (let i = listParent.children.length - 1; i >= 0; i--) {
      const itemNode = listParent.children[i];
      if (!("children" in itemNode)) {
        continue;
      }
      const item = itemNode as MdastParent;
      let depth = 0;
      for (let idx = 0; idx < item.children.length; idx++) {
        const child = item.children[idx];
        if (isOpeningFence(child)) {
          depth++;
        } else if (isClosingFence(child)) {
          if (depth > 0) {
            depth--;
          } else {
            const lifted = item.children.splice(idx);
            const subsequentItems = listParent.children.splice(i + 1);
            const flattenedSubsequent = subsequentItems.flatMap((it) =>
              "children" in it && Array.isArray((it as MdastParent).children)
                ? (it as MdastParent).children
                : [],
            );
            return [...lifted, ...flattenedSubsequent];
          }
        }
      }
    }
  }
  return [];
}

function processChildren(children: MdastNode[]): MdastNode[] {
  // 第一步：先对所有段落执行按行切分归一化
  const splitNodes = children.flatMap(splitFenceParagraph);

  // 第二步：处理列表/引用内部各层级并多节点提升误吞的外层节点
  const normalizedNodes: MdastNode[] = [];
  for (const node of splitNodes) {
    if ("children" in node && Array.isArray((node as MdastParent).children)) {
      const parent = node as MdastParent;
      parent.children = processChildren(parent.children);
      const liftedNodes = liftLazyContinuationNodes(node);
      normalizedNodes.push(node);
      if (liftedNodes.length > 0) {
        normalizedNodes.push(...liftedNodes);
      }
    } else {
      normalizedNodes.push(node);
    }
  }

  // 第三步：收集并递归解析围栏块
  const result: MdastNode[] = [];
  let i = 0;

  while (i < normalizedNodes.length) {
    const node = normalizedNodes[i];

    const opening = parseOpeningFence(node);
    if (opening) {
      const { type, props } = opening;
      const blockResult = collectFencedBlock(normalizedNodes, i + 1);

      if (blockResult) {
        const { contentNodes, endIndex } = blockResult;
        const processedContent = processChildren(contentNodes);

        const customNode: MdastNode = {
          type,
          data: {
            hName: type,
            hProperties: parseProps(props),
          },
          children: processedContent,
        };

        result.push(customNode);
        i = endIndex + 1;
        continue;
      }
    }

    result.push(node);
    i++;
  }

  return result;
}

export function remarkFencedBlocks() {
  return (tree: MdastNode) => {
    if (!("children" in tree) || !Array.isArray((tree as MdastParent).children))
      return;

    const parent = tree as MdastParent;
    parent.children = processChildren(parent.children);
  };
}
