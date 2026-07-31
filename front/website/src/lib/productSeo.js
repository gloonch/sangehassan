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
  const category = localizedField(product?.category || product?.categories?.[0], "title", locale);
  const mines = productValues(product, "mines", "mines", locale);
  const finishes = productValues(product, "finishes", "finishes", locale);
  const variants = productValues(product, "use_case_form", "variants", locale);
  const manualHTML = (locale === "fa"
    ? product?.description_html_fa || product?.short_description_html_fa
    : locale === "ar"
      ? product?.description_html_ar || product?.short_description_html_ar
      : product?.description_html_en || product?.short_description_html_en) ||
    product?.description_html || product?.short_description_html || product?.description || "";

  const normalizedTitle = title.toLocaleLowerCase();
  const detail = [category, mines[0], finishes[0]].filter((value, index, values) =>
    value && !normalizedTitle.includes(String(value).toLocaleLowerCase()) && values.indexOf(value) === index
  ).slice(0, 2).join(" - ");

  let generatedSummary;
  let seoTitle;
  if (locale === "fa") {
    generatedSummary = `${title}${category ? ` از دسته ${category}` : "، محصول سنگ طبیعی"}${mines.length ? ` با معدن ${mines.slice(0, 2).join(" و ")}` : ""}${finishes.length ? ` و فرآوری ${finishes.slice(0, 2).join(" و ")}` : ""}. تصاویر، مشخصات و گزینه‌های مناسب خرید و بررسی پروژه را در سنگ حسن مشاهده کنید.`;
    seoTitle = `${title}${detail ? ` | ${detail}` : " | مشخصات و خرید"} | سنگ حسن`;
  } else if (locale === "ar") {
    generatedSummary = `${title}${category ? ` من فئة ${category}` : " من الحجر الطبيعي"}${mines.length ? ` من محاجر ${mines.slice(0, 2).join(" و")}` : ""}${finishes.length ? ` بتشطيب ${finishes.slice(0, 2).join(" و")}` : ""}. شاهد الصور والمواصفات وخيارات التوريد للمشاريع من سانج حسن.`;
    seoTitle = `${title}${detail ? ` | ${detail}` : " | المواصفات والتوريد"} | سانج حسن`;
  } else {
    generatedSummary = `${title}${category ? ` is a ${category} product` : " is a natural stone product"}${mines.length ? ` sourced from ${mines.slice(0, 2).join(" and ")}` : ""}${finishes.length ? ` and available in ${finishes.slice(0, 2).join(" and ")} finishes` : ""}. View images, specifications, and project sourcing options from SangeHassan.`;
    seoTitle = `${title}${detail ? ` | ${detail}` : " | Specifications & Supply"} | SangeHassan`;
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
  const labels = locale === "fa"
    ? { mine: "معدن", finish: "فرآوری", form: "فرم عرضه" }
    : locale === "ar"
      ? { mine: "المحجر", finish: "التشطيب", form: "شكل التوريد" }
      : { mine: "Quarry", finish: "Finish", form: "Supply form" };
  return [
    [labels.mine, seo.mines],
    [labels.finish, seo.finishes],
    [labels.form, seo.variants]
  ].filter(([, values]) => values.length).map(([name, values]) => ({
    "@type": "PropertyValue",
    name,
    value: values.join(", ")
  }));
}
