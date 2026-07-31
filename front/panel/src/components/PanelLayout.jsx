import { NavLink, Outlet } from "react-router-dom";
import { useTranslation } from "../lib/i18n";
import { useAuth } from "../lib/auth";
import LanguageSwitch from "./LanguageSwitch";

const navItems = [
  { key: "dashboard", path: "/dashboard", label: "nav.dashboard", end: true },
  { key: "products", path: "/dashboard/products", label: "panelProducts.title", permission: "content.manage" },
  { key: "productTerms", path: "/dashboard/product-terms", label: "panelProductTerms.title", permission: "content.manage" },
  { key: "catalogFacetSEO", path: "/dashboard/catalog-facet-seo", label: "catalogFacetSEO.title", permission: "content.manage" },
  { key: "blocks", path: "/dashboard/blocks", label: "panelBlocks.title", permission: "content.manage" },
  { key: "projects", path: "/dashboard/projects", label: "panelProjects.title", permission: "content.manage" },
  { key: "categories", path: "/dashboard/categories", label: "categories.title", permission: "content.manage" },
  { key: "blogs", path: "/dashboard/blogs", label: "panelBlogs.title", permission: "content.manage" },
  { key: "templates", path: "/dashboard/templates", label: "templates.title", permission: "content.manage" },
  { key: "content", path: "/dashboard/content", label: "panelContent.title", permission: "content.manage" },
  { key: "team", path: "/dashboard/team", label: "panelTeam.title", permission: "content.manage" },
  { key: "ads", path: "/dashboard/ads", label: "panelAds.title", permission: "content.manage" },
  { key: "contactSubmissions", path: "/dashboard/contact-submissions", label: "panelContactSubmissions.title", permission: "content.manage" },
  { key: "sampleRequests", path: "/dashboard/sample-requests", label: "panelSampleRequests.title", permission: "content.manage" },
  { key: "users", path: "/dashboard/users", label: "panelUsers.title", permission: "users.view" },
  { key: "roles", path: "/dashboard/roles", text: "نقش‌ها و دسترسی‌ها", permission: "roles.view" },
  { key: "workflows", path: "/dashboard/workflows", text: "طراحی Workflow", permission: "workflow_templates.manage" },
  { key: "batches", path: "/dashboard/batches", text: "بچ‌ها", permissions: ["batches.view_assigned", "batches.view_all"] },
  { key: "inventory", path: "/dashboard/inventory", text: "موجودی", permission: "inventory.lots.view" },
  { key: "shipments", path: "/dashboard/shipments", text: "محموله‌ها", permissions: ["shipments.view_assigned", "shipments.view_all"] },
  { key: "audit", path: "/dashboard/audit", text: "گزارش تغییرات", permission: "audit.view" }
];

export default function PanelLayout() {
  const { t } = useTranslation();
  const { logout, user, hasPermission } = useAuth();

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
            <span className="text-xs text-primary/60">{[user?.first_name, user?.last_name].filter(Boolean).join(" ") || user?.phone}</span>
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
            {navItems.filter((item) => (!item.permission || hasPermission(item.permission)) && (!item.permissions || item.permissions.some(hasPermission))).map((item) => (
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
                {item.text || t(item.label)}
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
