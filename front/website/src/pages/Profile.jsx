import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";
import { formatPriceValue, formatPriceUnit } from "../lib/listings";
import { resolveImageUrl } from "../lib/assets";
import { usePageSeo } from "../lib/seo";

const PHONE_ALIAS_EMAIL_SUFFIX = "@phone.sangehassan.local";
const getLatestImageUrl = (ad) => {
  const images = Array.isArray(ad?.images) ? ad.images : [];
  for (let i = images.length - 1; i >= 0; i -= 1) {
    const imageUrl = images[i]?.image_url;
    if (imageUrl) return imageUrl;
  }
  return "";
};

export default function Profile() {
  const { t, lang } = useTranslation();
  const navigate = useNavigate();
  const [user, setUser] = useState(null);
  const [myAds, setMyAds] = useState([]);
  const [requests, setRequests] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [activeTab, setActiveTab] = useState("ads");
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [saving, setSaving] = useState(false);
  const [saveMessage, setSaveMessage] = useState("");

  usePageSeo({
    title: `${t("profile.title")} | SangeHassan`,
    description: t("profile.subtitle"),
    path: "/profile",
    lang,
    locale: lang === "fa" ? "fa_IR" : lang === "ar" ? "ar_SA" : "en_US",
    robots: "noindex,nofollow,noarchive"
  });

  const dateFormatter = useMemo(() => {
    const locale = lang === "fa" ? "fa-IR" : lang === "ar" ? "ar-SA" : "en-US";
    return new Intl.DateTimeFormat(locale, {
      dateStyle: "medium",
      timeStyle: "short"
    });
  }, [lang]);

  const tabs = useMemo(
    () => [
      { key: "ads", label: t("profile.tabAds") },
      { key: "requests", label: t("profile.tabRequests") },
      { key: "info", label: t("profile.tabInfo") }
    ],
    [t]
  );

  const tabContent = useMemo(
    () => ({
      ads: {
        title: t("profile.tabContent.adsTitle"),
        subtitle: t("profile.tabContent.adsSubtitle")
      },
      requests: {
        title: t("profile.tabContent.requestsTitle"),
        subtitle: t("profile.tabContent.requestsSubtitle")
      },
      info: {
        title: t("profile.tabContent.infoTitle"),
        subtitle: t("profile.tabContent.infoSubtitle")
      }
    }),
    [t]
  );

  const activeTabContent = tabContent[activeTab] || tabContent.info;

  useEffect(() => {
    let active = true;
    const load = async () => {
      try {
        const [meRes, reqRes, adsRes] = await Promise.all([
          fetchJSON("/api/v1/me"),
          fetchJSON("/api/v1/me/requests"),
          fetchJSON("/api/v1/me/listings")
        ]);
        const me = meRes?.data || meRes;
        const reqs = reqRes?.data || reqRes;
        const ads = adsRes?.data || adsRes;
        if (!active) return;
        setUser(me);
        setFullName(me?.full_name || "");
        setEmail(isPhoneAliasEmail(me?.email) ? "" : me?.email || "");
        setPhone(me?.phone || "");
        setRequests(Array.isArray(reqs) ? reqs : []);
        setMyAds(Array.isArray(ads) ? ads : []);
      } catch (err) {
        if (err?.status === 401) {
          navigate("/login", { replace: true });
          return;
        }
        setError(err?.message || t("messages.error"));
      } finally {
        if (active) setLoading(false);
      }
    };
    load();
    return () => {
      active = false;
    };
  }, [navigate, t]);

  const handleLogout = async () => {
    try {
      await fetchJSON("/api/v1/auth/logout", { method: "POST" });
    } catch (_) {
      // ignore
    } finally {
      navigate("/login", { replace: true });
    }
  };

  const handleSave = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError("");
    setSaveMessage("");
    try {
      const updatedRes = await fetchJSON("/api/v1/me", {
        method: "PUT",
        body: JSON.stringify({
          full_name: fullName || null,
          email: email || null,
          phone: phone || null
        })
      });
      const updated = updatedRes?.data || updatedRes;
      setUser(updated);
      setSaveMessage(t("messages.saved"));
      sessionStorage.setItem("sh_me", JSON.stringify(updated));
      window.dispatchEvent(new CustomEvent("sh:me-updated", { detail: updated }));
    } catch (err) {
      setError(err?.message || t("messages.error"));
    } finally {
      setSaving(false);
    }
  };

  const formatDate = (value) => {
    if (!value) return "—";
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) return String(value);
    return dateFormatter.format(parsed);
  };

  const formatValue = (value) => {
    if (value === undefined || value === null || value === "") return "—";
    if (typeof value === "object") return JSON.stringify(value);
    return String(value);
  };

  return (
    <>
      <div className="section-shell pt-4">
        <div className="mx-auto flex w-full max-w-3xl flex-col items-center text-center">
          <div className="flex w-full flex-wrap items-center justify-center gap-2 rounded-2xl border border-primary/10 bg-white/70 p-2 text-xs font-semibold text-primary/80 shadow-sm">
            {tabs.map((tab) => (
              <button
                key={tab.key}
                type="button"
                onClick={() => setActiveTab(tab.key)}
                className={`min-w-[7rem] rounded-xl px-4 py-2 transition ${
                  activeTab === tab.key
                    ? "bg-primary text-sand shadow"
                    : "hover:bg-primary/10 hover:text-primary"
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
          <h1 className="mt-6 font-display text-3xl">{activeTabContent.title}</h1>
          <p className="mt-2 text-sm text-primary/70">{activeTabContent.subtitle}</p>
        </div>
      </div>

      <section className="section-shell pb-16 pt-8">
        <div className="glass-panel rounded-3xl p-8 md:p-10">
          <div className="flex flex-col gap-4">
            {loading ? (
              <p className="text-sm text-primary/70">{t("messages.loading")}</p>
            ) : error ? (
              <p className="text-sm text-red-600">{error}</p>
            ) : (
              <>
                {activeTab === "ads" && (
                  <div className="rounded-2xl border border-primary/10 bg-white/70 p-4 shadow-sm">
                  <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
                    <p className="text-xs uppercase tracking-[0.2em] text-primary/60">
                      {t("ads.myListings")}
                    </p>
                    <button
                      type="button"
                      onClick={() => navigate("/ads/new")}
                      className="rounded-full bg-primary px-4 py-2 text-xs font-semibold text-sand shadow hover:bg-primary/90"
                    >
                      {t("ads.create")}
                    </button>
                  </div>
                  <div className="rounded-xl border border-primary/15 bg-primary/5 px-3 py-2 text-xs text-primary/80">
                    {t("ads.privacyNote")}
                  </div>

                  {myAds.length === 0 ? (
                    <p className="mt-4 text-sm text-primary/70">{t("ads.empty")}</p>
                  ) : (
                    <div className="mt-4 grid gap-4 sm:grid-cols-2">
                      {myAds.map((ad) => {
                        const extraEntries = Object.entries(ad.extra_props || {});
                        const location = [ad.province, ad.city].filter(Boolean).join(" / ");
                        const latestImageUrl = getLatestImageUrl(ad);
                        return (
                          <article
                            key={ad.id}
                            className="rounded-2xl border border-primary/10 bg-white/90 p-4 shadow-sm"
                          >
                            <div className="mb-3 overflow-hidden rounded-xl border border-primary/10 bg-primary/5">
                              {latestImageUrl ? (
                                <img
                                  src={resolveImageUrl(latestImageUrl)}
                                  alt={ad.title || ad.stone_type || t("ads.title")}
                                  loading="lazy"
                                  className="h-40 w-full object-cover"
                                />
                              ) : (
                                <div className="flex h-40 items-center justify-center text-xs font-semibold text-primary/55">
                                  {t("productDetail.noImages")}
                                </div>
                              )}
                            </div>
                            <div>
                              <h3 className="mt-1 text-lg font-semibold text-primary">
                                {ad.title || ad.stone_type || t("ads.title")}
                              </h3>
                            </div>

                            <p className="mt-2 text-sm text-primary/70 line-clamp-2">{ad.description || " "}</p>

                            <div className="mt-4 grid grid-cols-2 gap-x-3 gap-y-2">
                              <MetaRow label={t("profile.meta.stoneType")} value={formatValue(ad.stone_type)} />
                              <MetaRow label={t("profile.meta.form")} value={formatValue(ad.form)} />
                              <MetaRow
                                label={t("profile.meta.tonnage")}
                                value={ad.tonnage === undefined || ad.tonnage === null ? "—" : `${ad.tonnage}`}
                              />
                              <MetaRow label={t("ads.form.province")} value={formatValue(ad.province)} />
                              <MetaRow label={t("ads.form.city")} value={formatValue(ad.city)} />
                              <MetaRow label={t("profile.meta.location")} value={location || "—"} />
                              <MetaRow label={t("profile.meta.price")} value={formatPriceValue(ad.price_amount, ad.price_unit, t)} />
                              <MetaRow label={t("profile.meta.priceUnit")} value={formatPriceUnit(ad.price_unit, t)} />
                              <MetaRow
                                label={t("profile.meta.images")}
                                value={Array.isArray(ad.images) ? ad.images.length : 0}
                              />
                            </div>

                            {extraEntries.length > 0 && (
                              <div className="mt-4">
                                <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-primary/60">
                                  {t("ads.form.extra")}
                                </p>
                                <ul className="mt-2 space-y-1 text-xs text-primary/80">
                                  {extraEntries.map(([key, value]) => (
                                    <li key={key} className="rounded-lg border border-primary/10 px-2 py-1">
                                      <span className="font-semibold">{key}: </span>
                                      <span>{formatValue(value)}</span>
                                    </li>
                                  ))}
                                </ul>
                              </div>
                            )}

                            <div className="mt-4">
                              <Link
                                to={`/ads/${ad.id}`}
                                className="inline-flex rounded-full bg-primary px-4 py-2 text-xs font-semibold text-sand hover:bg-primary/90"
                              >
                                {t("ads.viewDetails")}
                              </Link>
                            </div>
                          </article>
                        );
                      })}
                    </div>
                  )}
                  </div>
                )}

                {activeTab === "requests" && (
                  <div className="rounded-2xl border border-primary/10 bg-white/70 p-4 shadow-sm">
                  <div className="mb-3 flex items-center justify-between">
                    <p className="text-xs uppercase tracking-[0.2em] text-primary/60">
                      {t("profile.requestsTitle")}
                    </p>
                  </div>
                  {requests.length === 0 ? (
                    <p className="text-sm text-primary/70">{t("profile.requestsEmpty")}</p>
                  ) : (
                    <div className="space-y-3">
                      {requests.map((item) => (
                        <article key={item.id || JSON.stringify(item)} className="rounded-xl border border-primary/10 p-3">
                          <div className="flex flex-wrap items-center justify-between gap-2">
                            <p className="text-xs uppercase tracking-[0.2em] text-primary/60">
                              {t("profile.meta.requestId")} #{formatValue(item.id)}
                            </p>
                            <span className="rounded-full bg-primary/10 px-3 py-1 text-[11px] font-semibold text-primary/80">
                              {formatValue(item.status)}
                            </span>
                          </div>
                          <div className="mt-3 grid grid-cols-2 gap-x-3 gap-y-2">
                            <MetaRow label={t("profile.meta.requestType")} value={formatValue(item.request_type)} />
                            <MetaRow label={t("profile.meta.listingId")} value={formatValue(item.listing_id)} />
                            <MetaRow label={t("profile.meta.meetingAt")} value={formatDate(item.meeting_at)} />
                          </div>
                          {item.buyer_note ? (
                            <p className="mt-3 rounded-lg border border-primary/10 bg-primary/5 p-2 text-xs text-primary/80">
                              <span className="font-semibold">{t("profile.meta.buyerNote")}: </span>
                              {item.buyer_note}
                            </p>
                          ) : null}
                          {item.admin_note ? (
                            <p className="mt-2 rounded-lg border border-primary/10 bg-primary/5 p-2 text-xs text-primary/80">
                              <span className="font-semibold">{t("profile.meta.adminNote")}: </span>
                              {item.admin_note}
                            </p>
                          ) : null}
                        </article>
                      ))}
                    </div>
                  )}
                  </div>
                )}

                {activeTab === "info" && (
                  <div className="rounded-2xl border border-primary/10 bg-white/70 p-4 shadow-sm">
                  <div className="w-full max-w-xl">
                    <p className="text-xs uppercase tracking-[0.2em] text-primary/60">{t("profile.basicInfo")}</p>
                    <form className="mt-3 space-y-3" onSubmit={handleSave}>
                      <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                        {t("auth.fullName")}
                        <input
                          type="text"
                          value={fullName}
                          onChange={(e) => setFullName(e.target.value)}
                          className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                        />
                      </label>
                      <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                        {t("auth.email")}
                        <input
                          type="email"
                          value={email}
                          onChange={(e) => setEmail(e.target.value)}
                          className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                        />
                      </label>
                      <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                        {t("auth.phone")}
                        <input
                          type="tel"
                          value={phone}
                          onChange={(e) => setPhone(e.target.value)}
                          className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                        />
                      </label>
                      <div className="flex items-center gap-3">
                        <button
                          type="submit"
                          disabled={saving}
                          className="rounded-full bg-primary px-5 py-2 text-xs font-semibold text-sand transition hover:bg-primary/90 disabled:opacity-60"
                        >
                          {saving ? t("messages.loading") : t("actions.save")}
                        </button>
                        <button
                          type="button"
                          onClick={handleLogout}
                          className="rounded-full border border-red-600 bg-red-600 px-5 py-2 text-xs font-semibold text-white transition hover:bg-white hover:text-red-600"
                        >
                          {t("auth.logout")}
                        </button>
                        {saveMessage && <span className="text-xs font-semibold text-green-700">{saveMessage}</span>}
                      </div>
                    </form>
                  </div>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      </section>
    </>
  );
}

function MetaRow({ label, value }) {
  return (
    <div>
      <p className="text-[11px] uppercase tracking-[0.16em] text-primary/60">{label}</p>
      <p className="text-xs font-semibold text-primary/90 break-words">{value}</p>
    </div>
  );
}

function isPhoneAliasEmail(email) {
  if (!email) return false;
  return String(email).toLowerCase().endsWith(PHONE_ALIAS_EMAIL_SUFFIX);
}
