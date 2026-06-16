const IMAGE_BASE = import.meta.env.VITE_IMAGE_BASE_URL || "";
const PROTECTED_IMAGE_REV = import.meta.env.VITE_PROTECTED_IMAGE_REV || "protected-2026-06-09-r3";

const toProtectedImagePath = (url) => {
  if (!url.startsWith("/images/")) return url;
  return url.replace(/^\/images\//, "/protected-images/");
};

const appendProtectedImageRev = (url) => {
  if (!url.startsWith("/protected-images/")) return url;

  const [urlWithoutHash, hash = ""] = url.split("#");
  const [pathname, query = ""] = urlWithoutHash.split("?");
  const params = new URLSearchParams(query);
  params.set("pv", PROTECTED_IMAGE_REV);
  const nextQuery = params.toString();
  return `${pathname}${nextQuery ? `?${nextQuery}` : ""}${hash ? `#${hash}` : ""}`;
};

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

export const resolveProtectedImageUrl = (url) => {
  if (!url) return "";
  if (url.startsWith("http")) return url;
  return `${IMAGE_BASE}${appendProtectedImageRev(toProtectedImagePath(url))}`;
};

export const resolveProtectedThumbnailUrl = (url) => {
  const protectedUrl = resolveProtectedImageUrl(url);
  if (!protectedUrl) return "";
  const separator = protectedUrl.includes("?") ? "&" : "?";
  return `${protectedUrl}${separator}variant=thumb`;
};
