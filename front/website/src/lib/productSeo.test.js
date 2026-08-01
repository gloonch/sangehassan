import { describe, expect, it } from "vitest";
import { getProductSeo, productAdditionalProperties } from "./productSeo";

const product = {
  title_en: "Hassan Beige Travertine",
  title_fa: "تراورتن حسن",
  slug: "hassan-travertine",
  price: 450000,
  is_popular: true,
  short_description_html_fa: "<p>قیمت و خرید تراورتن حسن برای نما و اسلب.</p>",
  category: { title_en: "Travertine", title_fa: "تراورتن" },
  terms: [
    { taxonomy: "use_case_form", label_en: "Slab", label_fa: "اسلب" },
    { taxonomy: "use_case_application", label_en: "Facade", label_fa: "نما" }
  ]
};

describe("product SEO", () => {
  it("uses commercial intent for priced Persian products", () => {
    const seo = getProductSeo(product, "fa");
    expect(seo.seoTitle).toBe("قیمت و خرید تراورتن حسن (Hassan Beige Travertine) | سنگ حسن");
    expect(seo.description).toContain("قیمت و خرید تراورتن حسن");
  });

  it("includes supply form and application in structured properties", () => {
    expect(productAdditionalProperties(product, "fa")).toEqual(expect.arrayContaining([
      expect.objectContaining({ name: "فرم عرضه", value: "اسلب" }),
      expect.objectContaining({ name: "کاربرد", value: "نما" })
    ]));
  });
});
