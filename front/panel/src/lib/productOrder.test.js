import { describe, expect, it } from "vitest";
import { moveProduct } from "./productOrder";

const products = [
  { id: 1, display_order: 0 },
  { id: 2, display_order: 1 },
  { id: 3, display_order: 2 },
  { id: 4, display_order: 3 }
];

describe("moveProduct", () => {
  it("moves a product before another product and renumbers the full list", () => {
    const result = moveProduct(products, 4, 2, "before");
    expect(result.map((product) => product.id)).toEqual([1, 4, 2, 3]);
    expect(result.map((product) => product.display_order)).toEqual([0, 1, 2, 3]);
  });

  it("moves a product after another product", () => {
    const result = moveProduct(products, 1, 3, "after");
    expect(result.map((product) => product.id)).toEqual([2, 3, 1, 4]);
  });

  it("keeps the original reference for an invalid move", () => {
    expect(moveProduct(products, 2, 2)).toBe(products);
    expect(moveProduct(products, 99, 2)).toBe(products);
  });
});
