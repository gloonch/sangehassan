import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";

const inputClass =
  "w-full rounded-xl border border-primary/20 bg-white px-3 py-2 text-sm focus:border-primary focus:outline-none";

export default function AdDetail() {
  const { id } = useParams();
  const { t } = useTranslation();
  const navigate = useNavigate();

  const [ad, setAd] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [user, setUser] = useState(null);
  const [requestType, setRequestType] = useState("INSPECTION");
  const [requestNote, setRequestNote] = useState("");
  const [reqStatus, setReqStatus] = useState("");
  const [formState, setFormState] = useState({});
  const [saveMsg, setSaveMsg] = useState("");

  const isOwner = useMemo(() => {
    if (!ad || !user) return false;
    return ad.created_by && ad.created_by === user.id;
  }, [ad, user]);

  useEffect(() => {
    let active = true;
    const loadUser = async () => {
      try {
        const res = await fetchJSON("/api/v1/me");
        const me = res?.data || res;
        if (active) setUser(me);
      } catch (_) {
        /* not logged */
      }
    };
    const loadAd = async () => {
      try {
        const res = await fetchJSON(`/api/ads/${id}`);
        const data = res?.data || res;
        if (!active) return;
        setAd(data);
        setFormState({
          title: data.title || "",
          stone_type: data.stone_type || "",
          form: data.form || "",
          tonnage: data.tonnage || "",
          province: data.province || "",
          city: data.city || "",
          price_amount: data.price_amount || "",
          price_unit: data.price_unit || "per_ton",
          description: data.description || "",
          extra_props: JSON.stringify(data.extra_props || {}, null, 2),
          images: (data.images || []).map((i) => i.image_url).join(", ")
        });
      } catch (err) {
        if (active) setError(err?.message || t("messages.error"));
      } finally {
        if (active) setLoading(false);
      }
    };
    loadUser();
    loadAd();
    return () => {
      active = false;
    };
  }, [id, t]);

  const handleRequest = async () => {
    setReqStatus("");
    try {
      await fetchJSON(`/api/ads/${id}/requests`, {
        method: "POST",
        body: JSON.stringify({ request_type: requestType, buyer_note: requestNote })
      });
      setReqStatus(t("messages.sent") || "Sent");
    } catch (err) {
      if (err?.status === 401) {
        sessionStorage.setItem("sh_after_login", window.location.pathname);
        navigate("/login", { replace: true });
      } else {
        setReqStatus(err?.message || t("messages.error"));
      }
    }
  };

  const handleSave = async () => {
    setSaveMsg("");
    try {
      let extra = {};
      try {
        extra = formState.extra_props ? JSON.parse(formState.extra_props) : {};
      } catch {
        setSaveMsg("Invalid JSON");
        return;
      }
      const payload = {
        title: formState.title || null,
        stone_type: formState.stone_type || null,
        form: formState.form || null,
        tonnage: formState.tonnage ? Number(formState.tonnage) : null,
        province: formState.province || null,
        city: formState.city || null,
        price_amount: formState.price_amount ? Number(formState.price_amount) : null,
        price_unit: formState.price_unit || null,
        description: formState.description || null,
        extra_props: extra,
        images: formState.images
          ? formState.images
              .split(",")
              .map((s) => s.trim())
              .filter(Boolean)
          : []
      };
      await fetchJSON(`/api/ads/${id}`, { method: "PUT", body: JSON.stringify(payload) });
      setSaveMsg(t("ads.saved"));
    } catch (err) {
      if (err?.status === 401) {
        sessionStorage.setItem("sh_after_login", window.location.pathname);
        navigate("/login", { replace: true });
      } else {
        setSaveMsg(err?.message || t("messages.error"));
      }
    }
  };

  const handleDelete = async () => {
    if (!window.confirm(t("ads.confirmDelete"))) return;
    try {
      await fetchJSON(`/api/ads/${id}`, { method: "DELETE" });
      navigate("/ads", { replace: true });
    } catch (err) {
      setSaveMsg(err?.message || t("messages.error"));
    }
  };

  if (loading) return <section className="section-shell py-16">{t("messages.loading")}</section>;
  if (error) return <section className="section-shell py-16 text-red-600">{error}</section>;
  if (!ad) return null;

  const extraEntries = Object.entries(ad.extra_props || {});

  return (
    <section className="section-shell py-16">
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 rounded-3xl border border-primary/10 bg-white/80 p-6 shadow-sm">
          <p className="text-xs uppercase tracking-[0.2em] text-primary/60">
            {ad.form || "—"} · {ad.city || ad.province || "—"}
          </p>
          <h1 className="mt-2 font-display text-3xl">{ad.title || ad.stone_type || t("ads.title")}</h1>
          <div className="mt-4 grid grid-cols-2 gap-3 text-sm text-primary/80">
            <Info label="Stone">{ad.stone_type || "—"}</Info>
            <Info label="Form">{ad.form || "—"}</Info>
            <Info label="Tonnage">{ad.tonnage ? `${ad.tonnage} t` : "—"}</Info>
            <Info label="Price">
              {ad.price_amount ? `${ad.price_amount} ${ad.price_unit}` : "—"}
            </Info>
            <Info label="Location">
              {[ad.province, ad.city].filter(Boolean).join(" / ") || "—"}
            </Info>
          </div>
          <p className="mt-4 text-sm text-primary/80 whitespace-pre-line">{ad.description}</p>
          {extraEntries.length > 0 && (
            <div className="mt-4">
              <p className="text-xs uppercase tracking-[0.2em] text-primary/60">More</p>
              <ul className="mt-2 grid grid-cols-2 gap-2 text-sm text-primary/80">
                {extraEntries.map(([k, v]) => (
                  <li key={k} className="rounded-lg border border-primary/10 bg-primary/5 px-3 py-2">
                    <span className="font-semibold">{k}: </span>
                    <span>{typeof v === "object" ? JSON.stringify(v) : String(v)}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>

        <div className="rounded-3xl border border-primary/10 bg-white/80 p-4 shadow-sm">
          <div className="rounded-xl border border-primary/15 bg-primary/5 px-3 py-2 text-xs text-primary/80">
            {t("ads.privacyNote")}
          </div>
          {isOwner ? (
            <>
              <p className="mt-3 text-xs font-semibold uppercase tracking-[0.2em] text-primary/60">
                {t("ads.owned")}
              </p>
              <div className="mt-3 space-y-3">
                {renderInput("title")}
                {renderInput("stone_type")}
                <div className="grid grid-cols-2 gap-2">
                  <select
                    value={formState.form || ""}
                    onChange={(e) => setFormState((p) => ({ ...p, form: e.target.value }))}
                    className={inputClass}
                  >
                    <option value="">—</option>
                    <option value="block">block</option>
                    <option value="finished">finished</option>
                  </select>
                  <input
                    className={inputClass}
                    placeholder="tonnage"
                    value={formState.tonnage || ""}
                    onChange={(e) => setFormState((p) => ({ ...p, tonnage: e.target.value }))}
                    type="number"
                    step="0.01"
                  />
                </div>
                <div className="grid grid-cols-2 gap-2">
                  {renderInput("province")}
                  {renderInput("city")}
                </div>
                <div className="grid grid-cols-2 gap-2">
                  <input
                    className={inputClass}
                    placeholder="price"
                    value={formState.price_amount || ""}
                    onChange={(e) => setFormState((p) => ({ ...p, price_amount: e.target.value }))}
                    type="number"
                    step="0.01"
                  />
                  <select
                    value={formState.price_unit || "per_ton"}
                    onChange={(e) => setFormState((p) => ({ ...p, price_unit: e.target.value }))}
                    className={inputClass}
                  >
                    <option value="per_ton">per_ton</option>
                    <option value="total">total</option>
                    <option value="negotiable">negotiable</option>
                  </select>
                </div>
                <textarea
                  className={inputClass}
                  rows={3}
                  placeholder="description"
                  value={formState.description}
                  onChange={(e) => setFormState((p) => ({ ...p, description: e.target.value }))}
                />
                <textarea
                  className={`${inputClass} font-mono text-xs`}
                  rows={3}
                  placeholder="extra props (JSON)"
                  value={formState.extra_props}
                  onChange={(e) => setFormState((p) => ({ ...p, extra_props: e.target.value }))}
                />
                <textarea
                  className={`${inputClass} text-xs`}
                  rows={2}
                  placeholder="image URLs comma separated"
                  value={formState.images}
                  onChange={(e) => setFormState((p) => ({ ...p, images: e.target.value }))}
                />
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={handleSave}
                    className="rounded-full bg-primary px-4 py-2 text-xs font-semibold text-sand hover:bg-primary/90"
                  >
                    {t("ads.edit")}
                  </button>
                  <button
                    type="button"
                    onClick={handleDelete}
                    className="rounded-full border border-red-300 px-4 py-2 text-xs font-semibold text-red-600 hover:bg-red-50"
                  >
                    {t("ads.delete")}
                  </button>
                  {saveMsg && <span className="text-xs font-semibold text-primary">{saveMsg}</span>}
                </div>
              </div>
            </>
          ) : (
            <>
              <p className="text-xs uppercase tracking-[0.2em] text-primary/60">{t("ads.request")}</p>
              <div className="mt-3 space-y-3">
                <select value={requestType} onChange={(e) => setRequestType(e.target.value)} className={inputClass}>
                  <option value="INSPECTION">INSPECTION</option>
                  <option value="PURCHASE">PURCHASE</option>
                  <option value="BOTH">BOTH</option>
                </select>
                <textarea
                  className={inputClass}
                  rows={3}
                  placeholder={t("ads.requestNote")}
                  value={requestNote}
                  onChange={(e) => setRequestNote(e.target.value)}
                />
                <button
                  type="button"
                  onClick={handleRequest}
                  className="w-full rounded-full bg-primary px-4 py-2 text-sm font-semibold text-sand hover:bg-primary/90"
                >
                  {t("ads.submit")}
                </button>
                {reqStatus && <p className="text-xs font-semibold text-primary">{reqStatus}</p>}
              </div>
            </>
          )}
        </div>
      </div>
    </section>
  );

  function renderInput(key) {
    return (
      <input
        className={inputClass}
        placeholder={key}
        value={formState[key] || ""}
        onChange={(e) => setFormState((p) => ({ ...p, [key]: e.target.value }))}
      />
    );
  }
}

function Info({ label, children }) {
  return (
    <div>
      <p className="text-[11px] uppercase tracking-[0.2em] text-primary/50">{label}</p>
      <p className="font-semibold text-primary">{children}</p>
    </div>
  );
}
