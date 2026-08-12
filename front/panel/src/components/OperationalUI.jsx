import { useEffect, useRef } from "react";

export function ListPage({ title, description, actions, filters, children }) {
  return <div className="space-y-5" dir="rtl">
    <header className="panel-card flex flex-wrap items-start justify-between gap-4"><div><h2 className="font-display text-2xl">{title}</h2>{description && <p className="mt-2 text-sm text-primary/55">{description}</p>}</div>{actions && <div className="flex flex-wrap gap-2">{actions}</div>}</header>
    {filters}
    {children}
  </div>;
}

export function EmptyState({ title = "موردی ثبت نشده است.", description, action }) {
  return <div className="rounded-2xl border border-dashed border-primary/20 bg-primary/[0.02] p-8 text-center" role="status">
    <p className="font-semibold text-primary/80">{title}</p>
    {description && <p className="mt-2 text-sm text-primary/55">{description}</p>}
    {action && <div className="mt-4">{action}</div>}
  </div>;
}

export function AsyncState({ loading, error, empty, emptyText, children, retry }) {
  if (loading) return <div className="panel-card animate-pulse text-sm text-primary/60" aria-busy="true">در حال دریافت اطلاعات…</div>;
  if (error) return <div className="panel-card border-red-200 bg-red-50 text-sm text-red-800" role="alert">
    <p>{error}</p>{retry && <button type="button" onClick={retry} className="mt-3 rounded-full border border-red-300 px-4 py-2">تلاش دوباره</button>}
  </div>;
  if (empty) return <EmptyState title={emptyText} />;
  return children;
}

export function Pagination({ page, pageSize, total, onChange }) {
  const pages = Math.max(1, Math.ceil((total || 0) / pageSize));
  if (pages <= 1) return null;
  return <nav className="mt-5 flex items-center justify-center gap-3" aria-label="صفحه‌بندی">
    <button type="button" disabled={page <= 1} onClick={() => onChange(page - 1)} className="rounded-full border px-4 py-2 disabled:opacity-40">قبلی</button>
    <span className="text-sm">صفحه {page} از {pages}</span>
    <button type="button" disabled={page >= pages} onClick={() => onChange(page + 1)} className="rounded-full border px-4 py-2 disabled:opacity-40">بعدی</button>
  </nav>;
}

export function FilterBar({ children, onClear }) {
  return <div className="grid gap-3 rounded-2xl border bg-white/70 p-4 md:grid-cols-4">{children}<button type="button" onClick={onClear} className="rounded-xl border px-4 py-3 text-sm">پاک‌کردن فیلترها</button></div>;
}

export function ConfirmDialog({ open, title, description, requireReason = false, reason, onReason, pending, onCancel, onConfirm, confirmLabel = "تأیید" }) {
  const dialog = useRef(null);
  useEffect(() => {
    if (!open) return undefined;
    const focusable = [...(dialog.current?.querySelectorAll("button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled])") || [])];
    focusable[0]?.focus();
    const key = (event) => {
      if (event.key === "Escape" && !pending) onCancel();
      if (event.key === "Tab" && focusable.length) {
        const first = focusable[0], last = focusable[focusable.length - 1];
        if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus(); }
        else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus(); }
      }
    };
    window.addEventListener("keydown", key);
    return () => window.removeEventListener("keydown", key);
  }, [open, pending, onCancel]);
  if (!open) return null;
  return <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/45 p-4" role="presentation" onMouseDown={(e) => e.target === e.currentTarget && !pending && onCancel()}>
    <section ref={dialog} role="dialog" aria-modal="true" aria-labelledby="confirm-title" className="w-full max-w-md rounded-3xl bg-white p-6 shadow-2xl" dir="rtl">
      <h2 id="confirm-title" className="text-xl font-semibold">{title}</h2>
      <p className="mt-3 text-sm leading-7 text-primary/65">{description}</p>
      {requireReason && <label className="mt-4 block text-sm font-medium">دلیل
        <textarea autoFocus value={reason} onChange={(e) => onReason(e.target.value)} className="mt-2 min-h-24 w-full rounded-xl border p-3" required />
      </label>}
      <div className="mt-5 flex gap-2">
        <button type="button" disabled={pending || (requireReason && !reason?.trim())} onClick={onConfirm} className="rounded-full bg-primary px-5 py-2 text-sand disabled:opacity-40">{pending ? "در حال ثبت…" : confirmLabel}</button>
        <button type="button" disabled={pending} onClick={onCancel} className="rounded-full border px-5 py-2">انصراف</button>
      </div>
    </section>
  </div>;
}

export function PersianDate({ value, timezone = "Asia/Tehran", dateOnly = false }) {
  if (!value) return <span>—</span>;
  const options = dateOnly ? { dateStyle: "medium", timeZone: timezone } : { dateStyle: "medium", timeStyle: "short", timeZone: timezone };
  return <time dateTime={new Date(value).toISOString()}>{new Intl.DateTimeFormat("fa-IR-u-ca-persian", options).format(new Date(value))}</time>;
}
