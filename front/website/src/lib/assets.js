const IMAGE_BASE = import.meta.env.VITE_IMAGE_BASE_URL || "";

export const resolveImageUrl = (url) => {
  if (!url) return "";
  if (url.startsWith("http")) return url;
  return `${IMAGE_BASE}${url}`;
};

export const appendImageVersion = (url, version, param = "v") => {
  if (!url || !version) return url || "";
  const [urlWithoutHash, hash = ""] = String(url).split("#");
  const [pathname, query = ""] = urlWithoutHash.split("?");
  const params = new URLSearchParams(query);
  params.set(param, version);
  const nextQuery = params.toString();
  return `${pathname}${nextQuery ? `?${nextQuery}` : ""}${hash ? `#${hash}` : ""}`;
};

export const resolveVersionedImageUrl = (url, version, param = "v") => {
  return appendImageVersion(resolveImageUrl(url), version, param);
};
