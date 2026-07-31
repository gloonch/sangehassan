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
import WorkflowBuilder from "./pages/WorkflowBuilder";
import WorkflowRuntime from "./pages/WorkflowRuntime";
import { Batches, BatchDetail, Inventory, Shipments, ShipmentDetail } from "./pages/OperationsPhase3";
import PanelLayout from "./components/PanelLayout";
import { useTranslation } from "./lib/i18n";

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
  return hasPermission(permission) ? children : <Navigate to="/dashboard" replace />;
};
const AnyPermissionRoute = ({ permissions, children }) => {
  const { hasPermission } = useAuth();
  return permissions.some(hasPermission) ? children : <Navigate to="/dashboard" replace />;
};

export default function AppRoutes() {
  return (
    <Routes>
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
        <Route path="categories" element={<Categories />} />
        <Route path="product-terms" element={<ProductTerms />} />
        <Route path="catalog-facet-seo" element={<CatalogFacetPages />} />
        <Route path="products" element={<Products />} />
        <Route path="blocks" element={<Blocks />} />
        <Route path="blogs" element={<Blogs />} />
        <Route path="projects" element={<Projects />} />
        <Route path="templates" element={<Templates />} />
        <Route path="content" element={<ContentSections />} />
        <Route path="team" element={<TeamMembers />} />
        <Route path="ads" element={<Ads />} />
        <Route path="contact-submissions" element={<ContactSubmissions />} />
        <Route path="sample-requests" element={<SampleRequests />} />
        <Route path="users" element={<Users />} />
        <Route path="roles" element={<Roles />} />
        <Route path="roles/:id" element={<RoleEditor />} />
        <Route path="audit" element={<Audit />} />
        <Route path="orders/:id" element={<Order />} />
        <Route path="workflows" element={<PermissionRoute permission="workflow_templates.manage"><WorkflowTemplates /></PermissionRoute>} />
        <Route path="workflows/:templateId/builder" element={<PermissionRoute permission="workflow_templates.manage"><WorkflowBuilder /></PermissionRoute>} />
        <Route path="workflows/:workflowInstanceId" element={<WorkflowRuntime />} />
        <Route path="batches" element={<AnyPermissionRoute permissions={["batches.view_assigned","batches.view_all"]}><Batches /></AnyPermissionRoute>} />
        <Route path="batches/:id" element={<AnyPermissionRoute permissions={["batches.view_assigned","batches.view_all"]}><BatchDetail /></AnyPermissionRoute>} />
        <Route path="inventory" element={<PermissionRoute permission="inventory.lots.view"><Inventory /></PermissionRoute>} />
        <Route path="shipments" element={<AnyPermissionRoute permissions={["shipments.view_assigned","shipments.view_all"]}><Shipments /></AnyPermissionRoute>} />
        <Route path="shipments/:id" element={<AnyPermissionRoute permissions={["shipments.view_assigned","shipments.view_all"]}><ShipmentDetail /></AnyPermissionRoute>} />
      </Route>
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}
