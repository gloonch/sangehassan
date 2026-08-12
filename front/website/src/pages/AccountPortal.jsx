import { useEffect, useMemo, useState } from "react";
import LegacyAccount from "./Account";
import { fetchJSON } from "../lib/api";

const persianDigits = value => String(value).replace(/\d/g, digit => "۰۱۲۳۴۵۶۷۸۹"[Number(digit)]);
const decimal = value => { const normalized=String(value??"0");const match=normalized.match(/^(-?)(\d+)(?:\.(\d+))?$/);if(!match)return normalized;const grouped=match[2].replace(/\B(?=(\d{3})+(?!\d))/g,"٬");return persianDigits(`${match[1]}${grouped}${match[3]?`٫${match[3]}`:""}`); };
const money = (value, currency) => `${decimal(value)} ${currency || ""}`;
const persianDate = value => value ? new Intl.DateTimeFormat("fa-IR-u-ca-persian", { dateStyle:"medium", timeStyle:"short", timeZone:"Asia/Tehran" }).format(new Date(value)) : "—";
const installationStatus = { DRAFT: "پیش‌نویس", PLANNED: "برنامه‌ریزی‌شده", READY: "آماده شروع", IN_PROGRESS: "در حال نصب", PAUSED: "متوقف", COMPLETED: "تکمیل‌شده", CANCELLED: "لغوشده" };

export default function AccountPortal() {
  const [tab, setTab] = useState("orders");
  const [orders, setOrders] = useState([]);
  const [financial, setFinancial] = useState({});
  const [schedules, setSchedules] = useState({});
  const [payments, setPayments] = useState({});
  const [documents, setDocuments] = useState({});
  const [installations, setInstallations] = useState({});
  const [notifications, setNotifications] = useState([]);
  const [acceptance, setAcceptance] = useState({});
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    async function load() {
      try {
        const response = await fetchJSON("/api/v1/account/orders");
        const items = response.data || [];
        if (!active) return;
        setOrders(items);
        const optional = (path) => fetchJSON(path).catch((err) => err.status === 404 ? { data: null } : Promise.reject(err));
        const [f, s, p, d, i, n] = await Promise.all([
          Promise.all(items.map((o) => fetchJSON(`/api/v1/account/orders/${o.id}/financial-summary`).then((r) => [o.id, r.data]))),
          Promise.all(items.map((o) => fetchJSON(`/api/v1/account/orders/${o.id}/payment-schedule`).then((r) => [o.id, r.data || []]))),
          Promise.all(items.map((o) => fetchJSON(`/api/v1/account/orders/${o.id}/payments`).then((r) => [o.id, r.data || []]))),
          Promise.all(items.map((o) => fetchJSON(`/api/v1/account/orders/${o.id}/documents`).then((r) => [o.id, r.data || []]))),
          Promise.all(items.map((o) => optional(`/api/v1/account/orders/${o.id}/installation`).then((r) => [o.id, r.data || null]))),
          fetchJSON("/api/v1/notifications?limit=100"),
        ]);
        if (!active) return;
        setFinancial(Object.fromEntries(f));
        setSchedules(Object.fromEntries(s));
        setPayments(Object.fromEntries(p));
        setDocuments(Object.fromEntries(d));
        setInstallations(Object.fromEntries(i));
        setNotifications(n.data?.items || []);
      } catch (err) {
        if (active) setError(err.message);
      }
    }
    void load();
    return () => { active = false; };
  }, []);

  const allPayments = useMemo(() => orders.flatMap((o) => (payments[o.id] || []).map((p) => ({ ...p, order_number: o.order_number }))), [orders, payments]);
  const allDocuments = useMemo(() => orders.flatMap((o) => (documents[o.id] || []).map((d) => ({ ...d, order_number: o.order_number }))), [orders, documents]);
  const installationOrders = orders.filter((o) => installations[o.id]);

  const confirmAcceptance = async (order) => {
    const installation = installations[order.id];
    const form = acceptance[order.id] || {};
    setError("");
    try {
      const response = await fetchJSON(`/api/v1/account/orders/${order.id}/acceptance`, {
        method: "POST",
        headers: { "Idempotency-Key": crypto.randomUUID() },
        body: JSON.stringify({ installation_job_id: installation.id, customer_name: form.customer_name, accepted: true, comment: form.comment || "" }),
      });
      setInstallations((current) => ({ ...current, [order.id]: { ...current[order.id], acceptances: [...(current[order.id]?.acceptances || []), response.data] } }));
      setMessage(`تأیید نهایی سفارش ${order.order_number} ثبت شد.`);
    } catch (err) { setError(err.message); }
  };

  const tabs = [["orders", "سفارش‌ها و محموله‌ها"], ["payments", "پرداخت‌ها"], ["documents", "اسناد"], ...(installationOrders.length ? [["installation", "نصب و تأیید نهایی"]] : []), ["notifications", "اعلان‌ها"]];
  return <div dir="rtl">
    <section className="section-shell pt-10"><div className="mx-auto flex max-w-5xl flex-wrap gap-2">{tabs.map(([key, label]) => <button key={key} onClick={() => setTab(key)} className={`rounded-full px-5 py-2 ${tab === key ? "bg-primary text-sand" : "border"}`}>{label}</button>)}</div></section>
    {error && <p className="section-shell mx-auto max-w-5xl py-4 text-red-700">{error}</p>}
    {message && <p className="section-shell mx-auto max-w-5xl py-4 text-green-700">{message}</p>}
    {tab === "orders" && <LegacyAccount />}
    {tab === "payments" && <section className="section-shell py-10"><div className="mx-auto max-w-5xl space-y-4"><h1 className="font-display text-3xl">پرداخت‌های من</h1>{orders.map((o) => <article key={o.id} className="glass-panel rounded-3xl p-6"><div className="flex flex-wrap justify-between"><b>{o.order_number}</b>{financial[o.id] && <span>مانده: {money(financial[o.id].outstanding_amount, financial[o.id].currency)} • سررسید گذشته: {money(financial[o.id].overdue_amount, financial[o.id].currency)}</span>}</div><div className="mt-4 space-y-2">{(schedules[o.id] || []).map((s) => <div key={s.id} className="flex flex-wrap justify-between rounded-xl bg-primary/5 p-3 text-sm"><span>{s.title_fa}</span><b>{money(s.amount, s.currency)}</b><span>{s.status}</span></div>)}</div>{(payments[o.id] || []).map((p) => <div key={p.id} className="mt-3 flex justify-between rounded-xl border p-3"><span>{p.payment_number} • {money(p.amount, p.currency)}</span><b>{p.status}</b></div>)}</article>)}{allPayments.length === 0 && <p>پرداخت تأییدشده‌ای وجود ندارد.</p>}</div></section>}
    {tab === "documents" && <section className="section-shell py-10"><div className="mx-auto max-w-5xl"><h1 className="font-display text-3xl">اسناد سفارش</h1><div className="mt-5 grid gap-4 md:grid-cols-2">{allDocuments.map((d) => <article key={d.id} className="glass-panel rounded-3xl p-6"><b>{d.document_type}</b><p className="text-sm">{d.document_number} • {d.order_number}</p><a href={`/api/v1/account/documents/${d.id}/download`} target="_blank" rel="noreferrer" className="mt-4 inline-block rounded-full border px-4 py-2">مشاهده PDF</a></article>)}{allDocuments.length === 0 && <p>سند قابل نمایشی وجود ندارد.</p>}</div></div></section>}
    {tab === "installation" && <section className="section-shell py-10"><div className="mx-auto max-w-5xl space-y-5"><h1 className="font-display text-3xl">نصب و تأیید نهایی</h1>{installationOrders.map((order) => { const job = installations[order.id]; const form = acceptance[order.id] || {}; const accepted = (job.acceptances || []).some((item) => item.accepted); return <article key={job.id} className="glass-panel rounded-3xl p-6"><div className="flex flex-wrap justify-between gap-3"><div><b>{order.order_number} • {job.project_name || "پروژه نصب"}</b><p className="text-sm">{job.project_address}</p></div><span className="rounded-full bg-primary/10 px-4 py-2">{installationStatus[job.status] || job.status}</span></div><p className="mt-4">پیشرفت ثبت‌شده: {decimal(job.installed_area)} {job.area_unit}{job.planned_area ? ` از ${decimal(job.planned_area)}` : ""}</p><div className="mt-4 space-y-2">{(job.updates || []).map((update) => <div key={update.id} className="rounded-xl border p-3"><b>{decimal(update.installed_quantity)} {update.quantity_unit}</b><p>{update.description}</p><time dateTime={update.created_at} className="text-xs opacity-60">{persianDate(update.created_at)}</time></div>)}</div>{job.files?.length > 0 && <div className="mt-4 flex flex-wrap gap-2">{job.files.map((file) => <a key={file.id} href={`/api/v1/workflow-files/${file.id}`} target="_blank" rel="noreferrer" className="rounded-full border px-4 py-2 text-sm">{file.original_file_name || "تصویر نصب"}</a>)}</div>}{accepted ? <p className="mt-5 rounded-xl bg-green-50 p-4 text-green-800">تأیید نهایی شما ثبت شده است.</p> : job.status === "COMPLETED" && <div className="mt-5 grid gap-3 border-t pt-5"><h3 className="font-semibold">تأیید عملیاتی پایان کار</h3><input required className="rounded-xl border p-3" placeholder="نام تأییدکننده" value={form.customer_name || ""} onChange={(e) => setAcceptance((current) => ({ ...current, [order.id]: { ...form, customer_name: e.target.value } }))}/><textarea className="rounded-xl border p-3" placeholder="نظر یا توضیح" value={form.comment || ""} onChange={(e) => setAcceptance((current) => ({ ...current, [order.id]: { ...form, comment: e.target.value } }))}/><button disabled={!form.customer_name} onClick={() => confirmAcceptance(order)} className="rounded-full bg-primary px-5 py-3 text-sand disabled:opacity-40">تأیید پایان عملیات</button><small>این تأیید صرفاً عملیاتی است و امضای حقوقی محسوب نمی‌شود.</small></div>}</article>; })}</div></section>}
    {tab === "notifications" && <section className="section-shell py-10"><div className="mx-auto max-w-5xl space-y-3"><h1 className="font-display text-3xl">اعلان‌های من</h1>{notifications.map((n) => <button key={n.id} onClick={async () => { if (n.status === "UNREAD") await fetchJSON(`/api/v1/notifications/${n.id}/read`, { method: "POST" }); if (n.deep_link?.startsWith("/account")) window.location.assign(n.deep_link); }} className={`glass-panel block w-full rounded-2xl p-5 text-right ${n.status === "UNREAD" ? "ring-1 ring-primary" : "opacity-70"}`}><b>{n.title}</b><p>{n.body}</p><time dateTime={n.created_at} className="text-xs text-primary/50">{persianDate(n.created_at)}</time></button>)}</div></section>}
  </div>;
}
