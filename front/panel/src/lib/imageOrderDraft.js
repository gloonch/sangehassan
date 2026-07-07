export const createImageOrderDraft = (baseImages = [], currentImages = []) => {
  const originalIndexes = new Map();

  baseImages.forEach((url, index) => {
    if (!originalIndexes.has(url)) {
      originalIndexes.set(url, index);
    }
  });

  return {
    type: "image_order",
    updated_at: new Date().toISOString(),
    original: baseImages.map((url, index) => ({
      url,
      index
    })),
    current: currentImages.map((url, index) => ({
      url,
      index,
      original_index: originalIndexes.has(url) ? originalIndexes.get(url) : null
    }))
  };
};

export const moveImage = (images = [], index, direction) => {
  const targetIndex = index + direction;
  if (targetIndex < 0 || targetIndex >= images.length) {
    return images;
  }

  const nextImages = [...images];
  [nextImages[index], nextImages[targetIndex]] = [nextImages[targetIndex], nextImages[index]];
  return nextImages;
};

export const selectedIndexAfterImageMove = (selectedIndex, fromIndex, direction) => {
  const targetIndex = fromIndex + direction;
  if (selectedIndex === fromIndex) return targetIndex;
  if (selectedIndex === targetIndex) return fromIndex;
  return selectedIndex;
};
