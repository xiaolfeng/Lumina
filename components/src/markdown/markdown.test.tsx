/** @vitest-environment jsdom */
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { MarkdownLite } from "./markdown-lite";

afterEach(() => {
  cleanup();
});

describe("MarkdownLite 渲染", () => {
  it("相邻行围栏（开栏后无空行）渲染为 callout 卡片", () => {
    const { container } = render(
      <MarkdownLite>
        {':::callout{type="info"}\n**核心思想**：先粗筛再精排。\n:::'}
      </MarkdownLite>,
    );

    const callout = container.querySelector("div.my-4");
    expect(callout).toBeTruthy();
    expect(callout?.className).toContain("border-lagoon/20");
    expect(screen.getByText("核心思想")).toBeTruthy();
    expect(container.textContent).not.toContain(":::");
  });

  it("表格包装在横向滚动容器内，行内代码不逐字符断行、过滤 node 属性防 DOM 污染并收敛外边距", () => {
    const { container } = render(
      <MarkdownLite>
        {
          "| 脚本 | 输出 |\n| --- | --- |\n| `02_gen_query_opt.py` | `law_names` |"
        }
      </MarkdownLite>,
    );

    const table = container.querySelector("table");
    expect(table).toBeTruthy();
    expect(table?.className).toContain("my-0!");
    // 杜绝 react-markdown 注入 node hast 对象造成 DOM 污染
    expect(table?.getAttribute("node")).toBeNull();
    const wrapper = table?.parentElement;
    expect(wrapper?.className).toContain("w-full");
    expect(wrapper?.className).toContain("overflow-x-auto");
    expect(wrapper?.className).toContain("my-4");
    expect(wrapper?.className).toContain("first:mt-0");
    expect(wrapper?.className).toContain("last:mb-0");
    expect(wrapper?.className).not.toContain("break-all");
    expect(screen.getByText("02_gen_query_opt.py")).toBeTruthy();
  });

  it("无空行嵌套 :::steps 与 :::step 正确渲染，不触发空标签崩溃", () => {
    const raw = [
      ":::steps",
      ":::step",
      "执行数据召回分析",
      ":::",
      ":::step",
      "执行二次排序验证",
      ":::",
      ":::",
    ].join("\n");

    const { container } = render(<MarkdownLite>{raw}</MarkdownLite>);

    expect(container.querySelector("ol")).toBeTruthy();
    expect(container.querySelectorAll("li").length).toBe(2);
    expect(screen.getByText("执行数据召回分析")).toBeTruthy();
    expect(screen.getByText("执行二次排序验证")).toBeTruthy();
    expect(container.textContent).not.toContain(":::");
  });

  it("连续相邻无空行围栏分别正确渲染为独立组件", () => {
    const raw = [
      ':::callout{type="info"}',
      "第一个提示卡片",
      ":::",
      ':::callout{type="warning"}',
      "第二个警告卡片",
      ":::",
    ].join("\n");

    const { container } = render(<MarkdownLite>{raw}</MarkdownLite>);

    expect(screen.getByText("第一个提示卡片")).toBeTruthy();
    expect(screen.getByText("第二个警告卡片")).toBeTruthy();
    const callouts = container.querySelectorAll("div.my-4");
    expect(callouts.length).toBe(2);
    expect(container.textContent).not.toContain(":::");
  });

  it("非标准开栏语法或未闭合开栏头安全回退为普通文本，无空标签白屏", () => {
    const raw = [
      ':::unknownComponent{foo="bar"}',
      "非标准组件内容",
      ":::",
      ":::step{unclosedBracket",
      "语法错误的内容",
    ].join("\n");

    const { container } = render(<MarkdownLite>{raw}</MarkdownLite>);

    expect(screen.getByText(/:::unknownComponent/)).toBeTruthy();
    expect(screen.getByText(/:::step{unclosedBracket/)).toBeTruthy();
    // 确保没有生成空标签 node
    expect(
      container.querySelector(":not([class])")?.tagName.toLowerCase(),
    ).not.toBe("");
  });

  it("列表尾行紧贴关栏（CommonMark 惰性延续）时外层围栏正常闭合", () => {
    const raw = [
      ':::callout{type="info"}',
      "* 第一项列表",
      "* 第二项列表",
      ":::",
    ].join("\n");

    const { container } = render(<MarkdownLite>{raw}</MarkdownLite>);

    const callout = container.querySelector("div.my-4");
    expect(callout).toBeTruthy();
    expect(screen.getByText("第一项列表")).toBeTruthy();
    expect(screen.getByText("第二项列表")).toBeTruthy();
    expect(container.textContent).not.toContain(":::");
  });

  it("引用块尾行紧贴关栏时外层围栏正常闭合", () => {
    const raw = [':::callout{type="info"}', "> 引用块正文内容", ":::"].join(
      "\n",
    );

    const { container } = render(<MarkdownLite>{raw}</MarkdownLite>);

    const callout = container.querySelector("div.my-4");
    expect(callout).toBeTruthy();
    expect(screen.getByText("引用块正文内容")).toBeTruthy();
    expect(container.textContent).not.toContain(":::");
  });

  it("card 带有 URL 属性不被 remark-gfm autolink literal 破坏，正确渲染卡片组件", () => {
    const raw = [
      ':::card{href="https://example.com/docs" title="开发文档"}',
      "点击跳转到外部文档",
      ":::",
    ].join("\n");

    const { container } = render(<MarkdownLite>{raw}</MarkdownLite>);

    expect(screen.getByText("开发文档")).toBeTruthy();
    expect(screen.getByText("点击跳转到外部文档")).toBeTruthy();
    const anchor = container.querySelector("a");
    expect(anchor).toBeTruthy();
    expect(anchor?.getAttribute("href")).toBe("https://example.com/docs");
    expect(container.textContent).not.toContain(":::");
  });

  it("围栏属性支持嵌套花括号与单双引号混合使用", () => {
    const raw = [
      ':::card{title="模板 {name}" quote=\'He said "ok"\' href="https://example.com"}',
      "带特殊符号属性的内容",
      ":::",
    ].join("\n");

    const { container } = render(<MarkdownLite>{raw}</MarkdownLite>);

    expect(screen.getByText("模板 {name}")).toBeTruthy();
    expect(screen.getByText("带特殊符号属性的内容")).toBeTruthy();
    expect(container.textContent).not.toContain(":::");
  });

  it("列表尾行紧贴关栏且后方紧跟连续 steps 时不吞行漏提，外层步骤正常闭合", () => {
    const raw = [
      ":::steps",
      ":::step",
      "* 第一步列表 A",
      "* 第一步列表 B",
      ":::",
      ":::step",
      "第二步普通正文",
      ":::",
      ":::",
    ].join("\n");

    const { container } = render(<MarkdownLite>{raw}</MarkdownLite>);

    expect(screen.getByText("第一步列表 A")).toBeTruthy();
    expect(screen.getByText("第一步列表 B")).toBeTruthy();
    expect(screen.getByText("第二步普通正文")).toBeTruthy();
    // 验证外层 steps 包含 2 个 step (ol > li)，且第一步内包含独立的 2 个无序列表项 (ul > li)
    expect(container.querySelectorAll("ol > li").length).toBe(2);
    expect(container.querySelectorAll("ul > li").length).toBe(2);
    expect(container.textContent).not.toContain(":::");
  });

  it("列表尾行紧贴关栏且后方紧跟普通段落时不吞行，普通段落正常作为兄弟节点渲染", () => {
    const raw = [
      ':::callout{type="info"}',
      "* 列表条目 1",
      "* 列表条目 2",
      ":::",
      "紧随其后的外层独立段落",
    ].join("\n");

    const { container } = render(<MarkdownLite>{raw}</MarkdownLite>);

    const callout = container.querySelector("div.my-4");
    expect(callout).toBeTruthy();
    expect(screen.getByText("列表条目 1")).toBeTruthy();
    expect(screen.getByText("列表条目 2")).toBeTruthy();
    expect(screen.getByText("紧随其后的外层独立段落")).toBeTruthy();
    expect(container.textContent).not.toContain(":::");
  });

  it("带编号 KaTeX 块级公式正常渲染且包含 tag 编号", async () => {
    const raw = "$$\nE = mc^2 \\tag{1}\n$$";

    const { container, findByText } = render(
      <MarkdownLite>{raw}</MarkdownLite>,
    );

    // KaTeX rehype 插件为异步动态加载，等待渲染完成
    const tagOpenParen = await findByText("(");
    expect(tagOpenParen).toBeTruthy();
    const katexDisplay = container.querySelector(".katex-display");
    expect(katexDisplay).toBeTruthy();
    const tag = container.querySelector(".tag");
    expect(tag).toBeTruthy();
  });

  it("多行推导 aligned 环境带 tag 编号正常渲染，保持流式排版与垂直居中", async () => {
    const raw = `$$
\\begin{aligned}
f(x) &= (x + 1)^2 \\\\
     &= x^2 + 2x + 1
\\end{aligned} \\tag{Eq.2}
$$`;

    const { container, findByText } = render(
      <MarkdownLite>{raw}</MarkdownLite>,
    );

    const tagText = await findByText("Eq.2");
    expect(tagText).toBeTruthy();
    const tag = container.querySelector(".tag");
    expect(tag).toBeTruthy();
    const katexDisplay = container.querySelector(".katex-display");
    expect(katexDisplay).toBeTruthy();
  });

  it("超复杂高维矩阵 pmatrix 块级公式正常渲染且不破坏页面版面", async () => {
    const raw = `$$
A = \\begin{pmatrix}
a_{11} & a_{12} & a_{13} \\\\
a_{21} & a_{22} & a_{23} \\\\
a_{31} & a_{32} & a_{33}
\\end{pmatrix}
$$`;

    const { container, findAllByText } = render(
      <MarkdownLite>{raw}</MarkdownLite>,
    );

    const matrixCells = await findAllByText("a");
    expect(matrixCells.length).toBeGreaterThan(0);
    const katexDisplay = container.querySelector(".katex-display");
    expect(katexDisplay).toBeTruthy();
  });
});
