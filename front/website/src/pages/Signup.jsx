import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";

export default function Signup() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    const checkSession = async () => {
      try {
        await fetchJSON("/api/v1/me");
        navigate("/profile", { replace: true });
      } catch (_) {
        // ignore
      }
    };
    checkSession();
  }, [navigate]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    setError("");
    try {
      await fetchJSON("/api/v1/auth/signup", {
        method: "POST",
        body: JSON.stringify({
          email,
          password,
          full_name: fullName || undefined,
          phone: phone || undefined
        })
      });
      navigate("/profile", { replace: true });
    } catch (err) {
      setError(err?.message || t("messages.error"));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <section className="section-shell py-16">
      <div className="max-w-xl rounded-3xl bg-white/80 p-8 shadow-xl">
        <h1 className="font-display text-3xl">{t("auth.signupTitle")}</h1>
        <p className="mt-2 text-sm text-primary/70">{t("auth.signupSubtitle")}</p>

        <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
          <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
            {t("auth.fullName")}
            <input
              type="text"
              className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
              placeholder={t("auth.fullName")}
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
            />
          </label>
          <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
            {t("auth.email")}
            <input
              type="email"
              className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
              placeholder={t("auth.email")}
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </label>
          <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
            {t("auth.phone")}
            <input
              type="tel"
              className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
              placeholder={t("auth.phone")}
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
            />
          </label>
          <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
            {t("auth.password")}
            <input
              type="password"
              className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
              placeholder={t("auth.password")}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </label>
          {error && <p className="text-sm text-red-600">{error}</p>}
          <button
            type="submit"
            disabled={submitting}
            className="w-full rounded-full bg-primary px-6 py-3 text-sm font-semibold text-sand disabled:opacity-60"
          >
            {submitting ? t("messages.loading") : t("auth.signupButton")}
          </button>
        </form>
      </div>
    </section>
  );
}
