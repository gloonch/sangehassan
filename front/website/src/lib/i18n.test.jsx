import { describe, expect, it } from "vitest";
import { DEFAULT_LANGUAGE, getLanguageFromPath } from "./i18n";

describe("website language selection", () => {
  it("uses English for routes without an explicit locale", () => {
    expect(DEFAULT_LANGUAGE).toBe("en");
    expect(getLanguageFromPath("/")).toBe("en");
    expect(getLanguageFromPath("/products")).toBe("en");
    expect(getLanguageFromPath("/blogs")).toBe("en");
  });

  it("keeps the locale declared by localized routes", () => {
    expect(getLanguageFromPath("/en/products")).toBe("en");
    expect(getLanguageFromPath("/fa/blogs/example")).toBe("fa");
    expect(getLanguageFromPath("/ar/products/travertine")).toBe("ar");
  });
});
