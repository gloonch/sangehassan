import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { useAuth } from "../lib/auth";
import LanguageSwitch from "../components/LanguageSwitch";

export default function Login() {
  const { t } = useTranslation();
  const { login } = useAuth();
  const navigate = useNavigate();
  const [form, setForm] = useState({ username: "", password: "" });
  const [error, setError] = useState("");

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");
    try {
      await login(form);
      navigate("/dashboard", { replace: true });
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  return (
    <section className="panel-shell flex min-h-screen items-center justify-center py-16">
      <div className="panel-card w-full max-w-md">
        <div className="mb-6 flex justify-end">
          <LanguageSwitch />
        </div>
        <h1 className="font-display text-3xl">{t("adminLogin.title")}</h1>
        <p className="mt-2 text-sm text-primary/70">{t("adminLogin.subtitle")}</p>

        <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
          <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
            {t("adminLogin.username")}
            <input
              type="text"
              className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
              value={form.username}
              onChange={(event) => setForm({ ...form, username: event.target.value })}
            />
          </label>
          <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
            {t("adminLogin.password")}
            <input
              type="password"
              className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
              value={form.password}
              onChange={(event) => setForm({ ...form, password: event.target.value })}
            />
          </label>
          {error && <p className="text-sm text-red-500">{error}</p>}
          <button
            type="submit"
            className="w-full rounded-full bg-primary px-6 py-3 text-sm font-semibold text-sand"
          >
            {t("adminLogin.button")}
          </button>
        </form>
      </div>
    </section>
  );
}
