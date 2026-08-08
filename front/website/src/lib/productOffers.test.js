import { describe, expect, it } from "vitest";
import { formatOfferPrice, getProductOfferStructuredData, getProductSku } from "./productOffers";

describe("product offers", () => {
  it("formats a starting square-metre price", () => {
    expect(formatOfferPrice(450000, "fa")).toContain("۴۵۰٬۰۰۰ تومان به‌ازای هر مترمربع");
  });

  it("converts Toman to IRR in Product offer schema", () => {
    const offer = getProductOfferStructuredData({ is_popular: true, price: 450000 }, "https://sangehassan.com/fa/products/sample");
    expect(offer).toMatchObject({
      "@type": "Offer",
      price: "4500000",
      priceCurrency: "IRR",
      itemCondition: "https://schema.org/NewCondition",
      seller: {
        "@type": "Organization",
        name: "SangeHassan"
      },
      url: "https://sangehassan.com/fa/products/sample",
      priceSpecification: {
        unitCode: "MTK",
        price: "4500000"
      }
    });
  });

  it("uses the stable product identifier as sku", () => {
    expect(getProductSku({ id: 1431, slug: "abbasabad-travertine" })).toBe("1431");
    expect(getProductSku({ slug: "abbasabad-travertine" })).toBe("abbasabad-travertine");
  });
});
