import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { useTranslation } from "../lib/i18n";

export default function BlocksCatalog({ embedded = false } = {}) {
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

  const sectionClass = embedded ? "section-shell pt-16 pb-16" : "section-shell pt-16 pb-12";

  const content = (
    <section className={sectionClass}>
      <div className="mb-8 flex flex-col gap-4">
        <p className="text-sm uppercase tracking-[0.3em] text-primary/60">{t("blocks.title")}</p>
        <h1 className="font-display text-3xl md:text-4xl">{t("blocks.subtitle")}</h1>
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
  );

  if (embedded) {
    return content;
  }

  return (
    <div className="bg-quiet-grid">
      {content}
    </div>
  );
}
