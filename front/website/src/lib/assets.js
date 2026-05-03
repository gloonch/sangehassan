const IMAGE_BASE = import.meta.env.VITE_IMAGE_BASE_URL || "";

export const resolveImageUrl = (url) => {
  if (!url) return "";
  if (url.startsWith("http")) return url;
  return `${IMAGE_BASE}${url}`;
};
