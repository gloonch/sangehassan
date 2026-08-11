import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useAuth } from "../lib/auth";

export default function ChangePassword() {
  const [form, setForm] = useState({ current_password: "", new_password: "" });
  const [error, setError] = useState("");
  const { refreshUser } = useAuth();
  const navigate = useNavigate();

  const submit = async (event) => {
    event.preventDefault();
    setError("");

    if (Array.from(form.new_password).length < 8) {
      setError("رمز جدید باید حداقل ۸ نویسه باشد.");
      return;
    }

    try {
      await fetchJSON("/api/v1/auth/change-password", {
        method: "POST",
        body: JSON.stringify(form)
      });
    } catch (requestError) {
      const messages = {
        PASSWORD_TOO_SHORT: "رمز جدید باید حداقل ۸ نویسه باشد.",
        PASSWORD_TOO_LONG: "رمز جدید برای ذخیره‌سازی امن بیش از حد طولانی است.",
        CURRENT_PASSWORD_INVALID: "رمز موقت فعلی صحیح نیست. لطفاً آن را دقیقاً مطابق رمز دریافتی وارد کنید."
      };
      setError(
        messages[requestError.code] ||
          (requestError.status === 401
            ? "نشست شما منقضی شده است؛ دوباره وارد شوید."
            : "تغییر رمز انجام نشد. لطفاً دوباره تلاش کنید.")
      );
      return;
    }

    try {
      await refreshUser();
      navigate("/dashboard", { replace: true });
    } catch {
      // The password was already changed successfully. Reloading obtains a
      // fresh session instead of reporting a false password-change failure.
      window.location.replace("/dashboard");
    }
  };

  return (
    <section className="panel-shell flex min-h-screen items-center justify-center">
      <form onSubmit={submit} className="panel-card w-full max-w-md space-y-4" dir="rtl">
        <h1 className="font-display text-2xl">تغییر رمز عبور</h1>
        <p className="text-sm text-primary/65">برای ادامه، رمز موقت را تغییر دهید.</p>
        <input
          required
          autoComplete="current-password"
          className="w-full rounded-xl border p-3"
          type="password"
          placeholder="رمز موقت"
          value={form.current_password}
          onChange={(event) => setForm({ ...form, current_password: event.target.value })}
        />
        <input
          required
          minLength={8}
          autoComplete="new-password"
          className="w-full rounded-xl border p-3"
          type="password"
          placeholder="رمز جدید (حداقل ۸ نویسه)"
          value={form.new_password}
          onChange={(event) => setForm({ ...form, new_password: event.target.value })}
        />
        {error && <p className="text-sm text-red-600">{error}</p>}
        <button className="w-full rounded-full bg-primary p-3 text-sand">ذخیره رمز جدید</button>
      </form>
    </section>
  );
}
