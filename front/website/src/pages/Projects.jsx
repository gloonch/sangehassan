import { Fragment, useEffect, useMemo, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

const getColumnsForWidth = (width) => {
  if (width >= 1024) return 3;
  if (width >= 768) return 2;
  return 1;
};

const pickLocalizedDescription = (project, lang) => {
  if (!project) return "";
  if (lang === "fa") return project.description_fa || project.description_en || project.description_ar || project.description || "";
  if (lang === "ar") return project.description_ar || project.description_en || project.description_fa || project.description || "";
  return project.description_en || project.description_fa || project.description_ar || project.description || "";
};

export default function Projects() {
  const { t, lang } = useTranslation();
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);
  const [columns, setColumns] = useState(() => {
    if (typeof window === "undefined") return 3;
    return getColumnsForWidth(window.innerWidth);
  });
  const [openProjectId, setOpenProjectId] = useState(null);
  const [detailsById, setDetailsById] = useState({});
  const [activeSlideIndex, setActiveSlideIndex] = useState(0);
  const [fullscreenImage, setFullscreenImage] = useState("");
  const [descriptionVisible, setDescriptionVisible] = useState(false);

  useEffect(() => {
    let active = true;
    const load = async () => {
      try {
        const response = await fetchJSON("/api/projects");
        if (!active) return;
        setProjects(response.data || []);
      } catch (_) {
        if (!active) return;
        setProjects([]);
      } finally {
        if (active) setLoading(false);
      }
    };
    load();
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    const onResize = () => setColumns(getColumnsForWidth(window.innerWidth));
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  const fetchProjectDetails = async (projectId) => {
    setDetailsById((prev) => {
      const cached = prev[projectId];
      if (cached?.status === "loading" || cached?.status === "loaded") return prev;
      return {
        ...prev,
        [projectId]: { status: "loading", data: null, error: "" }
      };
    });

    try {
      const response = await fetchJSON(`/api/projects/${projectId}`);
      const data = response?.data || null;
      setDetailsById((prev) => ({
        ...prev,
        [projectId]: { status: "loaded", data, error: "" }
      }));
    } catch (error) {
      setDetailsById((prev) => ({
        ...prev,
        [projectId]: { status: "error", data: null, error: error?.message || t("messages.error") }
      }));
    }
  };

  const handleProjectClick = (projectId) => {
    if (projectId === openProjectId) {
      setOpenProjectId(null);
      setActiveSlideIndex(0);
      return;
    }

    setOpenProjectId(projectId);
    setActiveSlideIndex(0);
    fetchProjectDetails(projectId);
  };

  const openIndex = useMemo(
    () => projects.findIndex((project) => project.id === openProjectId),
    [projects, openProjectId]
  );

  const insertAfterIndex = useMemo(() => {
    if (openIndex < 0) return -1;
    const row = Math.floor(openIndex / columns);
    return Math.min((row + 1) * columns - 1, projects.length - 1);
  }, [openIndex, columns, projects.length]);

  const openProjectDetails = openProjectId ? detailsById[openProjectId] : null;
  const galleryImages = openProjectDetails?.data?.gallery_images || [];
  const localizedDescription = pickLocalizedDescription(openProjectDetails?.data, lang) || t("projects.noDescription");

  useEffect(() => {
    if (!openProjectId || openProjectDetails?.status !== "loaded") return;
    setDescriptionVisible(false);
    const raf = window.requestAnimationFrame(() => setDescriptionVisible(true));
    return () => window.cancelAnimationFrame(raf);
  }, [openProjectId, activeSlideIndex, openProjectDetails?.status, lang]);

  useEffect(() => {
    if (!fullscreenImage) return undefined;

    const onKeyDown = (event) => {
      if (event.key === "Escape") setFullscreenImage("");
    };

    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", onKeyDown);

    return () => {
      document.body.style.overflow = prevOverflow;
      window.removeEventListener("keydown", onKeyDown);
    };
  }, [fullscreenImage]);

  useEffect(() => {
    if (!openProjectId || galleryImages.length === 0) return;
    if (activeSlideIndex > galleryImages.length - 1) setActiveSlideIndex(0);
  }, [openProjectId, galleryImages, activeSlideIndex]);

  return (
    <section className="section-shell ">
      <div className="mx-auto w-full max-w-5xl">
        <div className="mb-8 flex min-h-[30vh] flex-col justify-end border-b border-primary/25 pb-6 md:mb-10 md:pb-8">
          <p className="text-xs uppercase tracking-[0.42em] text-primary/55 md:text-sm">{t("projects.title")}</p>
          <h1 className="mt-3 font-display text-4xl leading-[1.02] tracking-tight text-primary md:text-6xl lg:text-7xl">
            {t("projects.subtitle")}
          </h1>
        </div>

        {loading ? (
          <div className="grid grid-cols-1 gap-0 border-l border-t border-primary/20 md:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 9 }).map((_, index) => (
              <div key={index} className="aspect-[3/4] animate-pulse border-b border-r border-primary/20 bg-primary/10" />
            ))}
          </div>
        ) : projects.length === 0 ? (
          <p className="text-sm text-primary/70">{t("projects.empty")}</p>
        ) : (
          <div className="grid grid-cols-1 gap-0 border-l border-t border-primary/20 md:grid-cols-2 lg:grid-cols-3">
            {projects.map((project, index) => (
              <Fragment key={project.id}>
                <button
                  type="button"
                  onClick={() => handleProjectClick(project.id)}
                  className={`group relative aspect-[3/4] overflow-hidden border-b border-r transition ${openProjectId === project.id
                    ? "border-accent"
                    : "border-primary/20 hover:border-primary/40"
                    }`}
                >
                  {project.cover_image_url ? (
                    <img
                      src={resolveImageUrl(project.cover_image_url)}
                      alt="Project cover"
                      className="h-full w-full object-cover object-center"
                      loading="lazy"
                    />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center text-xs text-primary/60">
                      {t("messages.empty")}
                    </div>
                  )}
                  <span className="pointer-events-none absolute inset-0 bg-gradient-to-t from-primary/20 via-transparent to-transparent opacity-0 transition group-hover:opacity-100" />
                </button>

                {openProjectId && index === insertAfterIndex && (
                  <div className="col-span-1 border-b border-r border-primary/20 md:col-span-2 lg:col-span-3">
                    {openProjectDetails?.status === "loading" || !openProjectDetails ? (
                      <div className="space-y-2 p-4 animate-pulse">
                        <div className="h-4 w-2/3 bg-primary/10" />
                        <div className="h-4 w-1/2 bg-primary/10" />
                        <div className="h-56 bg-primary/10" />
                      </div>
                    ) : openProjectDetails.status === "error" ? (
                      <p className="p-4 text-sm text-red-500">{openProjectDetails.error || t("messages.error")}</p>
                    ) : galleryImages.length > 0 ? (
                      <div className="relative w-full overflow-hidden">
                        <button
                          type="button"
                          onClick={() => setActiveSlideIndex((prev) => Math.max(prev - 1, 0))}
                          disabled={activeSlideIndex === 0}
                          className="absolute left-4 top-1/2 z-20 -translate-y-1/2 bg-transparent p-0 text-5xl font-light leading-none text-white drop-shadow-[0_3px_6px_rgba(0,0,0,0.55)] transition disabled:cursor-not-allowed disabled:opacity-35 md:text-6xl"
                          aria-label={t("projects.prev")}
                        >
                          {"<"}
                        </button>

                        <button
                          type="button"
                          onClick={() =>
                            setActiveSlideIndex((prev) => Math.min(prev + 1, galleryImages.length - 1))
                          }
                          disabled={activeSlideIndex === galleryImages.length - 1}
                          className="absolute right-4 top-1/2 z-20 -translate-y-1/2 bg-transparent p-0 text-5xl font-light leading-none text-white drop-shadow-[0_3px_6px_rgba(0,0,0,0.55)] transition disabled:cursor-not-allowed disabled:opacity-35 md:text-6xl"
                          aria-label={t("projects.next")}
                        >
                          {">"}
                        </button>

                        <button
                          type="button"
                          onClick={() => setFullscreenImage(galleryImages[activeSlideIndex])}
                          className="block h-full w-full"
                        >
                          <img
                            src={resolveImageUrl(galleryImages[activeSlideIndex])}
                            alt="Project gallery"
                            className="h-[420px] w-full cursor-zoom-in object-cover object-center md:h-[620px]"
                            loading="lazy"
                          />
                        </button>

                        <div
                          className={`absolute inset-x-0 bottom-0 z-10 bg-gradient-to-t from-primary/88 via-primary/50 to-transparent px-6 pb-7 pt-20 text-base leading-8 text-white md:text-xl md:leading-10 transition-opacity duration-500 ${descriptionVisible ? "opacity-100" : "opacity-0"
                            }`}
                        >
                          {localizedDescription}
                        </div>
                      </div>
                    ) : (
                      <div className="bg-primary/5 p-5 text-base leading-8 text-primary/80 md:text-lg md:leading-9">
                        {localizedDescription}
                      </div>
                    )}
                  </div>
                )}
              </Fragment>
            ))}
          </div>
        )}
      </div>

      {fullscreenImage && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-primary/88 px-4 py-8"
          onClick={() => setFullscreenImage("")}
        >
          <button
            type="button"
            onClick={() => setFullscreenImage("")}
            className="absolute right-4 top-4 border border-white/40 bg-white/10 px-3 py-2 text-lg text-white"
            aria-label={t("actions.close")}
          >
            x
          </button>
          <img
            src={resolveImageUrl(fullscreenImage)}
            alt="Project fullscreen"
            className="max-h-full max-w-full object-contain"
            onClick={(event) => event.stopPropagation()}
          />
        </div>
      )}
    </section>
  );
}
