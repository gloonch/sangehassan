import { useEffect, useMemo, useState } from "react";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";

const allowedTaxonomies = new Set(["tone", "use_case_application", "finishes", "use_case_form", "mines", "pattern", "availability"]);

const emptyForm = {
  category_id: "",
  term_id: "",
  title_en: "", title_fa: "", title_ar: "",
  description_en: "", description_fa: "", description_ar: "",
  h1_en: "", h1_fa: "", h1_ar: "",
  intro_en: "", intro_fa: "", intro_ar: "",
  is_active: true,
  is_indexable: true
};

export default function CatalogFacetPages() {
  const { t, lang } = useTranslation();
  const [pages, setPages] = useState([]);
  const [categories, setCategories] = useState([]);
  const [terms, setTerms] = useState([]);
  const [form, setForm] = useState(emptyForm);
  const [contentLang, setContentLang] = useState("en");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  const getTermLabel = (term) => term?.[`label_${lang}`] || term?.label_en || term?.term_key || "";
  const activeTerms = useMemo(() => terms.filter((term) => term.is_active !== false && allowedTaxonomies.has(term.taxonomy)), [terms]);

  const loadData = async () => {
    try {
      const [pageRes, categoryRes, termRes] = await Promise.all([
        fetchJSON("/api/admin/catalog-facet-pages"),
        fetchJSON("/api/admin/categories"),
        fetchJSON("/api/admin/product-terms")
      ]);
      setPages(pageRes.data || []);
      setCategories((categoryRes.data || []).filter((category) => category.is_active !== false && !category.parent_id));
      setTerms(termRes.data || []);
      setError("");
    } catch (_) {
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const submit = async (event) => {
    event.preventDefault();
    try {
      await fetchJSON("/api/admin/catalog-facet-pages", {
        method: "POST",
        body: JSON.stringify({ ...form, category_id: Number(form.category_id), term_id: Number(form.term_id) })
      });
      setForm(emptyForm);
      await loadData();
    } catch (_) {
      setError(t("messages.error"));
    }
  };

  const edit = (page) => {
    setForm({
      ...emptyForm,
      ...page,
      category_id: String(page.category_id),
      term_id: String(page.term_id)
    });
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const remove = async (id) => {
    try {
      await fetchJSON(`/api/admin/catalog-facet-pages/${id}`, { method: "DELETE" });
      await loadData();
    } catch (_) {
      setError(t("messages.error"));
    }
  };

  return (
    <div className="space-y-6">
      <section className="panel-card">
        <h2 className="font-display text-xl">{t("catalogFacetSEO.title")}</h2>
        <p className="mt-2 text-sm text-primary/60">{t("catalogFacetSEO.description")}</p>
        <form className="mt-6 space-y-5" onSubmit={submit}>
          <div className="grid gap-4 md:grid-cols-2">
            <label className="text-xs font-semibold uppercase tracking-wide text-primary/70">
              {t("form.category")}
              <select required value={form.category_id} onChange={(event) => setForm({ ...form, category_id: event.target.value })} className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm">
                <option value="">{t("catalogFacetSEO.selectCategory")}</option>
                {categories.map((category) => <option key={category.id} value={category.id}>{category[`title_${lang}`] || category.title_en} ({category.slug})</option>)}
              </select>
            </label>
            <label className="text-xs font-semibold uppercase tracking-wide text-primary/70">
              {t("catalogFacetSEO.facetValue")}
              <select required value={form.term_id} onChange={(event) => setForm({ ...form, term_id: event.target.value })} className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm">
                <option value="">{t("catalogFacetSEO.selectFacet")}</option>
                {activeTerms.map((term) => <option key={term.id} value={term.id}>{term.taxonomy}: {getTermLabel(term)} ({term.key})</option>)}
              </select>
            </label>
          </div>

          <div className="flex flex-wrap gap-2" role="tablist" aria-label="SEO content language">
            {[{ key: "en", label: "English" }, { key: "fa", label: "فارسی" }, { key: "ar", label: "العربية" }].map((item) => (
              <button key={item.key} type="button" role="tab" aria-selected={contentLang === item.key} onClick={() => setContentLang(item.key)} className={`rounded-full border px-4 py-2 text-xs font-semibold ${contentLang === item.key ? "border-primary bg-primary text-sand" : "border-primary/20 text-primary/70"}`}>
                {item.label}
              </button>
            ))}
          </div>

          <div className="grid gap-4" dir={contentLang === "en" ? "ltr" : "rtl"}>
            <label className="text-xs font-semibold uppercase tracking-wide text-primary/70">
              SEO title
              <input value={form[`title_${contentLang}`]} onChange={(event) => setForm({ ...form, [`title_${contentLang}`]: event.target.value })} className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm" />
            </label>
            <label className="text-xs font-semibold uppercase tracking-wide text-primary/70">
              H1
              <input value={form[`h1_${contentLang}`]} onChange={(event) => setForm({ ...form, [`h1_${contentLang}`]: event.target.value })} className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm" />
            </label>
            <label className="text-xs font-semibold uppercase tracking-wide text-primary/70">
              Meta description
              <textarea rows="3" value={form[`description_${contentLang}`]} onChange={(event) => setForm({ ...form, [`description_${contentLang}`]: event.target.value })} className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm" />
            </label>
            <label className="text-xs font-semibold uppercase tracking-wide text-primary/70">
              Introduction
              <textarea rows="4" value={form[`intro_${contentLang}`]} onChange={(event) => setForm({ ...form, [`intro_${contentLang}`]: event.target.value })} className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm" />
            </label>
          </div>

          <div className="flex flex-wrap gap-6">
            <label className="flex items-center gap-2 text-sm font-semibold"><input type="checkbox" checked={form.is_active} onChange={(event) => setForm({ ...form, is_active: event.target.checked })} />{t("form.active")}</label>
            <label className="flex items-center gap-2 text-sm font-semibold"><input type="checkbox" checked={form.is_indexable} onChange={(event) => setForm({ ...form, is_indexable: event.target.checked })} />{t("form.indexable")}</label>
          </div>
          {error ? <p className="text-sm text-red-500">{error}</p> : null}
          <div className="flex gap-3">
            <button type="submit" className="rounded-full bg-primary px-5 py-2 text-xs font-semibold text-sand">{t("actions.update")}</button>
            <button type="button" onClick={() => setForm(emptyForm)} className="rounded-full border border-primary/20 px-5 py-2 text-xs font-semibold text-primary/70">{t("actions.cancel")}</button>
          </div>
        </form>
      </section>

      <section className="panel-card">
        <h3 className="font-display text-xl">{t("catalogFacetSEO.saved")}</h3>
        {loading ? <p className="mt-4 text-sm text-primary/60">{t("messages.loading")}</p> : pages.length === 0 ? <p className="mt-4 text-sm text-primary/60">{t("catalogFacetSEO.empty")}</p> : (
          <div className="mt-5 space-y-3">
            {pages.map((page) => {
              const term = terms.find((item) => item.id === page.term_id);
              return (
                <div key={page.id} className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-primary/10 bg-white/80 px-4 py-3">
                  <div>
                    <p className="text-sm font-semibold">{page.category_slug} / {page.taxonomy} / {page.term_key}</p>
                    <p className="mt-1 text-xs text-primary/55">{getTermLabel(term)} · {page.is_active ? t("form.active") : t("form.inactive")} · {page.is_indexable ? t("form.indexable") : t("form.noindex")}</p>
                  </div>
                  <div className="flex gap-2">
                    <button type="button" onClick={() => edit(page)} className="rounded-full border border-primary/20 px-3 py-1 text-xs font-semibold">{t("actions.edit")}</button>
                    <button type="button" onClick={() => remove(page.id)} className="rounded-full border border-red-200 px-3 py-1 text-xs font-semibold text-red-500">{t("actions.delete")}</button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>
    </div>
  );
}
