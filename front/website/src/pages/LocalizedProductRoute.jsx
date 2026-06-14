import { useEffect, useState } from "react";
import { useLocation, useParams } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { usePrerenderData } from "../lib/prerenderData";
import { useTranslation } from "../lib/i18n";
import ProductCatalog from "./ProductCatalog";
import ProductDetail from "./ProductDetail";

export default function LocalizedProductRoute() {
  const { slug } = useParams();
  const { lang, t } = useTranslation();
  const location = useLocation();
  const prerenderedProduct = usePrerenderData("product");
  const prerenderedCatalog = usePrerenderData("catalogPage");
  const prerenderedKind = prerenderedProduct?.slug === slug
    ? "product"
    : prerenderedCatalog?.category?.slug === slug ? "category" : "";
  const stateKind = location.state?.catalogRouteKind || "";
  const [kind, setKind] = useState(prerenderedKind || stateKind);
  const [catalogPage, setCatalogPage] = useState(
    prerenderedKind === "category" ? prerenderedCatalog : null
  );

  useEffect(() => {
    if (prerenderedKind || stateKind) {
      setKind(prerenderedKind || stateKind);
      setCatalogPage(prerenderedKind === "category" ? prerenderedCatalog : null);
      return undefined;
    }

    let active = true;
    setKind("");
    setCatalogPage(null);
    fetchJSON(`/api/catalog/categories/${slug}?locale=${lang}&limit=24&offset=0`)
      .then((response) => {
        if (!active) return;
        setCatalogPage(response.data || null);
        setKind("category");
      })
      .catch(() => {
        if (active) setKind("product");
      });

    return () => {
      active = false;
    };
  }, [lang, prerenderedCatalog, prerenderedKind, slug, stateKind]);

  if (kind === "category") {
    return <ProductCatalog categorySlugOverride={slug} initialPageOverride={catalogPage} />;
  }
  if (kind === "product") {
    return <ProductDetail />;
  }
  return <div className="section-shell min-h-[55vh] py-20 text-center text-sm text-primary/60">{t("messages.loading")}</div>;
}
