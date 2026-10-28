import { describe, expect, it } from "vitest";
import { proseQuestion, proseHint, proseArticle } from "./prose";

describe("prose styles overflow and word-break rules", () => {
  it("proseQuestion contains break-words, table header nowrap, code break and pre overflow rules", () => {
    expect(proseQuestion).toContain("break-words");
    expect(proseQuestion).toContain("[&_th]:whitespace-nowrap");
    expect(proseQuestion).toContain("[&_code]:break-words");
    expect(proseQuestion).not.toContain("[&_code]:break-all");
    expect(proseQuestion).toContain("[&_pre]:overflow-x-auto");
  });

  it("proseHint contains break-words and table header nowrap rules", () => {
    expect(proseHint).toContain("break-words");
    expect(proseHint).toContain("[&_th]:whitespace-nowrap");
    expect(proseHint).not.toContain("[&_code]:break-all");
  });

  it("proseArticle contains break-words, table header nowrap and pre overflow rules", () => {
    expect(proseArticle).toContain("break-words");
    expect(proseArticle).toContain("[&_th]:whitespace-nowrap");
    expect(proseArticle).toContain("[&_code]:break-words");
    expect(proseArticle).not.toContain("[&_code]:break-all");
    expect(proseArticle).toContain("[&_pre]:overflow-x-auto");
  });

  it("proseArticle preserves core theme variables to prevent dark mode contrast regression", () => {
    expect(proseArticle).toContain("[&_table]:w-full");
    expect(proseArticle).toContain("[&_table]:text-sm");
    expect(proseArticle).toContain("[&_th]:bg-foam");
    expect(proseArticle).toContain("[&_th]:text-sea-ink");
    expect(proseArticle).toContain("[&_td]:text-sea-ink-soft");
    expect(proseArticle).toContain("[&_td]:border-line");
    expect(proseArticle).toContain("[&_hr]:border-line");
    expect(proseArticle).toContain("[&_ul]:text-sea-ink");
    expect(proseArticle).toContain("[&_ol]:text-sea-ink");
    expect(proseArticle).toContain("[&_li]:text-sea-ink");
    expect(proseArticle).toContain("[&_em]:text-sea-ink-soft");
  });
});
