import { render, screen } from "@testing-library/react";
import React from "react";
import { describe, expect, it } from "vitest";
import { LpwDocumentViewer, lpwRegistry, registerAll } from "./index";

describe("LPW registerAll 对齐测试", () => {
  it("应准确注册全部 31 种组件（1 种 Layout + 4 种 Container + 26 种 Block），且无 columns", () => {
    lpwRegistry.clear();
    expect(lpwRegistry.entries()).toHaveLength(0);

    registerAll();
    const all = lpwRegistry.entries();
    expect(all).toHaveLength(31);

    // Layout
    const layouts = lpwRegistry.entries("layout");
    expect(layouts).toHaveLength(1);
    expect(layouts[0].displayName).toBe("Layout");

    // Containers
    const containers = lpwRegistry.entries("container");
    expect(containers).toHaveLength(4);
    const containerTypes = containers.map((c) => c.displayName);
    expect(containerTypes).toContain("Section");
    expect(containerTypes).toContain("Panel");
    expect(containerTypes).toContain("Details");
    expect(containerTypes).toContain("Tabs");
    expect(containerTypes).not.toContain("Columns");

    // Blocks
    const blocks = lpwRegistry.entries("block");
    expect(blocks).toHaveLength(26);

    // columns 不在任何类别中
    expect(lpwRegistry.get("container", "columns")).toBeUndefined();
  });

  it("渲染时若未预先手动调用 registerAll，应具备自愈能力并正确渲染 section 与 markdown 节点", () => {
    lpwRegistry.clear();
    const source = JSON.stringify({
      version: "1.1",
      content: [
        {
          id: "s1",
          kind: "container",
          type: "section",
          props: { variant: "article", title: "章节测试" },
          children: [
            {
              id: "m1",
              kind: "block",
              type: "markdown",
              props: { content: "正文测试" },
            },
          ],
        },
      ],
    });
    const { container } = render(
      React.createElement(LpwDocumentViewer, { source }),
    );
    expect(container.textContent).not.toContain(
      "未注册的 container 类型: section",
    );
    expect(container.textContent).not.toContain("UNKNOWN_BLOCK");
    expect(container.textContent).toContain("章节测试");
    expect(container.textContent).toContain("正文测试");
  });
});
