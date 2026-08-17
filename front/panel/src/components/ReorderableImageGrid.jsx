import { ArrowLeft, ArrowRight, X } from "lucide-react";
import { resolveImageUrl } from "../lib/assets";

const buttonBaseClass =
  "flex h-8 w-8 items-center justify-center rounded-full border border-primary/15 bg-white text-primary/70 transition hover:border-primary/40 hover:text-primary disabled:cursor-not-allowed disabled:opacity-35 disabled:hover:border-primary/15 disabled:hover:text-primary/70";

export default function ReorderableImageGrid({
  images = [],
  selectedIndex = 0,
  onSelect,
  onRemove,
  onMove,
  showPreview = true,
  previewClassName = "h-40",
  gridClassName = "grid grid-cols-3 gap-3 md:grid-cols-5",
  thumbnailClassName = "h-20",
  showOrderNumbers = false,
  isItemLocked = () => false,
  labels = {}
}) {
  const resolvedSelectedIndex = images[selectedIndex] ? selectedIndex : 0;
  const selectedImage = images[resolvedSelectedIndex] || "";

  if (!images.length) {
    return <p className="text-xs text-primary/50">{labels.empty}</p>;
  }

  return (
    <div className="space-y-3" dir="ltr">
      {showPreview && selectedImage ? (
        <div className={`relative overflow-hidden rounded-2xl border border-primary/15 bg-primary/5 ${previewClassName}`}>
          <img src={resolveImageUrl(selectedImage)} alt="" className="h-full w-full object-cover" />
          {showOrderNumbers ? (
            <span className="absolute bottom-2 left-2 rounded-full bg-primary/85 px-3 py-1 text-[11px] font-semibold text-white shadow">
              {labels.slide || "Slide"} {resolvedSelectedIndex + 1} / {images.length}
            </span>
          ) : null}
        </div>
      ) : null}

      <div className={gridClassName}>
        {images.map((url, index) => {
          const itemLocked = isItemLocked(url, index);
          const previousLocked = index > 0 && isItemLocked(images[index - 1], index - 1);
          const nextLocked = index < images.length - 1 && isItemLocked(images[index + 1], index + 1);
          return (
            <div key={`${url}-${index}`} className="min-w-0 space-y-2">
              <div className="relative">
                <button
                  type="button"
                  onClick={() => onSelect?.(index)}
                  disabled={!onSelect}
                  className={`block w-full overflow-hidden rounded-xl border bg-white ${
                    onSelect && resolvedSelectedIndex === index ? "border-accent" : "border-primary/15"
                  } ${thumbnailClassName}`}
                >
                  <img
                    src={resolveImageUrl(url)}
                    alt={labels.slide ? `${labels.slide} ${index + 1}` : ""}
                    className="h-full w-full object-cover"
                  />
                </button>
                {showOrderNumbers ? (
                  <span className="pointer-events-none absolute bottom-1 left-1 flex h-6 min-w-6 items-center justify-center rounded-full bg-primary/85 px-1.5 text-[10px] font-bold text-white shadow">
                    {index + 1}
                  </span>
                ) : null}
                {onRemove && !itemLocked ? (
                  <button
                    type="button"
                    onClick={() => onRemove(index)}
                    className="absolute right-1 top-1 flex h-6 w-6 items-center justify-center rounded-full bg-white/95 text-primary shadow"
                    aria-label={labels.remove}
                    title={labels.remove}
                  >
                    <X size={13} />
                  </button>
                ) : null}
              </div>

              {onMove && !itemLocked ? (
                <div className="flex items-center justify-center gap-1">
                  <button
                    type="button"
                    onClick={() => onMove(index, -1)}
                    disabled={index === 0 || previousLocked}
                    className={buttonBaseClass}
                    aria-label={labels.moveLeft}
                    title={labels.moveLeft}
                  >
                    <ArrowLeft size={15} />
                  </button>
                  <button
                    type="button"
                    onClick={() => onMove(index, 1)}
                    disabled={index === images.length - 1 || nextLocked}
                    className={buttonBaseClass}
                    aria-label={labels.moveRight}
                    title={labels.moveRight}
                  >
                    <ArrowRight size={15} />
                  </button>
                </div>
              ) : null}
            </div>
          );
        })}
      </div>
    </div>
  );
}
