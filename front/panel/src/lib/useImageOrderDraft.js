import { useState } from "react";
import { createImageOrderDraft } from "./imageOrderDraft";

export const useImageOrderDraft = () => {
  const [imageOrderDraft, setImageOrderDraft] = useState(null);
  const [baseImages, setBaseImages] = useState([]);

  const stageImageOrder = (previousImages = [], nextImages = []) => {
    const base = imageOrderDraft ? baseImages : previousImages;
    if (!imageOrderDraft) {
      setBaseImages(base);
    }
    setImageOrderDraft(createImageOrderDraft(base, nextImages));
  };

  const resetImageOrderDraft = (nextBaseImages = []) => {
    setBaseImages(nextBaseImages);
    setImageOrderDraft(null);
  };

  return {
    imageOrderDraft,
    stageImageOrder,
    resetImageOrderDraft
  };
};
