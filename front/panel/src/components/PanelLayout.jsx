import { NavLink, Outlet } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { useAuth } from "../lib/auth";
import LanguageSwitch from "./LanguageSwitch";

const navItems = [
  { key: "dashboard", path: "/dashboard", label: "nav.dashboard", end: true },
  { key: "products", path: "/dashboard/products", label: "panelProducts.title" },
  { key: "categories", path: "/dashboard/categories", label: "categories.title" },
  { key: "blogs", path: "/dashboard/blogs", label: "panelBlogs.title" },
  { key: "templates", path: "/dashboard/templates", label: "templates.title" }
];

export default function PanelLayout() {
  const { t } = useTranslation();
  const { logout } = useAuth();

  return (
    <div className="min-h-screen">
      <div className="panel-shell flex flex-col gap-8 py-10 lg:flex-row">
        <aside className="panel-card w-full max-w-xs">
          <div className="mb-6">
            <p className="font-display text-xl">{t("brand")}</p>
            <p className="text-xs text-primary/60">{t("admin.subtitle")}</p>
          </div>
          <nav className="flex flex-col gap-3 text-sm font-semibold">
            {navItems.map((item) => (
              <NavLink
                key={item.key}
                to={item.path}
                end={item.end}
                className={({ isActive }) =>
                  `rounded-xl px-3 py-2 transition ${
                    isActive ? "bg-primary text-sand" : "text-primary/70 hover:text-primary"
                  }`
                }
              >
                {t(item.label)}
              </NavLink>
            ))}
          </nav>
        </aside>

        <section className="flex-1">
          <div className="mb-6 flex flex-wrap items-center justify-between gap-4">
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
          </div>

          <Outlet />
        </section>
      </div>
    </div>
  );
}
