import { describe, expect, it } from "vitest";
import {
  isOpeningFence,
  parseOpeningFence,
  remarkFencedBlocks,
  splitFenceParagraph,
} from "./remark-fenced-blocks";

function createAst(children: unknown[]) {
  return { type: "root", children };
}

function paragraph(text: string) {
  return { type: "paragraph", children: [{ type: "text", value: text }] };
}

describe("remarkFencedBlocks", () => {
  it("converts :::callout to custom callout node", () => {
    const tree = createAst([
      paragraph(":::callout"),
      paragraph("Hello world"),
      paragraph(":::"),
    ]);

    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children).toHaveLength(1);
    expect(tree.children[0]).toHaveProperty("type", "callout");
    expect(tree.children[0]).toHaveProperty("data.hName", "callout");
    expect(tree.children[0]).toHaveProperty("data.hProperties", {});
    expect((tree.children[0] as { children: unknown[] }).children).toHaveLength(
      1,
    );
    expect(
      (tree.children[0] as { children: { type: string }[] }).children[0].type,
    ).toBe("paragraph");
  });

  it("converts :::callout with props to custom node with hProperties", () => {
    const tree = createAst([
      paragraph(':::callout{type="info"}'),
      paragraph("Hello world"),
      paragraph(":::"),
    ]);

    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children).toHaveLength(1);
    expect(tree.children[0]).toHaveProperty("type", "callout");
    expect(tree.children[0]).toHaveProperty("data.hProperties", {
      type: "info",
    });
  });

  it("converts nested :::steps with :::step blocks", () => {
    const tree = createAst([
      paragraph(":::steps"),
      paragraph(":::step"),
      paragraph("Step 1"),
      paragraph(":::"),
      paragraph(":::step"),
      paragraph("Step 2"),
      paragraph(":::"),
      paragraph(":::"),
    ]);

    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children).toHaveLength(1);
    expect(tree.children[0]).toHaveProperty("type", "steps");
    const stepsChildren = (tree.children[0] as { children: unknown[] })
      .children;
    expect(stepsChildren).toHaveLength(2);
    expect((stepsChildren[0] as { type: string }).type).toBe("step");
    expect((stepsChildren[1] as { type: string }).type).toBe("step");
    expect(
      (stepsChildren[0] as { children: { type: string }[] }).children[0].type,
    ).toBe("paragraph");
    expect(
      (stepsChildren[1] as { children: { type: string }[] }).children[0].type,
    ).toBe("paragraph");
  });

  it("treats unclosed fence as plain text", () => {
    const tree = createAst([paragraph(":::callout"), paragraph("Hello world")]);

    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children).toHaveLength(2);
    expect(tree.children[0]).toHaveProperty("type", "paragraph");
    expect(tree.children[1]).toHaveProperty("type", "paragraph");
  });

  it("does not affect code blocks", () => {
    const tree = createAst([
      { type: "code", lang: "markdown", value: ":::callout\nHello\n:::\n" },
    ]);

    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children).toHaveLength(1);
    expect(tree.children[0]).toHaveProperty("type", "code");
  });

  it("converts :::card with title and href props", () => {
    const tree = createAst([
      paragraph(':::card{title="My Card" href="https://example.com"}'),
      paragraph("Card content"),
      paragraph(":::"),
    ]);

    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children).toHaveLength(1);
    expect(tree.children[0]).toHaveProperty("type", "card");
    expect(tree.children[0]).toHaveProperty("data.hProperties", {
      title: "My Card",
      href: "https://example.com",
    });
  });

  it("does not affect horizontal rules", () => {
    const tree = createAst([{ type: "thematicBreak" }]);

    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children).toHaveLength(1);
    expect(tree.children[0]).toHaveProperty("type", "thematicBreak");
  });
});

describe("remarkFencedBlocks with fences adjacent to content", () => {
  it("splits a paragraph where opening fence, content and closing fence share one paragraph", () => {
    const tree = createAst([
      {
        type: "paragraph",
        children: [
          { type: "text", value: ':::callout{type="info"}\n' },
          { type: "emphasis", children: [{ type: "text", value: "核心思想" }] },
          { type: "text", value: "：先粗筛再精排，保证召回。\n:::" },
        ],
      },
    ]);

    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children).toHaveLength(1);
    expect(tree.children[0]).toHaveProperty("type", "callout");
    expect(tree.children[0]).toHaveProperty("data.hProperties", {
      type: "info",
    });
    const calloutChildren = (tree.children[0] as { children: unknown[] })
      .children;
    expect(calloutChildren).toHaveLength(1);
    expect(JSON.stringify(calloutChildren)).not.toContain(":::");
  });

  it("splits the opening fence when only it is merged with the first content paragraph", () => {
    const tree = createAst([
      {
        type: "paragraph",
        children: [
          { type: "text", value: ':::callout{type="warning"}\n' },
          { type: "text", value: "bm25_rank 字段是恒等值" },
        ],
      },
      paragraph("`03_coarse.py` 中 bm25_rank 为恒等值"),
      paragraph(":::"),
    ]);

    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children).toHaveLength(1);
    expect(tree.children[0]).toHaveProperty("type", "callout");
    expect(tree.children[0]).toHaveProperty("data.hProperties", {
      type: "warning",
    });
    expect((tree.children[0] as { children: unknown[] }).children).toHaveLength(
      2,
    );
  });

  it("splits a closing fence merged into the last content paragraph", () => {
    const tree = createAst([
      paragraph(':::callout{type="warning"}'),
      {
        type: "paragraph",
        children: [
          { type: "inlineCode", value: "03_coarse.py" },
          { type: "text", value: " 修复完毕\n:::" },
        ],
      },
    ]);

    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children).toHaveLength(1);
    expect(tree.children[0]).toHaveProperty("type", "callout");
    const calloutChildren = (tree.children[0] as { children: unknown[] })
      .children;
    expect(calloutChildren).toHaveLength(1);
    expect(JSON.stringify(calloutChildren)).not.toContain(":::");
  });

  it("does not mutate original AST text nodes in place", () => {
    const originalTextNode = {
      type: "text",
      value: ':::callout{type="info"}\nContent text\n:::',
    };
    const inputParagraph = {
      type: "paragraph",
      children: [originalTextNode],
    };

    const splitResult = splitFenceParagraph(inputParagraph);
    expect(splitResult).toHaveLength(3);
    // 验证原输入对象的属性保持未变
    expect(originalTextNode.value).toBe(
      ':::callout{type="info"}\nContent text\n:::',
    );
  });

  it("safely rejects unsupported or non-standard fence opening without generating empty hName", () => {
    const invalidNode = paragraph(':::unknownType{title="test"}');
    expect(isOpeningFence(invalidNode)).toBe(false);
    expect(parseOpeningFence(invalidNode)).toBeNull();

    const unclosedPropsNode = paragraph(':::steps{title="test"');
    expect(isOpeningFence(unclosedPropsNode)).toBe(false);
    expect(parseOpeningFence(unclosedPropsNode)).toBeNull();
  });

  it("parses opening fence props containing single quotes, double quotes and nested braces", () => {
    const node = paragraph(
      ':::card{title="User {id}" quote=\'He said "hello"\' href="https://example.com" empty=""}',
    );
    expect(isOpeningFence(node)).toBe(true);
    const parsed = parseOpeningFence(node);
    expect(parsed).not.toBeNull();
    expect(parsed?.type).toBe("card");

    const tree = createAst([node, paragraph("Card content"), paragraph(":::")]);
    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children[0]).toHaveProperty("data.hProperties", {
      title: "User {id}",
      quote: 'He said "hello"',
      href: "https://example.com",
      empty: "",
    });
  });

  it("handles opening fence split by autolink literal into text and link nodes", () => {
    const tree = createAst([
      {
        type: "paragraph",
        children: [
          { type: "text", value: ':::card{href="' },
          {
            type: "link",
            url: "https://example.com/docs",
            children: [{ type: "text", value: "https://example.com/docs" }],
          },
          { type: "text", value: '" title="Doc"}' },
        ],
      },
      paragraph("Card body"),
      paragraph(":::"),
    ]);

    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children).toHaveLength(1);
    expect(tree.children[0]).toHaveProperty("type", "card");
    expect(tree.children[0]).toHaveProperty("data.hProperties", {
      href: "https://example.com/docs",
      title: "Doc",
    });
  });

  it("lifts closing fence and subsequent nodes mis-captured by lazy continuation", () => {
    const tree = createAst([
      paragraph(":::steps"),
      paragraph(":::step"),
      {
        type: "list",
        children: [
          {
            type: "listItem",
            children: [
              paragraph("item 1"),
              paragraph("item 2\n:::\n:::step\nstep 2 content\n:::\n:::"),
            ],
          },
        ],
      },
    ]);

    const plugin = remarkFencedBlocks();
    plugin(tree);

    expect(tree.children).toHaveLength(1);
    expect(tree.children[0]).toHaveProperty("type", "steps");
    const stepsChildren = (tree.children[0] as { children: { type: string }[] })
      .children;
    expect(stepsChildren).toHaveLength(2);
    expect(stepsChildren[0].type).toBe("step");
    expect(stepsChildren[1].type).toBe("step");
  });
});
