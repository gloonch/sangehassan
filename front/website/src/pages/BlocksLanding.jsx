import { useEffect } from "react";
import BlocksCatalog from "./BlocksCatalog";
import { useTranslation } from "../lib/i18n";
import { withIranAccessSeoNotice } from "../lib/seo";

const blocksSeoContent = {
  fa: {
    title: "کوپ سنگ | سنگ حسن",
    description:
      "مشاهده کوپ‌ها و بلوک‌های سنگ طبیعی سنگ حسن برای تامین پروژه‌های ساختمانی، تولید اسلب و همکاری عمده.",
    locale: "fa_IR"
  },
  en: {
    title: "Stone Blocks | SangeHassan",
    description:
      "Browse SangeHassan natural stone blocks for project supply, slab production, and wholesale B2B sourcing.",
    locale: "en_US"
  },
  ar: {
    title: "بلوكات الحجر | سانج حسن",
    description:
      "تصفح بلوكات الحجر الطبيعي من سانج حسن لتوريد المشاريع وإنتاج الألواح والتعاون بالجملة.",
    locale: "ar_SA"
  }
};

export default function BlocksLanding() {
  const { lang } = useTranslation();

  useEffect(() => {
    if (typeof window === "undefined" || typeof document === "undefined") return;

    const seo = withIranAccessSeoNotice(blocksSeoContent[lang] || blocksSeoContent.fa);
    const pageUrl = `${window.location.origin}/blocks`;
    const previousTitle = document.title;
    const previousDescription = document.head.querySelector('meta[name="description"]')?.getAttribute("content") ?? null;

    let description = document.head.querySelector('meta[name="description"]');
    if (!description) {
      description = document.createElement("meta");
      description.setAttribute("name", "description");
      document.head.appendChild(description);
    }

    let canonical = document.head.querySelector('link[rel="canonical"]');
    const createdCanonical = !canonical;
    const previousCanonical = canonical?.getAttribute("href") ?? null;
    if (!canonical) {
      canonical = document.createElement("link");
      canonical.setAttribute("rel", "canonical");
      document.head.appendChild(canonical);
    }

    document.title = seo.title;
    description.setAttribute("content", seo.description);
    canonical.setAttribute("href", pageUrl);

    return () => {
      document.title = previousTitle;
      if (previousDescription === null) description?.removeAttribute("content");
      else description?.setAttribute("content", previousDescription);

      if (createdCanonical) canonical?.remove();
      else if (previousCanonical === null) canonical?.removeAttribute("href");
      else canonical?.setAttribute("href", previousCanonical);
    };
  }, [lang]);

  return (
    <div className="bg-quiet-grid">
      <BlocksCatalog embedded />
    </div>
  );
}
