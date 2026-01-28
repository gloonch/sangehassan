import { useEffect, useMemo, useRef, useState } from "react";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { useTranslation } from "../lib/i18n";

const easeClass = "transition-[transform,opacity] duration-700 ease-[cubic-bezier(0.22,1,0.36,1)]";

const getRelativePosition = (index, activeIndex, total) => {
  if (total <= 1) return 0;
  const prevIndex = (activeIndex - 1 + total) % total;
  const nextIndex = (activeIndex + 1) % total;
  if (index === activeIndex) return 0;
  if (index === prevIndex) return -1;
  if (index === nextIndex) return 1;
  return 2;
};

export default function MaterialPreviewStudio() {
  const { t } = useTranslation();
  const [templates, setTemplates] = useState([]);
  const [stones, setStones] = useState([]);
  const [templateIndex, setTemplateIndex] = useState(0);
  const [stoneIndex, setStoneIndex] = useState(0);
  const autoStepRef = useRef(0);
  const swipeStart = useRef(null);

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      try {
        const [templateRes, stoneRes] = await Promise.all([
          fetchJSON("/api/templates?active=true"),
          fetchJSON("/api/products?popular=true")
        ]);
        if (!mounted) return;
        setTemplates(templateRes.data || []);
        setStones((stoneRes.data || []).filter((item) => item.image_url));
      } catch (error) {
        if (!mounted) return;
        setTemplates([]);
        setStones([]);
      }
    };
    load();
    return () => {
      mounted = false;
    };
  }, []);

  useEffect(() => {
    if (templateIndex >= templates.length) {
      setTemplateIndex(0);
    }
  }, [templates.length, templateIndex]);

  useEffect(() => {
    if (stoneIndex >= stones.length) {
      setStoneIndex(0);
    }
  }, [stones.length, stoneIndex]);

  const nextStone = (manual = false) => {
    if (stones.length <= 1) return;
    setStoneIndex((prev) => (prev + 1) % stones.length);
    if (manual) autoStepRef.current = 0;
  };

  const prevStone = (manual = false) => {
    if (stones.length <= 1) return;
    setStoneIndex((prev) => (prev - 1 + stones.length) % stones.length);
    if (manual) autoStepRef.current = 0;
  };

  const nextTemplate = (manual = false) => {
    if (templates.length <= 1) return;
    setTemplateIndex((prev) => (prev + 1) % templates.length);
    if (manual) autoStepRef.current = 0;
  };

  const prevTemplate = (manual = false) => {
    if (templates.length <= 1) return;
    setTemplateIndex((prev) => (prev - 1 + templates.length) % templates.length);
    if (manual) autoStepRef.current = 0;
  };

  useEffect(() => {
    if (stones.length <= 1 && templates.length <= 1) return;
    const interval = setInterval(() => {
      if (stones.length > 1) {
        nextStone();
        autoStepRef.current += 1;
        if (autoStepRef.current >= 3) {
          autoStepRef.current = 0;
          if (templates.length > 1) {
            nextTemplate();
          }
        }
      }
    }, 3000);
    return () => clearInterval(interval);
  }, [stones.length, templates.length]);

  const handlePointerDown = (event) => {
    swipeStart.current = {
      x: event.clientX,
      y: event.clientY
    };
  };

  const handlePointerUp = (event) => {
    if (!swipeStart.current) return;
    const deltaX = event.clientX - swipeStart.current.x;
    const deltaY = event.clientY - swipeStart.current.y;
    swipeStart.current = null;

    const threshold = 40;
    if (Math.abs(deltaX) < threshold && Math.abs(deltaY) < threshold) return;

    if (Math.abs(deltaX) > Math.abs(deltaY)) {
      if (deltaX < 0) {
        nextStone(true);
      } else {
        prevStone(true);
      }
    } else {
      if (deltaY < 0) {
        prevTemplate(true);
      } else {
        nextTemplate(true);
      }
    }
  };

  const handlePointerCancel = () => {
    swipeStart.current = null;
  };

  const templateItems = useMemo(() => templates, [templates]);
  const stoneItems = useMemo(() => stones, [stones]);

  return (
    <div className="relative">
      <div
        className="relative mx-auto flex w-full max-w-[640px] select-none items-center justify-center touch-none py-12"
        onPointerDown={handlePointerDown}
        onPointerUp={handlePointerUp}
        onPointerLeave={handlePointerCancel}
        onPointerCancel={handlePointerCancel}
      >
        <div className="relative aspect-square w-full max-w-[560px]">
          {/* Stone Layer */}
          <div className="relative h-full w-full overflow-hidden rounded-[36px]">
            {stoneItems.length === 0 && (
              <div className="flex h-full w-full items-center justify-center text-sm text-primary/50">
                {t("messages.empty")}
              </div>
            )}
            {stoneItems.map((stone, index) => {
              const position = getRelativePosition(index, stoneIndex, stoneItems.length);
              const baseStyle = "absolute inset-0";
              if (position === 2) {
                return (
                  <div key={stone.id || index} className={`${baseStyle} opacity-0`} />
                );
              }
              const translate = position === 0 ? "translate-x-0" : position === -1 ? "-translate-x-[90%]" : "translate-x-[90%]";
              const opacity = position === 0 ? "opacity-100" : "opacity-40";
              const scale = position === 0 ? "scale-100" : "scale-95";
              const zIndex = position === 0 ? "z-20" : "z-10";
              return (
                <div
                  key={stone.id || index}
                  className={`${baseStyle} ${translate} ${opacity} ${scale} ${zIndex} ${easeClass} bg-cover bg-center`}
                  style={{ backgroundImage: `url(${resolveImageUrl(stone.image_url)})` }}
                />
              );
            })}
          </div>

          {/* Template Layer */}
          <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
            <div className="relative h-[70%] w-[70%]">
              <div className="relative h-full w-full overflow-hidden">
                {templateItems.length === 0 && (
                  <div className="flex h-full w-full items-center justify-center rounded-3xl border border-dashed border-primary/30 text-xs text-primary/50">
                    {t("messages.empty")}
                  </div>
                )}
                {templateItems.map((template, index) => {
                  const position = getRelativePosition(index, templateIndex, templateItems.length);
                  if (position === 2) {
                    return <div key={template.id || index} className="absolute inset-0 opacity-0" />;
                  }
                  const translate = position === 0 ? "translate-y-0" : position === -1 ? "-translate-y-[90%]" : "translate-y-[90%]";
                  const opacity = position === 0 ? "opacity-100" : "opacity-40";
                  const scale = position === 0 ? "scale-100" : "scale-95";
                  const zIndex = position === 0 ? "z-20" : "z-10";
                  return (
                    <div key={template.id || index} className={`absolute inset-0 ${translate} ${opacity} ${scale} ${zIndex} ${easeClass}`}>
                      <img src={resolveImageUrl(template.image_url)} alt={template.name} className="h-full w-full object-contain" />
                    </div>
                  );
                })}
              </div>
            </div>
          </div>

          {/* Buttons */}
          <button
            type="button"
            onClick={() => prevTemplate(true)}
            className="pointer-events-auto absolute -top-12 left-1/2 -translate-x-1/2 rounded-full border border-white/20 bg-white/10 backdrop-blur-md p-3 text-primary/80 shadow-lg hover:bg-white/20 transition-colors"
            aria-label={t("preview.prevTemplate")}
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
              <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 15.75l7.5-7.5 7.5 7.5" />
            </svg>
          </button>

          <button
            type="button"
            onClick={() => nextTemplate(true)}
            className="pointer-events-auto absolute -bottom-12 left-1/2 -translate-x-1/2 rounded-full border border-white/20 bg-white/10 backdrop-blur-md p-3 text-primary/80 shadow-lg hover:bg-white/20 transition-colors"
            aria-label={t("preview.nextTemplate")}
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
              <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
            </svg>
          </button>

          <button
            type="button"
            onClick={() => prevStone(true)}
            className="pointer-events-auto absolute top-1/2 -left-16 -translate-y-1/2 rounded-full border border-white/20 bg-white/10 backdrop-blur-md p-3 text-primary/80 shadow-lg hover:bg-white/20 transition-colors"
            aria-label={t("preview.prevStone")}
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
                <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 19.5L8.25 12l7.5-7.5" />
            </svg>
          </button>

          <button
            type="button"
            onClick={() => nextStone(true)}
            className="pointer-events-auto absolute top-1/2 -right-16 -translate-y-1/2 rounded-full border border-white/20 bg-white/10 backdrop-blur-md p-3 text-primary/80 shadow-lg hover:bg-white/20 transition-colors"
            aria-label={t("preview.nextStone")}
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
              <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
            </svg>
          </button>

        </div>
      </div>
    </div>
  );
}
