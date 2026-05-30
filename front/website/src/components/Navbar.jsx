import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Link, NavLink, useLocation } from "react-router-dom";
import { gsap } from "gsap";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { getLiveDealsConfig, renderDealMessage } from "../lib/liveDeals";
import { getTileCompletionTime } from "../lib/routeReveal";
import logoImage from "@shared/assets/logo.png";
import logoWhiteImage from "@shared/assets/logo_white.png";

const navItems = [
  { key: "products", path: "/products" },
  { key: "blocks", path: "/blocks" },
  { key: "projects", path: "/projects" },
  { key: "ads", path: "/ads" },
  { key: "blogs", path: "/blogs" },
  { key: "about", path: "/about" },
];

const getUserDisplayName = (user) => {
  const fullName = typeof user?.full_name === "string" ? user.full_name.trim() : "";
  if (fullName) return fullName;
  const phone = typeof user?.phone === "string" ? user.phone.trim() : "";
  if (phone) return phone;
  return "";
};

export default function Navbar() {
  const { t, lang } = useTranslation();
  const location = useLocation();
  const headerRef = useRef(null);
  const mobileMenuPanelRef = useRef(null);
  const [open, setOpen] = useState(false);
  const [user, setUser] = useState(null);
  const [dealIndex, setDealIndex] = useState(0);
  const [dealVisible, setDealVisible] = useState(true);
  const isHome = location.pathname === "/";
  const isAbout = location.pathname === "/about";
  const navSubline = isHome || isAbout ? t("nav.sinceLine") : t("nav.sinceLineAlt");
  const liveDealsConfig = getLiveDealsConfig(lang);
  const liveDeals = liveDealsConfig.deals;
  const activeDeal = liveDeals.length > 0 ? liveDeals[dealIndex % liveDeals.length] : null;

  useEffect(() => {
    let active = true;
    const restore = () => {
      try {
        const stored = sessionStorage.getItem("sh_me");
        if (stored) {
          const parsed = JSON.parse(stored);
          if (parsed?.id) {
            setUser(parsed);
          }
        }
      } catch (_) {
        /* ignore */
      }
    };

    const fetchMe = async () => {
      try {
        const res = await fetchJSON("/api/v1/me");
        const me = res?.data || res;
        if (!active) return;
        setUser(me);
        sessionStorage.setItem("sh_me", JSON.stringify(me));
      } catch (_) {
        if (!active) return;
        setUser(null);
        sessionStorage.removeItem("sh_me");
      }
    };

    restore();
    fetchMe();

    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    const handler = (event) => {
      if (event?.detail) {
        setUser(event.detail);
        return;
      }
      try {
        const stored = sessionStorage.getItem("sh_me");
        if (stored) {
          setUser(JSON.parse(stored));
        }
      } catch (_) {
        /* ignore */
      }
    };
    window.addEventListener("sh:me-updated", handler);
    return () => window.removeEventListener("sh:me-updated", handler);
  }, []);

  useEffect(() => {
    setOpen(false);
  }, [location.pathname]);

  useEffect(() => {
    if (!liveDeals.length) return undefined;

    let fadeTimer;
    setDealIndex((current) => current % liveDeals.length);
    setDealVisible(true);

    const interval = window.setInterval(() => {
      setDealVisible(false);
      fadeTimer = window.setTimeout(() => {
        setDealIndex((current) => (current + 1) % liveDeals.length);
        setDealVisible(true);
      }, 280);
    }, 2000);

    return () => {
      window.clearInterval(interval);
      window.clearTimeout(fadeTimer);
    };
  }, [lang, liveDeals.length]);

  useEffect(() => {
    if (typeof document === "undefined") return;
    if (!open) return;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, [open]);

  useEffect(() => {
    if (!open || typeof window === "undefined" || !window.matchMedia) return;
    const panel = mobileMenuPanelRef.current;
    if (!panel) return;
    const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const rows = panel.querySelectorAll("[data-mobile-item='true']");

    const ctx = gsap.context(() => {
      gsap.fromTo(
        panel,
        { xPercent: 16 },
        {
          xPercent: 0,
          duration: reduceMotion ? 0.3 : 0.65,
          ease: "power3.out",
          force3D: true
        }
      );

      if (rows.length) {
        gsap.fromTo(
          rows,
          { x: 20, autoAlpha: 0 },
          {
            x: 0,
            autoAlpha: 1,
            duration: reduceMotion ? 0.22 : 0.42,
            delay: reduceMotion ? 0.06 : 0.14,
            stagger: reduceMotion ? 0.012 : 0.04,
            ease: "power2.out"
          }
        );
      }
    }, panel);

    return () => ctx.revert();
  }, [open]);

  useEffect(() => {
    const header = headerRef.current;
    if (!header || typeof window === "undefined" || !window.matchMedia) return;
    const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const isLandingPage = location.pathname === "/";

    const left = header.querySelector("[data-nav-block='left']");
    const center = header.querySelector("[data-nav-block='center']");
    const rightCandidates = Array.from(header.querySelectorAll("[data-nav-block='right']"));
    const right =
      rightCandidates.find((element) => element.offsetParent !== null) ||
      rightCandidates[0] ||
      null;
    const blocks = [left, center, right].filter(Boolean);
    if (!blocks.length) return;

    if (!isLandingPage) {
      gsap.set(blocks, { autoAlpha: 1, y: 0, clearProps: "transform,opacity" });
      return;
    }

    const ctx = gsap.context(() => {
      gsap.set(blocks, { autoAlpha: 0, y: 26, force3D: true });

      const timeline = gsap.timeline();
      blocks.forEach((block, index) => {
        timeline.to(
          block,
          {
            autoAlpha: 1,
            y: 0,
            duration: reduceMotion ? 0.38 : 0.82,
            ease: "power3.out",
            clearProps: "transform,opacity"
          },
          getTileCompletionTime(index, reduceMotion)
        );
      });
    }, header);

    return () => ctx.revert();
  }, [location.pathname]);

  const displayName = getUserDisplayName(user);
  const avatarSource = displayName.startsWith("+") ? displayName.slice(1) : displayName;
  const avatarChar = (avatarSource || "U").trim().charAt(0).toUpperCase();
  const profileLabel = displayName || t("nav.profile");
  const visibleNavItems = navItems;
  const navHeaderClass = isHome
    ? open
      ? "border-b border-sand/25 bg-primary/45 backdrop-blur-xl"
      : "border-transparent bg-transparent"
    : "border-b border-primary/10 bg-sand/80 backdrop-blur-lg";
  const navTextClass = isHome
    ? "text-sand/80 hover:text-sand"
    : "text-primary/70 hover:text-primary";
  const navActiveClass = isHome ? "text-sand" : "text-accent";
  const mobileOverlayBaseClass = "fixed inset-0 z-[9999] bg-[#E5E1DD] text-primary lg:hidden";

  const mobileMenu =
    open && typeof document !== "undefined"
      ? createPortal(
          <div className={mobileOverlayBaseClass} style={{ backgroundColor: "#E5E1DD", opacity: 1 }}>
            <div ref={mobileMenuPanelRef} className="h-full w-full">
              <div className="section-shell flex h-20 items-center justify-between gap-4">
                <Link to="/" onClick={() => setOpen(false)} className="inline-flex items-center" aria-label={t("brand")}>
                  <img src={logoImage} alt={t("brand")} className="h-12 w-auto object-contain" />
                </Link>
                <button
                  type="button"
                  className="inline-flex h-10 w-10 items-center justify-center rounded-full text-primary transition hover:bg-primary/10"
                  onClick={() => setOpen(false)}
                  aria-label={t("actions.close")}
                >
                  <span className="relative block h-4 w-4">
                    <span className="absolute left-0 top-[7px] h-0.5 w-full -rotate-45 rounded-full bg-primary" />
                    <span className="absolute left-0 top-[7px] h-0.5 w-full rotate-45 rounded-full bg-primary" />
                  </span>
                </button>
              </div>

              <div className="section-shell flex h-[calc(100dvh-5rem)] flex-col items-center justify-center overflow-y-auto py-8 text-center">
                <nav className="flex w-full max-w-md flex-col items-center">
                  {visibleNavItems.map((item) => (
                    <NavLink
                      key={item.key}
                      to={item.path}
                      onClick={() => setOpen(false)}
                      end={item.end}
                      data-mobile-item="true"
                      className={({ isActive }) =>
                        `w-full border-b border-primary/15 py-4 text-center text-xl font-semibold tracking-[0.02em] text-primary transition hover:text-primary ${
                          isActive ? "border-primary/35 text-primary" : "text-primary/80"
                        }`
                      }
                    >
                      {t(`nav.${item.key}`)}
                    </NavLink>
                  ))}
                </nav>

                <div className="mt-6 w-full max-w-md border-t border-primary/15 pt-6">
                  {user ? (
                    <NavLink
                      to="/profile"
                      onClick={() => setOpen(false)}
                      data-mobile-item="true"
                      className="mx-auto inline-flex items-center gap-3 rounded-full border border-primary/20 bg-primary px-4 py-2 text-sm font-semibold uppercase text-sand transition hover:bg-primary/90"
                      aria-label={profileLabel}
                      title={profileLabel}
                    >
                      <span className="flex h-8 w-8 items-center justify-center rounded-full bg-sand/20 text-xs">
                        {avatarChar}
                      </span>
                      <span>{profileLabel}</span>
                    </NavLink>
                  ) : (
                    <NavLink
                      to="/login"
                      onClick={() => setOpen(false)}
                      data-mobile-item="true"
                      className="w-full rounded-full bg-primary px-4 py-3 text-center text-sm font-semibold text-sand transition hover:bg-primary/90"
                    >
                      {t("actions.loginRegister")}
                    </NavLink>
                  )}
                </div>
              </div>
            </div>
          </div>,
          document.body
        )
      : null;

  return (
    <header ref={headerRef} className={`fixed left-0 right-0 top-0 z-50 transition-colors duration-300 ${navHeaderClass}`}>
      <div className="section-shell flex h-20 items-center justify-between gap-4">
        <Link to="/" className="inline-flex items-center will-change-transform" aria-label={t("brand")} data-nav-block="left">
          <img src={isHome ? logoWhiteImage : logoImage} alt={t("brand")} className="h-12 w-auto object-contain" />
        </Link>

        <nav className="hidden items-center gap-6 text-sm font-medium lg:flex will-change-transform" data-nav-block="center">
          {visibleNavItems.map((item) => (
            <NavLink
              key={item.key}
              to={item.path}
              end={item.end}
              className={({ isActive }) =>
                `transition ${isActive ? navActiveClass : navTextClass}`
              }
            >
              {t(`nav.${item.key}`)}
            </NavLink>
          ))}
        </nav>

        <div className="hidden items-center gap-3 lg:flex will-change-transform" data-nav-block="right">
          {user ? (
            <NavLink
              to="/profile"
              className={`flex h-10 w-10 items-center justify-center rounded-full text-sm font-semibold uppercase shadow transition ${isHome
                ? "border border-sand/30 bg-sand/15 text-sand hover:bg-sand/25"
                : "bg-primary text-sand hover:bg-primary/90"
                }`}
              aria-label={profileLabel}
              title={profileLabel}
            >
              {avatarChar}
            </NavLink>
          ) : (
            <NavLink
              to="/login"
              className={`rounded-full px-4 py-2 text-xs font-semibold transition ${isHome
                ? "border border-sand/35 bg-sand/15 text-sand hover:bg-sand/25"
                : "bg-primary text-sand hover:bg-primary/90"
                }`}
            >
              {t("actions.loginRegister")}
            </NavLink>
          )}
        </div>

        <button
          type="button"
          data-nav-block="right"
          className={`flex h-10 w-10 items-center justify-center rounded-full text-base font-semibold leading-none transition will-change-transform lg:hidden ${isHome
            ? "bg-sand/10 text-sand hover:bg-sand/20"
            : "bg-transparent text-primary hover:bg-primary/10"
            }`}
          onClick={() => setOpen((prev) => !prev)}
          aria-expanded={open}
          aria-label={t("actions.menu")}
        >
          <span className="sr-only">{t("actions.menu")}</span>
          <span className="relative block h-4 w-5">
            <span
              className={`absolute left-0 top-0 h-0.5 w-full rounded-full transition-all duration-300 ${isHome ? "bg-sand" : "bg-primary"} ${open ? "translate-y-[7px] rotate-45" : ""
                }`}
            />
            <span
              className={`absolute left-0 top-[7px] h-0.5 w-full rounded-full transition-all duration-300 ${isHome ? "bg-sand" : "bg-primary"} ${open ? "opacity-0" : "opacity-100"
                }`}
            />
            <span
              className={`absolute left-0 top-[14px] h-0.5 w-full rounded-full transition-all duration-300 ${isHome ? "bg-sand" : "bg-primary"} ${open ? "-translate-y-[7px] -rotate-45" : ""
                }`}
            />
          </span>
        </button>
      </div>

      <div className={`border-t ${isHome ? "border-sand/20" : "border-primary/10"}`}>
        <div className="section-shell relative flex h-8 items-center">
          <p className={`min-w-[7.5rem] text-[11px] font-semibold ${isHome ? "text-sand/75" : "text-primary/65"}`}>
            {navSubline}
          </p>
          {activeDeal ? (
            <Link
              to="/ads"
              className={`absolute left-1/2 top-1/2 hidden w-[min(48vw,34rem)] items-center justify-center overflow-hidden text-center text-[11px] font-semibold transition-all duration-300 ease-out lg:flex ${dealVisible ? "opacity-100" : "opacity-0"
                } ${isHome ? "text-sand/75 hover:text-sand/75" : "text-primary/70 hover:text-primary"}`}
              style={{ transform: `translate(-50%, ${dealVisible ? "-50%" : "calc(-50% + 0.25rem)"})` }}
              aria-label={t("ads.title")}
            >
              <span className="block truncate">{renderDealMessage(activeDeal, liveDealsConfig)}</span>
            </Link>
          ) : null}
        </div>
      </div>

      {mobileMenu}
    </header>
  );
}
