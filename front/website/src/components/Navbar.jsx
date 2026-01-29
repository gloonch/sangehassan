import { useState } from "react";
import { Link, NavLink } from "react-router-dom";
import LanguageSwitch from "./LanguageSwitch";
import { useTranslation } from "../lib/i18n";

const navItems = [
  { key: "home", path: "/", end: true },
  { key: "products", path: "/products" },
  { key: "blogs", path: "/blogs" },
  { key: "stoneFinishes", path: "/stone-finishes" },
  // { key: "gallery", path: "/gallery" },
  { key: "about", path: "/about" },
  // { key: "contact", path: "/contact" }
];

export default function Navbar() {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  return (
    <header className="fixed left-0 right-0 top-0 z-50 border-b border-primary/10 bg-sand/80 backdrop-blur-lg">
      <div className="section-shell flex h-20 items-center justify-between gap-4">
        <Link to="/" className="font-display text-xl tracking-wide">
          {t("brand")}
        </Link>

        <nav className="hidden items-center gap-6 text-sm font-medium lg:flex">
          {navItems.map((item) => (
            <NavLink
              key={item.key}
              to={item.path}
              end={item.end}
              className={({ isActive }) =>
                `transition ${isActive ? "text-accent" : "text-primary/70 hover:text-primary"}`
              }
            >
              {t(`nav.${item.key}`)}
            </NavLink>
          ))}
        </nav>

        <div className="hidden items-center gap-3 lg:flex">
          <LanguageSwitch />
          <NavLink
            to="/login"
            className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/80 transition hover:border-primary/50"
          >
            {t("actions.login")}
          </NavLink>
          <NavLink
            to="/signup"
            className="rounded-full bg-primary px-4 py-2 text-xs font-semibold text-sand transition hover:bg-primary/90"
          >
            {t("actions.signup")}
          </NavLink>
        </div>

        <button
          type="button"
          className="flex items-center gap-2 text-sm font-semibold text-primary lg:hidden"
          onClick={() => setOpen((prev) => !prev)}
          aria-expanded={open}
          aria-label={t("actions.menu")}
        >
          <span className="h-0.5 w-6 bg-primary" />
          <span className="h-0.5 w-4 bg-primary" />
        </button>
      </div>

      {open && (
        <div className="lg:hidden">
          <div className="section-shell flex flex-col gap-4 border-t border-primary/10 py-6">
            <div className="flex items-center justify-between">
              <LanguageSwitch />
            </div>
            <div className="flex flex-col gap-3 text-sm font-medium">
              {navItems.map((item) => (
                <NavLink
                  key={item.key}
                  to={item.path}
                  onClick={() => setOpen(false)}
                  end={item.end}
                  className={({ isActive }) =>
                    `transition ${isActive ? "text-accent" : "text-primary/70 hover:text-primary"}`
                  }
                >
                  {t(`nav.${item.key}`)}
                </NavLink>
              ))}
            </div>
            <div className="flex items-center gap-3">
              <NavLink
                to="/login"
                onClick={() => setOpen(false)}
                className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/80 transition hover:border-primary/50"
              >
                {t("actions.login")}
              </NavLink>
              <NavLink
                to="/signup"
                onClick={() => setOpen(false)}
                className="rounded-full bg-primary px-4 py-2 text-xs font-semibold text-sand transition hover:bg-primary/90"
              >
                {t("actions.signup")}
              </NavLink>
            </div>
          </div>
        </div>
      )}
    </header>
  );
}
