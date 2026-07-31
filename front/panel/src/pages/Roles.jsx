import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useAuth } from "../lib/auth";

export default function Roles() {
  const { hasPermission } = useAuth();
  const [roles, setRoles] = useState([]);
  const [form, setForm] = useState({ code: "", name_fa: "", description_fa: "", clone_role_id: null });
  const [show, setShow] = useState(false);
  const [error, setError] = useState("");

  const load = async (isCancelled = () => false) => {
    try {
      const response = await fetchJSON("/api/v1/admin/roles");
      if (isCancelled()) return;
      setRoles(response.data || []);
      setError("");
    } catch (loadError) {
      if (!isCancelled()) setError(loadError.message);
    }
  };

  useEffect(() => {
    let cancelled = false;
    void load(() => cancelled);
    return () => {
      cancelled = true;
    };
  }, []);

  const submit = async (event) => {
    event.preventDefault();
    await fetchJSON("/api/v1/admin/roles", { method: "POST", body: JSON.stringify(form) });
    setShow(false);
    setForm({ code: "", name_fa: "", description_fa: "", clone_role_id: null });
    void load();
  };

  return <div className="space-y-5" dir="rtl"><section className="panel-card"><div className="flex justify-between"><div><h2 className="font-display text-2xl">نقش‌ها و دسترسی‌ها</h2><p className="text-sm text-primary/60">Roleها hard-code نیستند و از permissionهای پایدار تشکیل می‌شوند.</p></div>{hasPermission("roles.create")&&<button className="rounded-full bg-primary px-5 text-sand" onClick={()=>setShow(v=>!v)}>نقش جدید</button>}</div></section>{show&&<form className="panel-card grid gap-3 md:grid-cols-2" onSubmit={submit}><input required className="rounded-xl border p-3" placeholder="کد انگلیسی پایدار" value={form.code} onChange={e=>setForm({...form,code:e.target.value})}/><input required className="rounded-xl border p-3" placeholder="عنوان فارسی" value={form.name_fa} onChange={e=>setForm({...form,name_fa:e.target.value})}/><input className="rounded-xl border p-3 md:col-span-2" placeholder="توضیح فارسی" value={form.description_fa} onChange={e=>setForm({...form,description_fa:e.target.value})}/><select className="rounded-xl border p-3" value={form.clone_role_id||""} onChange={e=>setForm({...form,clone_role_id:e.target.value?Number(e.target.value):null})}><option value="">بدون کپی دسترسی</option>{roles.map(r=><option key={r.id} value={r.id}>کپی از {r.name_fa}</option>)}</select><button className="rounded-full bg-primary p-3 text-sand">ایجاد</button></form>}{error&&<p className="text-red-600">{error}</p>}<div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">{roles.map(r=><article key={r.id} className="panel-card"><div className="flex justify-between"><h3 className="font-semibold">{r.name_fa}</h3><span className={`rounded-full px-2 py-1 text-xs ${r.is_active?"bg-green-100":"bg-gray-100"}`}>{r.is_active?"فعال":"غیرفعال"}</span></div><code className="mt-2 block text-xs" dir="ltr">{r.code}</code><p className="mt-3 text-sm text-primary/60">{r.description_fa}</p><p className="mt-3 text-xs">{r.user_count} کاربر {r.is_protected&&"• محافظت‌شده"}</p><Link className="mt-4 inline-block rounded-full border px-4 py-2 text-xs" to={`/dashboard/roles/${r.id}`}>ویرایش دسترسی‌ها</Link></article>)}</div></div>;
}
