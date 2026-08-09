import { describe, expect, it } from "vitest";
import { DEFAULT_LANGUAGE, getLanguageFromPath } from "./i18n";

describe("website language selection", () => {
  it("uses Persian for routes without an explicit locale", () => {
    expect(DEFAULT_LANGUAGE).toBe("fa");
    expect(getLanguageFromPath("/")).toBe("fa");
    expect(getLanguageFromPath("/products")).toBe("fa");
    expect(getLanguageFromPath("/blogs")).toBe("fa");
  });

  it("keeps the locale declared by localized routes", () => {
    expect(getLanguageFromPath("/en/products")).toBe("en");
    expect(getLanguageFromPath("/fa/blogs/example")).toBe("fa");
    expect(getLanguageFromPath("/ar/products/travertine")).toBe("ar");
  });
});
