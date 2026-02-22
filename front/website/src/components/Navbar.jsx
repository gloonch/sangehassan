import { useEffect, useState } from "react";
import { Link, NavLink, useLocation } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import logoImage from "@shared/assets/logo.png";
import logoWhiteImage from "@shared/assets/logo_white.png";

const navItems = [
  { key: "products", path: "/products" },
  { key: "blocks", path: "/blocks" },
  { key: "ads", path: "/ads" },
  { key: "blogs", path: "/blogs" },
  { key: "about", path: "/about" },
];

export default function Navbar() {
  const { t } = useTranslation();
  const location = useLocation();
  const [open, setOpen] = useState(false);
  const [user, setUser] = useState(null);
  const isHome = location.pathname === "/";

  useEffect(() => {
    let active = true;
    const restore = () => {
      try {
        const stored = sessionStorage.getItem("sh_me");
        if (stored) {
          const parsed = JSON.parse(stored);
          if (parsed?.email) {
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

  const displayName = user?.full_name || user?.email;
  const avatarChar = (displayName || "U").trim().charAt(0).toUpperCase();
  const visibleNavItems = navItems.filter((item) => item.key !== "ads" || Boolean(user));
  const navHeaderClass = isHome
    ? open
      ? "border-b border-sand/25 bg-primary/45 backdrop-blur-xl"
      : "border-transparent bg-transparent"
    : "border-b border-primary/10 bg-sand/80 backdrop-blur-lg";
  const navTextClass = isHome
    ? "text-sand/80 hover:text-sand"
    : "text-primary/70 hover:text-primary";
  const navActiveClass = isHome ? "text-sand" : "text-accent";

  return (
    <header className={`fixed left-0 right-0 top-0 z-50 transition-colors duration-300 ${navHeaderClass}`}>
      <div className="section-shell flex h-20 items-center justify-between gap-4">
        <Link to="/" className="inline-flex items-center" aria-label={t("brand")}>
          <img src={isHome ? logoWhiteImage : logoImage} alt={t("brand")} className="h-12 w-auto object-contain" />
        </Link>

        <nav className="hidden items-center gap-6 text-sm font-medium lg:flex">
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

        <div className="hidden items-center gap-3 lg:flex">
          {user ? (
            <NavLink
              to="/profile"
              className={`flex h-10 w-10 items-center justify-center rounded-full text-sm font-semibold uppercase shadow transition ${isHome
                ? "border border-sand/30 bg-sand/15 text-sand hover:bg-sand/25"
                : "bg-primary text-sand hover:bg-primary/90"
                }`}
              aria-label={displayName}
              title={displayName}
            >
              {avatarChar}
            </NavLink>
          ) : (
            <>
              <NavLink
                to="/login"
                className={`rounded-full border px-4 py-2 text-xs font-semibold transition ${isHome
                  ? "border-sand/35 text-sand/90 hover:border-sand/60 hover:text-sand"
                  : "border-primary/20 text-primary/80 hover:border-primary/50"
                  }`}
              >
                {t("actions.login")}
              </NavLink>
              <NavLink
                to="/signup"
                className={`rounded-full px-4 py-2 text-xs font-semibold transition ${isHome
                  ? "border border-sand/35 bg-sand/15 text-sand hover:bg-sand/25"
                  : "bg-primary text-sand hover:bg-primary/90"
                  }`}
              >
                {t("actions.signup")}
              </NavLink>
            </>
          )}
        </div>

        <button
          type="button"
          className={`flex h-10 w-10 items-center justify-center rounded-full border text-base font-semibold leading-none transition lg:hidden ${
            isHome
              ? "border-sand/40 bg-sand/10 text-sand hover:bg-sand/20"
              : "border-primary/25 bg-white/70 text-primary hover:bg-white"
          }`}
          onClick={() => setOpen((prev) => !prev)}
          aria-expanded={open}
          aria-label={t("actions.menu")}
        >
          i
        </button>
      </div>

      {open && (
        <div className={`lg:hidden ${isHome ? "bg-primary/50 backdrop-blur-xl" : "bg-sand/95 backdrop-blur-lg"}`}>
          <div className={`section-shell py-4 ${isHome ? "border-t border-sand/25" : "border-t border-primary/10"}`}>
            <div className="flex flex-nowrap items-center gap-2 overflow-x-auto pb-1">
              {visibleNavItems.map((item) => (
                <NavLink
                  key={item.key}
                  to={item.path}
                  onClick={() => setOpen(false)}
                  end={item.end}
                  className={({ isActive }) =>
                    `whitespace-nowrap rounded-full border px-3 py-1.5 text-xs font-semibold transition ${
                      isActive
                        ? isHome
                          ? "border-sand/50 bg-sand/15 text-sand"
                          : "border-primary/30 bg-primary text-sand"
                        : isHome
                          ? "border-sand/30 text-sand/85"
                          : "border-primary/20 text-primary/70"
                    }`
                  }
                >
                  {t(`nav.${item.key}`)}
                </NavLink>
              ))}
            </div>
            <div className="mt-3 flex flex-nowrap items-center gap-2 overflow-x-auto pb-1">
              {user ? (
                <NavLink
                  to="/profile"
                  onClick={() => setOpen(false)}
                  className={`flex h-9 min-w-9 items-center justify-center rounded-full text-xs font-semibold uppercase shadow transition ${
                    isHome
                      ? "border border-sand/30 bg-sand/15 text-sand hover:bg-sand/25"
                      : "bg-primary text-sand hover:bg-primary/90"
                  }`}
                  aria-label={displayName}
                  title={displayName}
                >
                  {avatarChar}
                </NavLink>
              ) : (
                <>
                  <NavLink
                    to="/login"
                    onClick={() => setOpen(false)}
                    className={`whitespace-nowrap rounded-full border px-3 py-1.5 text-xs font-semibold transition ${
                      isHome
                        ? "border-sand/35 text-sand/90 hover:border-sand/60 hover:text-sand"
                        : "border-primary/20 text-primary/80 hover:border-primary/50"
                    }`}
                  >
                    {t("actions.login")}
                  </NavLink>
                  <NavLink
                    to="/signup"
                    onClick={() => setOpen(false)}
                    className={`whitespace-nowrap rounded-full px-3 py-1.5 text-xs font-semibold transition ${
                      isHome
                        ? "border border-sand/35 bg-sand/15 text-sand hover:bg-sand/25"
                        : "bg-primary text-sand hover:bg-primary/90"
                    }`}
                  >
                    {t("actions.signup")}
                  </NavLink>
                </>
              )}
            </div>
          </div>
        </div>
      )}
    </header>
  );
}
