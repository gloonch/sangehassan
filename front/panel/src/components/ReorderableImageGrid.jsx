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
  labels = {}
}) {
  const selectedImage = images[selectedIndex] || images[0] || "";

  if (!images.length) {
    return <p className="text-xs text-primary/50">{labels.empty}</p>;
  }

  return (
    <div className="space-y-3" dir="ltr">
      {showPreview && selectedImage ? (
        <div className={`overflow-hidden rounded-2xl border border-primary/15 bg-primary/5 ${previewClassName}`}>
          <img src={resolveImageUrl(selectedImage)} alt="" className="h-full w-full object-cover" />
        </div>
      ) : null}

      <div className={gridClassName}>
        {images.map((url, index) => (
          <div key={`${url}-${index}`} className="min-w-0 space-y-2">
            <div className="relative">
              <button
                type="button"
                onClick={() => onSelect?.(index)}
                className={`block w-full overflow-hidden rounded-xl border bg-white ${
                  onSelect && selectedIndex === index ? "border-accent" : "border-primary/15"
                } ${thumbnailClassName}`}
              >
                <img src={resolveImageUrl(url)} alt="" className="h-full w-full object-cover" />
              </button>
              <button
                type="button"
                onClick={() => onRemove?.(index)}
                className="absolute right-1 top-1 flex h-6 w-6 items-center justify-center rounded-full bg-white/95 text-primary shadow"
                aria-label={labels.remove}
                title={labels.remove}
              >
                <X size={13} />
              </button>
            </div>

            <div className="flex items-center justify-center gap-1">
              <button
                type="button"
                onClick={() => onMove?.(index, -1)}
                disabled={index === 0}
                className={buttonBaseClass}
                aria-label={labels.moveLeft}
                title={labels.moveLeft}
              >
                <ArrowLeft size={15} />
              </button>
              <button
                type="button"
                onClick={() => onMove?.(index, 1)}
                disabled={index === images.length - 1}
                className={buttonBaseClass}
                aria-label={labels.moveRight}
                title={labels.moveRight}
              >
                <ArrowRight size={15} />
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
