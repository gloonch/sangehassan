import { Search, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";

const labels = { orders: "سفارش‌ها", customers: "مشتریان", proformas: "پیش‌فاکتورها", shipments: "محموله‌ها", batches: "بچ‌ها", payments: "پرداخت‌ها" };

export default function GlobalSearch() {
  const navigate = useNavigate();
  const box = useRef(null);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState(null);
  const [error, setError] = useState("");
  useEffect(() => {
    if (query.trim().length < 2) { setResults(null); setError(""); return undefined; }
    let active = true;
    const timer = setTimeout(() => fetchJSON(`/api/v1/search?q=${encodeURIComponent(query.trim())}`).then((response) => active && setResults(response.data || {})).catch((e) => active && setError(e.message)), 250);
    return () => { active = false; clearTimeout(timer); };
  }, [query]);
  useEffect(() => {
    const close = (event) => { if (event.key === "Escape") { setQuery(""); setResults(null); } if (box.current && !box.current.contains(event.target)) setResults(null); };
    document.addEventListener("keydown", close); document.addEventListener("mousedown", close);
    return () => { document.removeEventListener("keydown", close); document.removeEventListener("mousedown", close); };
  }, []);
  const choose = (item) => { setQuery(""); setResults(null); navigate(item.path.replace(/^\/panel/, "")); };
  const hasResults = results && Object.values(results).some((items) => items.length);
  return <div ref={box} className="relative w-full max-w-md" dir="rtl">
    <Search aria-hidden="true" className="absolute right-3 top-3 h-4 w-4 text-primary/45" />
    <label className="sr-only" htmlFor="global-search">جست‌وجوی سراسری</label>
    <input id="global-search" value={query} onChange={(e) => { setQuery(e.target.value); setError(""); }} className="w-full rounded-full border bg-white py-2.5 pl-10 pr-10 text-sm" placeholder="شماره سفارش، مشتری، محموله…" autoComplete="off" />
    {query && <button type="button" onClick={() => { setQuery(""); setResults(null); }} aria-label="پاک‌کردن جست‌وجو" className="absolute left-3 top-2.5 rounded-full p-1"><X className="h-4 w-4" /></button>}
    {(results || error) && <div className="absolute top-12 z-50 max-h-[70vh] w-full overflow-auto rounded-2xl border bg-white p-3 shadow-xl">
      {error && <p className="p-3 text-sm text-red-700" role="alert">{error}</p>}
      {!error && !hasResults && <p className="p-3 text-sm text-primary/55">نتیجه‌ای با دسترسی فعلی شما پیدا نشد.</p>}
      {!error && Object.entries(results || {}).map(([group, items]) => items.length ? <section key={group} className="mb-3 last:mb-0"><h3 className="px-2 py-1 text-xs font-semibold text-primary/50">{labels[group] || group}</h3>{items.map((item) => <button key={`${group}-${item.id}`} type="button" onClick={() => choose(item)} className="block w-full rounded-xl px-3 py-2 text-right hover:bg-primary/5"><span className="block text-sm font-medium">{item.title}</span><span className="text-xs text-primary/50">{item.subtitle}</span></button>)}</section> : null)}
    </div>}
  </div>;
}
