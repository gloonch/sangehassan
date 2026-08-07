import { describe, expect, it } from "vitest";
import { getLanguageTarget } from "./languageRoutes";

describe("language switch targets", () => {
  it("links blog hubs to real locale URLs", () => {
    expect(getLanguageTarget({ pathname: "/fa/blogs", nextLang: "en" })).toBe("/en/blogs");
  });

  it("links blog pagination to the target locale hub", () => {
    expect(getLanguageTarget({ pathname: "/fa/blogs/page/2", nextLang: "ar" })).toBe("/ar/blogs");
  });

  it("links a translated article to its published translation", () => {
    expect(getLanguageTarget({
      pathname: "/fa/blogs/persian-slug",
      nextLang: "en",
      blogAlternates: { en: "/en/blogs/english-slug" }
    })).toBe("/en/blogs/english-slug");
  });

  it("links a missing article translation to that language hub", () => {
    expect(getLanguageTarget({
      pathname: "/fa/blogs/persian-only",
      nextLang: "ar",
      blogAlternates: { fa: "/fa/blogs/persian-only" }
    })).toBe("/ar/blogs");
  });

  it("preserves catalog suffixes and query strings", () => {
    expect(getLanguageTarget({ pathname: "/fa/products/travertine", search: "?sort=price", nextLang: "en" }))
      .toBe("/en/products/travertine?sort=price");
  });
});
