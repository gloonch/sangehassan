import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Factory, Landmark, PackageOpen, Warehouse, Workflow } from "lucide-react";
import { fetchJSON, idempotentHeaders } from "../lib/api";
import { useAuth } from "../lib/auth";

const icons = { mine: Landmark, factory: Factory, warehouse: Warehouse, export: PackageOpen, showroom: Warehouse, project: Workflow };

export default function OperationsDashboard() {
  const navigate = useNavigate();
  const { hasPermission } = useAuth();
  const [actions, setActions] = useState([]);
  const [templates, setTemplates] = useState([]);
  const [summary, setSummary] = useState(null);
  const [operationsSummary, setOperationsSummary] = useState(null);
  const [home, setHome] = useState(null);
  const [selected, setSelected] = useState(null);
  const [form, setForm] = useState({ customer_phone: "", customer_name: "" });
  const [activation, setActivation] = useState("");
  const [createdWorkflow, setCreatedWorkflow] = useState(null);
  const [matchedCustomer, setMatchedCustomer] = useState(null);
  const [error, setError] = useState("");

  const load = async (isCancelled = () => false) => {
    const [a, w, s, operations, homeResult] = await Promise.allSettled([
      fetchJSON("/api/v1/dashboard/action-items"),
      fetchJSON("/api/v1/dashboard/workflow-templates"),
      fetchJSON("/api/v1/dashboard/workflow-summary"),
      fetchJSON("/api/v1/dashboard/operations-summary"),
      fetchJSON("/api/v1/dashboard/home")
    ]);
    if (isCancelled()) return;
    const homeData=homeResult.status === "fulfilled" ? homeResult.value.data : null;
    setHome(homeData);
    setActions(homeData?.my_actions || (a.status === "fulfilled" ? a.value.data || [] : []));
    setTemplates(w.status === "fulfilled" ? w.value.data || [] : []);
    setSummary(s.status === "fulfilled" ? s.value.data : null);
    setOperationsSummary(homeData?.alerts || (operations.status === "fulfilled" ? operations.value.data : null));
    if (a.status === "rejected" && w.status === "rejected") setError("بارگذاری داشبورد انجام نشد.");
  };

  useEffect(() => {
    let cancelled = false;
    void load(() => cancelled);
    return () => {
      cancelled = true;
    };
  }, []);

  const choose = (workflow) => {
    setSelected({ ...workflow, idempotencyKey: crypto.randomUUID(), excludedOptionalStepCodes: [] });
    setActivation(""); setCreatedWorkflow(null); setMatchedCustomer(null); setError("");
  };
  const findCustomer = async () => { if (!form.customer_phone) return; try { const response = await fetchJSON(`/api/v1/customers/search?phone=${encodeURIComponent(form.customer_phone)}`); setMatchedCustomer(response.data || null); } catch { setMatchedCustomer(null); } };
  const regenerate = async () => { const reason=window.prompt("دلیل صدور کد فعال‌سازی جدید:");if(!reason?.trim())return;try { const response = await fetchJSON(`/api/v1/admin/customers/${matchedCustomer.id}/regenerate-activation`, { method: "POST", headers:idempotentHeaders(), body:JSON.stringify({reason}) }); setActivation(response.data.activation_code); setCreatedWorkflow(null); } catch (e) { setError(e.message); } };
  const start = async (event) => {
    event.preventDefault(); setError("");
    try {
      const response = await fetchJSON("/api/v1/workflow-instances", {
        method: "POST",
        headers: { "Idempotency-Key": selected.idempotencyKey },
        body: JSON.stringify({ workflow_template_id: selected.id, excluded_optional_step_codes: selected.excludedOptionalStepCodes, ...form })
      });
      const data = response.data;
      if (data.activation_code) { setActivation(data.activation_code); setCreatedWorkflow(data.workflow_instance_id); }
      else { setSelected(null); navigate(`/dashboard/workflows/${data.workflow_instance_id}`); }
    } catch (e) { setError(e.message); }
  };
  const finishInvitation = () => { const workflow = createdWorkflow; setSelected(null); setActivation(""); load(); if (workflow) navigate(`/dashboard/workflows/${workflow}`); };

  return <div className="flex flex-col gap-6" dir="rtl">
    <section className="panel-card order-1 md:hidden"><h2 className="font-display text-2xl">کارهای امروز من</h2><div className="mt-4 grid grid-cols-2 gap-2">{(home?.quick_actions||[]).map(item=><button key={item.code} onClick={()=>navigate(item.path.replace(/^\/panel/,""))} className="min-h-20 rounded-2xl border p-4 text-sm font-semibold">{item.label}</button>)}</div></section>
    <section className="panel-card order-1"><h2 className="font-display text-2xl">اقدامات موردنیاز من</h2><p className="mt-1 text-sm text-primary/60">کارهای باز براساس نقش‌ها و دسترسی‌های شما</p>
      {actions.length === 0 ? <p className="mt-5 rounded-xl bg-primary/5 p-4 text-sm">اقدام بازی برای شما وجود ندارد.</p> : <div className="mt-5 grid gap-3 lg:grid-cols-2">{actions.map((a) => <article key={a.id} className="rounded-2xl border bg-white/70 p-4"><div className="flex justify-between"><h3 className="font-semibold">{a.title_fa}</h3><span className="rounded-full bg-amber-100 px-2 py-1 text-xs">{a.priority}</span></div><p className="mt-2 text-sm text-primary/60">{a.customer} • {a.order_number}</p><p className="text-xs text-primary/50">{a.workflow_name} — {a.step_name}</p><button onClick={() => navigate(`/dashboard/workflows/${a.workflow_instance_id}`)} className="mt-3 rounded-full bg-primary px-4 py-2 text-xs text-sand">انجام اقدام</button></article>)}</div>}
    </section>
    {home?.quick_actions?.length>0&&<section className="panel-card order-2 hidden md:block"><h2 className="font-display text-2xl">عملیات سریع</h2><div className="mt-4 flex flex-wrap gap-3">{home.quick_actions.map(item=><button key={item.code} onClick={()=>navigate(item.path.replace(/^\/panel/,""))} className="rounded-full border px-5 py-3 text-sm font-semibold">{item.label}</button>)}</div></section>}
    <section className="panel-card order-2"><h2 className="font-display text-2xl">شروع فرایند جدید</h2><div className="mt-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">{templates.map((w) => { const Icon = icons[w.icon_key] || Workflow; return <button key={w.id} onClick={() => choose(w)} className="rounded-2xl border bg-white/70 p-4 text-right transition hover:border-primary"><Icon className="mb-3 h-7 w-7"/><b>{w.name_fa}</b><p className="mt-1 text-xs text-primary/55">{w.description_fa}</p></button>; })}</div></section>
    {summary&&<section className="order-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-5">{[["فرایند فعال",summary.active_workflows],["در انتظار",summary.waiting_steps],["نیازمند تأیید",summary.approval_steps],["مرحله دیرکرد",summary.overdue_steps],["مغایرت بحرانی",summary.critical_discrepancies],["مغایرت باز",summary.open_discrepancies],["Task مسدود",summary.blocking_tasks],["بدون مسئول",summary.unassigned_steps],["نزدیک تحویل",summary.due_soon_workflows]].filter(([,value])=>value!==undefined).map(([label,value])=><article key={label} className="panel-card p-4"><strong className="text-2xl">{value}</strong><span className="mt-1 block text-xs text-primary/60">{label}</span></article>)}</section>}
    {operationsSummary&&<section className="order-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-5">{[["بچ فعال",operationsSummary.active_batches],["آماده ارسال",operationsSummary.ready_for_shipment],["حمل فعال",operationsSummary.active_shipments],["تحویل جزئی",operationsSummary.partial_deliveries],["موجودی کم",operationsSummary.low_stock],["رزرو منقضی",operationsSummary.expired_reservations],["خرید باز",operationsSummary.open_purchases],["خرید دیرکرد",operationsSummary.overdue_purchases],["خریدهای من",operationsSummary.assigned_purchases],["QC در انتظار",operationsSummary.pending_quality_inspections],["QC ناموفق",operationsSummary.failed_quality_inspections],["نیازمند اصلاح",operationsSummary.quality_rework],["QCهای من",operationsSummary.assigned_quality_inspections],["نصب فعال",operationsSummary.active_installations],["نصب امروز",operationsSummary.installations_today],["نصب دیرکرد",operationsSummary.overdue_installations],["Issue نصب",operationsSummary.open_installation_issues],["نصب‌های من",operationsSummary.assigned_installations],["آماده بستن",operationsSummary.orders_ready_to_close]].filter(([,value])=>value!==undefined).map(([label,value])=><article key={label} className="panel-card border-amber-200 p-4"><strong className="text-2xl">{value}</strong><span className="mt-1 block text-xs text-primary/60">{label}</span></article>)}</section>}
    {selected && <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"><div className="w-full max-w-md rounded-3xl bg-white p-6 shadow-xl">{activation ? <div className="text-center"><h3 className="text-xl font-semibold">کد فعال‌سازی مشتری</h3><p className="mt-3 text-sm">این کد فقط همین بار نمایش داده می‌شود و باید تلفنی به مشتری اعلام شود:</p><code className="my-5 block text-3xl tracking-[.4em]">{activation}</code><button className="rounded-full bg-primary px-6 py-2 text-sand" onClick={createdWorkflow ? finishInvitation : () => {setActivation("");setSelected(null)}}>{createdWorkflow ? "ادامه به فرایند" : "بستن"}</button></div> : <form onSubmit={start} className="space-y-4"><h3 className="text-xl font-semibold">{selected.name_fa}</h3><input required dir="ltr" className="w-full rounded-xl border p-3" placeholder="شماره مشتری" value={form.customer_phone} onBlur={findCustomer} onChange={(e) => {setForm({ ...form, customer_phone: e.target.value });setMatchedCustomer(null)}}/>{matchedCustomer&&<div className="rounded-xl bg-green-50 p-3 text-sm text-green-800">مشتری موجود پیدا شد: <b>{matchedCustomer.display_name}</b>. با ثبت فرم همین حساب استفاده می‌شود.{matchedCustomer.status==="INVITED"&&<button type="button" className="mt-2 block underline" onClick={regenerate}>صدور مجدد کد فعال‌سازی</button>}</div>}<input className="w-full rounded-xl border p-3" placeholder="نام مشتری (برای مشتری جدید)" value={form.customer_name} onChange={(e) => setForm({ ...form, customer_name: e.target.value })}/>{selected.optional_steps?.length>0&&<fieldset className="rounded-xl border p-3"><legend className="px-2 text-sm font-semibold">مراحل اختیاری (پیش‌فرض فعال)</legend>{selected.optional_steps.map(step=><label key={step.code} className="mt-2 flex items-center gap-2 text-sm"><input type="checkbox" checked={!selected.excludedOptionalStepCodes.includes(step.code)} onChange={e=>setSelected(current=>({...current,excludedOptionalStepCodes:e.target.checked?current.excludedOptionalStepCodes.filter(code=>code!==step.code):[...current.excludedOptionalStepCodes,step.code]}))}/>{step.title}</label>)}</fieldset>}{error && <p className="text-sm text-red-600">{error}</p>}<div className="flex gap-2"><button className="rounded-full bg-primary px-5 py-2 text-sand">شروع فرایند</button><button type="button" className="rounded-full border px-5 py-2" onClick={() => setSelected(null)}>انصراف</button></div></form>}</div></div>}
  </div>;
}
