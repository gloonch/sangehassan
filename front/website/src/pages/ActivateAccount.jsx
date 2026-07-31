import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";

export default function ActivateAccount() {
  const navigate=useNavigate();
  const [form,setForm]=useState({phone:"",activation_code:"",password:""});
  const [error,setError]=useState(""); const [done,setDone]=useState(false);
  const submit=async(event)=>{event.preventDefault();setError("");try{await fetchJSON("/api/v1/auth/customer/activate",{method:"POST",body:JSON.stringify(form)});sessionStorage.setItem("sh_after_login","/account");setDone(true);setTimeout(()=>navigate("/login"),1200)}catch{setError("کد فعال‌سازی نامعتبر یا منقضی شده است.")}};
  return <section className="section-shell flex min-h-[70vh] items-center justify-center py-16" dir="rtl"><div className="glass-panel w-full max-w-md rounded-3xl p-8"><h1 className="font-display text-3xl">فعال‌سازی حساب سفارش</h1><p className="mt-2 text-sm text-primary/65">شماره تماس، کدی که اپراتور تلفنی اعلام کرده و رمز دلخواهتان را وارد کنید.</p>{done?<div className="mt-6 rounded-xl bg-green-50 p-4 text-green-800">حساب فعال شد؛ در حال انتقال به صفحه ورود…</div>:<form className="mt-6 space-y-4" onSubmit={submit}><input required dir="ltr" className="w-full rounded-xl border p-3" placeholder="09121234567" value={form.phone} onChange={e=>setForm({...form,phone:e.target.value})}/><input required dir="ltr" inputMode="numeric" className="w-full rounded-xl border p-3 text-center tracking-[.35em]" placeholder="کد فعال‌سازی" value={form.activation_code} onChange={e=>setForm({...form,activation_code:e.target.value})}/><input required type="password" className="w-full rounded-xl border p-3" placeholder="رمز جدید (حداقل ۸ نویسه)" value={form.password} onChange={e=>setForm({...form,password:e.target.value})}/>{error&&<p className="text-sm text-red-600">{error}</p>}<button className="w-full rounded-full bg-primary p-3 font-semibold text-sand">فعال‌سازی</button></form>}<Link className="mt-4 block text-center text-sm underline" to="/login">بازگشت به ورود</Link></div></section>;
}
