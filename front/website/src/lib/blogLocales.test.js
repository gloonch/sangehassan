import { describe, expect, it } from "vitest";
import { blogHubAlternates, defaultBlogLocale, normalizeBlogLocales } from "./blogLocales";

describe("blog locale SEO helpers", () => {
  it("keeps only supported published locales", () => {
    expect(normalizeBlogLocales(["fa", "de", "fa", "ar"])).toEqual(["fa", "ar"]);
  });

  it("uses Persian as x-default when English has no published articles", () => {
    expect(defaultBlogLocale(["fa"])).toBe("fa");
    expect(blogHubAlternates(["fa"])).toEqual([
      { lang: "fa", path: "/fa/blogs" },
      { lang: "x-default", path: "/fa/blogs" }
    ]);
  });

  it("activates a locale and prefers English after its first translation is published", () => {
    expect(blogHubAlternates(["fa", "en"])).toEqual([
      { lang: "en", path: "/en/blogs" },
      { lang: "fa", path: "/fa/blogs" },
      { lang: "x-default", path: "/en/blogs" }
    ]);
  });
});
