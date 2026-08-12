import { lazy, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "./lib/auth";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import Categories from "./pages/Categories";
import ProductTerms from "./pages/ProductTerms";
import CatalogFacetPages from "./pages/CatalogFacetPages";
import Products from "./pages/Products";
import Blocks from "./pages/Blocks";
import Blogs from "./pages/Blogs";
import Projects from "./pages/Projects";
import Templates from "./pages/Templates";
import ContentSections from "./pages/ContentSections";
import TeamMembers from "./pages/TeamMembers";
import Ads from "./pages/Ads";
import SampleRequests from "./pages/SampleRequests";
import ContactSubmissions from "./pages/ContactSubmissions";
import Users from "./pages/Users";
import Roles from "./pages/Roles";
import RoleEditor from "./pages/RoleEditor";
import Audit from "./pages/Audit";
import Order from "./pages/Order";
import ChangePassword from "./pages/ChangePassword";
import WorkflowTemplates from "./pages/WorkflowTemplates";
import WorkflowRuntime from "./pages/WorkflowRuntime";
import Notifications from "./pages/Notifications";
import PanelLayout from "./components/PanelLayout";
import { useTranslation } from "./lib/i18n";

const WorkflowBuilder = lazy(() => import("./pages/WorkflowBuilder"));
const Reports = lazy(() => import("./pages/Reports"));
const Finance = lazy(() => import("./pages/Finance"));
const Settings = lazy(() => import("./pages/Settings"));
const AdminTools = lazy(() => import("./pages/AdminTools"));
const AccessDenied = lazy(() => import("./pages/AccessDenied"));
const PanelNotFound = lazy(() => import("./pages/PanelNotFound"));
const lazyNamed = (loader, name) => lazy(() => loader().then((module) => ({ default: module[name] })));
const Batches = lazyNamed(() => import("./pages/LogisticsOperations"), "Batches");
const BatchDetail = lazyNamed(() => import("./pages/LogisticsOperations"), "BatchDetail");
const Inventory = lazyNamed(() => import("./pages/LogisticsOperations"), "Inventory");
const Shipments = lazyNamed(() => import("./pages/LogisticsOperations"), "Shipments");
const ShipmentDetail = lazyNamed(() => import("./pages/LogisticsOperations"), "ShipmentDetail");
const Suppliers = lazyNamed(() => import("./pages/SupplierQualityInstallation"), "Suppliers");
const SupplierDetail = lazyNamed(() => import("./pages/SupplierQualityInstallation"), "SupplierDetail");
const Purchases = lazyNamed(() => import("./pages/SupplierQualityInstallation"), "Purchases");
const PurchaseDetail = lazyNamed(() => import("./pages/SupplierQualityInstallation"), "PurchaseDetail");
const Quality = lazyNamed(() => import("./pages/SupplierQualityInstallation"), "Quality");
const QualityDetail = lazyNamed(() => import("./pages/SupplierQualityInstallation"), "QualityDetail");
const Installations = lazyNamed(() => import("./pages/SupplierQualityInstallation"), "Installations");
const InstallationDetail = lazyNamed(() => import("./pages/SupplierQualityInstallation"), "InstallationDetail");

const ProtectedRoute = ({ children }) => {
  const { isAuthenticated, checking, user } = useAuth();
  const { t } = useTranslation();

  if (checking) {
    return (
      <div className="panel-shell py-16 text-sm text-primary/70">{t("messages.loading")}</div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  if (user?.must_change_password) return <Navigate to="/change-password" replace />;

  return children;
};

const PermissionRoute = ({ permission, children }) => {
  const { hasPermission } = useAuth();
  return hasPermission(permission) ? children : <Navigate to="/access-denied" replace />;
};
const AnyPermissionRoute = ({ permissions, children }) => {
  const { hasPermission } = useAuth();
  return permissions.some(hasPermission) ? children : <Navigate to="/access-denied" replace />;
};

export default function AppRoutes() {
  return (
    <Suspense fallback={<div className="panel-shell py-16 text-sm text-primary/60" aria-busy="true">در حال بارگذاری…</div>}><Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/change-password" element={<ChangePassword />} />
      <Route
        path="/dashboard"
        element={
          <ProtectedRoute>
            <PanelLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Dashboard />} />
		<Route path="categories" element={<PermissionRoute permission="content.manage"><Categories /></PermissionRoute>} />
		<Route path="product-terms" element={<PermissionRoute permission="content.manage"><ProductTerms /></PermissionRoute>} />
		<Route path="catalog-facet-seo" element={<PermissionRoute permission="content.manage"><CatalogFacetPages /></PermissionRoute>} />
		<Route path="products" element={<PermissionRoute permission="content.manage"><Products /></PermissionRoute>} />
		<Route path="blocks" element={<PermissionRoute permission="content.manage"><Blocks /></PermissionRoute>} />
		<Route path="blogs" element={<PermissionRoute permission="content.manage"><Blogs /></PermissionRoute>} />
		<Route path="projects" element={<PermissionRoute permission="content.manage"><Projects /></PermissionRoute>} />
		<Route path="templates" element={<PermissionRoute permission="content.manage"><Templates /></PermissionRoute>} />
		<Route path="content" element={<PermissionRoute permission="content.manage"><ContentSections /></PermissionRoute>} />
		<Route path="team" element={<PermissionRoute permission="content.manage"><TeamMembers /></PermissionRoute>} />
		<Route path="ads" element={<PermissionRoute permission="content.manage"><Ads /></PermissionRoute>} />
		<Route path="contact-submissions" element={<PermissionRoute permission="content.manage"><ContactSubmissions /></PermissionRoute>} />
		<Route path="sample-requests" element={<PermissionRoute permission="content.manage"><SampleRequests /></PermissionRoute>} />
		<Route path="users" element={<PermissionRoute permission="users.view"><Users /></PermissionRoute>} />
		<Route path="roles" element={<PermissionRoute permission="roles.view"><Roles /></PermissionRoute>} />
		<Route path="roles/:id" element={<PermissionRoute permission="roles.view"><RoleEditor /></PermissionRoute>} />
		<Route path="audit" element={<PermissionRoute permission="audit.view"><Audit /></PermissionRoute>} />
		<Route path="orders/:id" element={<AnyPermissionRoute permissions={["orders.view_all","orders.view_own"]}><Order /></AnyPermissionRoute>} />
        <Route path="workflows" element={<PermissionRoute permission="workflow_templates.manage"><WorkflowTemplates /></PermissionRoute>} />
        <Route path="workflows/:templateId/builder" element={<PermissionRoute permission="workflow_templates.manage"><WorkflowBuilder /></PermissionRoute>} />
		<Route path="workflows/:workflowInstanceId" element={<AnyPermissionRoute permissions={["workflow_instances.view_all","workflow_instances.view_assigned"]}><WorkflowRuntime /></AnyPermissionRoute>} />
        <Route path="batches" element={<AnyPermissionRoute permissions={["batches.view_assigned","batches.view_all"]}><Batches /></AnyPermissionRoute>} />
        <Route path="batches/:id" element={<AnyPermissionRoute permissions={["batches.view_assigned","batches.view_all"]}><BatchDetail /></AnyPermissionRoute>} />
        <Route path="inventory" element={<PermissionRoute permission="inventory.lots.view"><Inventory /></PermissionRoute>} />
        <Route path="shipments" element={<AnyPermissionRoute permissions={["shipments.view_assigned","shipments.view_all"]}><Shipments /></AnyPermissionRoute>} />
        <Route path="shipments/:id" element={<AnyPermissionRoute permissions={["shipments.view_assigned","shipments.view_all"]}><ShipmentDetail /></AnyPermissionRoute>} />
        <Route path="notifications" element={<PermissionRoute permission="notifications.view_own"><Notifications /></PermissionRoute>} />
        <Route path="finance" element={<AnyPermissionRoute permissions={["finance.exchange_rates.manage","finance.payments.view","finance.costs.view","finance.costs.view_assigned"]}><Finance /></AnyPermissionRoute>} />
        <Route path="reports" element={<AnyPermissionRoute permissions={["reports.overview.view","reports.receivables.view","reports.costs.view","reports.profitability.view","reports.operations.view","reports.sales.view"]}><Reports /></AnyPermissionRoute>} />
        <Route path="suppliers" element={<PermissionRoute permission="suppliers.view"><Suppliers /></PermissionRoute>} />
        <Route path="suppliers/:id" element={<PermissionRoute permission="suppliers.view"><SupplierDetail /></PermissionRoute>} />
        <Route path="purchases" element={<AnyPermissionRoute permissions={["purchases.view_assigned","purchases.view_all"]}><Purchases /></AnyPermissionRoute>} />
        <Route path="purchases/:id" element={<AnyPermissionRoute permissions={["purchases.view_assigned","purchases.view_all"]}><PurchaseDetail /></AnyPermissionRoute>} />
        <Route path="quality" element={<AnyPermissionRoute permissions={["quality.view_assigned","quality.view_all"]}><Quality /></AnyPermissionRoute>} />
        <Route path="quality/:id" element={<AnyPermissionRoute permissions={["quality.view_assigned","quality.view_all"]}><QualityDetail /></AnyPermissionRoute>} />
        <Route path="installations" element={<AnyPermissionRoute permissions={["installation.view_assigned","installation.view_all"]}><Installations /></AnyPermissionRoute>} />
        <Route path="installations/:id" element={<AnyPermissionRoute permissions={["installation.view_assigned","installation.view_all"]}><InstallationDetail /></AnyPermissionRoute>} />
		<Route path="settings" element={<PermissionRoute permission="settings.view"><Settings /></PermissionRoute>} />
		<Route path="admin-tools" element={<PermissionRoute permission="admin_tools.view"><AdminTools /></PermissionRoute>} />
      </Route>
      <Route path="/access-denied" element={<AccessDenied />} />
	  <Route path="*" element={<PanelNotFound />} />
    </Routes></Suspense>
  );
}
