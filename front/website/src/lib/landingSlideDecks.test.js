import { describe, expect, it } from "vitest";
import { buildLandingSlideDecks } from "./landingSlideDecks";

describe("landing slideshow decks", () => {
  it("uses uploaded section images in their configured order", () => {
    const decks = buildLandingSlideDecks({
      sections: [
        { key: "finished", images: ["/images/content/first.webp", "/images/content/second.webp"] },
        { key: "blocks", images: ["/images/content/block.webp"] }
      ],
      productSlides: [{ src: "default-product" }],
      blockSlides: [{ src: "default-block" }],
      fallbackSlide: { src: "fallback" },
      resolveImageUrl: (value) => `https://cdn.example${value}`
    });

    expect(decks.products.map((slide) => slide.src)).toEqual([
      "https://cdn.example/images/content/first.webp",
      "https://cdn.example/images/content/second.webp"
    ]);
    expect(decks.blocks[0]).toMatchObject({
      src: "https://cdn.example/images/content/block.webp",
      managed: true
    });
  });

  it("keeps the existing default deck when no images were uploaded", () => {
    const decks = buildLandingSlideDecks({
      sections: [{ key: "finished", images: [] }],
      productSlides: [{ src: "first" }, { src: "second" }, { src: "third" }],
      blockSlides: [],
      fallbackSlide: { src: "fallback" },
      shuffleSlides: (values) => [...values].reverse()
    });

    expect(decks.products.map((slide) => slide.src)).toEqual(["first", "third", "second"]);
    expect(decks.blocks.map((slide) => slide.src)).toEqual(["fallback"]);
  });
});
