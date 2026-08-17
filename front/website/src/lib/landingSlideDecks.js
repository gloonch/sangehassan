const managedSlides = (sections, key, resolveImageUrl) => {
  const section = sections.find((item) => item.key === key);
  return (section?.images || [])
    .filter(Boolean)
    .map((src) => ({ src: resolveImageUrl(src), managed: true }));
};

const defaultDeck = (slides, fallbackSlide, shuffleSlides) => {
  if (!slides.length) return [fallbackSlide];
  return [slides[0], ...shuffleSlides(slides.slice(1))];
};

export const buildLandingSlideDecks = ({
  sections = [],
  productSlides = [],
  blockSlides = [],
  fallbackSlide,
  resolveImageUrl = (value) => value,
  shuffleSlides = (values) => values
}) => {
  const products = managedSlides(sections, "finished", resolveImageUrl);
  const blocks = managedSlides(sections, "blocks", resolveImageUrl);
  return {
    products: products.length ? products : defaultDeck(productSlides, fallbackSlide, shuffleSlides),
    blocks: blocks.length ? blocks : defaultDeck(blockSlides, fallbackSlide, shuffleSlides)
  };
};
