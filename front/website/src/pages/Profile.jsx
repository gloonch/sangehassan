import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";

export default function Profile() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [user, setUser] = useState(null);
  const [requests, setRequests] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [activeTab, setActiveTab] = useState("info");
  const [fullName, setFullName] = useState("");
  const [phone, setPhone] = useState("");
  const [saving, setSaving] = useState(false);
  const [saveMessage, setSaveMessage] = useState("");

  useEffect(() => {
    let active = true;
    const load = async () => {
      try {
        const meRes = await fetchJSON("/api/v1/me");
        const reqRes = await fetchJSON("/api/v1/me/requests");
        const me = meRes?.data || meRes;
        const reqs = reqRes?.data || reqRes;
        if (!active) return;
        setUser(me);
        setFullName(me?.full_name || "");
        setPhone(me?.phone || "");
        setRequests(Array.isArray(reqs) ? reqs : []);
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

  return (
    <section className="section-shell py-16">
      <div className="glass-panel rounded-3xl p-10">
        <h1 className="font-display text-3xl">{t("profile.title")}</h1>
        <p className="mt-2 text-sm text-primary/70">{t("profile.subtitle")}</p>
        <div className="mt-6 flex flex-col gap-4">
          {loading ? (
            <p className="text-sm text-primary/70">{t("messages.loading")}</p>
          ) : error ? (
            <p className="text-sm text-red-600">{error}</p>
          ) : (
            <>
              <div className="flex flex-wrap gap-2 rounded-xl border border-primary/10 bg-white/70 p-2 text-xs font-semibold text-primary/80">
                <button
                  type="button"
                  onClick={() => setActiveTab("info")}
                  className={`rounded-lg px-3 py-2 transition ${
                    activeTab === "info"
                      ? "bg-primary text-sand shadow"
                      : "hover:bg-primary/10 hover:text-primary"
                  }`}
                >
                  {t("profile.tabInfo")}
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab("requests")}
                  className={`rounded-lg px-3 py-2 transition ${
                    activeTab === "requests"
                      ? "bg-primary text-sand shadow"
                      : "hover:bg-primary/10 hover:text-primary"
                  }`}
                >
                  {t("profile.tabRequests")}
                </button>
              </div>

              {activeTab === "info" && (
                <div className="rounded-2xl border border-primary/10 bg-white/70 p-4 shadow-sm">
                  <div className="flex items-start justify-between">
                    <div className="w-full max-w-xl">
                      <p className="text-xs uppercase tracking-[0.2em] text-primary/60">
                        {t("profile.basicInfo")}
                      </p>
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
                            value={user?.email || ""}
                            readOnly
                            className="mt-2 w-full cursor-not-allowed rounded-xl border border-primary/20 bg-primary/5 px-4 py-3 text-sm text-primary/60"
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
                          {saveMessage && (
                            <span className="text-xs font-semibold text-green-700">{saveMessage}</span>
                          )}
                        </div>
                      </form>
                    </div>
                    <button
                      type="button"
                      onClick={handleLogout}
                      className="rounded-full border border-primary/20 px-3 py-1 text-xs font-semibold text-primary hover:border-primary/40"
                    >
                      {t("auth.logout")}
                    </button>
                  </div>
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
                    <ul className="space-y-2 text-sm text-primary/80">
                      {requests.map((item, idx) => (
                        <li key={idx} className="rounded-lg border border-primary/10 p-3">
                          {JSON.stringify(item)}
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </section>
  );
}
