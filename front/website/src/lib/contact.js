const toLatinDigits = (value) =>
  String(value || "")
    .replace(/[۰-۹]/g, (digit) => String("۰۱۲۳۴۵۶۷۸۹".indexOf(digit)))
    .replace(/[٠-٩]/g, (digit) => String("٠١٢٣٤٥٦٧٨٩".indexOf(digit)));

export const normalizePhone = (value) => {
  if (!value || typeof value !== "string") return "";
  const cleaned = toLatinDigits(value).replace(/[^\d+]/g, "");
  if (!cleaned) return "";
  if (cleaned.startsWith("+")) {
    return `+${cleaned.slice(1).replace(/\+/g, "")}`;
  }
  return cleaned.replace(/\+/g, "");
};

export const getContactPhoneItems = (primaryPhone) => {
  const seen = new Set();
  return [primaryPhone, "09121193835", "09121193935"]
    .map((value) => ({ value, normalized: normalizePhone(value) }))
    .filter((item) => {
      if (!item.normalized || seen.has(item.normalized)) return false;
      seen.add(item.normalized);
      return true;
    });
};

export const getContactHref = (key, value) => {
  if (!value || typeof value !== "string") return "";
  if (key === "phone") {
    const normalized = value.startsWith("tel:") ? value.slice(4) : normalizePhone(value);
    return normalized ? `tel:${normalized}` : "";
  }
  if (value.startsWith("http://") || value.startsWith("https://")) return value;
  const lower = value.toLowerCase();
  if (
    lower.includes("linkedin.com") ||
    lower.includes("instagram.com") ||
    lower.includes("t.me") ||
    lower.includes("telegram.me")
  ) {
    return `https://${value.replace(/^https?:\/\//i, "")}`;
  }
  return "";
};
