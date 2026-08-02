import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";
import { formatListingQuantity, formatPriceValue, getListingCoverImageUrl, getListingProductTitle } from "../lib/listings";
import { resolveImageUrl } from "../lib/assets";
import { getCanonicalUrl, usePageSeo } from "../lib/seo";

const normalizeSearchText = (value) =>
  String(value || "")
    .normalize("NFKC")
    .replace(/\u200c/g, " ")
    .trim()
    .toLowerCase()
    .replace(/\s+/g, " ");

const getNumericValue = (value) => {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
};

const getCreatedTime = (ad) => {
  const time = Date.parse(ad?.created_at || "");
  return Number.isFinite(time) ? time : 0;
};

const getLocationValue = (ad) => [ad?.province, ad?.city].filter(Boolean).join(" / ");

const getFormLabel = (form, t) => {
  if (form === "block") return t("ads.formOptions.block");
  if (form === "finished") return t("ads.formOptions.finished");
  return form || "—";
};

const uniqueOptions = (items) => {
  const map = new Map();
  items.forEach((item) => {
    const value = String(item || "").trim();
    if (!value || map.has(value)) return;
    map.set(value, value);
  });
  return [...map.values()].sort((a, b) => a.localeCompare(b));
};

export default function Ads() {
  const { t, lang } = useTranslation();
  const navigate = useNavigate();
  const isRTL = lang === "fa" || lang === "ar";
  const gradientDir = isRTL ? "bg-gradient-to-tl" : "bg-gradient-to-tr";
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [searchInput, setSearchInput] = useState("");
  const [formFilter, setFormFilter] = useState("all");
  const [stoneTypeFilter, setStoneTypeFilter] = useState("all");
  const [locationFilter, setLocationFilter] = useState("all");
  const [priceStatusFilter, setPriceStatusFilter] = useState("all");
  const [minTonnageInput, setMinTonnageInput] = useState("");
  const [maxTonnageInput, setMaxTonnageInput] = useState("");
  const [sortMode, setSortMode] = useState("newest");

  const latestImage = useMemo(() => items.map(getListingCoverImageUrl).find(Boolean) || "", [items]);
  const jsonLd = useMemo(
    () => ({
      "@context": "https://schema.org",
      "@type": "CollectionPage",
      inLanguage: lang,
      name: t("ads.title"),
      description: t("ads.subtitle"),
      url: getCanonicalUrl("/ads")
    }),
    [lang, t]
  );

  usePageSeo({
    title: `${t("ads.title")} | SangeHassan`,
    description: t("ads.subtitle"),
    path: "/ads",
    lang,
    locale: lang === "fa" ? "fa_IR" : lang === "ar" ? "ar_SA" : "en_US",
    image: latestImage ? resolveImageUrl(latestImage) : "",
    jsonLdId: "ads-jsonld",
    jsonLd
  });

  useEffect(() => {
    let active = true;
    const restore = () => {
      try {
        const stored = sessionStorage.getItem("sh_me");
        if (stored) {
          const parsed = JSON.parse(stored);
          if (parsed?.id) setIsAuthenticated(true);
        }
      } catch (_) {
        /* ignore */
      }
    };

    const loadSession = async () => {
      try {
        const res = await fetchJSON("/api/v1/session");
        const session = res?.data || res;
        const me = session?.authenticated ? session.user : null;
        if (!active) return;
        setIsAuthenticated(Boolean(me?.id));
        if (me?.id) {
          sessionStorage.setItem("sh_me", JSON.stringify(me));
        } else {
          sessionStorage.removeItem("sh_me");
        }
      } catch (_) {
        if (!active) return;
        setIsAuthenticated(false);
      }
    };

    const load = async () => {
      try {
        const res = await fetchJSON("/api/ads?limit=100&sort=newest");
        const data = res?.data || res;
        if (!active) return;
        setItems(Array.isArray(data) ? data : []);
      } catch (err) {
        if (!active) return;
        setError(err?.message || t("messages.error"));
      } finally {
        if (active) setLoading(false);
      }
    };
    restore();
    loadSession();
    load();
    return () => {
      active = false;
    };
  }, [t]);

  const filterOptions = useMemo(
    () => ({
      stoneTypes: uniqueOptions(items.map((ad) => ad.stone_type)),
      locations: uniqueOptions(items.map(getLocationValue))
    }),
    [items]
  );

  useEffect(() => {
    if (stoneTypeFilter !== "all" && !filterOptions.stoneTypes.includes(stoneTypeFilter)) {
      setStoneTypeFilter("all");
    }
    if (locationFilter !== "all" && !filterOptions.locations.includes(locationFilter)) {
      setLocationFilter("all");
    }
  }, [filterOptions, locationFilter, stoneTypeFilter]);

  const filteredItems = useMemo(() => {
    const query = normalizeSearchText(searchInput);
    const parsedMinTonnage = Number(minTonnageInput);
    const parsedMaxTonnage = Number(maxTonnageInput);
    const hasMinTonnage = minTonnageInput.trim() !== "" && Number.isFinite(parsedMinTonnage) && parsedMinTonnage >= 0;
    const hasMaxTonnage = maxTonnageInput.trim() !== "" && Number.isFinite(parsedMaxTonnage) && parsedMaxTonnage >= 0;

    let base = items;

    if (query) {
      base = base.filter((ad) => {
        const extraText = Object.entries(ad.extra_props || {})
          .map(([key, value]) => `${key} ${value}`)
          .join(" ");
        const haystack = normalizeSearchText(
          [
            ad.title,
            ad.description,
            ad.form,
            ad.stone_type,
            ad.province,
            ad.city,
            getListingProductTitle(ad, lang),
            extraText
          ].join(" ")
        );
        return haystack.includes(query);
      });
    }

    if (formFilter !== "all") {
      base = base.filter((ad) => ad.form === formFilter);
    }

    if (stoneTypeFilter !== "all") {
      base = base.filter((ad) => ad.stone_type === stoneTypeFilter);
    }

    if (locationFilter !== "all") {
      base = base.filter((ad) => getLocationValue(ad) === locationFilter);
    }

    if (priceStatusFilter === "priced") {
      base = base.filter((ad) => getNumericValue(ad.price_amount) > 0);
    } else if (priceStatusFilter === "unpriced") {
      base = base.filter((ad) => getNumericValue(ad.price_amount) <= 0);
    }

    if (hasMinTonnage) {
      base = base.filter((ad) => getNumericValue(ad.tonnage) >= parsedMinTonnage);
    }
    if (hasMaxTonnage) {
      base = base.filter((ad) => getNumericValue(ad.tonnage) <= parsedMaxTonnage);
    }

    const sorted = [...base];
    if (sortMode === "price_asc") {
      sorted.sort((a, b) => {
        const priceA = getNumericValue(a.price_amount);
        const priceB = getNumericValue(b.price_amount);
        if (priceA <= 0 && priceB <= 0) return getCreatedTime(b) - getCreatedTime(a);
        if (priceA <= 0) return 1;
        if (priceB <= 0) return -1;
        return priceA - priceB;
      });
    } else if (sortMode === "price_desc") {
      sorted.sort((a, b) => getNumericValue(b.price_amount) - getNumericValue(a.price_amount));
    } else if (sortMode === "tonnage_asc") {
      sorted.sort((a, b) => getNumericValue(a.tonnage) - getNumericValue(b.tonnage));
    } else if (sortMode === "tonnage_desc") {
      sorted.sort((a, b) => getNumericValue(b.tonnage) - getNumericValue(a.tonnage));
    } else {
      sorted.sort((a, b) => getCreatedTime(b) - getCreatedTime(a));
    }

    return sorted;
  }, [
    formFilter,
    items,
    lang,
    locationFilter,
    maxTonnageInput,
    minTonnageInput,
    priceStatusFilter,
    searchInput,
    sortMode,
    stoneTypeFilter
  ]);

  const hasFilters =
    searchInput.trim() !== "" ||
    formFilter !== "all" ||
    stoneTypeFilter !== "all" ||
    locationFilter !== "all" ||
    priceStatusFilter !== "all" ||
    minTonnageInput.trim() !== "" ||
    maxTonnageInput.trim() !== "" ||
    sortMode !== "newest";

  const resetFilters = () => {
    setSearchInput("");
    setFormFilter("all");
    setStoneTypeFilter("all");
    setLocationFilter("all");
    setPriceStatusFilter("all");
    setMinTonnageInput("");
    setMaxTonnageInput("");
    setSortMode("newest");
  };

  const handleCreate = () => {
    if (isAuthenticated) {
      navigate("/ads/new");
      return;
    }
    sessionStorage.setItem("sh_after_login", "/ads/new");
    navigate("/login");
  };

  return (
    <section className="section-shell pt-16 pb-12">
      <div className="mb-8 flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div className="max-w-3xl">
          <p className="text-sm uppercase tracking-[0.3em] text-primary/60">{t("ads.title")}</p>
          <h1 className="mt-3 font-display text-3xl md:text-4xl">{t("ads.subtitle")}</h1>
        </div>
        <button
          type="button"
          onClick={handleCreate}
          className="inline-flex h-10 items-center justify-center rounded-full bg-primary px-4 text-xs font-semibold text-sand shadow hover:bg-primary/90"
        >
          {t("ads.create")}
        </button>
      </div>

      <div className="mb-8 space-y-3">
        <div className="flex flex-wrap items-center gap-3">
          {[
            { value: "all", label: t("ads.filters.allOffers") },
            { value: "block", label: t("ads.formOptions.block") },
            { value: "finished", label: t("ads.formOptions.finished") }
          ].map((option) => (
            <button
              key={option.value}
              type="button"
              onClick={() => setFormFilter(option.value)}
              className={`rounded-full border px-4 py-2 text-xs font-semibold transition ${
                formFilter === option.value
                  ? "border-primary bg-primary text-sand"
                  : "border-primary/20 text-primary/70 hover:border-primary/50"
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>

        <label className="sr-only" htmlFor="ads-search">
          {t("ads.filters.searchLabel")}
        </label>
        <input
          id="ads-search"
          type="search"
          value={searchInput}
          onChange={(event) => setSearchInput(event.target.value)}
          placeholder={t("ads.filters.searchPlaceholder")}
          className="w-full rounded-full border border-primary/20 bg-white/70 px-4 py-2.5 text-sm font-semibold text-primary outline-none transition focus:border-primary/60"
        />

        <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-5">
          <FilterSelect
            value={stoneTypeFilter}
            onChange={setStoneTypeFilter}
            label={t("ads.filters.allStoneTypes")}
            options={filterOptions.stoneTypes.map((value) => ({ value, label: value }))}
          />
          <FilterSelect
            value={locationFilter}
            onChange={setLocationFilter}
            label={t("ads.filters.allLocations")}
            options={filterOptions.locations.map((value) => ({ value, label: value }))}
          />
          <FilterSelect
            value={priceStatusFilter}
            onChange={setPriceStatusFilter}
            label={t("ads.filters.priceStatus")}
            options={[
              { value: "priced", label: t("ads.filters.priceWith") },
              { value: "unpriced", label: t("ads.filters.priceWithout") }
            ]}
          />
          <input
            type="number"
            min="0"
            value={minTonnageInput}
            onChange={(event) => setMinTonnageInput(event.target.value)}
            placeholder={t("ads.filters.minQuantity")}
            className="h-10 rounded-full border border-primary/20 bg-white/70 px-4 text-sm font-semibold text-primary outline-none transition focus:border-primary/60"
          />
          <input
            type="number"
            min="0"
            value={maxTonnageInput}
            onChange={(event) => setMaxTonnageInput(event.target.value)}
            placeholder={t("ads.filters.maxQuantity")}
            className="h-10 rounded-full border border-primary/20 bg-white/70 px-4 text-sm font-semibold text-primary outline-none transition focus:border-primary/60"
          />
        </div>

        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <select
            value={sortMode}
            onChange={(event) => setSortMode(event.target.value)}
            className="h-10 rounded-full border border-primary/20 bg-white/70 px-4 text-sm font-semibold text-primary outline-none transition focus:border-primary/60"
            aria-label={t("ads.filters.sort")}
          >
            <option value="newest">{t("ads.filters.sortNewest")}</option>
            <option value="price_asc">{t("ads.filters.sortPriceLowToHigh")}</option>
            <option value="price_desc">{t("ads.filters.sortPriceHighToLow")}</option>
            <option value="tonnage_desc">{t("ads.filters.sortQuantityHighToLow")}</option>
            <option value="tonnage_asc">{t("ads.filters.sortQuantityLowToHigh")}</option>
          </select>
          <div className="flex items-center gap-3 text-xs font-semibold text-primary/60">
            <span>{t("ads.filters.results").replace("{count}", String(filteredItems.length))}</span>
            {hasFilters && (
              <button type="button" onClick={resetFilters} className="rounded-full border border-primary/20 px-3 py-2 text-primary/70 transition hover:border-primary/50">
                {t("ads.filters.reset")}
              </button>
            )}
          </div>
        </div>

        <div className="border-t border-primary/10 pt-3 text-xs leading-6 text-primary/65">
          {t("ads.privacyNote")}
        </div>
      </div>

      {loading ? (
        <p className="text-sm text-primary/70">{t("messages.loading")}</p>
      ) : error ? (
        <p className="text-sm text-red-600">{error}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-primary/70">{t("ads.empty")}</p>
      ) : filteredItems.length === 0 ? (
        <p className="mt-10 text-center text-sm text-primary/70">{t("ads.filters.emptyFiltered")}</p>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filteredItems.map((ad) => {
            const latestImageUrl = getListingCoverImageUrl(ad);
            const productTitle = getListingProductTitle(ad, lang);
            const locationLabel = getLocationValue(ad) || "—";
            const priceLabel = ad.price_amount
              ? formatPriceValue(ad.price_amount, ad.price_unit, t)
              : t("ads.priceUnitOptions.negotiable");
            return (
              <Link
                to={`/ads/${ad.id}`}
                key={ad.id}
                className="group flex h-full flex-col overflow-hidden transition hover:-translate-y-1 hover:shadow-xl"
                style={{ contentVisibility: "auto", containIntrinsicSize: "360px" }}
              >
                <div className="relative aspect-square w-full overflow-hidden bg-primary/10">
                  <div className="absolute left-3 top-3 z-10 flex max-w-[calc(100%-1.5rem)] flex-wrap gap-2">
                    <span className="rounded-full border border-white/90 bg-white/35 px-3 py-1 text-[11px] font-semibold text-white shadow-sm backdrop-blur">
                      {getFormLabel(ad.form, t)}
                    </span>
                    {ad.tonnage ? (
                      <span className="rounded-full border border-white/90 bg-white/35 px-3 py-1 text-[11px] font-semibold text-white shadow-sm backdrop-blur">
                        {formatListingQuantity(ad, lang, t)}
                      </span>
                    ) : null}
                  </div>
                  {latestImageUrl ? (
                    <img
                      src={resolveImageUrl(latestImageUrl)}
                      alt={ad.title || productTitle || t("ads.title")}
                      loading="lazy"
                      decoding="async"
                      className="h-full w-full object-cover transition duration-500 group-hover:scale-105"
                    />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center px-4 text-center text-sm text-primary/60">
                      {t("productDetail.noImages")}
                    </div>
                  )}

                  <div className={`pointer-events-none absolute inset-x-0 bottom-0 ${gradientDir} from-black/75 via-black/30 to-transparent p-4`}>
                    <p className="mb-2 inline-flex max-w-full min-w-0 items-center gap-1 rounded-full bg-white/15 px-3 py-1 text-[11px] font-bold text-white/95 backdrop-blur">
                      <span className="min-w-0 truncate">{priceLabel}</span>
                    </p>
                    <h3 className="font-display text-xl leading-tight text-white drop-shadow-[0_10px_24px_rgba(0,0,0,0.55)]">
                      {ad.title || productTitle || t("ads.title")}
                    </h3>
                    <div className="mt-2 space-y-1 text-[11px] font-semibold text-white/85 drop-shadow-[0_8px_18px_rgba(0,0,0,0.45)]">
                      <p className="truncate">{productTitle || ad.stone_type || t("ads.form.product")}</p>
                      <p className="truncate">{locationLabel}</p>
                      {ad.description ? <p className="line-clamp-2 text-white/75">{ad.description}</p> : null}
                    </div>
                  </div>
                </div>
              </Link>
            );
          })}
        </div>
      )}
    </section>
  );
}

function FilterSelect({ value, onChange, label, options }) {
  return (
    <select
      value={value}
      onChange={(event) => onChange(event.target.value)}
      className="h-10 rounded-full border border-primary/20 bg-white/70 px-4 text-sm font-semibold text-primary outline-none transition focus:border-primary/60"
      aria-label={label}
    >
      <option value="all">{label}</option>
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}
