import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useAuth } from "../lib/auth";

export default function ChangePassword() {
  const [form, setForm] = useState({ current_password: "", new_password: "" });
  const [error, setError] = useState("");
  const { refreshUser } = useAuth();
  const navigate = useNavigate();
  const submit = async (event) => { event.preventDefault(); setError(""); try { await fetchJSON("/api/v1/auth/change-password", { method: "POST", body: JSON.stringify(form) }); await refreshUser(); navigate("/dashboard", { replace: true }); } catch { setError("تغییر رمز انجام نشد. رمز جدید باید حداقل ۸ نویسه باشد."); } };
  return <section className="panel-shell flex min-h-screen items-center justify-center"><form onSubmit={submit} className="panel-card w-full max-w-md space-y-4" dir="rtl"><h1 className="font-display text-2xl">تغییر رمز عبور</h1><p className="text-sm text-primary/65">برای ادامه، رمز موقت را تغییر دهید.</p><input className="w-full rounded-xl border p-3" type="password" placeholder="رمز موقت" value={form.current_password} onChange={(e)=>setForm({...form,current_password:e.target.value})}/><input className="w-full rounded-xl border p-3" type="password" placeholder="رمز جدید" value={form.new_password} onChange={(e)=>setForm({...form,new_password:e.target.value})}/>{error&&<p className="text-sm text-red-600">{error}</p>}<button className="w-full rounded-full bg-primary p-3 text-sand">ذخیره رمز جدید</button></form></section>;
}
