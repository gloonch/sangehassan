export const catalogLocales = ["en", "fa", "ar"];

export const catalogLocaleConfig = {
  en: { locale: "en_US", numberLocale: "en-US", dir: "ltr" },
  fa: { locale: "fa_IR", numberLocale: "fa-IR", dir: "rtl" },
  ar: { locale: "ar_SA", numberLocale: "ar-SA", dir: "rtl" }
};

export const catalogCopy = {
  en: {
    home: "Home", products: "Products", productCount: "products", emptyCategories: "No active categories are available.",
    loading: "Loading products...", notFound: "The requested page was not found", back: "Back to product categories",
    searchLabel: "Search products", searchPlaceholder: "Product name or code", search: "Search", filters: "Filters",
    removeSearch: "Remove search", clearAll: "Clear all", emptyProducts: "No products match these filters.",
    loadMore: "Load more products", loadingMore: "Loading...", guide: "Selection guide", related: "Other stone categories", relatedProjects: "Related projects", project: "Project",
    close: "Close", noImage: "No image", breadcrumb: "Breadcrumb", popularBadge: "Popular", offerLabel: "Offer"
  },
  fa: {
    home: "خانه", products: "محصولات", productCount: "محصول", emptyCategories: "دسته‌بندی فعالی برای نمایش وجود ندارد.",
    loading: "در حال بارگذاری محصولات...", notFound: "صفحه موردنظر پیدا نشد", back: "بازگشت به دسته‌بندی محصولات",
    searchLabel: "جستجوی محصول", searchPlaceholder: "نام یا کد محصول", search: "جستجو", filters: "فیلترها",
    removeSearch: "حذف جستجو", clearAll: "پاک کردن همه", emptyProducts: "محصولی مطابق این فیلترها پیدا نشد.",
    loadMore: "نمایش محصولات بیشتر", loadingMore: "در حال بارگذاری...", guide: "راهنمای انتخاب", related: "سایر دسته‌بندی‌های سنگ", relatedProjects: "پروژه‌های مرتبط", project: "پروژه",
    close: "بستن", noImage: "بدون تصویر", breadcrumb: "مسیر صفحه", popularBadge: "محبوب", offerLabel: "Offer"
  },
  ar: {
    home: "الرئيسية", products: "المنتجات", productCount: "منتج", emptyCategories: "لا توجد فئات نشطة للعرض.",
    loading: "جار تحميل المنتجات...", notFound: "لم يتم العثور على الصفحة المطلوبة", back: "العودة إلى فئات المنتجات",
    searchLabel: "البحث عن منتج", searchPlaceholder: "اسم المنتج أو رمزه", search: "بحث", filters: "التصفية",
    removeSearch: "إزالة البحث", clearAll: "مسح الكل", emptyProducts: "لا توجد منتجات مطابقة لهذه التصفية.",
    loadMore: "عرض المزيد", loadingMore: "جار التحميل...", guide: "دليل الاختيار", related: "فئات حجر أخرى", relatedProjects: "مشاريع ذات صلة", project: "مشروع",
    close: "إغلاق", noImage: "لا توجد صورة", breadcrumb: "مسار الصفحة", popularBadge: "شائع", offerLabel: "Offer"
  }
};

export const catalogBasePath = (lang) => `/${catalogLocales.includes(lang) ? lang : "en"}/products`;

export const localizedField = (item, field, lang) => {
  if (!item) return "";
  return item[`${field}_${lang}`] || item[`${field}_en`] || item[`${field}_fa`] || item[`${field}_ar`] || item[field] || "";
};

export const catalogAlternates = (suffix = "") => [
  { lang: "en", path: `/en/products${suffix}` },
  { lang: "fa", path: `/fa/products${suffix}` },
  { lang: "ar", path: `/ar/products${suffix}` },
  { lang: "x-default", path: `/en/products${suffix}` }
];
