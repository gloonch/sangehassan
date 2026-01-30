import { NavLink, Outlet } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { useAuth } from "../lib/auth";
import LanguageSwitch from "./LanguageSwitch";

const navItems = [
  { key: "dashboard", path: "/dashboard", label: "nav.dashboard", end: true },
  { key: "products", path: "/dashboard/products", label: "panelProducts.title" },
  { key: "blocks", path: "/dashboard/blocks", label: "panelBlocks.title" },
  { key: "categories", path: "/dashboard/categories", label: "categories.title" },
  { key: "blogs", path: "/dashboard/blogs", label: "panelBlogs.title" },
  { key: "templates", path: "/dashboard/templates", label: "templates.title" },
  { key: "content", path: "/dashboard/content", label: "panelContent.title" }
];

export default function PanelLayout() {
  const { t } = useTranslation();
  const { logout } = useAuth();

  return (
    <div className="min-h-screen">
      <div className="panel-shell flex flex-col gap-6 py-10">
        <header className="panel-card flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="text-sm uppercase tracking-[0.3em] text-primary/60">{t("admin.subtitle")}</p>
            <h1 className="font-display text-3xl">{t("admin.title")}</h1>
          </div>
          <div className="flex items-center gap-3">
            <LanguageSwitch />
            <button
              type="button"
              onClick={logout}
              className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/80 transition hover:border-primary/50"
            >
              {t("actions.logout")}
            </button>
          </div>
        </header>

        <nav className="rounded-3xl border border-primary/10 bg-primary/5 px-4 py-3 shadow-lg">
          <div className="flex flex-wrap items-center gap-3 overflow-x-auto">
            {navItems.map((item) => (
              <NavLink
                key={item.key}
                to={item.path}
                end={item.end}
                className={({ isActive }) =>
                  `rounded-full px-5 py-2 text-xs font-semibold uppercase tracking-[0.2em] transition ${
                    isActive
                      ? "bg-primary text-sand shadow-md"
                      : "border border-primary/20 text-primary/70 hover:border-primary/50 hover:text-primary"
                  }`
                }
              >
                {t(item.label)}
              </NavLink>
            ))}
          </div>
        </nav>

        <section className="flex-1">
          <Outlet />
        </section>
      </div>
    </div>
  );
}
