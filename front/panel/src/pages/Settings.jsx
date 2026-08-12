import { useEffect, useMemo, useState } from "react";
import { fetchJSON } from "../lib/api";
import { AsyncState } from "../components/OperationalUI";
import { useAuth } from "../lib/auth";

const sections = {
  "عمومی": ["default_currency", "default_country_code", "default_phone_country", "default_timezone"],
  "سفارش‌ها": ["default_payment_due_days", "default_workflow_warning_hours", "allow_manager_force_close", "allow_manager_workflow_override"],
  "اعلان‌ها": ["sms_enabled", "customer_portal_enabled"],
  "فایل‌ها": ["max_upload_size_mb"],
  "ماژول‌ها": ["inventory_module_enabled", "supplier_module_enabled", "installation_module_enabled"],
};
const labels = { default_currency: "ارز پیش‌فرض", default_country_code: "کشور پیش‌فرض", default_phone_country: "پیش‌شماره تلفن", default_timezone: "منطقه زمانی", customer_portal_enabled: "حساب مشتری", sms_enabled: "پیامک", installation_module_enabled: "نصب", inventory_module_enabled: "موجودی", supplier_module_enabled: "تأمین‌کننده و خرید", allow_manager_force_close: "بستن سفارش با هشدار", allow_manager_workflow_override: "Override مدیریتی Workflow", default_payment_due_days: "روز سررسید پیش‌فرض", default_workflow_warning_hours: "ساعت هشدار Workflow", max_upload_size_mb: "حداکثر فایل (MB)" };

export default function Settings() {
  const { hasPermission } = useAuth();
  const [items, setItems] = useState([]), [system, setSystem] = useState(null), [loading, setLoading] = useState(true), [error, setError] = useState(""), [saving, setSaving] = useState("");
  const load = () => { setLoading(true); setError(""); Promise.all([fetchJSON("/api/v1/admin/settings"), fetchJSON("/api/v1/admin/system-info")]).then(([a, b]) => { setItems(a.data || []); setSystem(b.data || null); }).catch((e) => setError(e.message)).finally(() => setLoading(false)); };
  useEffect(load, []);
  const byKey = useMemo(() => Object.fromEntries(items.map((item) => [item.key, item])), [items]);
  const save = async (key, value) => { const reason = window.prompt("دلیل تغییر این تنظیم را وارد کنید:"); if (!reason?.trim()) return; setSaving(key); try { const response = await fetchJSON("/api/v1/admin/settings", { method: "PUT", body: JSON.stringify({ key, value, reason }) }); setItems((current) => current.map((item) => item.key === key ? response.data : item)); } catch (e) { setError(e.message); } finally { setSaving(""); } };
  return <div className="space-y-5" dir="rtl"><header className="panel-card"><h2 className="font-display text-2xl">تنظیمات سیستم</h2><p className="mt-2 text-sm text-primary/60">فقط تنظیمات شناخته‌شده و بدون اطلاعات محرمانه در این بخش نگهداری می‌شوند.</p></header>
    <AsyncState loading={loading} error={error} retry={load}>{Object.entries(sections).map(([section, keys]) => <section key={section} className="panel-card"><h3 className="text-lg font-semibold">{section}</h3><div className="mt-4 divide-y">{keys.map((key) => { const item = byKey[key]; if (!item) return null; const value = item.value; const disabled = !hasPermission("settings.manage") || saving === key; return <div key={key} className="flex flex-wrap items-center justify-between gap-4 py-4"><span><b className="block text-sm">{labels[key] || key}</b><small className="text-primary/50">{item.description}</small></span>{typeof value === "boolean" ? <input aria-label={labels[key]} type="checkbox" checked={value} disabled={disabled} onChange={(e) => save(key, e.target.checked)} className="h-5 w-5" /> : <div className="flex gap-2"><input aria-label={labels[key]} disabled={disabled} value={value ?? ""} onChange={(e) => setItems((current) => current.map((x) => x.key === key ? { ...x, value: typeof value === "number" ? Number(e.target.value) : e.target.value } : x))} className="rounded-xl border p-2" /><button type="button" disabled={disabled} onClick={() => save(key, value)} className="rounded-xl border px-4">ذخیره</button></div>}</div>; })}</div></section>)}</AsyncState>
    {system && <section className="panel-card"><h3 className="text-lg font-semibold">اطلاعات سیستم</h3><dl className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-5">{Object.entries(system).map(([key, value]) => <div key={key} className="rounded-xl bg-primary/5 p-3"><dt className="text-xs text-primary/50">{key}</dt><dd className="mt-1 break-all text-sm font-medium">{String(value)}</dd></div>)}</dl></section>}
  </div>;
}
