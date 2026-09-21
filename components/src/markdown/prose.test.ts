import { describe, expect, it } from "vitest";
import { proseQuestion, proseHint, proseArticle } from "./prose";

describe("prose styles overflow and word-break rules", () => {
  it("proseQuestion contains break-words, overflow-wrap, code break and pre overflow rules", () => {
    expect(proseQuestion).toContain("break-words");
    expect(proseQuestion).toContain("[overflow-wrap:anywhere]");
    expect(proseQuestion).toContain("[&_code]:break-all");
    expect(proseQuestion).toContain("[&_pre]:overflow-x-auto");
  });

  it("proseHint contains break-words and overflow-wrap rules", () => {
    expect(proseHint).toContain("break-words");
    expect(proseHint).toContain("[overflow-wrap:anywhere]");
  });

  it("proseArticle contains break-words and pre overflow rules", () => {
    expect(proseArticle).toContain("break-words");
    expect(proseArticle).toContain("[&_pre]:overflow-x-auto");
  });
});
