const stripHTML = (value = "") => String(value)
  .replace(/<[^>]*>/g, " ")
  .replace(/&nbsp;/g, " ")
  .replace(/&amp;/g, "&")
  .replace(/\s+/g, " ")
  .trim();

const localizedField = (item, field, locale) =>
  item?.[`${field}_${locale}`] || item?.[`${field}_en`] || item?.[`${field}_fa`] || item?.[`${field}_ar`] || item?.[field] || "";

const localizedTermValues = (product, taxonomy, locale) => (product?.terms || [])
  .filter((term) => term.taxonomy === taxonomy)
  .map((term) => localizedField(term, "label", locale) || term.key)
  .filter(Boolean);

const productValues = (product, taxonomy, legacyField, locale) => {
  const terms = localizedTermValues(product, taxonomy, locale);
  return terms.length ? terms : (Array.isArray(product?.[legacyField]) ? product[legacyField].filter(Boolean) : []);
};

const clip = (value, limit) => {
  const text = stripHTML(value);
  if (text.length <= limit) return text;
  return `${text.slice(0, limit - 1).trim()}…`;
};

export function getProductSeo(product, locale = "en") {
  const title = localizedField(product, "title", locale) || product?.slug || "Natural stone";
  const englishTitle = String(product?.title_en || "").trim();
  const genericSlugReference = /^product-\d+$/i.test(product?.slug || "")
    ? String(product.slug).replace(/-/g, " ").replace(/^./, (letter) => letter.toUpperCase())
    : "";
  const identity = genericSlugReference || (
    locale !== "en" && englishTitle && englishTitle.toLocaleLowerCase() !== title.toLocaleLowerCase()
      ? englishTitle
      : ""
  );
  const qualifiedTitle = identity ? `${title} (${identity})` : title;
  const category = localizedField(product?.category || product?.categories?.[0], "title", locale);
  const mines = productValues(product, "mines", "mines", locale);
  const finishes = productValues(product, "finishes", "finishes", locale);
  const variants = productValues(product, "use_case_form", "variants", locale);
  const localizedManualHTML = locale === "fa"
    ? product?.short_description_html_fa || product?.description_html_fa
    : locale === "ar"
      ? product?.short_description_html_ar || product?.description_html_ar
      : product?.short_description_html_en || product?.description_html_en;
  const legacyManualHTML = locale === "en"
    ? product?.short_description_html || product?.description_html || product?.description
    : "";
  const manualHTML = localizedManualHTML || legacyManualHTML || "";

  const numericPrice = typeof product?.price === "number" ? product.price : Number(product?.price);
  const hasPrice = product?.is_popular && Number.isFinite(numericPrice) && numericPrice > 0;

  let generatedSummary;
  let seoTitle;
  if (locale === "fa") {
    generatedSummary = `${qualifiedTitle}${category ? ` از دسته ${category}` : "، محصول سنگ طبیعی"}${mines.length ? ` با معدن ${mines.slice(0, 2).join(" و ")}` : ""}${finishes.length ? ` و فرآوری ${finishes.slice(0, 2).join(" و ")}` : ""}. تصاویر، مشخصات و گزینه‌های مناسب خرید و بررسی پروژه را در سنگ حسن مشاهده کنید.`;
    seoTitle = hasPrice
      ? `قیمت و خرید ${qualifiedTitle} | سنگ حسن`
      : `${qualifiedTitle} | مشخصات و کاربرد | سنگ حسن`;
  } else if (locale === "ar") {
    generatedSummary = `${qualifiedTitle}${category ? ` من فئة ${category}` : " من الحجر الطبيعي"}${mines.length ? ` من محاجر ${mines.slice(0, 2).join(" و")}` : ""}${finishes.length ? ` بتشطيب ${finishes.slice(0, 2).join(" و")}` : ""}. شاهد الصور والمواصفات وخيارات التوريد للمشاريع من سانج حسن.`;
    seoTitle = hasPrice
      ? `سعر وشراء ${qualifiedTitle} | سانج حسن`
      : `${qualifiedTitle} | المواصفات والاستخدام | سانج حسن`;
  } else {
    generatedSummary = `${qualifiedTitle}${category ? ` is a ${category} product` : " is a natural stone product"}${mines.length ? ` sourced from ${mines.slice(0, 2).join(" and ")}` : ""}${finishes.length ? ` and available in ${finishes.slice(0, 2).join(" and ")} finishes` : ""}. View images, specifications, and project sourcing options from SangeHassan.`;
    seoTitle = hasPrice
      ? `${qualifiedTitle} Price & Supply | SangeHassan`
      : `${qualifiedTitle} | Specifications & Applications | SangeHassan`;
  }

  return {
    title,
    seoTitle: clip(seoTitle, 65),
    description: clip(manualHTML || generatedSummary, 160),
    summary: stripHTML(manualHTML) || generatedSummary,
    manualHTML,
    category,
    mines,
    finishes,
    variants
  };
}

export function productAdditionalProperties(product, locale = "en") {
  const seo = getProductSeo(product, locale);
  const applications = localizedTermValues(product, "use_case_application", locale);
  const labels = locale === "fa"
    ? { mine: "معدن", finish: "فرآوری", form: "فرم عرضه", application: "کاربرد" }
    : locale === "ar"
      ? { mine: "المحجر", finish: "التشطيب", form: "شكل التوريد", application: "الاستخدام" }
      : { mine: "Quarry", finish: "Finish", form: "Supply form", application: "Application" };
  return [
    [labels.mine, seo.mines],
    [labels.finish, seo.finishes],
    [labels.form, seo.variants],
    [labels.application, applications]
  ].filter(([, values]) => values.length).map(([name, values]) => ({
    "@type": "PropertyValue",
    name,
    value: values.join(", ")
  }));
}
