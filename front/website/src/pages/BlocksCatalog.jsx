import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { useTranslation } from "../lib/i18n";

export default function BlocksCatalog() {
  const { t } = useTranslation();
  const [blocks, setBlocks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filterType, setFilterType] = useState("all");
  const [filterStatus, setFilterStatus] = useState("all");

  useEffect(() => {
    fetchJSON("/api/blocks")
      .then((res) => setBlocks(res.data || []))
      .catch(() => setBlocks([]))
      .finally(() => setLoading(false));
  }, []);

  const stoneTypes = useMemo(() => {
    const set = new Set();
    blocks.forEach((block) => {
      if (block.stone_type) set.add(block.stone_type);
    });
    return Array.from(set);
  }, [blocks]);

  const filtered = useMemo(() => {
    return blocks.filter((block) => {
      const typeMatch = filterType === "all" || block.stone_type === filterType;
      const statusMatch = filterStatus === "all" || block.status === filterStatus;
      return typeMatch && statusMatch;
    });
  }, [blocks, filterType, filterStatus]);

  const statusLabel = (status) => {
    if (status === "available") return t("blocks.statusAvailable");
    if (status === "reserved") return t("blocks.statusReserved");
    if (status === "sold") return t("blocks.statusSold");
    return status;
  };

  return (
    <div className="bg-quiet-grid">
      <section className="section-shell py-12">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.4em] text-primary/50">{t("blocks.catalogTitle")}</p>
            <h1 className="mt-2 font-display text-3xl text-primary">{t("blocks.subtitle")}</h1>
          </div>
          <div className="flex flex-wrap items-center gap-3 text-sm">
            <label className="flex flex-col gap-1 text-xs text-primary/60">
              {t("blocks.filterType")}
              <select
                className="rounded-full border border-primary/20 bg-white px-4 py-2 text-xs font-semibold text-primary/80"
                value={filterType}
                onChange={(event) => setFilterType(event.target.value)}
              >
                <option value="all">{t("blocks.filterAll")}</option>
                {stoneTypes.map((type) => (
                  <option key={type} value={type}>
                    {type}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1 text-xs text-primary/60">
              {t("blocks.filterStatus")}
              <select
                className="rounded-full border border-primary/20 bg-white px-4 py-2 text-xs font-semibold text-primary/80"
                value={filterStatus}
                onChange={(event) => setFilterStatus(event.target.value)}
              >
                <option value="all">{t("blocks.filterAll")}</option>
                <option value="available">{t("blocks.statusAvailable")}</option>
                <option value="reserved">{t("blocks.statusReserved")}</option>
                <option value="sold">{t("blocks.statusSold")}</option>
              </select>
            </label>
          </div>
        </div>

        <div className="mt-8">
          {loading ? (
            <p className="text-sm text-primary/70">{t("messages.loading")}</p>
          ) : filtered.length === 0 ? (
            <p className="text-sm text-primary/70">{t("blocks.empty")}</p>
          ) : (
            <div className="grid gap-5 md:grid-cols-2 lg:grid-cols-3">
              {filtered.map((block) => (
                <Link
                  to={`/blocks/${block.slug}`}
                  key={block.id}
                  className="group relative overflow-hidden rounded-3xl border border-primary/10 bg-white/70 shadow-[0_18px_40px_rgba(8,58,79,0.12)]"
                >
                  <div className="h-48 overflow-hidden">
                    {block.image_url ? (
                      <img
                        src={resolveImageUrl(block.image_url)}
                        alt=""
                        className="h-full w-full object-cover transition duration-500 group-hover:scale-105"
                      />
                    ) : (
                      <div className="flex h-full items-center justify-center text-xs uppercase tracking-[0.3em] text-primary/40">
                        {t("blocks.title")}
                      </div>
                    )}
                  </div>
                  <div className="space-y-2 p-5">
                    <div className="flex items-center justify-between gap-3">
                      <p className="truncate text-sm font-semibold text-primary">{block.title_fa || block.title_en}</p>
                      <span className="rounded-full border border-primary/20 px-3 py-1 text-[10px] font-semibold uppercase tracking-[0.2em] text-primary/70">
                        {statusLabel(block.status)}
                      </span>
                    </div>
                    <div className="text-xs text-primary/70">
                      {block.stone_type ? `${t("blocks.stoneType")}: ${block.stone_type}` : null}
                    </div>
                    <div className="flex flex-wrap gap-3 text-xs text-primary/60">
                      {block.weight_ton ? <span>{t("blocks.weightTon")}: {block.weight_ton}</span> : null}
                      {block.dimensions ? <span>{t("blocks.dimensions")}: {block.dimensions}</span> : null}
                    </div>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>
      </section>
    </div>
  );
}
